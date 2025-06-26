FROM liquibase/liquibase

# Install PostgreSQL client tools
RUN apt-get update && \
    apt-get install -y postgresql-client && \
    rm -rf /var/lib/apt/lists/*

# Copy changelog files
COPY changelog/ /liquibase/changelog/