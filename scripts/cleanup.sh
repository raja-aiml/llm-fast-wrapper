#!/usr/bin/env bash
set -euo pipefail

CLUSTER_NAME="${CLUSTER_NAME:-llm-cluster}"
helm uninstall llm -n llm || true
kubectl delete ns llm || true
k3d cluster delete "$CLUSTER_NAME" || true
