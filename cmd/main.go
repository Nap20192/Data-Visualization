package main

import (
	"context"
	"dv/db"
	"dv/internal"
	"dv/pkg/logger"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	logger.InitLogger("debug")
	// use a cancellable context that is cancelled on SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	postgres, err := db.NewPostgres(ctx, db.Config{
		Host:     "localhost",
		Port:     5432,
		Dbname:   "movies",
		Username: "postgres",
		Password: "postgres",
	})

	if err != nil {
		slog.Error("failed to connect to Postgres", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer postgres.Close()


	// Start Binance exporter for multiple cryptocurrency pairs
	symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT", "SOLUSDT", "XRPUSDT"}
	binanceExporter := internal.NewBinanceExporter(symbols)

	// Start Prometheus HTTP server
	http.Handle("/metrics", promhttp.Handler())
	server := &http.Server{Addr: ":8000"}

	go func() {
		if err := binanceExporter.Start(ctx); err != nil {
			slog.Error("failed to start Binance exporter", slog.String("error", err.Error()))
			os.Exit(1)
		}
		slog.Info("Binance exporter started", slog.Any("symbols", symbols))
		slog.Info("Starting Prometheus metrics server", slog.String("addr", ":8000"))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Prometheus server error", slog.String("error", err.Error()))
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Error shutting down server", slog.String("error", err.Error()))
	}
}
