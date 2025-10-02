#!/bin/bash
set -e

echo "🚀 Starting Minikube cluster..."
minikube start --cpus=4 --memory=8192 --driver=docker

echo "📦 Enabling Ingress addon..."
minikube addons enable ingress

echo "⏳ Waiting for Ingress controller to be ready..."
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=300s

echo "🏗️  Creating namespace..."
kubectl apply -f k8s/namespace.yaml

echo "📥 Pulling images from GHCR..."
echo "   (This may take a few minutes on first run)"
minikube ssh "docker pull ghcr.io/zimengzhou1/asocial-backend:latest"
minikube ssh "docker pull ghcr.io/zimengzhou1/asocial-frontend:latest"

echo "🔧 Applying Kubernetes manifests..."
kubectl apply -f k8s/redis/
kubectl apply -f k8s/backend/
kubectl apply -f k8s/frontend/
kubectl apply -f k8s/ingress.yaml

echo "⏳ Waiting for deployments to be ready..."
kubectl wait --namespace asocial \
  --for=condition=ready pod \
  --selector=app=redis \
  --timeout=300s

kubectl wait --namespace asocial \
  --for=condition=ready pod \
  --selector=app=backend \
  --timeout=300s

kubectl wait --namespace asocial \
  --for=condition=ready pod \
  --selector=app=frontend \
  --timeout=300s

echo ""
echo "✅ Deployment complete!"
echo ""
echo "📊 Status:"
kubectl get pods -n asocial
echo ""
echo "🌐 To access the application:"
echo "   Run: minikube tunnel"
echo "   Then visit: http://localhost"
echo ""
echo "📝 View logs:"
echo "   Run: ./scripts/k8s-logs.sh"
