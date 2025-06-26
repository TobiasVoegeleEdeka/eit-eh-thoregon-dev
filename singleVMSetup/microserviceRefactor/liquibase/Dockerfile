FROM liquibase/liquibase:4.23.1

USER root
RUN microdnf install -y postgresql15 && \
    microdnf clean all
USER liquibase

COPY changelog/ /liquibase/changelog/