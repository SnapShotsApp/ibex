FROM postgres:alpine
ENV POSTGRES_USER gotest
ENV POSTGRES_PASSWORD gotest

ADD test_data.sql /docker-entrypoint-initdb.d/test_data.sql
