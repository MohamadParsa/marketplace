package httptransport

import (
	"encoding/json"
	"errors"
	"github.com/cherriva/marketplace/apps/api/internal/application/auth"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"io"
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
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&in) != nil || strings.TrimSpace(in.Phone) == "" {
			writeError(w, http.StatusBadRequest, "invalid_request")
			return
		}
		id, err := h.service.RequestOTP(r.Context(), surface, in.Phone)
		if err != nil {
			writeAuthError(w, err)
			return
		}
		out := map[string]any{"request_id": id}
		if available, err := h.service.PasswordAvailable(r.Context(), surface, in.Phone); err == nil {
			out["password_available"] = available
		}
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
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&in) != nil || in.RequestID == "" || in.Code == "" {
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
func (h *AuthHandler) PasswordLogin(surface identity.Surface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Phone    string `json:"phone"`
			Password string `json:"password"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&in) != nil || strings.TrimSpace(in.Phone) == "" || in.Password == "" {
			writeError(w, http.StatusBadRequest, "invalid_request")
			return
		}
		session, err := h.service.PasswordLogin(r.Context(), surface, in.Phone, in.Password)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid_credentials")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"token": session.Token, "expires_at": session.Expires, "user": session.User})
	}
}
func (h *AuthHandler) Profile(surface identity.Surface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "private, no-store")
		session, err := h.service.Authenticate(r.Context(), bearer(r.Header.Get("Authorization")), surface)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		writeJSON(w, http.StatusOK, profileJSON(session.User))
	}
}
func (h *AuthHandler) UpdateProfile(surface identity.Surface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "private, no-store")
		var in struct {
			NationalID *string `json:"national_id"`
			Name       string  `json:"name"`
			Email      string  `json:"email"`
			Password   string  `json:"password"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&in) != nil {
			writeError(w, http.StatusBadRequest, "invalid_request")
			return
		}
		token := bearer(r.Header.Get("Authorization"))
		current, authErr := h.service.Authenticate(r.Context(), token, surface)
		if authErr != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if in.Name == "" {
			in.Name = current.User.Name
		}
		if in.Email == "" {
			in.Email = current.User.Email
		}
		var nationalID []string
		if in.NationalID != nil {
			nationalID = []string{*in.NationalID}
		}
		user, err := h.service.UpdateProfile(r.Context(), surface, token, in.Name, in.Email, in.Password, nationalID...)
		if err != nil {
			writeProfileError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, profileJSON(user))
	}
}
func (h *AuthHandler) RequestRecoveryOTP(surface identity.Surface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Phone       string `json:"phone"`
			Destination string `json:"destination"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&in) != nil || strings.TrimSpace(in.Phone) == "" {
			writeError(w, http.StatusBadRequest, "invalid_request")
			return
		}
		id, _, err := h.service.RequestRecoveryOTP(r.Context(), surface, in.Phone, in.Destination)
		if err != nil {
			writeRecoveryError(w, err)
			return
		}
		out := map[string]any{"request_id": id}
		if h.exposeTestCode {
			if code, ok := h.notifier.Code(surface, identity.NormalizePhone(in.Phone)); ok {
				out["test_code"] = code
			}
		}
		writeJSON(w, http.StatusAccepted, out)
	}
}
func (h *AuthHandler) RecoverPassword(surface identity.Surface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			RequestID string `json:"request_id"`
			Code      string `json:"code"`
			Password  string `json:"password"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&in) != nil || in.RequestID == "" || in.Code == "" || in.Password == "" {
			writeError(w, http.StatusBadRequest, "invalid_request")
			return
		}
		if err := h.service.RecoverPassword(r.Context(), surface, in.RequestID, in.Code, in.Password); err != nil {
			writeRecoveryError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
func (h *AuthHandler) AdminResetPassword(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Surface  string `json:"surface"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&in) != nil || strings.TrimSpace(in.Phone) == "" || in.Password == "" {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}
	surface := identity.SurfaceBackoffice
	if in.Surface != "" {
		surface = identity.Surface(in.Surface)
	}
	if err := h.service.AdminResetPassword(r.Context(), bearer(r.Header.Get("Authorization")), surface, in.Phone, in.Password); err != nil {
		switch {
		case errors.Is(err, identity.ErrNotAdmin):
			writeError(w, http.StatusForbidden, "forbidden")
		case errors.Is(err, identity.ErrInvalidProfile):
			writeError(w, http.StatusBadRequest, "invalid_password")
		case errors.Is(err, identity.ErrUnknown):
			writeError(w, http.StatusNotFound, "user_not_found")
		default:
			writeError(w, http.StatusUnauthorized, "unauthorized")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h *AuthHandler) RequestAdminOTP(w http.ResponseWriter, r *http.Request) {
	id, code, err := h.service.RequestAdminOTP(r.Context(), bearer(r.Header.Get("Authorization")))
	switch {
	case errors.Is(err, identity.ErrNotAdmin):
		writeError(w, http.StatusForbidden, "forbidden")
	case errors.Is(err, identity.ErrPasswordChangeRequired), errors.Is(err, identity.ErrInactive), errors.Is(err, identity.ErrInvalidOTP):
		writeError(w, http.StatusUnauthorized, "unauthorized")
	case errors.Is(err, identity.ErrResendSoon) || errors.Is(err, identity.ErrRateLimit):
		writeError(w, http.StatusTooManyRequests, "otp_rate_limit")
	case err != nil:
		writeError(w, http.StatusInternalServerError, "internal_error")
	default:
		out := map[string]string{"request_id": id}
		if h.exposeTestCode {
			out["test_code"] = code
		}
		writeJSON(w, http.StatusAccepted, out)
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
func profileJSON(user identity.User) map[string]any {
	return map[string]any{"user_id": user.ID, "name": user.Name, "phone": user.Phone, "email": user.Email, "national_id": user.NationalID, "password_available": user.PasswordHash != ""}
}
func writeProfileError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, identity.ErrConflict):
		writeError(w, http.StatusConflict, "email_conflict")
	case errors.Is(err, identity.ErrInvalidProfile):
		writeError(w, http.StatusBadRequest, "invalid_profile")
	default:
		writeError(w, http.StatusUnauthorized, "unauthorized")
	}
}
func writeRecoveryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, identity.ErrEmailRecoveryUnavailable):
		writeError(w, http.StatusNotImplemented, "email_recovery_unavailable")
	case errors.Is(err, identity.ErrInvalidProfile):
		writeError(w, http.StatusBadRequest, "invalid_password")
	case errors.Is(err, identity.ErrResendSoon), errors.Is(err, identity.ErrRateLimit):
		writeError(w, http.StatusTooManyRequests, "otp_rate_limit")
	default:
		writeError(w, http.StatusUnauthorized, "invalid_or_expired_otp")
	}
}
func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
func writeError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]string{"code": code})
}

func (h *AuthHandler) CompleteTemporaryPassword(surface identity.Surface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		var in struct {
			Password string `json:"password"`
		}
		d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
		d.DisallowUnknownFields()
		if d.Decode(&in) != nil || d.Decode(new(any)) != io.EOF {
			writeError(w, 400, "invalid_request")
			return
		}
		session, e := h.service.CompleteTemporaryPassword(r.Context(), bearer(r.Header.Get("Authorization")), surface, in.Password)
		if e != nil {
			if errors.Is(e, identity.ErrInvalidProfile) {
				writeError(w, 400, "new_password_required")
			} else {
				writeError(w, 401, "unauthorized")
			}
			return
		}
		writeJSON(w, 200, map[string]any{"token": session.Token, "expires_at": session.Expires, "user": session.User})
	}
}
