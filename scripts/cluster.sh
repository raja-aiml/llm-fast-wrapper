#!/usr/bin/env bash
################################################################################
# cluster.sh – One‑touch lifecycle management for a local Kubernetes stack     #
################################################################################
# ❑  Creates / tears down a k3d cluster (Docker‑in‑Docker)
# ❑  Installs PostgreSQL, Argo CD, and an LLM Helm chart
# ❑  Auto‑installs missing CLI deps (k3d, kubectl, helm, docker) where possible
# ❑  Colourised, timestamped logging with selectable verbosity
# ❑  Resilient cleanup logic & graceful error handling
################################################################################
# Usage: ./cluster.sh [setup|delete|status|restart] [--quiet] [--debug]
# Env vars:
#   CLUSTER_NAME  k3d cluster name               (default: llm-cluster)
#   NAMESPACE     K8s namespace for the stack    (default: llm)
#   LOG_LEVEL     debug|info|warn|error          (default: info)
#   PORT          Local app port                 (default: 8080)
################################################################################

set -Eeuo pipefail
shopt -s inherit_errexit 2>/dev/null || true

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
readonly DEPLOY_DIR="$REPO_ROOT/deploy"
readonly CHARTS_DIR="$DEPLOY_DIR/charts"
readonly CLUSTER_NAME="${CLUSTER_NAME:-llm-cluster}"
readonly NAMESPACE="${NAMESPACE:-llm}"
readonly LOG_LEVEL="${LOG_LEVEL:-info}"
readonly PORT="${PORT:-8080}"
readonly ACTION="${1:-setup}"
readonly TIMEOUT_CLUSTER="300s"
readonly TIMEOUT_DEPLOY="300s"
readonly REQUIRED_CMDS=(docker k3d kubectl helm)
readonly REQUIRED_DIRS=("$DEPLOY_DIR" "$CHARTS_DIR")
readonly REQUIRED_FILES=(
    "$DEPLOY_DIR/postgres.yaml"
    "$DEPLOY_DIR/argocd/install.yaml"
    "$DEPLOY_DIR/argocd-app.yaml"
)

function set_colors() {
  if [[ -t 2 ]]; then
    export RED="$(tput setaf 1)" YEL="$(tput setaf 3)" GRN="$(tput setaf 2)" BLU="$(tput setaf 4)" RST="$(tput sgr0)"
  else
    export RED="" YEL="" GRN="" BLU="" RST=""
  fi
}
set_colors

function log() {
  local level="$1"
  local color="$2"
  local message="$3"
  local timestamp="$(date '+%Y-%m-%dT%H:%M:%S%z')"

  local current_level_num message_level_num

  case "$LOG_LEVEL" in
    debug) current_level_num=0 ;;
    info)  current_level_num=1 ;;
    warn)  current_level_num=2 ;;
    error) current_level_num=3 ;;
    *)     current_level_num=1 ;;
  esac

  case "$level" in
    debug) message_level_num=0 ;;
    info)  message_level_num=1 ;;
    warn)  message_level_num=2 ;;
    error) message_level_num=3 ;;
    *)     message_level_num=1 ;;
  esac

  if (( message_level_num >= current_level_num )); then
    printf '%s %b[%5s]%b %s\n' "$timestamp" "$color" "$level" "$RST" "$message" >&2
  fi
}

log_debug() { log "debug" "$BLU" "$*"; }
log_info()  { log "info" "$GRN" "$*"; }
log_warn()  { log "warn" "$YEL" "$*"; }
log_error() { log "error" "$RED" "$*"; }

CLEANUPS=()
register_cleanup() { CLEANUPS+=("$1"); }

