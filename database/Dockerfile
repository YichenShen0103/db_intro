FROM mysql:8.0

ENV MYSQL_DATABASE=db_intro
ENV MYSQL_ROOT_PASSWORD=root

COPY my.cnf /etc/mysql/conf.d/charset.cnf
COPY db_schema.sql /docker-entrypoint-initdb.d/01-schema.sql
COPY init_data.sql /docker-entrypoint-initdb.d/02-data.sql
