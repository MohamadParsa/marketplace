package bootstrap

import (
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"net/http"
	"time"
)

func NewHTTPServer(cfg config.Config) *http.Server {
	return &http.Server{Addr: cfg.Address, Handler: NewHTTPHandler(cfg), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
}
