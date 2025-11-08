package main

import (
	"context"
	"dv/db"
	"dv/internal"
	"dv/pkg/logger"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	internal.StartDBPoller(ctx, postgres.Pool(), 500*time.Millisecond)

	<-ctx.Done()
	slog.Info("shutting down")

}
