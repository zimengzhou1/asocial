#!/bin/bash
# Check status of remote k3s deployment
# Run this from your local Mac with KUBECONFIG pointing to remote k3s

set -e

echo "📊 Remote k3s Cluster Status"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if ! kubectl cluster-info &> /dev/null; then
    echo "❌ kubectl is not configured or cluster is not reachable"
    echo "Please set KUBECONFIG to your remote k3s config:"
    echo "  export KUBECONFIG=~/.kube/k3s-config"
    exit 1
fi

echo ""
echo "🖥️  Cluster Info:"
kubectl cluster-info | head -1

echo ""
echo "📦 Pods:"
kubectl get pods -n asocial -o wide

echo ""
echo "🔌 Services:"
kubectl get svc -n asocial

echo ""
echo "🌐 Ingress:"
kubectl get ingress -n asocial

echo ""
echo "💾 Persistent Volumes:"
kubectl get pvc -n asocial

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
