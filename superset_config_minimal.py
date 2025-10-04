"""
Minimal Superset Configuration File
"""
import os

# Database configuration
SQLALCHEMY_DATABASE_URI = os.environ.get('SUPERSET_DATABASE_URI', 'postgresql+psycopg2://postgres:postgres@postgres:5432/movies')

# Security
SECRET_KEY = os.environ.get('SUPERSET_SECRET_KEY', 'superset')

# Disable problematic features for initial setup
FEATURE_FLAGS = {
    "ENABLE_TEMPLATE_PROCESSING": False,
}

# Simple cache configuration
CACHE_CONFIG = {
    'CACHE_TYPE': 'SimpleCache',
}

# SQL Lab configuration
SQLLAB_TIMEOUT = 300

# Disable CSV upload for now
CSV_TO_HIVE_UPLOAD_S3_BUCKET = None

# Simple logging - no file rotation
ENABLE_TIME_ROTATE = False
LOG_LEVEL = 'INFO'

# Webserver configuration
WEBSERVER_TIMEOUT = 60
