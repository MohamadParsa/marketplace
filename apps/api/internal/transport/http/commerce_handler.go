package httptransport

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/application/auth"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

type CommerceHandler struct {
	service     *backoffice.Service
	auth        *auth.Service
	mu          sync.Mutex
	guestLimits map[string]guestLimit
}
type guestLimit struct {
	At    time.Time
	Count int
}

func NewCommerceHandler(s *backoffice.Service, a *auth.Service) *CommerceHandler {
	return &CommerceHandler{service: s, auth: a, guestLimits: map[string]guestLimit{}}
}
func commerceError(w http.ResponseWriter, e error) {
	status, code := 503, "commerce_unavailable"
	switch {
	case errors.Is(e, backoffice.ErrReauthenticationRequired):
		status, code = 403, "reauthentication_required"
	case errors.Is(e, backoffice.ErrUserContactConflict):
		status, code = 409, "duplicate_user_contact"
	case errors.Is(e, backoffice.ErrLastAdministrator):
		status, code = 409, "last_administrator"
	case errors.Is(e, backoffice.ErrForbidden):
		status, code = 403, "forbidden"
	case errors.Is(e, backoffice.ErrNotFound):
		status, code = 404, "not_found"
	case errors.Is(e, backoffice.ErrProfileIncomplete):
		status, code = 400, "recipient_profile_incomplete"
	case errors.Is(e, backoffice.ErrAddressIncomplete):
		status, code = 400, "address_incomplete"
	case errors.Is(e, backoffice.ErrAddressUnavailable):
		status, code = 400, "address_unavailable"
	case errors.Is(e, backoffice.ErrWithdrawalUnavailable):
		status, code = 409, "withdrawal_credit_unavailable"
	case errors.Is(e, backoffice.ErrWalletExceedsTotal):
		status, code = 400, "wallet_exceeds_total"
	case errors.Is(e, backoffice.ErrCouponUnknown):
		status, code = 400, "coupon_invalid"
	case errors.Is(e, backoffice.ErrCouponExpired):
		status, code = 400, "coupon_expired"
	case errors.Is(e, backoffice.ErrCouponNotStarted):
		status, code = 400, "coupon_not_started"
	case errors.Is(e, backoffice.ErrCouponUserQuota):
		status, code = 400, "coupon_user_limit_reached"
	case errors.Is(e, backoffice.ErrCouponQuota):
		status, code = 400, "coupon_quota_reached"
	case errors.Is(e, backoffice.ErrCouponNoDiscount):
		status, code = 400, "coupon_no_discount"
	case errors.Is(e, backoffice.ErrCoupon):
		status, code = 400, "coupon_not_eligible"
	case errors.Is(e, backoffice.ErrInvalid):
		status, code = 400, "invalid_request"
	case errors.Is(e, backoffice.ErrConflict):
		status, code = 409, "state_conflict"
	case errors.Is(e, backoffice.ErrStock):
		status, code = 409, "insufficient_stock"
	case errors.Is(e, backoffice.ErrCheckoutChanged):
		status, code = 409, "checkout_changed"
	case errors.Is(e, backoffice.ErrInsufficientWallet):
		status, code = 409, "insufficient_wallet"
	case errors.Is(e, backoffice.ErrInvalidMethod):
		status, code = 400, "payment_method_not_eligible"
	case errors.Is(e, backoffice.ErrIconTooLarge):
		status, code = 413, "receipt_too_large"
	case errors.Is(e, backoffice.ErrInvalidIcon):
		status, code = 400, "invalid_receipt"
	}
	writeError(w, status, code)
}
func commerceBody(w http.ResponseWriter, r *http.Request, out any) bool {
	media, _, e := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if e != nil || media != "application/json" {
		writeError(w, 400, "json_required")
		return false
	}
	limit := int64(128 * 1024)
	if strings.HasPrefix(r.URL.Path, "/backoffice/stores/") && strings.HasSuffix(r.URL.Path, "/eligibility") {
		limit = 1024 * 1024
	}
	if strings.HasPrefix(r.URL.Path, "/backoffice/axis-conversions/") {
		limit = 8 * 1024 * 1024
	}
	if r.URL.Path == "/backoffice/commerce/sales" || r.URL.Path == "/backoffice/users/export" {
		limit = 2 * 1024 * 1024
	}
	if r.URL.Path == "/backoffice/commerce/settings" {
		limit = 1024 * 1024
	}
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	var body io.Reader = r.Body
	if strings.HasPrefix(r.URL.Path, "/storefront/") {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, 400, "invalid_request")
			return false
		}
		dec := json.NewDecoder(bytes.NewReader(b))
		dec.UseNumber()
		var value any
		if dec.Decode(&value) != nil || dec.Decode(new(any)) != io.EOF {
			writeError(w, 400, "invalid_request")
			return false
		}
		value, err = normalizeReferenceValue(value, "")
		if err != nil {
			writeError(w, 400, "invalid_catalog_reference")
			return false
		}
		b, err = json.Marshal(value)
		if err != nil {
			writeError(w, 400, "invalid_request")
			return false
		}
		body = bytes.NewReader(b)
	}
	d := json.NewDecoder(body)
	d.DisallowUnknownFields()
	if d.Decode(out) != nil || d.Decode(new(any)) != io.EOF {
		writeError(w, 400, "invalid_request")
		return false
	}
	return true
}
func (h *CommerceHandler) actor(w http.ResponseWriter, r *http.Request, surface identity.Surface) (identity.User, bool) {
	session, e := h.auth.Authenticate(r.Context(), bearer(r.Header.Get("Authorization")), surface)
	if e != nil {
		writeError(w, 401, "unauthorized")
		return identity.User{}, false
	}
	return session.User, true
}
func (h *CommerceHandler) owner(w http.ResponseWriter, r *http.Request) (backoffice.CartOwner, bool) {
	if r.Header.Get("Authorization") != "" {
		a, ok := h.actor(w, r, identity.SurfaceCustomer)
		return backoffice.CartOwner{Actor: a}, ok
	}
	token := r.Header.Get("X-Cart-Token")
	if token == "" {
		writeError(w, 401, "cart_token_required")
		return backoffice.CartOwner{}, false
	}
	return backoffice.CartOwner{GuestToken: token}, true
}
func commerceResult(w http.ResponseWriter, r *http.Request, data any, e error) {
	if e != nil {
		commerceError(w, e)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/storefront/") {
		b, err := publicCatalogJSON(data)
		if err != nil {
			commerceError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(200)
		_, _ = w.Write(b)
		return
	}
	writeJSON(w, 200, data)
}
func (h *CommerceHandler) Config(w http.ResponseWriter, r *http.Request) {
	d, e := h.service.PublicCommerceConfig(r.Context())
	if e != nil {
		commerceError(w, e)
		return
	}
	writeCatalogJSON(w, r, d)
}
func (h *CommerceHandler) CreateGuest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	if host == "" {
		host = r.RemoteAddr
	}
	h.mu.Lock()
	now := time.Now()
	limit := h.guestLimits[host]
	if now.Sub(limit.At) > time.Hour {
		limit = guestLimit{At: now}
	}
	limit.Count++
	h.guestLimits[host] = limit
	if len(h.guestLimits) > 10000 {
		for ip, v := range h.guestLimits {
			if now.Sub(v.At) > time.Hour {
				delete(h.guestLimits, ip)
			}
		}
	}
	h.mu.Unlock()
	if limit.Count > 60 {
		writeError(w, 429, "guest_cart_rate_limited")
		return
	}
	d, e := h.service.CreateGuestCart(r.Context())
	commerceResult(w, r, d, e)
}
func (h *CommerceHandler) Cart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	o, ok := h.owner(w, r)
	if !ok {
		return
	}
	if r.Method == "GET" {
		v, e := h.service.ReadCart(r.Context(), o)
		commerceResult(w, r, v, e)
		return
	}
	var in backoffice.CartMutation
	if !commerceBody(w, r, &in) {
		return
	}
	d, e := h.service.MutateCart(r.Context(), o, in, r.Header.Get("Idempotency-Key"))
	commerceResult(w, r, d, e)
}
func (h *CommerceHandler) Customer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	a, ok := h.actor(w, r, identity.SurfaceCustomer)
	if !ok {
		return
	}
	key := r.Header.Get("Idempotency-Key")
	var result any
	var e error
	switch r.URL.Path {
	case "/storefront/cart/merge":
		var in struct {
			GuestToken string `json:"guest_token"`
		}
		if !commerceBody(w, r, &in) {
			return
		}
		result, e = h.service.MergeCart(r.Context(), a, in.GuestToken, key)
	case "/storefront/account/addresses":
		if r.Method == "GET" {
			result, e = h.service.ReadAddresses(r.Context(), a)
		} else {
			var in backoffice.AddressInput
			if !commerceBody(w, r, &in) {
				return
			}
			result, e = h.service.SaveAddress(r.Context(), a, in, key)
		}
	case "/storefront/checkout":
		result, e = h.service.PrepareCheckout(r.Context(), a)
	case "/storefront/checkout/preview":
		var in backoffice.QuoteInput
		if !commerceBody(w, r, &in) {
			return
		}
		result, e = h.service.PreviewCheckout(r.Context(), a, in)
	case "/storefront/checkout/quotes":
		var in backoffice.QuoteInput
		if !commerceBody(w, r, &in) {
			return
		}
		result, e = h.service.CreateQuote(r.Context(), a, in, key)
	case "/storefront/checkout/purchases":
		var in backoffice.AcceptQuoteInput
		if !commerceBody(w, r, &in) {
			return
		}
		result, e = h.service.AcceptQuote(r.Context(), a, in, key)
	case "/storefront/account/orders":
		result, e = h.service.ReadPurchases(r.Context(), a, "")
	case "/storefront/wallet":
		result, e = h.service.ReadWallet(r.Context(), a)
	case "/storefront/wallet/deposits":
		var in backoffice.DepositInput
		if !commerceBody(w, r, &in) {
			return
		}
		result, e = h.service.CreateDeposit(r.Context(), a, in, key)
	case "/storefront/wallet/refunds":
		var in backoffice.RefundInput
		if !commerceBody(w, r, &in) {
			return
		}
		result, e = h.service.RequestRefund(r.Context(), a, in, key)
	default:
		if r.PathValue("id") != "" {
			if r.Method == "GET" {
				rows, err := h.service.ReadPurchases(r.Context(), a, r.PathValue("id"))
				e = err
				if e == nil {
					result = rows[0]
				}
			} else {
				var in struct {
					Reason string `json:"reason"`
				}
				if !commerceBody(w, r, &in) {
					return
				}
				result, e = h.service.CancelPurchase(r.Context(), a, r.PathValue("id"), in.Reason, key)
			}
		} else {
			e = backoffice.ErrNotFound
		}
	}
	commerceResult(w, r, result, e)
}
func (h *CommerceHandler) Admin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	a, ok := h.actor(w, r, identity.SurfaceBackoffice)
	if !ok {
		return
	}
	key := r.Header.Get("Idempotency-Key")
	var result any
	var e error
	switch r.URL.Path {
	case "/backoffice/commerce/sales":
		if r.Method == "GET" {
			result, e = h.service.ReadSales(r.Context(), a)
		} else {
			var in commerce.SalesSettings
			if !commerceBody(w, r, &in) {
				return
			}
			result, e = h.service.SaveSales(r.Context(), a, in, key)
		}
	case "/backoffice/commerce/wallet/credits":
		if r.Method == "GET" {
			result, e = h.service.ReadCustomerCreditWallet(r.Context(), a, r.URL.Query().Get("user_id"), r.Header.Get("Idempotency-Key"))
		} else {
			var in backoffice.AdminCreditInput
			if !commerceBody(w, r, &in) {
				return
			}
			result, e = h.service.CreditCustomerWallet(r.Context(), a, in, key)
		}

	case "/backoffice/commerce/coupons/activate":
		var in backoffice.CouponActivationInput
		if !commerceBody(w, r, &in) {
			return
		}
		result, e = h.service.ActivateCoupon(r.Context(), a, in, key)
	case "/backoffice/commerce/coupons/preview":
		var in backoffice.CouponPreviewInput
		if !commerceBody(w, r, &in) {
			return
		}
		result, e = h.service.PreviewCoupon(r.Context(), a, in)
	case "/backoffice/commerce/coupons":
		if r.Method == "GET" {
			result, e = h.service.ReadCoupons(r.Context(), a)
		} else {
			var in backoffice.CouponInput
			if !commerceBody(w, r, &in) {
				return
			}
			result, e = h.service.SaveCoupon(r.Context(), a, in, key)
		}
	case "/backoffice/commerce":
		result, e = h.service.ReadCommerceAdmin(r.Context(), a)
	case "/backoffice/commerce/settings":
		if r.Method == "GET" {
			result, e = h.service.ReadCommerceSettings(r.Context(), a)
		} else {
			var in backoffice.CommerceSettingsInput
			if !commerceBody(w, r, &in) {
				return
			}
			result, e = h.service.SaveCommerceSettings(r.Context(), a, in, key)
		}
	case "/backoffice/commerce/deposits/review":
		var in backoffice.DepositReviewInput
		if !commerceBody(w, r, &in) {
			return
		}
		result, e = h.service.ReviewDeposit(r.Context(), a, in, key)
	case "/backoffice/commerce/orders/confirm":
		var in backoffice.PurchaseAdminInput
		if !commerceBody(w, r, &in) {
			return
		}
		result, e = h.service.ConfirmPurchase(r.Context(), a, in, key)
	case "/backoffice/commerce/orders/cancel":
		var in backoffice.PurchaseAdminInput
		if !commerceBody(w, r, &in) {
			return
		}
		result, e = h.service.AdminCancelPurchase(r.Context(), a, in, key)
	case "/backoffice/commerce/refunds/resolve":
		var in backoffice.RefundResolveInput
		if !commerceBody(w, r, &in) {
			return
		}
		result, e = h.service.ResolveRefund(r.Context(), a, in, key)
	default:
		e = backoffice.ErrNotFound
	}
	commerceResult(w, r, result, e)
}
func (h *CommerceHandler) Seller(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	a, ok := h.actor(w, r, identity.SurfaceSeller)
	if !ok {
		return
	}
	out, e := h.service.ReadSellerCheckoutOrders(r.Context(), a, r.PathValue("id"))
	commerceResult(w, r, out, e)
}
func (h *CommerceHandler) Receipt(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	admin := strings.HasPrefix(r.URL.Path, "/backoffice/")
	surface := identity.SurfaceCustomer
	if admin {
		surface = identity.SurfaceBackoffice
	}
	a, ok := h.actor(w, r, surface)
	if !ok {
		return
	}
	data, e := h.service.ReadReceipt(r.Context(), a, r.PathValue("id"), admin)
	if e != nil {
		commerceError(w, e)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Write(data)
}
func (h *CommerceHandler) UploadReceipt(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	a, ok := h.actor(w, r, identity.SurfaceCustomer)
	if !ok {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, (2<<20)+16*1024)
	if e := r.ParseMultipartForm((2 << 20) + 16*1024); e != nil {
		var large *http.MaxBytesError
		if errors.As(e, &large) {
			commerceError(w, backoffice.ErrIconTooLarge)
		} else {
			commerceError(w, backoffice.ErrInvalid)
		}
		return
	}
	defer r.MultipartForm.RemoveAll()
	if len(r.MultipartForm.Value) != 1 || len(r.MultipartForm.Value["version"]) != 1 || len(r.MultipartForm.File) != 1 || len(r.MultipartForm.File["file"]) != 1 {
		commerceError(w, backoffice.ErrInvalid)
		return
	}
	version, e := strconv.Atoi(r.FormValue("version"))
	if e != nil {
		commerceError(w, backoffice.ErrInvalid)
		return
	}
	f, _, e := r.FormFile("file")
	if e != nil {
		commerceError(w, backoffice.ErrInvalid)
		return
	}
	defer f.Close()
	data, e := io.ReadAll(io.LimitReader(f, (2<<20)+1))
	if e != nil {
		commerceError(w, backoffice.ErrInvalid)
		return
	}
	d, e := h.service.AttachReceipt(r.Context(), a, r.PathValue("id"), version, r.Header.Get("Idempotency-Key"), data)
	commerceResult(w, r, d, e)
}

func (h *CommerceHandler) AuthorizeExport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	surface := identity.SurfaceBackoffice
	if strings.HasPrefix(r.URL.Path, "/seller/") {
		surface = identity.SurfaceSeller
	}
	a, ok := h.actor(w, r, surface)
	if !ok {
		return
	}
	if err := h.service.AuthorizeGridExport(r.Context(), a); err != nil {
		commerceError(w, err)
		return
	}
	w.WriteHeader(204)
}

func (h *CommerceHandler) Users(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	actor, ok := h.actor(w, r, identity.SurfaceBackoffice)
	if !ok {
		return
	}
	if strings.HasSuffix(r.URL.Path, "/save") {
		var in backoffice.UserInput
		if !commerceBody(w, r, &in) {
			return
		}
		out, e := h.service.SaveUser(r.Context(), actor, in, r.Header.Get("Idempotency-Key"))
		commerceResult(w, r, out, e)
		return
	}
	var in struct {
		Filter backoffice.UserFilter `json:"filter"`
		IDs    []string              `json:"ids"`
	}
	if !commerceBody(w, r, &in) {
		return
	}
	out, e := h.service.ListUsers(r.Context(), actor, in.Filter, strings.HasSuffix(r.URL.Path, "/export"), in.IDs)
	commerceResult(w, r, out, e)
}
