#!/usr/bin/env bash

set -e

GCP_TOKEN=$(curl -sf "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token" \
  -H "Metadata-Flavor: Google" | python3 -c "import sys,json; print(json.load(sys.stdin)['access_token'])")

PROJECT_ID=$(curl -sf "http://metadata.google.internal/computeMetadata/v1/project/project-id" -H "Metadata-Flavor: Google")

DESEC_TOKEN=$(curl -sf "https://secretmanager.googleapis.com/v1/projects/$PROJECT_ID/secrets/DESEC_TOKEN/versions/latest:access" \
  -H "Authorization: Bearer $GCP_TOKEN" | python3 -c "import sys,json,base64; print(base64.b64decode(json.load(sys.stdin)['payload']['data']).decode())")

DESEC_DOMAIN="sgame.dedyn.io"

curl -sf --user "$DESEC_DOMAIN:$DESEC_TOKEN" https://update.dedyn.io/

echo $GCP_TOKEN | docker login -u oauth2accesstoken --password-stdin europe-central2-docker.pkg.dev

git -C /opt/sgame fetch origin
git -C /opt/sgame reset --hard origin/main

docker compose -f /opt/sgame/docker-compose.prod.yaml pull
docker compose -f /opt/sgame/docker-compose.prod.yaml up -d --remove-orphans
echo "App started"
