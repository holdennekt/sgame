#!/usr/bin/env bash

set -e

DESEC_TOKEN=$(gcloud secrets versions access latest --secret="DESEC_TOKEN")
DESEC_DOMAIN="sgame.dedyn.io"

git -C /opt/sgame pull

curl --user "$DESEC_DOMAIN:$DESEC_TOKEN" https://update.dedyn.io/

docker compose -f /opt/sgame/docker-compose.prod.yaml pull
docker compose -f /opt/sgame/docker-compose.prod.yaml up -d --remove-orphans
