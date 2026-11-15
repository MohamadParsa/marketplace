package main

import (
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthReady(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	res := httptest.NewRecorder()
	bootstrap.NewHTTPHandler(config.Config{Address: ":8080", Environment: "test"}).ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status=%d", res.Code)
	}
}
