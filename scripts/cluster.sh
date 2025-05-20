#!/usr/bin/env bash
################################################################################
# cluster.sh – One‑touch lifecycle management for a local       sKubernetes stack #
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

# Set -Eeuo pipefail:  Exit immediately if a command exits with a non-zero status,
#                       treat unset variables as an error, and prevent errors in a pipeline
#                       from being masked.
set -Eeuo pipefail
# Attempt to enable 'inherit_errexit' (inherit errexit in command substitutions) if supported.
shopt -s inherit_errexit 2>/dev/null || true

################################# CONSTANTS ###################################
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
readonly DEPLOY_DIR="$REPO_ROOT/deploy"
readonly CHARTS_DIR="$DEPLOY_DIR/charts"
readonly CLUSTER_NAME="${CLUSTER_NAME:-llm-cluster}"
readonly NAMESPACE="${NAMESPACE:-llm}"
readonly LOG_LEVEL="${LOG_LEVEL:-info}"
readonly PORT="${PORT:-8080}"
readonly ACTION="${1:-setup}" # Default action is setup
readonly TIMEOUT_CLUSTER="300s"
readonly TIMEOUT_DEPLOY="300s"

# Declare arrays as constants.  This prevents accidental modification.
readonly REQUIRED_CMDS=(docker k3d kubectl helm)
readonly REQUIRED_DIRS=("$DEPLOY_DIR" "$CHARTS_DIR")
readonly REQUIRED_FILES=(
    "$DEPLOY_DIR/postgres.yaml"
    "$DEPLOY_DIR/argocd/install.yaml"
    "$DEPLOY_DIR/argocd-app.yaml"
)

################################# COLOURS #####################################
# Use a function to set colors, improving readability and maintainability.
# This also centralizes the color definitions.
function set_colors() {
  if [[ -t 2 ]]; then
    export RED="$(tput setaf 1)"
    export YEL="$(tput setaf 3)"
    export GRN="$(tput setaf 2)"
    export BLU="$(tput setaf 4)"
    export RST="$(tput sgr0)"
  else
    export RED=""
    export YEL=""
    export GRN=""
    export BLU=""
    export RST=""
  fi
}
set_colors # Call the function to initialize the color variables.

################################# LOGGING #####################################
#  Refactored logging function for better readability and maintainability.
#  Uses printf for consistent formatting.
function log() {
    local level="$1"
    local color="$2"
    local message="$3"
    local timestamp="$(date '+%Y-%m-%dT%H:%M:%S%z')"

    # Map log levels to numeric priorities.
    local current_level_num
    case "$LOG_LEVEL" in
      debug) current_level_num=0 ;;  
      info)  current_level_num=1 ;;  
      warn)  current_level_num=2 ;;  
      error) current_level_num=3 ;;  
      *)      current_level_num=1 ;; # default to info
    esac
    local message_level_num
    case "$level" in
      debug) message_level_num=0 ;;  
      info)  message_level_num=1 ;;  
      warn)  message_level_num=2 ;;  
      error) message_level_num=3 ;;  
      *)      message_level_num=1 ;;
    esac

    # Only log if the message level is at or above the configured log level.
    if [[ "$message_level_num" -ge "$current_level_num" ]]; then
      printf '%s %b[%5s]%b %s\n' "$timestamp" "$color" "$level" "$RST" "$message" >&2
    fi
}

#  Define helper functions for each log level.  This improves code clarity
#  and reduces the chance of errors in log calls.
function log_debug() { log "debug" "$BLU" "$*"; }
function log_info()  { log "info"  "$GRN" "$*"; }
function log_warn()  { log "warn"  "$YEL" "$*"; }
function log_error() { log "error" "$RED" "$*"; }

################################ CLEAN‑UP #####################################
#  Use an array to store cleanup commands, which is more robust and flexible
#  than using a string and eval.
CLEANUPS=()

#  Function to register cleanup commands.  Uses printf for safer string
#  handling.
function register_cleanup() {
  CLEANUPS+=("$1")
  # printf "Registered cleanup: %s\n" "$1" # debugging
}

