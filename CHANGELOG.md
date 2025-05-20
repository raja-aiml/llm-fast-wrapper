# Changelog

## Unreleased
- Introduced context-aware streaming in OpenAI mock streamer and updated HTTP servers accordingly
- Switched local Kubernetes setup to k3d with bundled Postgres and Argo CD manifests
- Updated k3d Postgres deployment to use pgvector image
- Exposed Postgres service on host at `localhost:5432`

## v0.1.0 - 2025-05-17
- Initial project scaffold with Fiber and Gin servers
- CLI client
- Docker compose infra
- Helm chart and kind scripts
- GitHub workflows

