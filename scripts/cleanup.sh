#!/usr/bin/env bash
set -euo pipefail

CLUSTER_NAME="llm-cluster"
helm uninstall llm -n llm || true
kubectl delete ns llm || true
kind delete cluster --name "$CLUSTER_NAME" || true
