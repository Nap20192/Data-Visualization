package internal

import (
	"context"
	"fmt"
	"time"

	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

// StartDBPoller launches a goroutine that every `interval` polls the database
// for all tables in the public schema and logs their row counts. It stops when
// the provided ctx is done.
func StartDBPoller(ctx context.Context, pool *pgxpool.Pool, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		slog.Info("db poller started", slog.Duration("interval", interval))
		for {
			select {
			case <-ctx.Done():
				slog.Info("db poller stopping")
				return
			case <-ticker.C:
				pollOnce(ctx, pool)
			}
		}
	}()
}

func pollOnce(ctx context.Context, pool *pgxpool.Pool) {
	rows, err := pool.Query(ctx, "SELECT tablename FROM pg_tables WHERE schemaname='public'")
	if err != nil {
		slog.Error("failed to list tables", slog.String("error", err.Error()))
		return
	}
	defer rows.Close()

	var table string
	for rows.Next() {
		if err := rows.Scan(&table); err != nil {
			slog.Error("failed to scan table name", slog.String("error", err.Error()))
			continue
		}
		// Use qualified identifier with double quotes to handle names with caps/underscores safely
		sql := fmt.Sprintf("SELECT COUNT(*) FROM \"%s\"", table)
		var count int64
		if err := pool.QueryRow(ctx, sql).Scan(&count); err != nil {
			slog.Error("failed to count rows", slog.String("table", table), slog.String("error", err.Error()))
			continue
		}
		slog.Info("table row count", slog.String("table", table), slog.Int64("count", count))
	}
	if err := rows.Err(); err != nil {
		slog.Error("tables rows iteration error", slog.String("error", err.Error()))
	}
}
