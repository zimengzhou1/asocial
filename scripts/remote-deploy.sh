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

echo "🔐 Checking for required secrets..."
MISSING_SECRETS=false

if ! kubectl get secret firebase-credentials -n asocial &> /dev/null; then
    echo "❌ ERROR: firebase-credentials secret not found!"
    MISSING_SECRETS=true
fi

if ! kubectl get secret postgres-secret -n asocial &> /dev/null; then
    echo "❌ ERROR: postgres-secret not found!"
    MISSING_SECRETS=true
fi

if [ "$MISSING_SECRETS" = true ]; then
    echo ""
    echo "You MUST create the production secrets before deploying."
    echo ""
    echo "Run the automated setup script:"
    echo "  make remote-secrets"
    echo ""
    echo "Or manually create secrets (see k8s/README.md)"
    exit 1
fi

echo "✅ Production secrets found"

echo "🔧 Creating database migrations ConfigMap..."
kubectl create configmap db-migrations \
  --from-file=migrations/ \
  --namespace=asocial \
  --dry-run=client -o yaml | kubectl apply -f -

echo "📦 Applying Kubernetes manifests (prod overlay)..."
kubectl apply -f k8s/namespace.yaml
kubectl apply -k k8s/postgres/overlays/prod
kubectl apply -f k8s/redis/
kubectl apply -k k8s/backend/overlays/prod
kubectl apply -k k8s/frontend/overlays/prod
kubectl apply -f k8s/ingress.yaml

echo "⏳ Waiting for pods to be ready..."
echo "   (This may take a few minutes on first deploy while images are pulled)"

kubectl wait --namespace asocial \
  --for=condition=ready pod \
  --selector=app=postgres \
  --timeout=300s || true

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
