#!/usr/bin/env bash

set -e

GCP_TOKEN=$(curl -sf "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token" \
  -H "Metadata-Flavor: Google" | python3 -c "import sys,json; print(json.load(sys.stdin)['access_token'])")
DESEC_TOKEN=$(curl -sf "https://secretmanager.googleapis.com/v1/projects/smart-altar-496311-p4/secrets/DESEC_TOKEN/versions/latest:access" \
  -H "Authorization: Bearer $GCP_TOKEN" | python3 -c "import sys,json,base64; print(base64.b64decode(json.load(sys.stdin)['payload']['data']).decode())")
DESEC_DOMAIN="sgame.dedyn.io"

git -C /opt/sgame pull

curl --user "$DESEC_DOMAIN:$DESEC_TOKEN" https://update.dedyn.io/

docker compose -f /opt/sgame/docker-compose.prod.yaml pull
docker compose -f /opt/sgame/docker-compose.prod.yaml up -d --remove-orphans
