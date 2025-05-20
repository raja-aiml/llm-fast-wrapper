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

## Repository Layout

- `cmd/` – main entrypoints for the API and subcommands
- `api/` – HTTP servers for Fiber and Gin
- `client/` – CLI client code
- `internal/` – shared libraries (embeddings, telemetry, logging, etc.)
- `docker/` – Docker Compose files and helper scripts
- `tests/` – unit and integration tests

For detailed CLI examples see `client/internal/help/usage.md`.
