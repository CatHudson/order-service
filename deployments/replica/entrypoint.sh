#!/bin/bash
set -e

# Wait for primary to be ready
echo "Replica: waiting for primary..."
until pg_isready -h postgres-primary -U replicator -d orders; do
    sleep 1
done

# If data directory is empty, clone from primary
if [ -z "$(ls -A /var/lib/postgresql/data)" ]; then
    echo "Replica: cloning from primary..."
    chown postgres:postgres /var/lib/postgresql/data
    chmod 700 /var/lib/postgresql/data
    gosu postgres pg_basebackup \
        -h postgres-primary \
        -U replicator \
        -D /var/lib/postgresql/data \
        -Fp -Xs -P -R
    echo "Replica: base backup complete."
fi

# Start Postgres as the postgres user
exec gosu postgres postgres