package bootstrap

import (
	"github.com/cherriva/marketplace/apps/api/internal/adapters/authmemory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficeidentity"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
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
	service := auth.NewService(users, store, store, notifier, authmemory.Clock{}, authmemory.Random{})
	testData := cfg.Environment == "local" || cfg.Environment == "test"
	handler := httptransport.NewAuthHandler(service, notifier, testData)
	workspace := backoffice.NewService(repository, authmemory.Clock{}, authmemory.Random{})
	router := httptransport.RouterConfig{Auth: handler, Backoffice: httptransport.NewBackofficeHandler(workspace, service)}
	if testData {
		router.SellerOffers = httptransport.NewSellerOffersHandler(seller.NewOffers(repository), service)
		router.CustomerAccount = httptransport.NewCustomerAccountHandler(customer.NewAccounts(repository), service)
	}
	return httptransport.NewRouter(router)
}
