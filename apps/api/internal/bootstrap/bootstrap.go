package bootstrap

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/authmemory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficeidentity"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/kavenegar"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/neshan"
	"github.com/cherriva/marketplace/apps/api/internal/application/auth"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/application/customer"
	"github.com/cherriva/marketplace/apps/api/internal/application/seller"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	httptransport "github.com/cherriva/marketplace/apps/api/internal/transport/http"
	"net/http"
	"strings"
	"sync"
	"time"
)

func NewHTTPHandler(cfg config.Config) http.Handler {
	return NewHTTPHandlerWithRepository(cfg, backofficememory.New(backofficememory.Fixture()))
}

func NewHTTPHandlerWithRepository(cfg config.Config, repository backoffice.Repository, clocks ...backoffice.Clock) http.Handler {
	h, _ := newCatalogHandler(cfg, repository, nil, clocks...)
	return h
}
func NewHTTPHandlerWithPublicCache(cfg config.Config, repository backoffice.Repository, c *backoffice.PublicCache) http.Handler {
	h, _ := newCatalogHandler(cfg, repository, c)
	return h
}
func newCatalogHandler(cfg config.Config, repository backoffice.Repository, c *backoffice.PublicCache, clocks ...backoffice.Clock) (http.Handler, *backoffice.Service) {
	return newCatalogHandlerRuntime(cfg, repository, c, nil, clocks...)
}
func newCatalogHandlerRuntime(cfg config.Config, repository backoffice.Repository, c *backoffice.PublicCache, logs *OperationLogging, clocks ...backoffice.Clock) (http.Handler, *backoffice.Service) {
	if logs == nil {
		var err error
		logs, err = buildOperationLogging(cfg)
		if err != nil {
			panic("operation log directory must be private")
		}
	}
	store := authmemory.NewStore(nil)
	users := backofficeidentity.New(repository)
	notifier := authmemory.NewNotifier()
	var sender auth.OTPNotifier = notifier
	if cfg.KavenegarAPIKey != "" && cfg.KavenegarTemplate != "" {
		sender = kavenegar.New(cfg.KavenegarAPIKey, cfg.KavenegarTemplate, cfg.KavenegarBaseURL, nil)
	}
	service := auth.NewService(users, store, store, sender, authmemory.Clock{}, authmemory.Random{})
	testData := cfg.Environment == "local" || cfg.Environment == "test"
	handler := httptransport.NewAuthHandler(service, notifier, testData)
	var clock backoffice.Clock = authmemory.Clock{}
	if len(clocks) > 0 && clocks[0] != nil {
		clock = clocks[0]
	}
	workspace := backoffice.NewService(repository, clock, authmemory.Random{})
	workspace.UseOperationLogs(logs.Recorder, logs.Store, func() any {
		if logs.Store != nil {
			return map[string]any{"status": logs.Status, "recorder": logs.Recorder.Stats(), "storage": logs.Store.Stats()}
		}
		return map[string]any{"status": logs.Status, "recorder": logs.Recorder.Stats()}
	})
	if err := workspace.ConfigureGatewaySettingsKey(cfg.PaymentGatewayKey); err != nil {
		panic("PAYMENT_GATEWAY_ENCRYPTION_KEY must be a base64-encoded 32-byte key")
	}
	if err := workspace.ConfigureGatewayCallbackBase(cfg.PaymentPublicBaseURL, testData); err != nil {
		panic("PAYMENT_PUBLIC_BASE_URL must be a valid public HTTPS origin (local HTTP only in local/test)")
	}
	workspace.UsePublicCache(c)
	storefrontRoot := ""
	if cfg.WebSurface == "storefront" {
		storefrontRoot = cfg.WebRoot
	}
	router := httptransport.RouterConfig{ProductDetail: httptransport.NewProductDetailHandler(workspace, storefrontRoot, cfg.Environment != "production"), StaticRoot: cfg.WebRoot, Listing: httptransport.NewListingHandler(workspace, storefrontRoot), StorefrontSurface: cfg.WebSurface == "storefront", Home: httptransport.NewHomeHandler(workspace), Auth: handler, Backoffice: httptransport.NewBackofficeHandler(workspace, service), CatalogIcons: httptransport.NewCatalogIconHandler(workspace, service), ProductWorkspace: httptransport.NewProductWorkspaceHandler(workspace, service), Banners: httptransport.NewBannerHandler(workspace, service)}
	if testData {
		router.SellerOffers = httptransport.NewSellerOffersHandler(seller.NewOffers(repository), service)
	}
	router.ProductReviews = httptransport.NewProductReviewsHandler(workspace, service)
	router.CustomerAccount = httptransport.NewCustomerAccountHandler(customer.NewAccounts(repository), service)
	router.ProductQA = httptransport.NewProductQAHandler(workspace, service)
	router.Commerce = httptransport.NewCommerceHandler(workspace, service)
	var reverse httptransport.ReverseGeocoder
	if cfg.NeshanServiceKey != "" {
		reverse = neshan.New(cfg.NeshanServiceKey)
	}
	router.AddressMap = httptransport.NewAddressMapHandler(workspace, service, cfg.NeshanWebKey, reverse, cfg.AddressMapDailyLimit)
	next := httptransport.NewRouter(router)
	var expiryMu sync.Mutex
	var lastExpiry time.Time
	root := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && (r.URL.Path == "/" || strings.HasPrefix(r.URL.Path, "/storefront/")) {
			expiryMu.Lock()
			if time.Since(lastExpiry) > time.Second {
				ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
				e := workspace.ReleaseExpiredCommerce(ctx)
				cancel()
				if e != nil {
					expiryMu.Unlock()
					http.Error(w, "temporarily unavailable", http.StatusServiceUnavailable)
					return
				}
				lastExpiry = time.Now()
			}
			expiryMu.Unlock()
		}
		next.ServeHTTP(w, r)
	})
	return &loggedHandler{Handler: httptransport.OperationMiddleware(logs.Recorder, root), logs: logs}, workspace
}
