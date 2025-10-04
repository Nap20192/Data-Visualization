FROM apache/superset:latest

USER root

# Copy and make executable the init script
COPY init-superset.sh /init-superset.sh
RUN chmod +x /init-superset.sh

# Copy Superset configuration
COPY superset_config.py /app/pythonpath/superset_config.py

# Create directories with proper permissions
RUN mkdir -p /app/superset_home /tmp/superset_uploads && \
    chown -R superset:superset /app/superset_home /tmp/superset_uploads

# Install Postgres driver in the superset virtual environment
RUN /app/.venv/bin/pip install --no-cache-dir psycopg2-binary

# Switch back to superset user
USER superset

# Set default environment
ENV SUPERSET_SECRET_KEY=mysecretkey123

# Run initialization and then start superset
CMD ["/bin/bash", "-c", "/init-superset.sh && superset run -h 0.0.0.0 -p 8088"]
