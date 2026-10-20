# SGame infrastructure (GCP)

Single-VM deployment: one Compute Engine instance runs the whole Docker Compose stack
(`docker-compose.prod.yaml`); GCS, Pub/Sub, Artifact Registry, Secret Manager and Vertex AI
are the managed pieces around it.

| Area | Resources |
|---|---|
| Compute | `sgame` VM (Debian 12, `e2-standard-2`), separate `sgame-data` disk with daily snapshots, custom VPC, firewall for 80/443 and IAP-only SSH |
| Storage | `<project>-attachments-raw` (CORS, 1-day delete) and `<project>-attachments-processed`, both with object ACLs enabled |
| Upload pipeline | Pub/Sub topic `sgame-attachment-uploads`, GCS `OBJECT_FINALIZE` notification, pull subscription `sgame-attachment-uploads-sub` |
| Images | Artifact Registry `sgame-images` with cleanup policies |
| Answer validator | Claude Haiku 5.5 on Vertex AI (`VALIDATOR_TYPE=claude-vertex`) |
| Secrets | `sgame-env` (generated passwords and settings, becomes `/opt/sgame/.env`), `DESEC_TOKEN` (filled by hand) |
| Identity | `sgame-vm` runtime SA (keyless, signs URLs through IAM), `github-actions-sa` plus Workload Identity Federation limited to this repo on `refs/heads/main` |

## One-time setup

Prerequisites: an empty GCP project with billing enabled, `gcloud auth application-default login`, Terraform >= 1.9.

```bash
cd terraform/gcp

# 1. State bucket (local state, tiny)
cd bootstrap
terraform init && terraform apply -var project_id=YOUR_PROJECT
terraform output -raw backend_hcl > ../backend.hcl
cd ..

# 2. Main config
cp terraform.tfvars.example terraform.tfvars   # set project_id (and admin_members, letsencrypt_email)
terraform init -backend-config=backend.hcl
terraform plan
terraform apply
```

## After the first apply

1. **Claude on Vertex AI.** In the console, open Vertex AI → Model Garden, find the model in `claude_model_id`
   (Claude Haiku 5.5 by default) and enable it for the project. The refusal fallback, `claude-sonnet-5`, needs
   the same step.
2. **deSEC token** (kept out of Terraform state):
   ```bash
   printf '%s' "$DESEC_TOKEN" | gcloud secrets versions add DESEC_TOKEN --data-file=- --project YOUR_PROJECT
   ```
   Then on the VM re-run the startup script so it can issue the certificate:
   `gcloud compute ssh sgame --tunnel-through-iap --zone europe-central2-c -- sudo google_metadata_script_runner startup`
   (your account needs to be in `admin_members`).
3. **Point GitHub Actions at the new project** (once, from the repo):
   ```bash
   terraform output -raw github_setup_commands | sh
   ```
4. **First deploy**: push to `main` (or run the workflow manually). CI runs, images are built into the
   new registry, the VM pulls them and starts. Until then `sgame.service` fails on image pulls, which is expected.
5. **Smoke test the keyless storage setup**, which I could not verify without a real project:
   create a pack with an image and a video (signed upload policy + signed GET URLs), and watch an
   attachment go `pending -> ready`. If signing fails with a permission error, the culprit is the
   self-`serviceAccountTokenCreator` binding in `iam.tf`.

## Things to know

- **Bucket names are global.** Defaults are prefixed with the project ID; the old `sgame-attachments-*` names may still be held by the old project.
- **Data is not migrated.** MongoDB and Redis start empty. Copy volumes or use `mongodump`/`mongorestore` if you need the old data.
- **`deletion_protection` and `prevent_destroy`** guard the VM and the data disk. To tear everything down, set `deletion_protection=false`, apply, remove `prevent_destroy` from `compute.tf`, then `terraform destroy`.
- **Startup script changes** update instance metadata only; they take effect on the next boot or via `google_metadata_script_runner startup`.
- **The image registry comes from `IMAGE_REPOSITORY`** in the generated `.env`, so `region` can change without touching the compose file.
- **`vertex_location`** defaults to `global`. If Vertex reports location errors, set it to a region where the Claude model is served. `VALIDATOR_TYPE=gemini` still works if you want Gemini back.
- **Not managed here:** the DNS record (deSEC, updated at boot by `start-app.sh`), billing budgets and GitHub repository settings.
