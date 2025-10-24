#!/bin/bash
# Setup secrets for development (Minikube)
set -e

echo "🔐 Setting up secrets for DEVELOPMENT (Minikube)"
echo ""

# Check current context
CURRENT_CONTEXT=$(kubectl config current-context)
echo "Current kubectl context: $CURRENT_CONTEXT"

# Verify we're on minikube context
if [ "$CURRENT_CONTEXT" != "minikube" ]; then
    echo ""
    echo "⚠️  ERROR: Current context is '$CURRENT_CONTEXT', not 'minikube'"
    echo ""
    echo "This script is for development (Minikube) only."
    echo "To switch context: kubectl config use-context minikube"
    echo ""
    echo "For production (k3s) setup, use: ./scripts/setup-secrets-prod.sh"
    exit 1
fi

echo "✅ Correct context (minikube) - proceeding with dev setup"
echo ""

# Check if namespace exists
if ! kubectl get namespace asocial &> /dev/null; then
    echo "📦 Creating namespace 'asocial'..."
    kubectl apply -f k8s/namespace.yaml
    echo ""
fi

# Firebase credentials
echo "🔥 Setting up Firebase credentials..."
FIREBASE_DEV_PATH="/Users/zimeng/Home/asocial-dev-89a09-firebase-adminsdk-fbsvc-ff5fc8ef22.json"

if [ ! -f "$FIREBASE_DEV_PATH" ]; then
    echo ""
    echo "⚠️  WARNING: Dev Firebase credentials not found at:"
    echo "   $FIREBASE_DEV_PATH"
    echo ""
    read -p "Enter path to Firebase service account JSON: " FIREBASE_DEV_PATH

    if [ ! -f "$FIREBASE_DEV_PATH" ]; then
        echo "❌ File not found: $FIREBASE_DEV_PATH"
        exit 1
    fi
fi

# Check if secret already exists
if kubectl get secret firebase-credentials -n asocial &> /dev/null; then
    echo "   Secret 'firebase-credentials' already exists"
    read -p "   Replace it? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        kubectl delete secret firebase-credentials -n asocial
        kubectl create secret generic firebase-credentials \
            --from-file=serviceAccount.json="$FIREBASE_DEV_PATH" \
            -n asocial
        echo "   ✅ Firebase secret updated"
    else
        echo "   ⏭️  Skipping Firebase secret"
    fi
else
    kubectl create secret generic firebase-credentials \
        --from-file=serviceAccount.json="$FIREBASE_DEV_PATH" \
        -n asocial
    echo "   ✅ Firebase secret created"
fi

echo ""

# PostgreSQL credentials
echo "🐘 Setting up PostgreSQL credentials..."

if kubectl get secret postgres-secret -n asocial &> /dev/null; then
    echo "   Secret 'postgres-secret' already exists"
    read -p "   Replace it? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        kubectl delete secret postgres-secret -n asocial
        kubectl apply -f k8s/postgres/overlays/dev/secret.yaml
        echo "   ✅ PostgreSQL secret updated"
    else
        echo "   ⏭️  Skipping PostgreSQL secret"
    fi
else
    kubectl apply -f k8s/postgres/overlays/dev/secret.yaml
    echo "   ✅ PostgreSQL secret created"
fi

echo ""

# JWT Secret for anonymous sessions (dev uses a simple fixed secret)
echo "🔑 Setting up JWT secret for anonymous sessions (DEV)..."

if kubectl get secret backend-secret -n asocial &> /dev/null; then
    echo "   Secret 'backend-secret' already exists"
    read -p "   Replace it? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        kubectl delete secret backend-secret -n asocial
        kubectl create secret generic backend-secret \
            --from-literal=jwt-secret="dev_jwt_secret_do_not_use_in_production_12345" \
            -n asocial
        echo "   ✅ Backend secret updated"
    else
        echo "   ⏭️  Skipping backend secret"
    fi
else
    kubectl create secret generic backend-secret \
        --from-literal=jwt-secret="dev_jwt_secret_do_not_use_in_production_12345" \
        -n asocial
    echo "   ✅ Backend secret created"
fi

echo ""
echo "📋 Verifying secrets..."
kubectl get secrets -n asocial

echo ""
echo "✅ Development secrets setup complete!"
echo ""
echo "Next steps:"
echo "  1. Deploy to minikube: make k8s-setup"
echo "  2. Or deploy manually: make k8s-deploy"
