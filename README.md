# llm-fast-wrapper

`llm-fast-wrapper` provides a thin streaming wrapper around OpenAI compatible LLM APIs. It ships two HTTP servers (Fiber and Gin) and a CLI client. The project focuses on observability and disaster recovery while keeping the code simple.

## Features

- SSE based streaming completions with either Fiber or Gin
- CLI client with audit logging of prompts and responses
- Embedding service with optional pgvector storage and in-memory cache
- Prometheus metrics, Jaeger tracing and Zap structured logging
- Docker Compose environment with PostgreSQL and Splunk for local testing
- Kubernetes manifests and Helm chart (under `deploy/`) for cluster deployments

## Getting Started

### Prerequisites

- Go 1.23+ (toolchain 1.24.3)
- Docker and Docker Compose
- [`task`](https://taskfile.dev) command runner
- OpenAI API key exported as `OPENAI_API_KEY`

### Quick Start

```bash
# start local services (Postgres, Prometheus, Splunk)
task docker:up

# run the API using Fiber
task dev

# alternatively run the Gin server
task dev -- --gin

# interact via the CLI
task client:run -- --prompt "hello"
```

All commands are defined in `Taskfile.yaml`. Logs are written under `logs/`.

To start a local Kubernetes cluster with Postgres and Argo CD:

```bash
task k3d:up
```

Tear it down with:

```bash
task k3d:down
```

## Kubernetes Cluster with k3d, Argo CD & Helm

Follow these steps to deploy, migrate the database, seed embeddings, and test the service end-to-end:

1. Clean up any existing cluster:
   ```bash
   scripts/cluster.sh delete || true
   k3d cluster delete llm-cluster || true
   ```
2. Create the cluster and deploy services:
   ```bash
   scripts/cluster.sh setup
   ```
3. Verify cluster and pods:
   ```bash
   kubectl get nodes
   kubectl get pods -n llm
   kubectl get pods -n argocd
   helm list -n llm
   ```
4. Port-forward Argo CD and log in:
   ```bash
   kubectl -n argocd port-forward svc/argocd-server 8080:443
   PASSWORD=$(kubectl -n argocd get secret argocd-initial-admin-secret \
     -o jsonpath="{.data.password}" | base64 --decode)
   argocd login localhost:8080 --username admin --password "$PASSWORD" --insecure
   ```
5. Sync the Argo CD application (runs DB migrations automatically):
   ```bash
   argocd app sync llm-fast-wrapper
   ```
6. Check the migration Job logs:
   ```bash
   kubectl get jobs -n llm
   kubectl logs job/llm-fast-wrapper-db-migrate -n llm
   ```
7. Validate the Postgres schema:
   ```bash
   psql "postgresql://llm:llm@localhost:5432/llmlogs?sslmode=disable" -c "\dt"
   ```
8. (Optional) Verify embeddings seeding:
   ```bash
   kubectl logs job/llm-fast-wrapper-intent-seed -n llm
   psql "postgresql://llm:llm@localhost:5432/llmlogs?sslmode=disable" -c "SELECT count(*) FROM embeddings;"
   ```
9. Smoke-test the LLM service:
   ```bash
   kubectl port-forward svc/llm-fast-wrapper 8080:8080 -n llm
   curl localhost:8080/health
   ```
10. Run the intent-matching CLI locally:
   ```bash
   go run cmd/intent/main.go \
     --db-dsn "postgresql://llm:llm@localhost:5432/llmlogs?sslmode=disable" \
     --db-dim 1536 \
     --use-db \
     --dir ./prompting-strategies \
     "Explain TCP"
   ```
11. Tear down the application and cluster:
   ```bash
   argocd app delete llm-fast-wrapper --cascade
   scripts/cluster.sh delete
   k3d cluster delete llm-cluster
   ```

## Repository Layout

- `cmd/` – main entrypoints for the API and subcommands
- `api/` – HTTP servers for Fiber and Gin
- `client/` – CLI client code
- `internal/` – shared libraries (embeddings, telemetry, logging, etc.)
- `docker/` – Docker Compose files and helper scripts
- `tests/` – unit and integration tests

For detailed CLI examples see `client/internal/help/usage.md`.
