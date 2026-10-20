#!/usr/bin/env bash
# AWS counterpart of start-app.sh; the sgame.service drop-in written by terraform/aws sets AWS_REGION and DESEC_PARAM.

set -e

DESEC_TOKEN=$(aws ssm get-parameter --region "$AWS_REGION" --name "$DESEC_PARAM" --with-decryption --query Parameter.Value --output text)

DESEC_DOMAIN="sgame.dedyn.io"

curl -sf --user "$DESEC_DOMAIN:$DESEC_TOKEN" https://update.dedyn.io/

REGISTRY=$(grep '^IMAGE_REPOSITORY=' /opt/sgame/.env | cut -d= -f2 | cut -d/ -f1)
aws ecr get-login-password --region "$AWS_REGION" | docker login -u AWS --password-stdin "$REGISTRY"

git -C /opt/sgame fetch origin
git -C /opt/sgame reset --hard origin/main

docker compose -f /opt/sgame/docker-compose.prod.yaml pull
docker compose -f /opt/sgame/docker-compose.prod.yaml up -d --remove-orphans
echo "App started"
