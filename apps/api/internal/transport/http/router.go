package httptransport

import (
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type RouterConfig struct {
	Auth             *AuthHandler
	Backoffice       *BackofficeHandler
	CatalogIcons     *CatalogIconHandler
	ProductWorkspace *ProductWorkspaceHandler
	Banners          *BannerHandler
}

func NewRouter(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()
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
		mux.HandleFunc("POST /backoffice/catalog/{kind}/{target}/icon", cfg.CatalogIcons.Upload)
	}
	if root := os.Getenv("WEB_ROOT"); root != "" {
		mux.Handle("GET /", staticSite(root))
	} else {
		mux.HandleFunc("GET /{$}", health("marketplace-api"))
	}
	mux.HandleFunc("POST /backoffice/auth/request-otp", cfg.Auth.RequestOTP(identity.SurfaceBackoffice))
	mux.HandleFunc("POST /backoffice/auth/verify-otp", cfg.Auth.VerifyOTP(identity.SurfaceBackoffice))
	mux.HandleFunc("POST /backoffice/auth/logout", cfg.Auth.Logout)
	mux.HandleFunc("POST /seller/auth/request-otp", cfg.Auth.RequestOTP(identity.SurfaceSeller))
	mux.HandleFunc("POST /seller/auth/verify-otp", cfg.Auth.VerifyOTP(identity.SurfaceSeller))
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
		if strings.HasPrefix(r.URL.Path, "/backoffice/") || strings.HasPrefix(r.URL.Path, "/seller/") {
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
		if strings.HasPrefix(r.URL.Path, "/backoffice/") {
			w.Header().Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(w, r)
	})
}
