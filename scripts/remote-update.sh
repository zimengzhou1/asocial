#!/bin/bash
# Update remote k3s deployment with latest images
# Run this from your local Mac with KUBECONFIG pointing to remote k3s

set -e

echo "🔄 Updating remote k3s deployment..."

if ! kubectl cluster-info &> /dev/null; then
    echo "❌ kubectl is not configured or cluster is not reachable"
    echo "Please set KUBECONFIG to your remote k3s config:"
    echo "  export KUBECONFIG=~/.kube/k3s-config"
    exit 1
fi

echo "📋 Current cluster:"
kubectl cluster-info | head -1

echo ""
echo "This will pull the latest images from GHCR and restart pods."
read -p "Continue? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Update cancelled"
    exit 1
fi

echo "🔄 Restarting deployments to pull latest images..."

# Restart deployments (this will pull latest images if imagePullPolicy is Always)
kubectl rollout restart deployment/backend -n asocial
kubectl rollout restart deployment/frontend -n asocial
kubectl rollout restart statefulset/redis -n asocial

echo "⏳ Waiting for rollout to complete..."
kubectl rollout status deployment/backend -n asocial
kubectl rollout status deployment/frontend -n asocial
kubectl rollout status statefulset/redis -n asocial

echo ""
echo "✅ Update complete!"
echo ""
echo "📊 Current status:"
kubectl get pods -n asocial
