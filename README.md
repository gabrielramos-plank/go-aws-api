# go-aws-api

A small Go REST API deployed to AWS with a production-style setup:
**CloudFront (HTTPS) in front of an ALB → ECS Fargate → RDS Postgres, with Terraform and GitHub Actions using OIDC.**
The same CloudFront URL also serves a static FE from S3, so the browser calls the API on the same origin, with no CORS and no custom domain needed.

```
Browser ──HTTPS──> CloudFront  https://<id>.cloudfront.net
                      ├── /*      → S3 bucket (FE build, private, Origin Access Control)
                      └── /api/*  → ALB :80 (SG: CloudFront IPs only + secret header)
                                      │
                                      ▼
                                   ECS Fargate task (arm64, image from ECR)
                                      │ :5432 (TLS)
                                      ▼
                                   RDS Postgres 17 (private subnets)

SSM Parameter Store → DB password injected as a secret
CloudWatch Logs     → /ecs/go-aws-api
GitHub Actions      → CI on PRs, deploy on push to main, Terraform on demand
```

## API

| Method | Path            | Description                      |
|--------|-----------------|----------------------------------|
| GET    | `/health`       | Liveness (ALB health check only) |
| GET    | `/api/v1/items` | List the latest 100 items        |
| POST   | `/api/v1/items` | Create an item: `{"name":"..."}` |

Only `/api/*` is routed through CloudFront to the API.

## Repository layout

```
cmd/api/            entrypoint (config, DB, graceful shutdown)
internal/           config, database, HTTP server, items feature
infra/bootstrap/    one-time, ~free: tfstate bucket, FE bucket, ECR, GitHub OIDC roles (local state)
infra/app/          the paid stack: VPC, ALB, CloudFront, ECS, RDS (remote state in S3)
examples/           frontend-deploy.yml: copy into your FE repo
.github/workflows/
  ci.yml            PRs: go vet/test, golangci-lint, docker build, terraform fmt/validate
  deploy.yml        push to main: CI → build & push image → new task def → roll ECS service
  infra.yml         manual: terraform plan / apply / destroy of infra/app
```

## Run locally

Requires Go 1.25+ and Docker.

```bash
go mod tidy          # first time only: generates go.sum (commit it)
make up              # API + Postgres on http://localhost:8080
curl localhost:8080/health
curl -X POST localhost:8080/api/v1/items -d '{"name":"hello"}'
curl localhost:8080/api/v1/items
make test
```

In your FE dev server, proxy `/api` to `http://localhost:8080` (e.g. Vite `server.proxy`) so
the FE code uses relative URLs (`fetch("/api/v1/items")`) both locally and in AWS.

## Deploy to AWS (first time)

Prerequisites: AWS CLI logged in as an admin user (`aws sts get-caller-identity`),
Terraform ≥ 1.10, and this code pushed to a GitHub repo.

### 1. Bootstrap (once, from your machine)

```bash
cd infra/bootstrap
cp terraform.tfvars.example terraform.tfvars   # set github_repo (and frontend_github_repo)
terraform init && terraform apply
```

The bootstrap state is stored locally in `infra/bootstrap/terraform.tfstate` (gitignored), so keep that file.

### 2. Configure GitHub (this repo)

**Settings → Environments:** create `production` and `infrastructure`.
These names are part of the IAM trust policy, so only jobs in these environments can assume the roles.

**Settings → Secrets and variables → Actions → Variables:**

| Variable                 | Value             |
|--------------------------|-------------------|
| `AWS_REGION`             | e.g. `us-east-1`  |
| `AWS_DEPLOY_ROLE_ARN`    | bootstrap output  |
| `AWS_TERRAFORM_ROLE_ARN` | bootstrap output  |
| `TF_STATE_BUCKET`        | bootstrap output  |

No AWS access keys are stored in GitHub. Workflows get short-lived credentials through OIDC.

### 3. Create the infrastructure

**Actions → Infrastructure → Run workflow → `apply`** takes about 10–15 minutes (RDS and CloudFront are slow).
The `app_url` output is your HTTPS URL.

### 4. Deploy the API

**Actions → Deploy → Run workflow**. After that, it runs on every push to `main`.

```bash
curl https://<id>.cloudfront.net/api/v1/items
```

### 5. Deploy the FE (separate repo)

Copy `examples/frontend-deploy.yml` into the FE repo and follow its header comments.
Until a FE is uploaded, `https://<id>.cloudfront.net/` returns 403 (empty bucket), which is expected.

## Daily study loop (protect your credits)

| When            | Do                                         |
|-----------------|--------------------------------------------|
| Start a session | Infrastructure → `apply`                   |
| Change code     | push to `main` → auto deploy               |
| End a session   | Infrastructure → `destroy`                 |

The state bucket, FE bucket, ECR and IAM roles live in the bootstrap stack and survive `destroy`.
On the next `apply`, the API starts from the last `:latest` image and the FE from the files already in S3.
**The `*.cloudfront.net` URL changes on every recreate**, so check the `app_url` output.
To destroy automatically every night, uncomment `schedule` in `infra.yml`.

Locally, the same thing works with `make tf-apply` / `make tf-destroy`
(after `cp infra/app/backend.hcl.example infra/app/backend.hcl` and
`cp infra/app/terraform.tfvars.example infra/app/terraform.tfvars`).

**Approximate cost while running:** ALB ~$0.025/h, Fargate 0.25 vCPU arm64 ~$0.01/h,
RDS t4g.micro ~$0.016/h, public IPv4 ~$0.005/h each. That's about **$0.07/hour**, or ~$50/month if left on 24/7.
CloudFront and S3 are effectively free at study traffic. Set an AWS Budget alert.

## How config flows

- **Terraform owns configuration**: env vars, secrets, CPU/memory and IAM roles in the task definition.
- **CI owns the image**: `deploy.yml` downloads the latest task definition revision,
  swaps only the image tag (`:<git sha>`) and registers a new revision.
  The ECS service ignores `task_definition` in Terraform, so `apply` never rolls back a deploy.
- A Terraform config change (e.g. a new env var) goes live on the next deploy.
- The deployment circuit breaker rolls back automatically if new tasks fail health checks.

## Security notes

- The ALB only accepts CloudFront's IP ranges, **and** requests must carry an
  `X-Origin-Verify` header that only this distribution adds. Anything else gets a 403.
- CloudFront → ALB is plain HTTP over AWS's network, because the ALB has no certificate without a domain.
  Add a domain + ACM certificate on the ALB later if you want end-to-end TLS.
- The API itself is public to anyone with the URL. Add auth (e.g. an API key header or Cognito) before storing anything real.

## Next steps to learn

- Custom domain: Route 53 + ACM certificate (in `us-east-1`) on CloudFront
- Auto scaling on CPU / request count (`aws_appautoscaling_*`)
- Versioned DB migrations (goose / golang-migrate) as a one-off ECS task before deploy
- Private subnets + NAT Gateway or VPC endpoints (the "full" production network)
- CloudWatch alarms (5xx rate, unhealthy hosts) → SNS email
