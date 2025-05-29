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
# Usage: cluster.sh [options] <action>
# Env vars:
#   CLUSTER_NAME  k3d cluster name               (default: llm-cluster)
#   NAMESPACE     K8s namespace for the stack    (default: llm)
#   LOG_LEVEL     debug|info|warn|error          (default: info)
#   PORT          Local app port                 (default: 8080)
################################################################################
set -Eeuo pipefail
shopt -s inherit_errexit 2>/dev/null || true
# Program name for messages
PROGNAME="$(basename "$0")"

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
readonly DEPLOY_DIR="$REPO_ROOT/deploy"
readonly CHARTS_DIR="$DEPLOY_DIR/charts"
readonly CLUSTER_NAME="${CLUSTER_NAME:-llm-cluster}"
readonly NAMESPACE="${NAMESPACE:-llm}"
LOG_LEVEL="${LOG_LEVEL:-info}"
readonly PORT="${PORT:-8080}"
readonly TIMEOUT_CLUSTER="300s"
readonly TIMEOUT_DEPLOY="300s"
readonly REQUIRED_CMDS=(docker k3d kubectl helm)
readonly REQUIRED_DIRS=("$DEPLOY_DIR" "$CHARTS_DIR")
readonly REQUIRED_FILES=(
    "$DEPLOY_DIR/postgres.yaml"
    "$DEPLOY_DIR/argocd/install.yaml"
    "$DEPLOY_DIR/argocd-app.yaml"
)
# Local Docker image settings for LLM chart
readonly IMAGE_NAME="${IMAGE_NAME:-llm-fast-wrapper}"
readonly IMAGE_TAG="${IMAGE_TAG:-local}"

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

trap 'ec=$?; if (( ec )); then log_error "Script failed (exit $ec)"; execute_cleanup; else log_info "Script completed successfully."; fi' EXIT INT TERM

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
  k3d cluster create "$CLUSTER_NAME" --port 5432:5432@loadbalancer --timeout "$TIMEOUT_CLUSTER"
  register_cleanup "delete_cluster"
  kubectl wait --for=condition=Ready nodes --all --timeout=60s
}

deploy_argocd() {
  # Ensure ArgoCD namespace exists
  ns_ensure argocd
  log_info "Deploying Argo CD via official manifest"
  # Use the official Argo CD install manifest for full component and CRD installation
  kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

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

  # Register cleanup to remove Argo CD resources
  register_cleanup "delete_argocd"
}

ns_ensure() { kubectl get ns "$1" &>/dev/null || kubectl create ns "$1"; }
deploy_postgres() { ns_ensure "$NAMESPACE"; kubectl apply -f "$DEPLOY_DIR/postgres.yaml" -n "$NAMESPACE"; }

# Build and import the llm-fast-wrapper Docker image into the k3d cluster
build_image() {
  log_info "Building Docker image ${IMAGE_NAME}:${IMAGE_TAG}"
  docker build -t "${IMAGE_NAME}:${IMAGE_TAG}" "${REPO_ROOT}"
  log_info "Importing image into k3d cluster ${CLUSTER_NAME}"
  k3d image import "${IMAGE_NAME}:${IMAGE_TAG}" -c "${CLUSTER_NAME}"
  register_cleanup "docker rmi ${IMAGE_NAME}:${IMAGE_TAG} || true"
}

deploy_llm() {
  ns_ensure "$NAMESPACE"
  build_image
  helm upgrade --install llm "$CHARTS_DIR/llm-fast-wrapper" \
    -n "$NAMESPACE" \
    --set image.repository="${IMAGE_NAME}" \
    --set image.tag="${IMAGE_TAG}" \
    --wait \
    --timeout "$TIMEOUT_DEPLOY"
}

delete_cluster() {
  log_warn "Removing cluster '$CLUSTER_NAME'..."
  # Uninstall LLM Helm release
  helm uninstall llm -n "$NAMESPACE" --wait || true
  # Delete Argo CD Application
  kubectl delete -f "$DEPLOY_DIR/argocd-app.yaml" --ignore-not-found || true
  # Delete Argo CD core by removing its namespace
  delete_argocd || true
  # Delete Postgres resources
  kubectl delete -f "$DEPLOY_DIR/postgres.yaml" -n "$NAMESPACE" --ignore-not-found || true
  # Remove the k3d cluster
  k3d cluster delete "$CLUSTER_NAME" || true
}

# Remove ArgoCD installation by deleting its namespace
delete_argocd() {
  log_warn "Removing Argo CD resources (namespace 'argocd')..."
  kubectl delete namespace argocd --ignore-not-found || true
}
 
handle_action() {
  case "$ACTION" in
    setup|create)
        create_cluster
        deploy_postgres
        deploy_argocd
        deploy_llm
        ;;
    delete|teardown)
        delete_cluster
        ;;
    status)
        k3d cluster list
        ;;
    restart)
        delete_cluster
        create_cluster
        deploy_postgres
        deploy_argocd
        deploy_llm
        ;;
    *) log_error "Unknown action: $ACTION"; exit 1 ;;
  esac
}

show_help() {
  cat <<EOF
Usage: ${PROGNAME} [options] <action>
Options:
  -h, --help       Show this help message and exit
  -q, --quiet      Only log warnings and errors
  -d, --debug      Log debug messages and enable shell debug tracing
Actions:
  setup|create     Create cluster and deploy services (default)
  delete|teardown  Tear down the cluster and deployed services
  status           Show cluster status
  restart          Recreate cluster and redeploy services
EOF
}

parse_args() {
  ACTION="setup"
  while [[ $# -gt 0 ]]; do
    case "$1" in
      -h|--help)
        show_help
        exit 0
        ;;
      -q|--quiet)
        LOG_LEVEL="warn"
        shift
        ;;
      -d|--debug)
        LOG_LEVEL="debug"
        set -x
        shift
        ;;
      --)
        shift
        break
        ;;
      -*)
        log_error "Unknown option: $1"
        show_help
        exit 1
        ;;
      *)
        ACTION="$1"
        shift
        break
        ;;
    esac
  done
}

main() {
  parse_args "$@"
  ensure_deps
  ensure_artifacts
  handle_action
}
main "$@"