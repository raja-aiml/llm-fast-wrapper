#!/usr/bin/env bash
set -euo pipefail

CLUSTER_NAME="${CLUSTER_NAME:-llm-cluster}"

ensure_k3d() {
  if ! command -v k3d >/dev/null 2>&1; then
    echo "k3d not found. Please install it." >&2
    exit 1
  fi
}

create_cluster() {
  echo "Creating k3d cluster: $CLUSTER_NAME"
  k3d cluster create "$CLUSTER_NAME"
}

deploy_postgres() {
  echo "Deploying Postgres"
  kubectl apply -f deploy/postgres.yaml
}

deploy_argocd() {
  echo "Deploying Argo CD"
  kubectl apply -f deploy/argocd/install.yaml
  kubectl apply -f deploy/argocd-app.yaml
}

main() {
  ensure_k3d
  create_cluster
  deploy_postgres
  deploy_argocd
}

main "$@"
