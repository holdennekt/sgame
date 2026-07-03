#!/usr/bin/env bash
# Deletes leftover load-test data that k6's own teardown() can't reach.
#
# teardown() in k6.js deletes host user accounts over HTTP, but archived room
# history (backend/internal/eventsprocessor/server/user_disconnected.go writes
# a mongo document for any room that gets cleaned up mid-game, i.e. state not
# WaitingForStart/GameOver) has no REST endpoint to delete it through — k6 has
# no Mongo driver and no shell access, so this can only be done from outside
# the test. Run this after a run, especially one that got interrupted (Ctrl+C,
# or teardown() itself timing out) and skipped or only partially ran cleanup.
#
# Usage: loadtest/cleanup.sh [docker-compose-file]

set -euo pipefail

COMPOSE_FILE="${1:-docker-compose.dev.yaml}"

run_mongo() {
  # Passed as an env var (not interpolated into a quoted string) so $regex
  # and embedded quotes in the JS snippet survive both shell layers untouched.
  docker compose -f "$COMPOSE_FILE" exec -T -e EVAL_SCRIPT="$1" mongodb sh -c \
    'mongosh --quiet -u "$MONGO_INITDB_ROOT_USERNAME" -p "$MONGO_INITDB_ROOT_PASSWORD" --authenticationDatabase admin "$MONGO_INITDB_DATABASE" --eval "$EVAL_SCRIPT"'
}

echo "Deleting load-test users (login starting with 'lt')..."
run_mongo 'db.users.deleteMany({login: {$regex: "^lt"}})'

echo "Deleting archived load-test rooms (name starting with 'LT Room')..."
run_mongo 'db.rooms.deleteMany({name: {$regex: "^LT Room"}})'
