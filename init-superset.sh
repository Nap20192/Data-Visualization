#!/bin/bash
set -e

echo "Initializing Superset..."

# Wait for postgres to be ready using a simple approach
echo "Waiting for postgres..."
sleep 10

# Initialize the database
echo "Upgrading Superset database..."
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
echo "Initializing Superset..."
superset init

echo "Superset initialization complete!"
