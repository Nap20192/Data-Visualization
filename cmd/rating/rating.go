package main

import (
	"context"
	"dv/db"
	"dv/pkg/logger"
	"log/slog"
	"math/rand/v2"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger.InitLogger("debug")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

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
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	slog.Info("Rating service started")

	for {
		select {
		case <-sigChan:
			slog.Info("Shutdown signal received, stopping gracefully...")
			cancel()
			return
		case <-ctx.Done():
			slog.Info("Context cancelled, exiting...")
			return
		case <-ticker.C:
			if err := processRatings(ctx, queries, postgres); err != nil {
				slog.Error("failed to process ratings", slog.String("error", err.Error()))
			}
		}
	}
}

func processRatings(ctx context.Context, queries *db.Queries, postgres *db.Postgres) error {
	var limit int64 = 100
	offset := rand.IntN(1000)
	movies, err := queries.ListMovies(ctx, db.ListMoviesParams{
		Limit:  limit,
		Offset: int64(offset),
	})
	if err != nil {
		return err
	}

	tx, err := postgres.Pool().Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := queries.WithTx(tx)

	for _, m := range movies {

		times := rand.IntN(100)

		for i := 0; i < times; i++ {
			_, err := qtx.InsertMovieRating(ctx, db.InsertMovieRatingParams{
				RatingValue: 1.0 + (float64(rand.Float32()) * 8.9),
				MovieID:     m.MovieID,
			})

			slog.Debug("inserted rating", slog.String("movie", m.Title.String), slog.Float64("rating", 1.0+(float64(rand.Float32())*8.9)))

			if err != nil {
				return err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	slog.Info("processed ratings", slog.Int("count", len(movies)))
	return nil
}
