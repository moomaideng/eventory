#!/bin/bash
set -euo pipefail

# Runs on first Postgres data-dir init only (docker-entrypoint-initdb.d).
: "${POSTGRES_MULTIPLE_DATABASES:?POSTGRES_MULTIPLE_DATABASES must be set}"

IFS=',' read -ra DATABASES <<< "$POSTGRES_MULTIPLE_DATABASES"
for db in "${DATABASES[@]}"; do
	db="${db#"${db%%[![:space:]]*}"}"
	db="${db%"${db##*[![:space:]]}"}"
	if [ -z "$db" ]; then
		continue
	fi
	psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
		SELECT 'CREATE DATABASE ${db}'
		WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '${db}')\gexec
	EOSQL
done
