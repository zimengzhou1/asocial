#!/bin/bash
# Tail logs from remote k3s pods
# Run this from your local Mac with KUBECONFIG pointing to remote k3s

set -e

if ! kubectl cluster-info &> /dev/null; then
    echo "❌ kubectl is not configured or cluster is not reachable"
    echo "Please set KUBECONFIG to your remote k3s config:"
    echo "  export KUBECONFIG=~/.kube/k3s-config"
    exit 1
fi

echo "📝 Tailing logs from remote k3s pods..."
echo "   (Press Ctrl+C to stop)"
echo ""

# Tail logs from all pods in asocial namespace
kubectl logs -n asocial -l app=backend --all-containers=true --follow --tail=50 --max-log-requests=10 &
kubectl logs -n asocial -l app=frontend --all-containers=true --follow --tail=50 --max-log-requests=10 &
kubectl logs -n asocial -l app=redis --all-containers=true --follow --tail=50 --max-log-requests=10 &
kubectl logs -n asocial -l app=postgres --all-containers=true --follow --tail=50 --max-log-requests=10 &

wait
