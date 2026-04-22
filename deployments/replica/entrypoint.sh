#!/bin/bash
set -e

PGDATA=/var/lib/postgresql/data/pgdata

echo "Replica: waiting for primary..."
until pg_isready -h postgres-primary -U replicator -d orders; do
    sleep 1
done

if [ -z "$(ls -A $PGDATA 2>/dev/null)" ]; then
    echo "Replica: cloning from primary..."
    mkdir -p "$PGDATA"
    chown postgres:postgres "$PGDATA"
    chmod 700 "$PGDATA"
    gosu postgres pg_basebackup \
        -h postgres-primary \
        -U replicator \
        -D "$PGDATA" \
        -Fp -Xs -P -R
    echo "Replica: base backup complete."
fi

exec gosu postgres postgres
