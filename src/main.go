package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mantit0985/go-template/src/core"
	"github.com/mantit0985/go-template/src/core/api"
	corelog "github.com/mantit0985/go-template/src/core/log"
	"github.com/mantit0985/go-template/src/core/telemetry"
	"github.com/mantit0985/go-template/src/core/validate"
	"github.com/mantit0985/go-template/src/features/example"
	"github.com/mantit0985/go-template/src/features/health"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// 1. Basic Platform Init
	corelog.InitLogger()
	core.InitConfig()

	// 2. Telemetry Setup
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tp, err := telemetry.InitTracer(ctx, "go-template-service", "localhost:4317")
	if err != nil {
		log.Fatalf("failed to initialize tracer: %v", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), telemetry.GetShutdownTimeout())
		defer cancel()
		if err := tp.Shutdown(shutdownCtx); err != nil {
			log.Printf("error shutting down tracer provider: %v", err)
		}
	}()
	telemetry.SetupGlobalTelemetry()

	// 3. Router & Validator Setup
	router := api.NewRouter()
	validator := validate.NewValidator()
	router.SetValidator(validator)

	// 4. Feature Registration
	router.Route("/api", func(sub *api.Router) {
		example.RegisterHandlers(sub)
	})

	// Public Observability Endpoints
	healthSvc := health.NewService()
	healthHdl := health.NewHandler(healthSvc)
	router.PublicHandle("/health", http.HandlerFunc(healthHdl.HealthCheck))
	router.PublicHandle("/healthz", http.HandlerFunc(healthHdl.HealthCheck))
	router.PublicHandle("/readyz", http.HandlerFunc(healthHdl.ReadyCheck))
	router.PublicHandle("/metrics", promhttp.Handler())
	router.PublicHandle("/version", http.HandlerFunc(healthHdl.VersionCheck))

	// 5. Server Start
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		fmt.Println("Server starting on :8080...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelShutdown()
	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
