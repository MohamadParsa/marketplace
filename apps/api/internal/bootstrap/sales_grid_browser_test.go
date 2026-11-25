package bootstrap_test

import (
	"fmt"
	"github.com/cherriva/marketplace/apps/api/internal/adapters/backofficememory"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/bootstrap"
	"github.com/cherriva/marketplace/apps/api/internal/config"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSalesGridBrowserHarness(t *testing.T) {
	if os.Getenv("CHERRIVA_SALES_GRID_BROWSER") != "1" {
		t.Skip("isolated sales/grid browser fixture")
	}
	w := commerceSeed()
	w.Users = append(w.Users, backoffice.User{ID: "inactive-admin-fixture", Name: "ادمین غیرفعال آزمون", Phone: "09120000888", Role: identity.RoleSuperAdmin})
	w.IconMedia[0].Data = iconPNG(t, 12)
	w.Categories = append(w.Categories, catalog.Category{ID: "child", Name: "فرزند دیجیتال", ParentID: "regular", Published: true, Version: 1})
	w.Products[0].CategoryID = "child"
	for i := 0; i < 150; i++ {
		name := fmt.Sprintf("کاربر نمونه %03d", i)
		if i == 100 {
			name = "=SUM(1,1)"
		}
		w.Users = append(w.Users, backoffice.User{ID: fmt.Sprintf("grid-user-%03d", i), Name: name, Phone: fmt.Sprintf("09%09d", 10000+i), Active: true, Role: identity.RoleViewer})
	}
	repo := backofficememory.New(w)
	for _, entry := range []struct{ port, surface, app string }{{"8097", "backoffice", "backoffice-web"}, {"5199", "storefront", "storefront-web"}} {
		listener, e := net.Listen("tcp", "127.0.0.1:"+entry.port)
		if e != nil {
			t.Fatal(e)
		}
		h := bootstrap.NewHTTPHandlerWithRepository(config.Config{Environment: "test", WebSurface: entry.surface, WebRoot: filepath.Join("..", "..", "..", entry.app, "build")}, repo)
		server := &http.Server{Handler: h, ReadHeaderTimeout: 5 * time.Second}
		t.Cleanup(func() { server.Close() })
		go server.Serve(listener)
	}
	t.Log("sales/grid harness ready:8097 and5199; memory only")
	time.Sleep(15 * time.Minute)
}
