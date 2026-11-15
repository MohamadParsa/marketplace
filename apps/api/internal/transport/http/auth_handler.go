package httptransport

import (
	"encoding/json"
	"errors"
	"github.com/cherriva/marketplace/apps/api/internal/application/auth"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"net/http"
	"strings"
)

type AuthHandler struct {
	service        *auth.Service
	exposeTestCode bool
	notifier       interface {
		Code(identity.Surface, string) (string, bool)
	}
}

func NewAuthHandler(service *auth.Service, notifier interface {
	Code(identity.Surface, string) (string, bool)
}, exposeTestCode bool) *AuthHandler {
	return &AuthHandler{service: service, notifier: notifier, exposeTestCode: exposeTestCode}
}
func (h *AuthHandler) RequestOTP(surface identity.Surface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Phone string `json:"phone"`
		}
		if json.NewDecoder(r.Body).Decode(&in) != nil || strings.TrimSpace(in.Phone) == "" {
			writeError(w, http.StatusBadRequest, "invalid_request")
			return
		}
		id, err := h.service.RequestOTP(r.Context(), surface, in.Phone)
		if err != nil {
			writeAuthError(w, err)
			return
		}
		out := map[string]string{"request_id": id}
		if h.exposeTestCode {
			if code, ok := h.notifier.Code(surface, identity.NormalizePhone(in.Phone)); ok {
				out["test_code"] = code
			}
		}
		writeJSON(w, http.StatusAccepted, out)
	}
}
func (h *AuthHandler) VerifyOTP(surface identity.Surface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			RequestID string `json:"request_id"`
			Code      string `json:"code"`
		}
		if json.NewDecoder(r.Body).Decode(&in) != nil || in.RequestID == "" || in.Code == "" {
			writeError(w, http.StatusBadRequest, "invalid_request")
			return
		}
		session, err := h.service.VerifyOTP(r.Context(), surface, in.RequestID, in.Code)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid_or_expired_otp")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"token": session.Token, "expires_at": session.Expires, "user": session.User})
	}
}
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	_ = h.service.Logout(r.Context(), bearer(r.Header.Get("Authorization")))
	w.WriteHeader(http.StatusNoContent)
}
func (h *AuthHandler) Require(surface identity.Surface, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := h.service.Authenticate(r.Context(), bearer(r.Header.Get("Authorization")), surface); err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next.ServeHTTP(w, r)
	})
}
func bearer(value string) string {
	scheme, token, ok := strings.Cut(value, " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return ""
	}
	return strings.TrimSpace(token)
}
func writeAuthError(w http.ResponseWriter, err error) {
	status := http.StatusUnauthorized
	code := err.Error()
	if errors.Is(err, identity.ErrInactive) {
		status = http.StatusForbidden
	}
	if errors.Is(err, identity.ErrResendSoon) || errors.Is(err, identity.ErrRateLimit) {
		status = http.StatusTooManyRequests
	}
	if errors.Is(err, identity.ErrUnknown) {
		code = "invalid_credentials"
	}
	writeError(w, status, code)
}
func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
func writeError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]string{"code": code})
}