execute_cleanup() {
  (( ${#CLEANUPS[@]} > 0 )) && log_info "Running cleanup actions..."
  for ((i=${#CLEANUPS[@]}-1; i>=0; i--)); do
    log_debug "Executing cleanup: ${CLEANUPS[$i]}"
    ${CLEANUPS[$i]} || log_error "Cleanup failed: ${CLEANUPS[$i]}"
  done
}

trap 'ec=$?; (( ec )) && log_error "Script failed (exit $ec)" || log_info "Script completed successfully."; execute_cleanup' EXIT INT TERM

cmd_exists() { command -v "$1" &>/dev/null; }
install_cmd() {
  log_info "Installing $1..."
  cmd_exists brew && brew install "$1" || cmd_exists apt-get && sudo apt-get install -y "$1" || cmd_exists yum && sudo yum install -y "$1" || return 1
}
ensure_deps() {
  log_info "Checking CLI dependencies..."
  for cmd in "${REQUIRED_CMDS[@]}"; do cmd_exists "$cmd" || install_cmd "$cmd" || return 1; done
  log_info "All CLI dependencies are satisfied."
}
ensure_artifacts() {
  log_info "Checking required artifacts..."
  for dir in "${REQUIRED_DIRS[@]}"; do [[ -d "$dir" ]] || { log_error "Missing directory: $dir"; return 1; }; done
  for file in "${REQUIRED_FILES[@]}"; do [[ -f "$file" ]] || { log_error "Missing file: $file"; return 1; }; done
  log_info "All required artifacts are present."
}

create_cluster() {
  k3d cluster list | grep -q "^$CLUSTER_NAME" && { log_warn "Cluster '$CLUSTER_NAME' already exists."; return; }
  log_info "Creating k3d cluster '$CLUSTER_NAME'..."
  k3d cluster create "$CLUSTER_NAME" --timeout "$TIMEOUT_CLUSTER"
  register_cleanup "delete_cluster"
  kubectl wait --for=condition=Ready nodes --all --timeout=60s
}

function deploy_argocd() {
  ns_ensure argocd

  log_info "Deploying Argo CD..."
  kubectl apply -f "$DEPLOY_DIR/argocd/install.yaml"

  log_info "Waiting for ArgoCD CRDs to become available..."
  for i in {1..30}; do
    if kubectl get crd applications.argoproj.io &>/dev/null; then
      log_info "ArgoCD CRDs are established."
      break
    fi
    log_info "Waiting for ArgoCD CRDs to be created ($i/30)..."
    sleep 5
  done

  if ! kubectl get crd applications.argoproj.io &>/dev/null; then
    log_error "Timed out waiting for ArgoCD CRDs."
    exit 1
  fi

  log_info "Waiting for ArgoCD Server deployment to be ready..."
  kubectl wait --for=condition=Available deployment/argocd-server -n argocd --timeout=120s

  log_info "Deploying ArgoCD Application manifest..."
  kubectl apply -f "$DEPLOY_DIR/argocd-app.yaml"

  register_cleanup "delete_argocd"
}

ns_ensure() { kubectl get ns "$1" &>/dev/null || kubectl create ns "$1"; }
deploy_postgres() { ns_ensure "$NAMESPACE"; kubectl apply -f "$DEPLOY_DIR/postgres.yaml" -n "$NAMESPACE"; }
deploy_llm() { ns_ensure "$NAMESPACE"; helm upgrade --install llm "$CHARTS_DIR/llm-fast-wrapper" -n "$NAMESPACE" --wait --timeout "$TIMEOUT_DEPLOY"; }

function delete_cluster() {
  log_warn "Removing cluster '$CLUSTER_NAME'..."
  helm uninstall llm -n "$NAMESPACE" --wait || true
  kubectl delete -f "$DEPLOY_DIR/argocd-app.yaml" --ignore-not-found || true
  kubectl delete -f "$DEPLOY_DIR/argocd/install.yaml" --ignore-not-found || true
  kubectl delete -f "$DEPLOY_DIR/postgres.yaml" -n "$NAMESPACE" --ignore-not-found || true
  k3d cluster delete "$CLUSTER_NAME" || true
}

handle_action() {
  case "$ACTION" in
    setup|create) 
        create_cluster; 
        deploy_postgres; 
        # deploy_argocd; 
        # deploy_llm; 
        ;;
    delete|teardown) delete_cluster ;;
    status) k3d cluster list ;;
    restart) 
        delete_cluster; 
        create_cluster; 
        deploy_postgres; 
        # deploy_argocd; 
        # deploy_llm 
        ;;
    *) log_error "Unknown action: $ACTION"; exit 1 ;;
  esac
}

main() { ensure_deps; ensure_artifacts; handle_action; CLEANUPS=(); }
main "$@"