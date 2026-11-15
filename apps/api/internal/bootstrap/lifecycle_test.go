package bootstrap

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"

	"go.uber.org/fx"
)

func serverApp(server *http.Server) *fx.App {
	return fx.New(fx.NopLogger, fx.Supply(server, slog.New(slog.NewTextHandler(io.Discard, nil))), fx.Invoke(RegisterHTTPServer))
}

func TestStartReportsOccupiedPort(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	app := serverApp(&http.Server{Addr: listener.Addr().String()})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := app.Start(ctx); err == nil {
		t.Fatal("expected occupied port to fail startup")
	}
}

func TestStopDrainsActiveRequest(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	listener.Close()
	entered := make(chan struct{})
	release := make(chan struct{})
	server := &http.Server{Addr: address, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-release
		_, _ = io.WriteString(w, "completed")
	})}
	app := serverApp(server)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	response := make(chan error, 1)
	go func() {
		client := &http.Client{Timeout: 4 * time.Second}
		res, err := client.Get("http://" + address)
		if err == nil {
			defer res.Body.Close()
			var body []byte
			body, err = io.ReadAll(res.Body)
			if string(body) != "completed" {
				response <- io.ErrUnexpectedEOF
				return
			}
		}
		response <- err
	}()
	select {
	case <-entered:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	stopped := make(chan error, 1)
	go func() { stopped <- app.Stop(ctx) }()
	select {
	case err := <-stopped:
		t.Fatalf("stopped before request finished: %v", err)
	case <-time.After(50 * time.Millisecond):
	}
	close(release)
	if err := <-response; err != nil {
		t.Fatal(err)
	}
	if err := <-stopped; err != nil {
		t.Fatal(err)
	}
}
