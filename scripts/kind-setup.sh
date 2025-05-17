#!/usr/bin/env bash
set -euo pipefail

CLUSTER_NAME="llm-cluster"
kind create cluster --name "$CLUSTER_NAME"
# Install ingress controller
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.10.1/deploy/static/provider/kind/deploy.yaml
kubectl wait --namespace ingress-nginx --for=condition=Ready pods --selector=app.kubernetes.io/component=controller --timeout=120s

# Deploy Kubernetes dashboard
echo "Deploying dashboard"
kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v2.7.0/aio/deploy/recommended.yaml
kubectl create serviceaccount admin-user -n kubernetes-dashboard || true
kubectl create clusterrolebinding admin-user-binding --clusterrole=cluster-admin --serviceaccount=kubernetes-dashboard:admin-user || true
kubectl patch svc kubernetes-dashboard -n kubernetes-dashboard -p '{"spec": {"type": "NodePort"}}'
kubectl -n kubernetes-dashboard create token admin-user
