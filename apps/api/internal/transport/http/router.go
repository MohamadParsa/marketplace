package httptransport

import (
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type RouterConfig struct {
	Listing           *ListingHandler
	StaticRoot        string
	StorefrontSurface bool
	Home              *HomeHandler
	Auth              *AuthHandler
	Backoffice        *BackofficeHandler
	CatalogIcons      *CatalogIconHandler
	ProductWorkspace  *ProductWorkspaceHandler
	Banners           *BannerHandler
	SellerOffers      *SellerOffersHandler
	CustomerAccount   *CustomerAccountHandler
}

func NewRouter(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()
	if cfg.Listing != nil {
		mux.HandleFunc("GET /storefront/catalog", cfg.Listing.Read)
		mux.HandleFunc("GET /storefront/suggestions", cfg.Listing.Suggest)
		if cfg.StorefrontSurface {
			mux.HandleFunc("GET /category/{id}", cfg.Listing.Page)
			mux.HandleFunc("GET /search", cfg.Listing.Page)
			mux.HandleFunc("GET /offers", cfg.Listing.Page)
		}
	}
	if cfg.Home != nil {
		mux.HandleFunc("GET /storefront/home", cfg.Home.Read)
	}
	mux.HandleFunc("GET /health/live", health("ok"))
	mux.HandleFunc("GET /health/ready", health("ready"))
	if cfg.Banners != nil {
		mux.HandleFunc("GET /backoffice/banners", cfg.Banners.List)
		mux.HandleFunc("POST /backoffice/banners/{id}/media/{slot}", cfg.Banners.Upload)
		mux.HandleFunc("GET /backoffice/banner-media/{id}", func(w http.ResponseWriter, r *http.Request) {
			r.SetPathValue("access", "private")
			cfg.Banners.Media(w, r)
		})
		mux.HandleFunc("GET /content/banners", cfg.Banners.PublicList)
		mux.HandleFunc("GET /content/banners/{id}/view", cfg.Banners.View)
		mux.HandleFunc("GET /content/banner-media/{id}", cfg.Banners.Media)
	}
	if cfg.Backoffice != nil {
		mux.Handle("GET /backoffice/workspace", cfg.Backoffice)
		mux.Handle("POST /backoffice/commands", cfg.Backoffice)
		mux.Handle("GET /seller/workspace", cfg.Backoffice)
		mux.Handle("POST /seller/commands", cfg.Backoffice)
	}
	if cfg.ProductWorkspace != nil {
		for _, surface := range []string{"backoffice", "seller"} {
			mux.Handle("GET /"+surface+"/products/{id}/workspace", cfg.ProductWorkspace)
			mux.Handle("GET /"+surface+"/products/{id}/resolve", cfg.ProductWorkspace)
		}
	}
	if cfg.CatalogIcons != nil {
		mux.HandleFunc("GET /catalog/icons/{id}", cfg.CatalogIcons.Read)
		mux.HandleFunc("GET /backoffice/catalog/media/{id}", cfg.CatalogIcons.Preview)
		mux.HandleFunc("POST /backoffice/catalog/{kind}/{target}/icon", cfg.CatalogIcons.Upload)
	}
	if cfg.SellerOffers != nil {
		mux.Handle("GET /seller/stores/{storeID}/offers", cfg.SellerOffers)
		mux.Handle("PUT /seller/stores/{storeID}/offers/{offerID}/stock", cfg.SellerOffers)
	}
	if cfg.CustomerAccount != nil {
		mux.Handle("GET /storefront/account/profile", cfg.CustomerAccount)
		mux.HandleFunc("PATCH /storefront/account/profile", cfg.Auth.UpdateProfile(identity.SurfaceCustomer))
		mux.Handle("GET /storefront/account/purchases", cfg.CustomerAccount)
		mux.HandleFunc("POST /storefront/auth/request-otp", cfg.Auth.RequestOTP(identity.SurfaceCustomer))
		mux.HandleFunc("POST /storefront/auth/verify-otp", cfg.Auth.VerifyOTP(identity.SurfaceCustomer))
		mux.HandleFunc("POST /storefront/auth/password-login", cfg.Auth.PasswordLogin(identity.SurfaceCustomer))
		mux.HandleFunc("GET /storefront/account/me", cfg.Auth.Profile(identity.SurfaceCustomer))
		mux.HandleFunc("POST /storefront/auth/password-recovery/request-otp", cfg.Auth.RequestRecoveryOTP(identity.SurfaceCustomer))
		mux.HandleFunc("POST /storefront/auth/password-recovery/reset", cfg.Auth.RecoverPassword(identity.SurfaceCustomer))
		mux.HandleFunc("POST /storefront/auth/recover/request-otp", cfg.Auth.RequestRecoveryOTP(identity.SurfaceCustomer))
		mux.HandleFunc("POST /storefront/auth/recover/reset", cfg.Auth.RecoverPassword(identity.SurfaceCustomer))
		mux.HandleFunc("POST /storefront/auth/logout", cfg.Auth.Logout)
	}
	root := cfg.StaticRoot
	if root == "" {
		root = os.Getenv("WEB_ROOT")
	}
	if root != "" {
		mux.Handle("GET /", staticSite(root))
	} else {
		mux.HandleFunc("GET /{$}", health("marketplace-api"))
	}
	mux.HandleFunc("POST /backoffice/auth/request-otp", cfg.Auth.RequestOTP(identity.SurfaceBackoffice))
	mux.HandleFunc("POST /backoffice/auth/verify-otp", cfg.Auth.VerifyOTP(identity.SurfaceBackoffice))
	mux.HandleFunc("POST /backoffice/auth/password-login", cfg.Auth.PasswordLogin(identity.SurfaceBackoffice))
	mux.HandleFunc("GET /backoffice/account/profile", cfg.Auth.Profile(identity.SurfaceBackoffice))
	mux.HandleFunc("PATCH /backoffice/account/profile", cfg.Auth.UpdateProfile(identity.SurfaceBackoffice))
	mux.HandleFunc("POST /backoffice/auth/password-recovery/request-otp", cfg.Auth.RequestRecoveryOTP(identity.SurfaceBackoffice))
	mux.HandleFunc("POST /backoffice/auth/password-recovery/reset", cfg.Auth.RecoverPassword(identity.SurfaceBackoffice))
	mux.HandleFunc("POST /backoffice/auth/admin-password-reset", cfg.Auth.AdminResetPassword)
	mux.HandleFunc("POST /backoffice/auth/request-admin-otp", cfg.Auth.RequestAdminOTP)
	mux.HandleFunc("POST /backoffice/auth/logout", cfg.Auth.Logout)
	mux.HandleFunc("POST /seller/auth/request-otp", cfg.Auth.RequestOTP(identity.SurfaceSeller))
	mux.HandleFunc("POST /seller/auth/verify-otp", cfg.Auth.VerifyOTP(identity.SurfaceSeller))
	mux.HandleFunc("POST /seller/auth/password-login", cfg.Auth.PasswordLogin(identity.SurfaceSeller))
	mux.HandleFunc("GET /seller/account/profile", cfg.Auth.Profile(identity.SurfaceSeller))
	mux.HandleFunc("PATCH /seller/account/profile", cfg.Auth.UpdateProfile(identity.SurfaceSeller))
	mux.HandleFunc("POST /seller/auth/password-recovery/request-otp", cfg.Auth.RequestRecoveryOTP(identity.SurfaceSeller))
	mux.HandleFunc("POST /seller/auth/password-recovery/reset", cfg.Auth.RecoverPassword(identity.SurfaceSeller))
	mux.HandleFunc("POST /seller/auth/logout", cfg.Auth.Logout)
	return securityHeaders(mux)
}
func health(status string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": status})
	}
}

func staticSite(root string) http.Handler {
	files := http.FileServer(http.Dir(root))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if privateRoute(r.URL.Path) {
			http.NotFound(w, r)
			return
		}
		path := filepath.Join(root, filepath.Clean("/"+r.URL.Path))
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			files.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFile(w, r, filepath.Join(root, "index.html"))
	})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "same-origin")
		if privateRoute(r.URL.Path) {
			w.Header().Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(w, r)
	})
}

func privateRoute(path string) bool {
	return strings.HasPrefix(path, "/backoffice/") || strings.HasPrefix(path, "/seller/") || strings.HasPrefix(path, "/storefront/")
}
