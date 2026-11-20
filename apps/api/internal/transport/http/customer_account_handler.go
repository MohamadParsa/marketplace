package httptransport

import (
	"errors"
	"net/http"

	"github.com/cherriva/marketplace/apps/api/internal/application/auth"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/application/customer"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

type CustomerAccountHandler struct {
	accounts *customer.Accounts
	auth     *auth.Service
}

func NewCustomerAccountHandler(accounts *customer.Accounts, authService *auth.Service) *CustomerAccountHandler {
	return &CustomerAccountHandler{accounts: accounts, auth: authService}
}

func (h *CustomerAccountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, err := h.auth.Authenticate(r.Context(), bearer(r.Header.Get("Authorization")), identity.SurfaceCustomer)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var result any
	if r.URL.Path == "/storefront/account/profile" {
		result, err = h.accounts.Profile(r.Context(), session.User.ID)
	} else {
		result, err = h.accounts.Purchases(r.Context(), session.User.ID)
	}
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, result)
	case errors.Is(err, backoffice.ErrForbidden):
		writeError(w, http.StatusForbidden, "forbidden")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error")
	}
}