#  Function to execute cleanup commands in reverse order.  Handles errors
#  more gracefully.
function execute_cleanup() {
  if ((${#CLEANUPS[@]} > 0)); then
    log_info "Running cleanup actions..."
    for ((i = ${#CLEANUPS[@]} - 1; i >= 0; i--)); do
      # Log the command before executing it.
      log_debug "Executing cleanup: ${CLEANUPS[$i]}"
      # Execute the command and check its exit status.
      if ! ${CLEANUPS[$i]}; then
        log_error "Cleanup command failed: ${CLEANUPS[$i]}"
        #  Do NOT use exit here.  The goal is to execute all cleanup
        #  commands, even if some fail.  The trap will handle the
        #  overall script exit status.
      fi
    done
  else
    log_debug "No cleanup actions registered."
  fi
}

# Trap signals (EXIT, INT, TERM) to ensure cleanup is performed.
# Use a single trap command for all signals.
trap '
  ec=$?;
  if [[ $ec -ne 0 ]]; then
    log_error "Script failed (exit $ec)";
  else
    log_info "Script завершено успешно."; # Script completed successfully.
  fi
  execute_cleanup
' EXIT INT TERM

############################### VALIDATION ####################################
#  Checks if a command exists.  Uses command -v, which is more portable
#  than which.
function cmd_exists() {
  command -v "$1" &>/dev/null
}

#  Installs a command using the appropriate package manager.
#  Improves error handling and logging.
function install_cmd() {
  local cmd="$1"
  log_info "Installing $cmd..."

  if cmd_exists brew; then
    brew install "$cmd"
  elif cmd_exists apt-get; then
    sudo apt-get update -qq && sudo apt-get install -y "$cmd"
  elif cmd_exists yum; then
    sudo yum install -y "$cmd"
  else
    log_error "No package manager (brew/apt/yum) found"
    return 1 # Explicitly return a non-zero status on failure
  fi
  #  No need to check cmd_exists here; the set -e option will cause
  #  the script to exit if the installation fails.
}

#  Ensures that all required dependencies are installed.
function ensure_deps() {
  log_info "Checking CLI dependencies..."
  local missing=()
  for cmd in "${REQUIRED_CMDS[@]}"; do
    if ! cmd_exists "$cmd"; then
      missing+=("$cmd")
    fi
  done

  if ((${#missing[@]} > 0)); then
    for cmd in "${missing[@]}"; do
      if ! install_cmd "$cmd"; then
        log_error "Failed to install required command: $cmd"
        return 1 # Exit if any installation fails
      fi
    done
  fi

  # Final check after attempting installation
  for cmd in "${REQUIRED_CMDS[@]}"; do
    if ! cmd_exists "$cmd"; then
      log_error "Missing required command: $cmd"
      return 1
    fi
  done
  log_info "All CLI dependencies are satisfied."
}

#  Ensures that all required directories and files exist.
function ensure_artifacts() {
  log_info "Checking required artifacts..."
  local missing_artifact=0 # Use a flag

  for dir in "${REQUIRED_DIRS[@]}"; do
    if [[ ! -d "$dir" ]]; then
      log_error "Directory missing: $dir"
      missing_artifact=1
    fi
  done

  for file in "${REQUIRED_FILES[@]}"; do
    if [[ ! -f "$file" ]]; then
      log_error "File missing: $file"
      missing_artifact=1
    fi
  done

  if [[ $missing_artifact -eq 1 ]]; then
    return 1 # Return non-zero if any are missing
  fi
  log_info "All required artifacts are present."
}

#  Checks if the required port is in use.
function check_port() {
  if ss -tuln 2>/dev/null | grep -q ":$PORT "; then
    log_warn "Port $PORT is in use"
  fi
}

############################ K8s UTILITIES ####################################
#  Ensures that a Kubernetes namespace exists.
function ns_ensure() {
  kubectl get ns "$1" &>/dev/null || kubectl create ns "$1"
  # No need to check the return code of kubectl create ns,
  # as the script will exit if it fails, due to set -e.
}

#  Waits for a Kubernetes resource to be ready.
function resource_ready() {
  local resource_type="$1"
  local resource_name="$2"
  local namespace="${3:-default}" # Default namespace
  local timeout="${4:-120s}"     # Default timeout

  log_info "Waiting for $resource_type/$resource_name in namespace $namespace to be ready..."
  kubectl wait --for=condition=ready "$resource_type/$resource_name" -n "$namespace" --timeout="$timeout"
}

############################### ACTIONS ######################################
#  Functions for each action (setup, delete, status, restart).  This
#  improves code organization and readability.

function create_cluster() {
  if k3d cluster list | grep -q "^$CLUSTER_NAME"; then
    log_warn "Cluster '$CLUSTER_NAME' already exists."
    return # Do not create if it exists.
  fi

  log_info "Creating k3d cluster '$CLUSTER_NAME'..."
  k3d cluster create "$CLUSTER_NAME" --timeout "$TIMEOUT_CLUSTER"
  register_cleanup "delete_cluster" # Register for cleanup

  resource_ready node --all default 60s
}

function deploy_postgres() {
  ns_ensure "$NAMESPACE"
  log_info "Deploying PostgreSQL to namespace '$NAMESPACE'..."
  kubectl apply -f "$DEPLOY_DIR/postgres.yaml" -n "$NAMESPACE"
  resource_ready deployment/postgres "$NAMESPACE" "$TIMEOUT_DEPLOY"
  register_cleanup "delete_postgres"
}

function delete_postgres() {
  log_info "Deleting PostgreSQL from namespace '$NAMESPACE'..."
  kubectl delete -f "$DEPLOY_DIR/postgres.yaml" -n "$NAMESPACE" --ignore-not-found
}

function deploy_argocd() {
  ns_ensure argocd
  log_info "Deploying Argo CD..."
  kubectl apply -f "$DEPLOY_DIR/argocd/install.yaml"
  resource_ready deployment/argocd-server argocd "$TIMEOUT_DEPLOY"
  kubectl apply -f "$DEPLOY_DIR/argocd-app.yaml"
  register_cleanup "delete_argocd"
}

function delete_argocd() {
  log_info "Deleting Argo CD..."
  kubectl delete -f "$DEPLOY_DIR/argocd-app.yaml" --ignore-not-found
  kubectl delete -f "$DEPLOY_DIR/argocd/install.yaml" --ignore-not-found
}

function deploy_llm() {
  ns_ensure "$NAMESPACE"
  log_info "Deploying LLM chart to namespace '$NAMESPACE'..."
  helm upgrade --install llm "$CHARTS_DIR/llm-fast-wrapper" -n "$NAMESPACE" --wait --timeout "$TIMEOUT_DEPLOY"
  register_cleanup "delete_llm"
}

function delete_llm() {
  log_info "Deleting LLM deployment..."
  helm uninstall llm -n "$NAMESPACE" --wait --timeout 60s || true
}

function delete_cluster() {
  log_warn "Removing cluster '$CLUSTER_NAME'..."
  #  Remove all deployments and services
  delete_llm
  delete_argocd
  delete_postgres
  kubectl delete ns "$NAMESPACE" argocd --ignore-not-found &>/dev/null || true
  k3d cluster delete "$CLUSTER_NAME" || true # Ignore errors, k3d might be gone
}

function status() {
  if ! k3d cluster list | grep -q "^$CLUSTER_NAME"; then
    log_error "Cluster '$CLUSTER_NAME' is not running."
    return 1
  fi

  log_info "Cluster Status for '$CLUSTER_NAME':"
  k3d cluster list | grep "$CLUSTER_NAME"
  echo

  log_info "Nodes:"
  kubectl get nodes
  echo

  log_info "Namespaces:"
  kubectl get ns
  echo

  log_info "Deployments in namespace '$NAMESPACE':"
  kubectl get deploy -n "$NAMESPACE" || true
  echo

  log_info "Services in namespace '$NAMESPACE':"
  kubectl get svc -n "$NAMESPACE" || true
  echo

  log_info "Applications:"
  kubectl get applications -A || true
  echo

  check_port
}

#  Function to handle the main action of the script.
function handle_action() {
  case "$ACTION" in
    setup|create)
      create_cluster
      deploy_postgres
      deploy_argocd
      deploy_llm
      status
      log_info "Ready → http://localhost:$PORT"
      ;;
    delete|teardown)
      delete_cluster
      ;;
    status)
      status
      ;;
    restart)
      delete_cluster
      create_cluster
      deploy_postgres
      deploy_argocd
      deploy_llm
      status
      ;;
    *)
      log_error "Unknown action: $ACTION"
      echo "Usage: $0 [setup|delete|status|restart]" >&2
      return 1 # Return non-zero for invalid action
      ;;
  esac
}

################################# MAIN #######################################
#  Main function to orchestrate the script execution.
function main() {
  ensure_deps
  ensure_artifacts
  handle_action
  CLEANUPS=() # Clear to prevent double cleanup on success
}

#  Call the main function.  This is necessary in shell scripts.
main "$@"
