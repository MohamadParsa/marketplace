package httptransport

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/application/auth"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

type ProductReviewsHandler struct {
	service     *backoffice.Service
	auth        *auth.Service
	mu          sync.Mutex
	submissions map[string]guestLimit
	lastCleanup time.Time
}

func NewProductReviewsHandler(service *backoffice.Service, authService *auth.Service) *ProductReviewsHandler {
	return &ProductReviewsHandler{service: service, auth: authService, submissions: make(map[string]guestLimit)}
}

func (h *ProductReviewsHandler) Public(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method == http.MethodGet {
		page, err := h.service.PublicProductReviews(r.Context(), r.PathValue("id"), r.URL.Query().Get("cursor"))
		if err != nil {
			reviewError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, page)
		return
	}
	var input backoffice.ProductReviewInput
	if !decodeReviewRequest(w, r, &input) {
		return
	}
	if !h.allowSubmission(r) {
		writeError(w, http.StatusTooManyRequests, "review_rate_limited")
		return
	}
	if err := h.service.SubmitProductReview(r.Context(), r.PathValue("id"), input); err != nil {
		reviewError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductReviewsHandler) Admin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	session, err := h.auth.Authenticate(r.Context(), bearer(r.Header.Get("Authorization")), identity.SurfaceBackoffice)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if session.User.Role != identity.RoleSuperAdmin {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	if r.Method == http.MethodGet {
		status := r.URL.Query().Get("status")
		if status == "" {
			status = "pending"
		}
		page, err := h.service.AdminProductReviews(r.Context(), session.User, status, r.URL.Query().Get("cursor"))
		if err != nil {
			reviewError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, page)
		return
	}
	var decision backoffice.ProductReviewDecision
	if !decodeReviewRequest(w, r, &decision) {
		return
	}
	if err := h.service.DecideProductReview(r.Context(), session.User, r.PathValue("id"), decision); err != nil {
		reviewError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func decodeReviewRequest(w http.ResponseWriter, r *http.Request, input any) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeError(w, http.StatusUnsupportedMediaType, "json_required")
		return false
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16*1024))
	decoder.DisallowUnknownFields()
	if decoder.Decode(input) != nil || decoder.Decode(new(any)) != io.EOF {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return false
	}
	return true
}

func reviewError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, backoffice.ErrInvalid):
		writeError(w, http.StatusBadRequest, "invalid_request")
	case errors.Is(err, backoffice.ErrForbidden):
		writeError(w, http.StatusForbidden, "forbidden")
	case errors.Is(err, backoffice.ErrReviewSubmissionsClosed):
		writeError(w, http.StatusForbidden, "review_submissions_closed")
	case errors.Is(err, backoffice.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found")
	case errors.Is(err, backoffice.ErrConflict):
		writeError(w, http.StatusConflict, "review_conflict")
	default:
		writeError(w, http.StatusServiceUnavailable, "reviews_unavailable")
	}
}

func (h *ProductReviewsHandler) allowSubmission(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	now := time.Now()
	h.mu.Lock()
	defer h.mu.Unlock()
	if now.Sub(h.lastCleanup) >= time.Minute {
		for source, window := range h.submissions {
			if now.Sub(window.At) >= time.Minute {
				delete(h.submissions, source)
			}
		}
		h.lastCleanup = now
	}
	window, exists := h.submissions[host]
	if !exists && len(h.submissions) >= 10000 {
		return false
	}
	if now.Sub(window.At) >= time.Minute {
		window = guestLimit{At: now}
	}
	if window.Count >= 60 {
		return false
	}
	window.Count++
	h.submissions[host] = window
	return true
}
