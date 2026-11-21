package bootstrap

import (
	"github.com/cherriva/marketplace/apps/api/internal/adapters/authmemory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficeidentity"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/kavenegar"
	"github.com/cherriva/marketplace/apps/api/internal/application/auth"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/application/customer"
	"github.com/cherriva/marketplace/apps/api/internal/application/seller"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	httptransport "github.com/cherriva/marketplace/apps/api/internal/transport/http"
	"net/http"
)

func NewHTTPHandler(cfg config.Config) http.Handler {
	return NewHTTPHandlerWithRepository(cfg, backofficememory.New(backofficememory.Fixture()))
}

func NewHTTPHandlerWithRepository(cfg config.Config, repository backoffice.Repository) http.Handler {
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
	workspace := backoffice.NewService(repository, authmemory.Clock{}, authmemory.Random{})
	storefrontRoot := ""
	if cfg.WebSurface == "storefront" {
		storefrontRoot = cfg.WebRoot
	}
	router := httptransport.RouterConfig{ProductDetail: httptransport.NewProductDetailHandler(workspace, storefrontRoot, cfg.Environment != "production"), StaticRoot: cfg.WebRoot, Listing: httptransport.NewListingHandler(workspace, storefrontRoot), StorefrontSurface: cfg.WebSurface == "storefront", Home: httptransport.NewHomeHandler(workspace), Auth: handler, Backoffice: httptransport.NewBackofficeHandler(workspace, service), CatalogIcons: httptransport.NewCatalogIconHandler(workspace, service), ProductWorkspace: httptransport.NewProductWorkspaceHandler(workspace, service), Banners: httptransport.NewBannerHandler(workspace, service)}
	if testData {
		router.SellerOffers = httptransport.NewSellerOffersHandler(seller.NewOffers(repository), service)
		router.CustomerAccount = httptransport.NewCustomerAccountHandler(customer.NewAccounts(repository), service)
	}
	return httptransport.NewRouter(router)
}
