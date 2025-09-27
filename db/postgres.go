package db

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host     string
	Port     int32
	Dbname   string
	Username string
	Password string
}

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(ctx context.Context, cfg Config) (*Postgres, error) {
	port := strconv.Itoa(int(cfg.Port))
	dsn := "postgres://" + cfg.Username + ":" + cfg.Password + "@" + cfg.Host + ":" + port + "/" + cfg.Dbname + ""
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	return &Postgres{
		pool: pool,
	}, nil
}

func (p *Postgres) Close() {
	p.pool.Close()
}
func (p *Postgres) Pool() *pgxpool.Pool {
	return p.pool
}
