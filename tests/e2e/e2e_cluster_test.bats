#!/usr/bin/env bats

# End-to-end cluster.sh and Kubernetes deployments tests
# Configuration for all tests
# Absolute path to cluster.sh script
CLUSTER_SH="${BATS_TEST_DIRNAME}/../../scripts/cluster.sh"
CLUSTER_NAME="${CLUSTER_NAME:-llm-cluster}"
NAMESPACE="${NAMESPACE:-llm}"
# Directory to store per-test logs under this test directory
LOG_DIR="${BATS_TEST_DIRNAME}/logs"
mkdir -p "$LOG_DIR"

@test "Help output" {
  run "$CLUSTER_SH" -h
  [ "$status" -eq 0 ]
  echo "$output" > "$LOG_DIR/help.log"
  echo "$output" | grep -q "Usage:"
}

@test "Unknown action" {
  run "$CLUSTER_SH" foo
  [ "$status" -eq 1 ]
  echo "$output" > "$LOG_DIR/unknown_action.log"
  echo "$output" | grep -q "Unknown action"
}

@test "Status before cluster" {
  run "$CLUSTER_SH" status
  [ "$status" -eq 0 ]
  echo "$output" > "$LOG_DIR/status_before_cluster.log"
}

@test "Setup cluster and deploy" {
  run "$CLUSTER_SH" setup
  [ "$status" -eq 0 ]
  echo "$output" > "$LOG_DIR/setup.log"
}

@test "Verify cluster nodes" {
  run kubectl get nodes
  [ "$status" -eq 0 ]
  echo "$output" > "$LOG_DIR/nodes.log"
  echo "$output" | grep -q "$CLUSTER_NAME"
}

@test "Verify Postgres deployment" {
  run kubectl get deployment/postgres -n "$NAMESPACE"
  [ "$status" -eq 0 ]
  echo "$output" > "$LOG_DIR/postgres_deployment.log"
}

@test "Verify Argo CD namespace and server" {
  run kubectl get namespace argocd
  [ "$status" -eq 0 ]
  echo "$output" > "$LOG_DIR/argocd_namespace.log"
  run kubectl get deployment argocd-server -n argocd
  [ "$status" -eq 0 ]
  echo "$output" > "$LOG_DIR/argocd_server.log"
}

@test "Verify LLM Helm release" {
  run helm status llm -n "$NAMESPACE"
  [ "$status" -eq 0 ]
  echo "$output" > "$LOG_DIR/helm_status.log"
  run kubectl get deployment -l app=llm-fast-wrapper -n "$NAMESPACE"
  [ "$status" -eq 0 ]
  echo "$output" > "$LOG_DIR/llm_deployment.log"
}

@test "Restart workflow" {
  run "$CLUSTER_SH" restart
  [ "$status" -eq 0 ]
  echo "$output" > "$LOG_DIR/restart.log"
}

@test "Delete cluster" {
  run "$CLUSTER_SH" delete
  [ "$status" -eq 0 ]
  echo "$output" > "$LOG_DIR/delete.log"
}

@test "Ensure cluster is removed" {
  run k3d cluster list
  [ "$status" -eq 0 ]
  echo "$output" > "$LOG_DIR/cluster_list.log"
  ! echo "$output" | grep -q "$CLUSTER_NAME"
}