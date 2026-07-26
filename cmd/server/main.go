// Command server is the ONLY assembly point of the app:
// config → infrastructure → domain services → transports → serve.
// Everything below main is injectable and testable; nothing below main
// knows how the pieces are wired.
package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/L-Rocket/my-project/internal/config"
	"github.com/L-Rocket/my-project/internal/infrastructure/logger"
	"github.com/L-Rocket/my-project/internal/infrastructure/memory"
	"github.com/L-Rocket/my-project/internal/infrastructure/postgrest"
	"github.com/L-Rocket/my-project/internal/message"
	"github.com/L-Rocket/my-project/internal/session"
	transportgrpc "github.com/L-Rocket/my-project/internal/transport/grpc"
	transporthttp "github.com/L-Rocket/my-project/internal/transport/http"
	"github.com/L-Rocket/my-project/internal/user"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	log := logger.New(cfg.LogLevel)

	// --- infrastructure: memory by default, PostgREST when configured ---
	var (
		userStore    user.Store
		sessionStore session.Store
		messageStore message.Store
		storeName    string
	)
	if cfg.UsePostgREST() {
		pc := postgrest.New(cfg.SupabaseURL, cfg.SupabaseServiceKey)
		userStore = postgrest.NewUserStore(pc)
		sessionStore = postgrest.NewSessionStore(pc)
		messageStore = postgrest.NewMessageStore(pc)
		storeName = "postgrest"
	} else {
		userStore = memory.NewUserStore()
		sessionStore = memory.NewSessionStore()
		messageStore = memory.NewMessageStore()
		storeName = "memory"
	}
	log.Info("storage selected", "store", storeName)

	// --- domain services: message depends on session rules, not storage ---
	userSvc := user.NewService(userStore)
	sessionSvc := session.NewService(sessionStore)
	messageSvc := message.NewService(messageStore, sessionSvc)

	// --- transports ---
	engine := transporthttp.NewEngine(transporthttp.Deps{
		Logger:   log,
		Users:    userSvc,
		Sessions: sessionSvc,
		Messages: messageSvc,
		Store:    storeName,
	})
	httpSrv := transporthttp.NewServer(fmt.Sprintf(":%d", cfg.Port), engine)

	grpcSrv := transportgrpc.NewServer(log)
	grpcLis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		return fmt.Errorf("grpc listen: %w", err)
	}

	// --- serve both, shut down gracefully on SIGINT/SIGTERM ---
	errCh := make(chan error, 2)
	go func() {
		log.Info("http listening", "addr", httpSrv.Addr)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("http serve: %w", err)
		}
	}()
	go func() {
		log.Info("grpc listening", "addr", grpcLis.Addr().String())
		if err := grpcSrv.Serve(grpcLis); err != nil {
			errCh <- fmt.Errorf("grpc serve: %w", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Info("shutdown signal", "signal", sig.String())
	case err := <-errCh:
		// one transport died: shut the other down too, then report
		log.Error("transport failed", "error", err)
	}

	// in-flight requests get a bounded window to finish
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	grpcSrv.GracefulStop()
	if err := httpSrv.Shutdown(ctx); err != nil {
		return fmt.Errorf("http shutdown: %w", err)
	}
	log.Info("shutdown complete")
	return nil
}
