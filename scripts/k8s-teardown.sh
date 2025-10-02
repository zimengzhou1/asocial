#!/bin/bash
set -e

echo "🧹 Deleting Kubernetes resources..."
kubectl delete -f k8s/ingress.yaml --ignore-not-found=true
kubectl delete -f k8s/frontend/ --ignore-not-found=true
kubectl delete -f k8s/backend/ --ignore-not-found=true
kubectl delete -f k8s/redis/ --ignore-not-found=true
kubectl delete -f k8s/namespace.yaml --ignore-not-found=true

echo "🛑 Stopping Minikube..."
minikube stop

echo "✅ Teardown complete!"
echo ""
echo "To delete the cluster entirely (including all images and data):"
echo "   Run: minikube delete"
