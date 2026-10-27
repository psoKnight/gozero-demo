package healthing

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
)

// Handler serves the shared readiness endpoint used by all services.
func Handler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("ok\n"))
}

// Start serves /status.html and stops gracefully with the go-zero process.
func Start(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/status.html", Handler)
	server := &http.Server{
		Addr:              address,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logx.Errorf("health server stopped unexpectedly: %v", err)
		}
	}()

	proc.AddShutdownListener(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	})

	return nil
}
