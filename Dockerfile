FROM apache/superset:latest

USER root

# Copy and make executable the init script
COPY init-superset.sh /init-superset.sh
RUN chmod +x /init-superset.sh

# Switch back to superset user
USER superset

# Set default environment
ENV SUPERSET_SECRET_KEY=mysecretkey123

# Run initialization and then start superset
CMD ["/bin/bash", "-c", "/init-superset.sh && superset run -h 0.0.0.0 -p 8088"]
