#!/bin/bash
set -e

echo "Testing Superset and PostgreSQL setup..."

# Test if psycopg2 is available
echo "Testing psycopg2 import..."
/app/.venv/bin/python -c "import psycopg2; print('psycopg2 imported successfully')"

# Test basic superset import
echo "Testing Superset import..."
/app/.venv/bin/python -c "import superset; print('Superset imported successfully')"

# Test database connection
echo "Testing PostgreSQL connection..."
/app/.venv/bin/python -c "
import psycopg2
import os
try:
    conn = psycopg2.connect(
        host=os.environ.get('POSTGRES_HOST', 'postgres'),
        port=os.environ.get('POSTGRES_PORT', '5432'),
        user=os.environ.get('POSTGRES_USER', 'postgres'),
        password=os.environ.get('POSTGRES_PASSWORD', 'postgres'),
        database=os.environ.get('POSTGRES_DB', 'movies')
    )
    print('PostgreSQL connection successful')
    conn.close()
except Exception as e:
    print(f'PostgreSQL connection failed: {e}')
"

echo "All tests completed!"
