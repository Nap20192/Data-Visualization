package main

import (
	"context"
	"dv/db"
	"dv/logger"
	"encoding/csv"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
)

const reportsDir = "reports"

type reportDefinition struct {
	filename string
	headers  []string
	fetch    func(context.Context, *db.Queries) ([][]string, error)
}

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
		slog.Error("failed to connect to postgres", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer postgres.Close()
	querier := db.New(postgres.Pool())

	if err := os.MkdirAll(reportsDir, 0o755); err != nil {
		slog.Error("failed to prepare reports directory", slog.String("error", err.Error()))
		os.Exit(1)
	}

	reports := []reportDefinition{
		{
			filename: "actor_role_counts.csv",
			headers:  []string{"person_name", "roles_count", "avg_rating", "avg_popularity"},
			fetch: func(ctx context.Context, q *db.Queries) ([][]string, error) {
				rows, err := q.ActorRoleCounts(ctx)
				if err != nil {
					return nil, err
				}
				records := make([][]string, len(rows))
				for i, row := range rows {
					records[i] = []string{
						row.PersonName,
						formatInt(row.RolesCount),
						formatFloat(row.AvgMovieRating, 2),
						formatFloat(row.AvgMoviePopularity, 2),
					}
				}
				return records, nil
			},
		},
		{
			filename: "country_production_stats.csv",
			headers:  []string{"country_name", "movies_count", "avg_budget", "avg_revenue", "avg_rating"},
			fetch: func(ctx context.Context, q *db.Queries) ([][]string, error) {
				rows, err := q.CountryProductionStats(ctx)
				if err != nil {
					return nil, err
				}
				records := make([][]string, len(rows))
				for i, row := range rows {
					records[i] = []string{
						row.CountryName,
						formatInt(row.MoviesCount),
						formatFloat(row.AvgBudget, 0),
						formatFloat(row.AvgRevenue, 0),
						formatFloat(row.AvgRating, 2),
					}
				}
				return records, nil
			},
		},
		{
			filename: "director_performance.csv",
			headers:  []string{"director_name", "directed_movies", "avg_rating", "avg_revenue", "avg_budget", "total_box_office"},
			fetch: func(ctx context.Context, q *db.Queries) ([][]string, error) {
				rows, err := q.DirectorPerformance(ctx)
				if err != nil {
					return nil, err
				}
				records := make([][]string, len(rows))
				for i, row := range rows {
					records[i] = []string{
						row.DirectorName,
						formatInt(row.DirectedMovies),
						formatFloat(row.AvgRating, 2),
						formatFloat(row.AvgRevenue, 0),
						formatFloat(row.AvgBudget, 0),
						formatInt(row.TotalBoxOffice),
					}
				}
				return records, nil
			},
		},
		{
			filename: "genre_average_metrics.csv",
			headers:  []string{"genre_name", "movies_count", "avg_rating", "avg_popularity", "avg_revenue"},
			fetch: func(ctx context.Context, q *db.Queries) ([][]string, error) {
				rows, err := q.GenreAverageMetrics(ctx)
				if err != nil {
					return nil, err
				}
				records := make([][]string, len(rows))
				for i, row := range rows {
					records[i] = []string{
						row.GenreName,
						formatInt64(row.MoviesCount),
						formatFloat(row.AvgRating, 2),
						formatFloat(row.AvgPopularity, 2),
						formatFloat(row.AvgRevenue, 0),
					}
				}
				return records, nil
			},
		},
		{
			filename: "keyword_trends.csv",
			headers:  []string{"keyword_name", "movies_count", "avg_rating", "avg_revenue"},
			fetch: func(ctx context.Context, q *db.Queries) ([][]string, error) {
				rows, err := q.KeywordTrends(ctx)
				if err != nil {
					return nil, err
				}
				records := make([][]string, len(rows))
				for i, row := range rows {
					records[i] = []string{
						row.KeywordName,
						formatInt(row.MoviesCount),
						formatFloat(row.AvgRating, 2),
						formatFloat(row.AvgRevenue, 0),
					}
				}
				return records, nil
			},
		},
		{
			filename: "language_popularity.csv",
			headers:  []string{"language_name", "movies_count", "avg_rating", "avg_revenue", "avg_popularity"},
			fetch: func(ctx context.Context, q *db.Queries) ([][]string, error) {
				rows, err := q.LanguagePopularity(ctx)
				if err != nil {
					return nil, err
				}
				records := make([][]string, len(rows))
				for i, row := range rows {
					records[i] = []string{
						row.LanguageName,
						formatInt(row.MoviesCount),
						formatFloat(row.AvgRating, 2),
						formatFloat(row.AvgRevenue, 0),
						formatFloat(row.AvgPopularity, 2),
					}
				}
				return records, nil
			},
		},
		{
			filename: "top_profitable_movies.csv",
			headers:  []string{"title", "budget", "revenue", "profit", "roi_percent", "vote_average"},
			fetch: func(ctx context.Context, q *db.Queries) ([][]string, error) {
				rows, err := q.ListTopProfitableMovies(ctx)
				if err != nil {
					return nil, err
				}
				records := make([][]string, len(rows))
				for i, row := range rows {
					records[i] = []string{
						row.Title,
						formatInt(row.Budget),
						formatInt64(row.Revenue),
						formatInt(row.Profit),
						formatFloat(row.RoiPercent, 2),
						formatFloat(row.VoteAverage, 2),
					}
				}
				return records, nil
			},
		},
		{
			filename: "movies_by_decade.csv",
			headers:  []string{"decade", "movies_count", "avg_budget", "avg_revenue", "avg_rating", "avg_runtime"},
			fetch: func(ctx context.Context, q *db.Queries) ([][]string, error) {
				rows, err := q.MoviesByDecade(ctx)
				if err != nil {
					return nil, err
				}
				records := make([][]string, len(rows))
				for i, row := range rows {
					records[i] = []string{
						formatInt(row.Decade),
						formatInt(row.MoviesCount),
						formatFloat(row.AvgBudget, 0),
						formatFloat(row.AvgRevenue, 0),
						formatFloat(row.AvgRating, 2),
						formatFloat(row.AvgRuntime, 0),
					}
				}
				return records, nil
			},
		},
		{
			filename: "runtime_success_segments.csv",
			headers:  []string{"duration_category", "movies_count", "avg_revenue", "avg_rating", "avg_popularity"},
			fetch: func(ctx context.Context, q *db.Queries) ([][]string, error) {
				rows, err := q.RuntimeSuccessSegments(ctx)
				if err != nil {
					return nil, err
				}
				records := make([][]string, len(rows))
				for i, row := range rows {
					records[i] = []string{
						row.DurationCategory,
						formatInt(row.MoviesCount),
						formatFloat(row.AvgRevenue, 0),
						formatFloat(row.AvgRating, 2),
						formatFloat(row.AvgPopularity, 2),
					}
				}
				return records, nil
			},
		},
		{
			filename: "studio_performance.csv",
			headers:  []string{"company_name", "movies_count", "avg_revenue", "avg_rating", "total_revenue"},
			fetch: func(ctx context.Context, q *db.Queries) ([][]string, error) {
				rows, err := q.StudioPerformance(ctx)
				if err != nil {
					return nil, err
				}
				records := make([][]string, len(rows))
				for i, row := range rows {
					records[i] = []string{
						row.CompanyName,
						formatInt(row.MoviesCount),
						formatFloat(row.AvgRevenue, 0),
						formatFloat(row.AvgRating, 2),
						formatInt(row.TotalRevenue),
					}
				}
				return records, nil
			},
		},
	}

	for _, report := range reports {
		if err := exportReport(ctx, querier, reportsDir, report); err != nil {
			slog.Error("failed to export report", slog.String("report", report.filename), slog.String("error", err.Error()))
			os.Exit(1)
		}
	}
}

func exportReport(ctx context.Context, querier *db.Queries, dir string, def reportDefinition) error {
	records, err := def.fetch(ctx, querier)
	if err != nil {
		return err
	}

	path := filepath.Join(dir, def.filename)
	writer, closeFn, err := newFileMultiWriter(path, os.Stdout)
	if err != nil {
		return err
	}
	defer closeFn()

	csvWriter := csv.NewWriter(writer)
	if err := csvWriter.Write(def.headers); err != nil {
		return err
	}
	for _, record := range records {
		if err := csvWriter.Write(record); err != nil {
			return err
		}
	}
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return err
	}

	slog.Info("report exported", slog.String("path", path), slog.String("filename", def.filename), slog.Int("rows", len(records)))
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

func formatFloat(val float64, decimals int) string {
	return strconv.FormatFloat(val, 'f', decimals, 64)
}

func formatInt(val int) string {
	return strconv.Itoa(val)
}

func formatInt64(val int64) string {
	return strconv.FormatInt(val, 10)
}
