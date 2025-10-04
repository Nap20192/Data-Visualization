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

Once the containers finish booting (≈1–2 minutes), Superset is ready to use with the movie warehouse already wired in.

## Access

- Superset UI: <http://localhost:8088> (admin / admin)
- PostgreSQL DSN: `postgresql://postgres:postgres@localhost:5432/movies`
- Superset database connection: **Movies Warehouse** → points at the same PostgreSQL instance so you can explore it from SQL Lab and build dashboards immediately.

> Need a different Postgres host, port, or credentials? Override `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, or `POSTGRES_DB` in `docker-compose.yml` and the Superset container will register the matching connection on start-up.

## Superset Configuration

The Superset instance is automatically configured with:

- **Database Connection**: Automatically connects to the PostgreSQL movies database
- **Admin User**: username: `admin`, password: `admin`
- **Features**: Enhanced with template processing, cross filters, and advanced data types
- **SQL Lab**: Enabled for direct SQL queries with 5-minute timeout
- **CSV Upload**: Enabled for importing additional data

### Custom Configuration

The setup includes:

- `superset_config.py` - Main configuration file with database settings and feature flags
- `init-superset.sh` - Initialization script that waits for Postgres and sets up the connection
- PostgreSQL driver (`psycopg2-binary`) pre-installed for database connectivity

### Troubleshooting

```bash
# Check Superset logs
docker compose logs -f superset

# Check PostgreSQL connection
docker compose exec superset superset shell -c "from superset import db; print(db.engine.execute('SELECT 1').scalar())"

# Restart just Superset
docker compose restart superset
```

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
