"""
Superset Configuration File
"""
import os

# Database configuration
SQLALCHEMY_DATABASE_URI = os.environ.get('SUPERSET_DATABASE_URI', 'postgresql+psycopg2://postgres:postgres@postgres:5432/movies')

# Security
SECRET_KEY = os.environ.get('SUPERSET_SECRET_KEY', 'superset')

# Feature flags
FEATURE_FLAGS = {
    "ENABLE_TEMPLATE_PROCESSING": True,
    "DASHBOARD_CROSS_FILTERS": True,
    "DASHBOARD_RBAC": True,
    "ENABLE_ADVANCED_DATA_TYPES": True,
}

# Cache configuration
CACHE_CONFIG = {
    'CACHE_TYPE': 'simple',
}

# Celery configuration (for async queries) - simplified for Docker environment
class CeleryConfig(object):
    BROKER_URL = 'sqla+' + SQLALCHEMY_DATABASE_URI
    CELERY_IMPORTS = ('superset.sql_lab',)
    CELERY_RESULT_BACKEND = 'db+' + SQLALCHEMY_DATABASE_URI
    CELERYD_LOG_LEVEL = 'INFO'
    CELERYD_PREFETCH_MULTIPLIER = 1
    CELERY_ACKS_LATE = False

CELERY_CONFIG = CeleryConfig

# SQL Lab configuration
SQLLAB_CTAS_NO_LIMIT = True
SQLLAB_TIMEOUT = 300
SQLLAB_DEFAULT_DBID = None

# Enable uploading CSV files to database
CSV_TO_HIVE_UPLOAD_S3_BUCKET = None
UPLOAD_FOLDER = '/tmp/superset_uploads/'
CSV_TO_HIVE_UPLOAD_DIRECTORY = '/tmp/'

# Logging configuration - use superset home directory which is writable
ENABLE_TIME_ROTATE = True
TIME_ROTATE_LOG_LEVEL = 'INFO'
FILENAME = '/app/superset_home/superset.log'

# Dashboard and chart configuration
SUPERSET_DASHBOARD_POSITION_DATA_LIMIT = 65535
DASHBOARD_AUTO_REFRESH_MODE = "change"
DASHBOARD_AUTO_REFRESH_INTERVALS = [
    [0, "Don't refresh"],
    [10, "10 seconds"],
    [30, "30 seconds"],
    [60, "1 minute"],
    [300, "5 minutes"],
    [1800, "30 minutes"],
    [3600, "1 hour"],
]

# Row level security
ROW_LEVEL_SECURITY_FILTERS_MAX_COUNT = 1000

# Webserver configuration
WEBSERVER_TIMEOUT = 60
