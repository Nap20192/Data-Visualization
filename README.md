# Movie Analytics Platform

A compact Docker stack for exploring a curated movie dataset with PostgreSQL and Apache Superset.

## What’s inside

- **PostgreSQL 15** seeded with schema and data from `postgres/`
- **Apache Superset** for SQL exploration, charts, and dashboards
- **Docker Compose** orchestrating the services

## Quick start

```bash
git clone <repository-url>
cd dv
docker compose up --build -d
```

Once the containers finish booting (≈1–2 minutes), Superset is ready to use.

## Access

- Superset UI: <http://localhost:8088> (admin / admin)
- PostgreSQL DSN: `postgresql://postgres:postgres@localhost:5432/movies`

## Handy commands

```bash
# Stop services
docker compose down

# Rebuild from scratch (drops data volumes)
docker compose down -v && docker compose up --build -d

# Inspect logs
docker compose logs -f superset
```

## ERD
![erd](./image.png)
