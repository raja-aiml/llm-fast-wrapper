#!/usr/bin/env bash
set -euo pipefail

CLUSTER_NAME="llm-cluster"

# ───────────────────────────────────────────────
# 🔍 Check for `kind` and install if missing
# ───────────────────────────────────────────────
if ! command -v kind &>/dev/null; then
  echo "🔧 'kind' not found. Installing..."
  curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.22.0/kind-$(uname)-amd64
  chmod +x ./kind
  sudo mv ./kind /usr/local/bin/kind
  echo "✅ 'kind' installed successfully"
else
  echo "✅ 'kind' is already installed"
fi

# ───────────────────────────────────────────────
# 🚀 Create the cluster
# ───────────────────────────────────────────────
echo "⛴️  Creating kind cluster: $CLUSTER_NAME"
kind create cluster --name "$CLUSTER_NAME"

# ───────────────────────────────────────────────
# 🌐 Install ingress controller
# ───────────────────────────────────────────────
echo "🌐 Installing ingress-nginx controller"
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.10.1/deploy/static/provider/kind/deploy.yaml
kubectl wait --namespace ingress-nginx \
  --for=condition=Ready pods \
  --selector=app.kubernetes.io/component=controller \
  --timeout=120s

# ───────────────────────────────────────────────
# 🖥️  Deploy Kubernetes dashboard
# ───────────────────────────────────────────────
echo "📊 Deploying Kubernetes Dashboard"
kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v2.7.0/aio/deploy/recommended.yaml
kubectl create serviceaccount admin-user -n kubernetes-dashboard || true
kubectl create clusterrolebinding admin-user-binding \
  --clusterrole=cluster-admin \
  --serviceaccount=kubernetes-dashboard:admin-user || true
kubectl patch svc kubernetes-dashboard -n kubernetes-dashboard -p '{"spec": {"type": "NodePort"}}'

# 🔑 Output access token
echo "🔑 Admin token for dashboard access:"
kubectl -n kubernetes-dashboard create token admin-user