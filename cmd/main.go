package main

import (
	"context"
	"dv/db"
	"fmt"
	"log/slog"
	"os"
)

func main() {
	postgres, err := db.NewPostgres(context.Background(), db.Config{
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
	res, err := querier.ActorRoleCounts(context.Background())

	if err != nil {
		slog.Error("failed to get actor role counts", slog.String("error", err.Error()))
		os.Exit(1)
	}
	fmt.Println(len(res))
	for _, row := range res {
		fmt.Printf("Actor: %s, Roles: %d, Avg Rating: %.2f, Avg Popularity: %.2f\n", row.PersonName, row.RolesCount, row.AvgMovieRating, row.AvgMoviePopularity)
	}
	secondRes, err := querier.GenreAverageMetrics(context.Background())
	if err != nil {
		slog.Error("failed to get genre average metrics", slog.String("error", err.Error()))
		os.Exit(1)
	}
	fmt.Println(len(secondRes))
	for _, row := range secondRes {
		fmt.Printf("Genre: %s, Movies: %d, Avg Rating: %.2f, Avg Popularity: %.2f, Avg Revenue: %.2f\n", row.GenreName, row.MoviesCount, row.AvgRating, row.AvgPopularity, row.AvgRevenue)
	}


}
