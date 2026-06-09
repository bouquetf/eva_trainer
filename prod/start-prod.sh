#!/bin/sh
set -e

# If a dump file exists, clear the database and restore it
if [ -d /app/dump ]; then
  for dump_file in /app/dump/*.dump /app/dump/*.sql; do
    [ -f "$dump_file" ] || continue

    echo "Clearing existing database data..."
    psql "$DATABASE_URL" <<'EOSQL'
DROP EXTENSION IF EXISTS postgis_tiger_geocoder CASCADE;
DROP EXTENSION IF EXISTS postgis_topology CASCADE;
DROP EXTENSION IF EXISTS postgis CASCADE;
DROP EXTENSION IF EXISTS vector CASCADE;
DROP SCHEMA IF EXISTS tiger CASCADE;
DROP SCHEMA IF EXISTS topology CASCADE;
DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public;
EOSQL

    echo "Restoring dump: $dump_file"
    case "$dump_file" in
      *.sql)
        psql "$DATABASE_URL" < "$dump_file"
        ;;
      *.dump)
        pg_restore --section=pre-data --no-owner --no-acl -d "$DATABASE_URL" "$dump_file"
        psql "$DATABASE_URL" -c "ALTER TABLE topology.topology ADD COLUMN IF NOT EXISTS useslargeids boolean;"
        pg_restore --section=data --no-owner --no-acl -d "$DATABASE_URL" "$dump_file"
        pg_restore --section=post-data --no-owner --no-acl -d "$DATABASE_URL" "$dump_file"
        ;;
    esac
    echo "Dump restore complete."
    break
  done
fi

echo "Running migrations..."
/app/migrate

echo "Starting server..."
exec /app/server
