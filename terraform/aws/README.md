# SGame infrastructure (AWS)

The AWS twin of [`../gcp`](../gcp): one EC2 instance runs the whole Docker Compose stack
(`docker-compose.prod.yaml`); S3, SQS, ECR, SSM Parameter Store and Bedrock are the managed pieces around it.

| Area | Resources | GCP counterpart |
|---|---|---|
| Compute | `sgame` instance (Ubuntu 24.04, `t3.large`), separate `sgame-data` EBS volume with daily DLM snapshots, VPC with a public subnet, security group for 80/443 only | VM, data disk + snapshot policy, VPC/firewall |
| Shell access | SSM Session Manager (no SSH port) | IAP SSH |
| Storage | `sgame-<account>-attachments-raw` (CORS, 1-day expiry) and `sgame-<account>-attachments-processed` (`public/*` readable through the bucket policy, versioned for 7-day soft delete) | GCS buckets |
| Upload pipeline | S3 `ObjectCreated` notification → SQS `sgame-attachment-uploads` (+ dead-letter queue) | GCS notification → Pub/Sub |
| Images | ECR `sgame-images/sgame-backend`, `sgame-images/sgame-frontend` with lifecycle policies | Artifact Registry |
| Secrets | SSM SecureString `/sgame/env` (generated, becomes `/opt/sgame/.env`), `/sgame/DESEC_TOKEN` (added by hand) | Secret Manager |
| Answer validator | Claude Haiku 5.5 on Bedrock (`VALIDATOR_TYPE=claude-bedrock`) | Claude Haiku 5.5 on Vertex AI |
| Identity | `sgame-vm` instance role, `sgame-github-actions` role through GitHub OIDC limited to this repo on `refs/heads/main` | service accounts + Workload Identity Federation |

## One-time setup

Prerequisites: an AWS account, the AWS CLI logged in with admin rights (`aws configure sso` or access keys), Terraform >= 1.10.

```bash
cd terraform/aws

# 1. State bucket (local state, tiny)
cd bootstrap
terraform init && terraform apply
terraform output -raw backend_hcl > ../backend.hcl
cd ..

# 2. Main config
cp terraform.tfvars.example terraform.tfvars   # optional overrides
terraform init -backend-config=backend.hcl
terraform plan
terraform apply
```

## After the first apply

1. **Bedrock model access.** In the Bedrock console (same region), open the model catalog and request access to the
   Claude model in `claude_model_id` (Anthropic asks for a short use-case form once per account). The fallback
   used when that model refuses, `anthropic.claude-sonnet-5`, is open to all Bedrock customers.
2. **deSEC token** (kept out of Terraform state):
   ```bash
   read -rs DESEC_TOKEN
   aws ssm put-parameter --name /sgame/DESEC_TOKEN --type SecureString --value "$DESEC_TOKEN"
   ```
   Then re-run the boot script on the instance so it can issue the certificate:
   ```bash
   aws ssm start-session --target "$(terraform output -raw instance_id)"
   sudo /var/lib/cloud/instance/scripts/sgame-startup.sh
   ```
   (`start-session` needs the [Session Manager plugin](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html).)
3. **Point GitHub Actions at AWS** (once, from the repo):
   ```bash
   terraform output -raw github_setup_commands | sh
   ```
   This sets `DEPLOY_TARGET=aws`, which switches CI from `build-and-deploy.yaml` (GCP) to `build-and-deploy-aws.yaml`.
   Delete the variable to go back to GCP.
4. **First deploy**: push to `main` (or run the workflow manually). Until then `sgame.service` fails on image pulls, which is expected.
5. **Smoke test**: create a pack with an image and a video (browser POST upload to S3 + presigned GETs), watch an
   attachment go `pending -> ready` (S3 → SQS → attachment-processor), and answer a question that needs the AI validator.

## Things to know

- **Presigned URLs live at most as long as the instance role credentials that signed them.** The S3 adapter keeps at
  least one hour of credential lifetime, so a presigned GET is valid for `min(GET_URL_TTL, ~1h+)`. Room and pack
  responses re-sign URLs on every fetch, so this only matters for a tab left open for hours.
- **SQS visibility timeout is 5 minutes.** An attachment that takes longer to process is delivered again; after
  10 receives it lands in `sgame-attachment-uploads-dlq`.
- **Data is not migrated.** MongoDB and Redis start empty, and objects stay in GCS. Copy volumes (`mongodump`/`mongorestore`) and
  buckets (`gcloud storage rsync` / `aws s3 sync`) if you need the old data.
- **`termination_protection` and `prevent_destroy`** guard the instance and the data volume. To tear everything down,
  set `termination_protection=false`, apply, remove `prevent_destroy` from `compute.tf`, then `terraform destroy`.
- **One GitHub OIDC provider per account.** If the account already has one, set `create_github_oidc_provider = false`.
