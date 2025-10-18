package main

import (
	"context"
	"dv/db"
	"dv/pkg/logger"
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
	"strconv"
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

	// Export all ratings
	if err := exportRatings(ctx, queries); err != nil {
		slog.Error("failed to export ratings", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Export all movie ratings
	if err := exportMovieRatings(ctx, queries); err != nil {
		slog.Error("failed to export movie ratings", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Export movie_rating table
	if err := exportMovieRatingTable(ctx, queries); err != nil {
		slog.Error("failed to export movie_rating table", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("Successfully exported all ratings to CSV files")
}

func exportRatings(ctx context.Context, queries *db.Queries) error {
	slog.Info("Fetching all ratings...")

	ratings, err := queries.GetAllRatings(ctx)
	if err != nil {
		return fmt.Errorf("failed to get ratings: %w", err)
	}

	// Create reports directory if it doesn't exist
	if err := os.MkdirAll("./reports", 0755); err != nil {
		return fmt.Errorf("failed to create reports directory: %w", err)
	}

	// Create CSV file
	file, err := os.Create("./reports/all_ratings.csv")
	if err != nil {
		return fmt.Errorf("failed to create ratings CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"ID", "Rating Value", "Created At"}); err != nil {
		return fmt.Errorf("failed to write ratings header: %w", err)
	}

	// Write data
	for _, rating := range ratings {
		record := []string{
			strconv.Itoa(int(rating.ID)),
			fmt.Sprintf("%.1f", rating.RatingValue),
			rating.CreatedAt.Time.Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write rating record: %w", err)
		}
	}

	slog.Info("Exported ratings", slog.Int("count", len(ratings)))
	return nil
}

func exportMovieRatings(ctx context.Context, queries *db.Queries) error {
	slog.Info("Fetching all movie ratings...")

	movieRatings, err := queries.GetAllMovieRatings(ctx)
	if err != nil {
		return fmt.Errorf("failed to get movie ratings: %w", err)
	}

	// Create CSV file
	file, err := os.Create("./reports/all_movie_ratings.csv")
	if err != nil {
		return fmt.Errorf("failed to create movie ratings CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"Movie ID", "Movie Title", "Rating ID", "Rating Value", "Created At"}); err != nil {
		return fmt.Errorf("failed to write movie ratings header: %w", err)
	}

	// Write data
	for _, mr := range movieRatings {
		record := []string{
			strconv.Itoa(int(mr.MovieID)),
			mr.Title.String,
			strconv.Itoa(int(mr.RatingID)),
			fmt.Sprintf("%.1f", mr.RatingValue),
			mr.CreatedAt.Time.Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write movie rating record: %w", err)
		}
	}

	slog.Info("Exported movie ratings", slog.Int("count", len(movieRatings)))
	return nil
}

func exportMovieRatingTable(ctx context.Context, queries *db.Queries) error {
	slog.Info("Fetching movie_rating table...")

	movieRatingTable, err := queries.GetAllMovieRatingTable(ctx)
	if err != nil {
		return fmt.Errorf("failed to get movie_rating table: %w", err)
	}

	// Create CSV file
	file, err := os.Create("./reports/movie_rating_table.csv")
	if err != nil {
		return fmt.Errorf("failed to create movie_rating table CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"Movie ID", "Rating ID"}); err != nil {
		return fmt.Errorf("failed to write movie_rating table header: %w", err)
	}

	// Write data
	for _, mr := range movieRatingTable {
		record := []string{
			strconv.Itoa(int(mr.MovieID)),
			strconv.Itoa(int(mr.RatingID)),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write movie_rating table record: %w", err)
		}
	}

	slog.Info("Exported movie_rating table", slog.Int("count", len(movieRatingTable)))
	return nil
}
