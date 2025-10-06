#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

cd "${REPO_ROOT}"

echo "🚀 Starting Minikube cluster..."
minikube start --cpus=4 --memory=8192 --driver=docker

echo "🔍 Verifying kubectl context is set to minikube..."
CURRENT_CONTEXT=$(kubectl config current-context)
if [ "$CURRENT_CONTEXT" != "minikube" ]; then
    echo "⚠️  WARNING: Current context is '$CURRENT_CONTEXT', not 'minikube'"
    echo "Switching to minikube context..."
    kubectl config use-context minikube
fi
echo "✅ Using context: $(kubectl config current-context)"

echo "📦 Enabling Ingress addon..."
minikube addons enable ingress

echo "⏳ Waiting for Ingress controller to be ready..."
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=300s

echo "🏗️  Creating namespace..."
kubectl apply -f k8s/namespace.yaml

echo "🔐 Checking for required secrets..."
if ! kubectl get secret firebase-credentials -n asocial &> /dev/null || \
   ! kubectl get secret postgres-secret -n asocial &> /dev/null; then
    echo ""
    echo "⚠️  Missing required secrets!"
    echo ""
    echo "Please run secret setup first:"
    echo "  make k8s-secrets-dev"
    echo ""
    read -p "Do you want to run it now? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        ./scripts/setup-secrets-dev.sh
    else
        echo "❌ Cannot proceed without secrets"
        exit 1
    fi
fi
echo "✅ Secrets found"

# Check if we should use local images
if [ "${USE_LOCAL_IMAGES}" = "true" ]; then
    echo "🔨 Building images locally using minikube's Docker daemon..."
    echo "   (Setting DOCKER_HOST to minikube)"

    # Point Docker to minikube's daemon
    eval $(minikube docker-env)

    # Build backend
    echo "   Building backend..."
    docker build -f go.Dockerfile -t ghcr.io/zimengzhou1/asocial-backend:latest .

    # Build frontend
    echo "   Building frontend..."
    docker build -f frontend/dev.Dockerfile \
      --build-arg NEXT_PUBLIC_BACKEND_URL=http://localhost \
      --build-arg NEXT_PUBLIC_BACKEND_WS_URL=ws://localhost/api/chat \
      --build-arg NEXT_PUBLIC_FIREBASE_API_KEY=AIzaSyDl_ozuvDLnMOe7tLwY2pk_3BHIutHMHcY \
      --build-arg NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN=asocial-dev-89a09.firebaseapp.com \
      --build-arg NEXT_PUBLIC_FIREBASE_PROJECT_ID=asocial-dev-89a09 \
      -t ghcr.io/zimengzhou1/asocial-frontend:latest ./frontend

    echo "✅ Local images built successfully"
else
    echo "📥 Pulling images from GHCR..."
    echo "   (This may take a few minutes on first run)"
    minikube ssh "docker pull ghcr.io/zimengzhou1/asocial-backend:latest"
    minikube ssh "docker pull ghcr.io/zimengzhou1/asocial-frontend:latest"
fi

echo "🔧 Creating database migrations ConfigMap..."
kubectl create configmap db-migrations \
  --from-file=migrations/ \
  --namespace=asocial \
  --dry-run=client -o yaml | kubectl apply -f -

echo "🔧 Applying Kubernetes manifests (dev overlay)..."
kubectl apply -k k8s/postgres/overlays/dev
kubectl apply -f k8s/redis/
kubectl apply -k k8s/backend/overlays/dev
kubectl apply -k k8s/frontend/overlays/dev
kubectl apply -f k8s/ingress.yaml

echo "⏳ Waiting for deployments to be ready..."
kubectl wait --namespace asocial \
  --for=condition=ready pod \
  --selector=app=postgres \
  --timeout=300s

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
echo ""
echo "✅ Deployment complete!"
echo ""
echo "📊 Status:"
kubectl get pods -n asocial
echo ""
echo "🌐 To access the application:"
echo "   Run: make k8s-tunnel"
echo "   Then visit: http://localhost"
echo ""
echo "📝 View logs:"
echo "   Run: make k8s-logs"
echo ""
if [ "${USE_LOCAL_IMAGES}" = "true" ]; then
    echo "ℹ️  Using locally built images"
    echo "   To rebuild: USE_LOCAL_IMAGES=true make k8s-setup"
else
    echo "ℹ️  Using images from GHCR"
    echo "   To use local images: USE_LOCAL_IMAGES=true make k8s-setup"
fi
