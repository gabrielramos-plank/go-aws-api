.PHONY: run test lint typecheck check up down tf-fmt tf-bootstrap tf-plan tf-apply tf-destroy

DATABASE_URL ?= postgres://app:app@localhost:5432/app?sslmode=disable

# --- App --------------------------------------------------------------------
run:            ## Run the API locally against the compose database
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/api

test:
	go test -race ./...

lint:
	golangci-lint run

typecheck:
	go build ./... && go vet ./...

check: lint typecheck test   ## Everything CI checks, locally

up:             ## Start API + Postgres in Docker
	docker compose up --build

down:
	docker compose down

# --- Infrastructure ---------------------------------------------------------
tf-fmt:
	terraform fmt -recursive infra

tf-bootstrap:   ## One-time: state bucket, ECR, GitHub OIDC roles
	cd infra/bootstrap && terraform init && terraform apply

tf-plan:
	cd infra/app && terraform init -backend-config=backend.hcl && terraform plan

tf-apply:
	cd infra/app && terraform init -backend-config=backend.hcl && terraform apply

tf-destroy:     ## Tear down the paid resources at the end of a study session
	cd infra/app && terraform init -backend-config=backend.hcl && terraform destroy
