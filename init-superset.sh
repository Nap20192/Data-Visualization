#!/bin/bash
set -e

echo "Initializing Superset..."

POSTGRES_HOST=${POSTGRES_HOST:-postgres}
POSTGRES_PORT=${POSTGRES_PORT:-5432}
POSTGRES_USER=${POSTGRES_USER:-postgres}
POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-postgres}
POSTGRES_DB=${POSTGRES_DB:-movies}
export CONNECTION_URI="postgresql+psycopg2://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}"

# Wait for postgres to be ready using pg_isready if available, otherwise retrying TCP connection
echo "Waiting for postgres at ${POSTGRES_HOST}:${POSTGRES_PORT}..."
for i in {1..30}; do
    if command -v pg_isready &>/dev/null; then
        if pg_isready -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT}" -U "${POSTGRES_USER}" >/dev/null 2>&1; then
            echo "Postgres is ready!"
            break
        fi
    else
        if (echo >"/dev/tcp/${POSTGRES_HOST}/${POSTGRES_PORT}") >/dev/null 2>&1; then
            echo "Postgres is ready!"
            break
        fi
    fi
    echo "Postgres not ready yet (${i}/30)..."
    sleep 2
done

if [ $i -eq 30 ]; then
    echo "ERROR: Postgres failed to become ready after 60 seconds"
    exit 1
fi

# Initialize the database
echo "Upgrading Superset metadata database..."
superset db upgrade

# Create admin user (only if doesn't exist)
echo "Creating admin user..."
superset fab create-admin \
        --username admin \
        --firstname Admin \
        --lastname User \
        --email admin@example.com \
        --password admin || echo "Admin user already exists"

# Initialize Superset
echo "Running superset init..."
superset init

# Register the Movies analytics database in Superset
echo "Configuring Superset connection to Postgres..."
superset shell <<PY
import json
import os
from superset import db
from superset.models.core import Database

database_name = "Movies Warehouse"
sqlalchemy_uri = os.environ["CONNECTION_URI"]

database = db.session.query(Database).filter_by(database_name=database_name).one_or_none()

if not database:
        extra = {
                "metadata_params": {},
                "engine_params": {},
                "metadata_cache_timeout": {},
                "schemas_allowed_for_csv_upload": []
        }
        database = Database(
                database_name=database_name,
                sqlalchemy_uri=sqlalchemy_uri,
                extra=json.dumps(extra),
                expose_in_sqllab=True,
                allow_csv_upload=True,
        )
        db.session.add(database)
        db.session.commit()
        print("Created Superset database connection for Movies Warehouse")
else:
        updated = False
        if database.sqlalchemy_uri != sqlalchemy_uri:
                database.set_sqlalchemy_uri(sqlalchemy_uri)
                updated = True
        if not database.expose_in_sqllab:
                database.expose_in_sqllab = True
                updated = True
        if not database.allow_csv_upload:
                database.allow_csv_upload = True
                updated = True
        if updated:
                db.session.commit()
                print("Updated existing Superset database connection for Movies Warehouse")
        else:
                print("Superset database connection already configured")
PY

echo "Superset initialization complete!"
