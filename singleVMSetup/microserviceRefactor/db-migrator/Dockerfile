FROM liquibase/liquibase:4.23.1

WORKDIR /liquibase
COPY ./migrations /liquibase/changelog

ENTRYPOINT ["/liquibase/liquibase"]
CMD ["update"]