package main

import (
	"context"
	"dv/db"
	"dv/internal"
	"dv/pkg/logger"
	"log/slog"
	"os"
)

func main() {
	logger.InitLogger("debug")
	ctx := context.Background()

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
	queries := db.New(postgres.Pool())

	chartsService := internal.NewCharts(queries,"./charts/")

	if err := chartsService.GenerateAllCharts(); err != nil {
		slog.Error("failed to generate charts", slog.String("error", err.Error()))
		os.Exit(1)
	}

}
