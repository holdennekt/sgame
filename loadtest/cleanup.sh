#!/usr/bin/env bash
# Usage: loadtest/cleanup.sh [docker-compose-file]. Deletes archived room history left by load tests (it has no REST endpoint).

set -euo pipefail

COMPOSE_FILE="${1:-docker-compose.dev.yaml}"

run_mongo() {
  # Passed via env, not interpolated, so $regex and the JS quotes survive both shell layers.
  docker compose -f "$COMPOSE_FILE" exec -T -e EVAL_SCRIPT="$1" mongodb sh -c \
    'mongosh --quiet -u "$MONGO_INITDB_ROOT_USERNAME" -p "$MONGO_INITDB_ROOT_PASSWORD" --authenticationDatabase admin "$MONGO_INITDB_DATABASE" --eval "$EVAL_SCRIPT"'
}

echo "Deleting load-test users (login starting with 'lt')..."
run_mongo 'db.users.deleteMany({login: {$regex: "^lt"}})'

echo "Deleting archived load-test rooms (name starting with 'LT Room')..."
run_mongo 'db.rooms.deleteMany({name: {$regex: "^LT Room"}})'
