package main

import (
	"context"
	"dv/db"
	"dv/pkg/logger"
	"encoding/json"
	"fmt"
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

	// Export all movies with ratings
	if err := exportMoviesWithRatings(ctx, queries); err != nil {
		slog.Error("failed to export movies with ratings", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("Successfully exported movies with ratings to JSON file")
}

func exportMoviesWithRatings(ctx context.Context, queries *db.Queries) error {
	slog.Info("Fetching all movies with ratings...")

	moviesWithRatings, err := queries.GetAllMoviesWithRatings(ctx)
	if err != nil {
		return fmt.Errorf("failed to get movies with ratings: %w", err)
	}

	// Create reports directory if it doesn't exist
	if err := os.MkdirAll("./reports", 0755); err != nil {
		return fmt.Errorf("failed to create reports directory: %w", err)
	}

	// Create JSON file
	file, err := os.Create("./reports/movies_with_ratings.json")
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer file.Close()

	// Marshal to JSON with indentation
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "")

	if err := encoder.Encode(moviesWithRatings); err != nil {
		return fmt.Errorf("failed to encode to JSON: %w", err)
	}

	slog.Info("Exported movies with ratings", slog.Int("count", len(moviesWithRatings)))
	return nil
}
