package main

import (
	"context"
	"dv/db"
	"encoding/csv"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
)

const reportsDir = "reports"

func main() {
	ctx := context.Background()

	postgres, err := db.NewPostgres(ctx, db.Config{
		Host:     "localhost",
		Port:     5432,
		Dbname:   "movies",
		Username: "postgres",
		Password: "postgres",
	})
	if err != nil {
		slog.Error("failed to connect to postgres", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer postgres.Close()
	querier := db.New(postgres.Pool())

	if err := os.MkdirAll(reportsDir, 0o755); err != nil {
		slog.Error("failed to prepare reports directory", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := exportActorRoleCounts(ctx, querier, reportsDir); err != nil {
		slog.Error("failed to export actor role counts", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := exportGenreAverageMetrics(ctx, querier, reportsDir); err != nil {
		slog.Error("failed to export genre metrics", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := exportCountryProductionStats(ctx, querier, reportsDir); err != nil {
		slog.Error("failed to export country production stats", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func exportActorRoleCounts(ctx context.Context, querier *db.Queries, dir string) error {
	results, err := querier.ActorRoleCounts(ctx)
	if err != nil {
		return err
	}

	path := filepath.Join(dir, "actor_role_counts.csv")
	writer, closeFn, err := newFileMultiWriter(path, os.Stdout)
	if err != nil {
		return err
	}
	defer closeWithLog(path, closeFn)

	csvWriter := csv.NewWriter(writer)
	defer flushWithLog(path, csvWriter)

	if err := csvWriter.Write([]string{"person_name", "roles_count", "avg_rating", "avg_popularity"}); err != nil {
		return err
	}

	for _, row := range results {
		record := []string{
			row.PersonName,
			strconv.Itoa(row.RolesCount),
			formatFloat(row.AvgMovieRating, 2),
			formatFloat(row.AvgMoviePopularity, 2),
		}
		if err := csvWriter.Write(record); err != nil {
			return err
		}
	}

	slog.Info("actor role counts exported", slog.String("path", path), slog.Int("rows", len(results)))
	return nil
}

func exportGenreAverageMetrics(ctx context.Context, querier *db.Queries, dir string) error {
	results, err := querier.GenreAverageMetrics(ctx)
	if err != nil {
		return err
	}

	path := filepath.Join(dir, "genre_average_metrics.csv")
	writer, closeFn, err := newFileMultiWriter(path, os.Stdout)
	if err != nil {
		return err
	}
	defer closeWithLog(path, closeFn)

	csvWriter := csv.NewWriter(writer)
	defer flushWithLog(path, csvWriter)

	if err := csvWriter.Write([]string{"genre_name", "movies_count", "avg_rating", "avg_popularity", "avg_revenue"}); err != nil {
		return err
	}

	for _, row := range results {
		record := []string{
			row.GenreName,
			strconv.FormatInt(row.MoviesCount, 10),
			formatFloat(row.AvgRating, 2),
			formatFloat(row.AvgPopularity, 2),
			formatFloat(row.AvgRevenue, 0),
		}
		if err := csvWriter.Write(record); err != nil {
			return err
		}
	}

	slog.Info("genre metrics exported", slog.String("path", path), slog.Int("rows", len(results)))
	return nil
}

func exportCountryProductionStats(ctx context.Context, querier *db.Queries, dir string) error {
	results, err := querier.CountryProductionStats(ctx)
	if err != nil {
		return err
	}

	path := filepath.Join(dir, "country_production_stats.csv")
	writer, closeFn, err := newFileMultiWriter(path, os.Stdout)
	if err != nil {
		return err
	}
	defer closeWithLog(path, closeFn)

	csvWriter := csv.NewWriter(writer)
	defer flushWithLog(path, csvWriter)

	if err := csvWriter.Write([]string{"country_name", "movies_count", "avg_budget", "avg_revenue", "avg_rating"}); err != nil {
		return err
	}

	for _, row := range results {
		record := []string{
			row.CountryName,
			strconv.Itoa(row.MoviesCount),
			formatFloat(row.AvgBudget, 0),
			formatFloat(row.AvgRevenue, 0),
			formatFloat(row.AvgRating, 2),
		}
		if err := csvWriter.Write(record); err != nil {
			return err
		}
	}

	slog.Info("country production stats exported", slog.String("path", path), slog.Int("rows", len(results)))
	return nil
}

func newFileMultiWriter(path string, extra ...io.Writer) (io.Writer, func() error, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, nil, err
	}

	writers := append([]io.Writer{file}, extra...)
	return io.MultiWriter(writers...), file.Close, nil
}

func closeWithLog(path string, closeFn func() error) {
	if err := closeFn(); err != nil {
		slog.Warn("failed to close report file", slog.String("path", path), slog.String("error", err.Error()))
	}
}

func flushWithLog(path string, w *csv.Writer) {
	w.Flush()
	if err := w.Error(); err != nil {
		slog.Warn("failed to flush csv writer", slog.String("path", path), slog.String("error", err.Error()))
	}
}

func formatFloat(val float64, decimals int) string {
	return strconv.FormatFloat(val, 'f', decimals, 64)
}
