# Infrastructure

Two independent Terraform root modules, one per cloud. Each has its own state; `cd` into one and run
`terraform init/plan/apply` there.

- [`gcp/`](gcp/README.md): Compute Engine + GCS + Pub/Sub + Artifact Registry + Vertex AI (Claude)
- [`aws/`](aws/README.md): EC2 + S3 + SQS + ECR + Bedrock (Claude)

Both run the same `docker-compose.prod.yaml`; the generated `.env` picks the storage provider, queue,
answer validator and image registry. CI deploys to GCP unless the repo variable `DEPLOY_TARGET` is `aws`.
