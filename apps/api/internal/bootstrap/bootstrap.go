package bootstrap

import (
	"github.com/cherriva/marketplace/apps/api/internal/adapters/authmemory"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficeidentity"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/application/auth"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	httptransport "github.com/cherriva/marketplace/apps/api/internal/transport/http"
	"net/http"
)

func NewHTTPHandler(cfg config.Config) http.Handler {
	repository := backofficememory.New(backofficememory.Fixture())
	store := authmemory.NewStore(nil)
	users := backofficeidentity.New(repository)
	notifier := authmemory.NewNotifier()
	service := auth.NewService(users, store, store, notifier, authmemory.Clock{}, authmemory.Random{})
	handler := httptransport.NewAuthHandler(service, notifier, cfg.Environment == "local" || cfg.Environment == "test")
	workspace := backoffice.NewService(repository, authmemory.Clock{}, authmemory.Random{})
	return httptransport.NewRouter(httptransport.RouterConfig{Auth: handler, Backoffice: httptransport.NewBackofficeHandler(workspace, service), CatalogIcons: httptransport.NewCatalogIconHandler(workspace, service), ProductWorkspace: httptransport.NewProductWorkspaceHandler(workspace, service), Banners: httptransport.NewBannerHandler(workspace, service)})
}
