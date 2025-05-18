# Agent Instructions for llm-fast-wrapper

## Overview
This repository contains a Go-based streaming API wrapper for Large Language Models (LLMs). The code provides two HTTP server implementations (Fiber and Gin), a CLI client, logging backends, telemetry helpers, and Kubernetes/Docker infrastructure. Unit tests live under `tests/` and the project uses a `Taskfile` to simplify common workflows.

The agent should understand the directory layout and follow the contribution workflow described below when modifying this code base.

## Repository Structure

- `cmd/` – Cobra CLI commands for starting the API server.
- `api/` – Framework-specific HTTP server implementations (`fiber` and `gin`).
- `client/` – CLI client and supporting packages (`audit`, `chat`, `printer`, `ui`).
- `internal/` – Reusable internal packages:
  - `config` – CLI configuration helpers and OpenAI client creation.
  - `llm` – Interfaces and an in-memory OpenAI streamer stub.
  - `logging` – Zap-based logging utilities.
  - `logs` – Pluggable prompt logging backends (memory and PostgreSQL).
  - `telemetry` – OpenTelemetry setup utilities.
- `docker/` – `docker-compose.yaml` and Prometheus configuration for local infra.
- `k8s/` and `scripts/` – Helper manifests and scripts for running on Kubernetes via kind.
- `tests/` – Ginkgo/Golang unit tests.
- `Taskfile.yaml` – Defines commands for development, testing, and infrastructure.

## Development Guidelines

1. **Go Version** – Modules target Go 1.23 and build using toolchain `go1.24.3`. Always run `go mod tidy` if dependencies change.
2. **Formatting** – Run `gofmt -w` on all Go files before committing. Go code should compile without `go vet` issues.
3. **Taskfile** – Use `task` for common workflows:
   - `task dev` – run the API using Fiber.
   - `task dev -- --gin` – run the API using Gin.
   - `task docker:up` / `task docker:down` – start/stop local PostgreSQL, Prometheus, and Jaeger containers.
   - `task test` – execute `go test -cover ./...`.
   - `task test:ginkgo` – run Ginkgo tests (`ginkgo -r --cover --randomize-all --fail-on-pending`).
   - `task client:run -- --prompt "hello"` – exercise the client.
4. **Environment Variables** – The CLI expects `OPENAI_API_KEY`. Tests and dev commands may require the local services started via docker compose. Docker compose exposes PostgreSQL on `5432` and Prometheus on `9090`.
5. **Kubernetes** – Use `task kind:up` and `task kind:down` to create or destroy a local kind cluster and associated helm deployment. See `scripts/kind-setup.sh` and `scripts/cleanup.sh` for details.
6. **Logging & Telemetry** – The API and client use Zap for structured logs under `logs/llm-client.log` and OpenTelemetry for tracing. Ensure log file paths exist or are created when running locally.

## Testing

- Unit tests reside in `tests/unit`. Use `task test` for the full suite. For behavior-driven tests, run `task test:ginkgo`.
- Tests cover the OpenAI streamer stub (`internal/llm`) and the in-memory logger (`internal/logs`). When adding new features, include appropriate tests in the same style.
- Pull requests should pass `go test ./...` and `gofmt` checks before merging.

## Pull Requests

- Branch from `main` and create descriptive commit messages.
- Ensure the PR template checklist is satisfied: tests pass and documentation is updated when needed.
- Update `CHANGELOG.md` for any noteworthy changes and bump versions in `ROADMAP.md` if relevant.

## Documentation

- The primary usage instructions live in `README.md` and `client/internal/help/usage.md`.
- Keep these files in sync when changing CLI flags or API behavior.

## Summary

Follow Go best practices, keep the Taskfile workflows functional, and maintain test coverage. Provide detailed PR descriptions and ensure Docker/Kubernetes artifacts remain up to date when modifying infrastructure.
