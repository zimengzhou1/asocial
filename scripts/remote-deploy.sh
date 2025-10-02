#!/bin/bash
# Deploy application to remote k3s cluster
# Run this from your local Mac with KUBECONFIG pointing to remote k3s

set -e

echo "🚀 Deploying to remote k3s cluster..."

# Check if kubectl is configured
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ kubectl is not configured or cluster is not reachable"
    echo "Please set KUBECONFIG to your remote k3s config:"
    echo "  export KUBECONFIG=~/.kube/k3s-config"
    exit 1
fi

echo "📋 Current cluster:"
kubectl cluster-info | head -1

echo ""
read -p "Is this the correct cluster? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Deployment cancelled"
    exit 1
fi

echo "📦 Applying Kubernetes manifests..."
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/redis/
kubectl apply -f k8s/backend/
kubectl apply -f k8s/frontend/
kubectl apply -f k8s/ingress.yaml

echo "⏳ Waiting for pods to be ready..."
echo "   (This may take a few minutes on first deploy while images are pulled)"

kubectl wait --namespace asocial \
  --for=condition=ready pod \
  --selector=app=redis \
  --timeout=300s || true

kubectl wait --namespace asocial \
  --for=condition=ready pod \
  --selector=app=backend \
  --timeout=300s || true

kubectl wait --namespace asocial \
  --for=condition=ready pod \
  --selector=app=frontend \
  --timeout=300s || true

echo ""
echo "✅ Deployment complete!"
echo ""
echo "📊 Current status:"
kubectl get pods,svc,ingress -n asocial

echo ""
echo "🌐 Next steps:"
echo "1. Set up Cloudflare Tunnel to expose the application:"
echo "   make remote-tunnel-setup"
echo ""
echo "2. View logs:"
echo "   make remote-logs"
