#!/bin/bash
# Setup secrets for production (K3s)
set -e

echo "🔐 Setting up secrets for PRODUCTION (K3s)"
echo ""
echo "⚠️  WARNING: You are setting up production secrets!"
echo "   This will create secrets on your production cluster."
echo ""

# Check current context
CURRENT_CONTEXT=$(kubectl config current-context)
if [ -z "$CURRENT_CONTEXT" ]; then
    echo ""
    echo "⚠️  ERROR: No kubectl context is set"
    echo ""
    echo "Please set your kubectl context:"
    echo "  kubectl config use-context k3s"
    echo ""
    echo "Or if you haven't merged your k3s config yet, see docs/KUBECTL_CONTEXTS.md"
    exit 1
fi

echo "Current kubectl context: $CURRENT_CONTEXT"

# Verify we're NOT on minikube context
if [ "$CURRENT_CONTEXT" = "minikube" ]; then
    echo ""
    echo "⚠️  ERROR: Current context is 'minikube'"
    echo ""
    echo "This script is for PRODUCTION (k3s) only."
    echo "You are currently connected to your local Minikube cluster."
    echo ""
    echo "Switch to production context:"
    echo "  kubectl config use-context k3s"
    echo ""
    echo "For development setup, use: make k8s-secrets-dev"
    exit 1
fi

echo ""
read -p "Is this the correct production cluster? (yes/no) " -r
if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
    echo "Cancelled"
    exit 1
fi

echo ""
echo "✅ Proceeding with production setup on context: $CURRENT_CONTEXT"
echo ""

# Check if namespace exists
if ! kubectl get namespace asocial &> /dev/null; then
    echo "📦 Creating namespace 'asocial'..."
    kubectl apply -f k8s/namespace.yaml
    echo ""
fi

# Firebase credentials
echo "🔥 Setting up Firebase credentials (PRODUCTION)..."
echo ""

FIREBASE_PROD_PATH=""
if [ -f "/Users/zimeng/Home/asocial-6522d-firebase-adminsdk-fbsvc-984608b01c.json" ]; then
    FIREBASE_PROD_PATH="/Users/zimeng/Home/asocial-6522d-firebase-adminsdk-fbsvc-984608b01c.json"
    echo "Found production Firebase credentials at:"
    echo "  $FIREBASE_PROD_PATH"
    echo ""
    read -p "Use this file? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        FIREBASE_PROD_PATH=""
    fi
fi

if [ -z "$FIREBASE_PROD_PATH" ]; then
    echo "Enter path to PRODUCTION Firebase service account JSON:"
    read -p "Path: " FIREBASE_PROD_PATH

    if [ ! -f "$FIREBASE_PROD_PATH" ]; then
        echo "❌ File not found: $FIREBASE_PROD_PATH"
        exit 1
    fi
fi

# Verify it's a valid JSON
if ! jq empty "$FIREBASE_PROD_PATH" 2>/dev/null; then
    echo "❌ Invalid JSON file: $FIREBASE_PROD_PATH"
    exit 1
fi

# Check project_id to ensure it's production
PROJECT_ID=$(jq -r '.project_id' "$FIREBASE_PROD_PATH")
echo ""
echo "Firebase Project ID: $PROJECT_ID"
if [[ $PROJECT_ID == *"dev"* ]] || [[ $PROJECT_ID == *"test"* ]]; then
    echo "⚠️  WARNING: This looks like a development Firebase project!"
    read -p "Continue anyway? (yes/no) " -r
    if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        echo "Cancelled"
        exit 1
    fi
fi

# Create/update Firebase secret
if kubectl get secret firebase-credentials -n asocial &> /dev/null; then
    echo ""
    echo "⚠️  Secret 'firebase-credentials' already exists in production"
    read -p "Replace it? (yes/no) " -r
    if [[ $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        kubectl delete secret firebase-credentials -n asocial
        kubectl create secret generic firebase-credentials \
            --from-file=serviceAccount.json="$FIREBASE_PROD_PATH" \
            -n asocial
        echo "✅ Firebase secret updated"
    else
        echo "⏭️  Skipping Firebase secret"
    fi
else
    kubectl create secret generic firebase-credentials \
        --from-file=serviceAccount.json="$FIREBASE_PROD_PATH" \
        -n asocial
    echo "✅ Firebase secret created"
fi

echo ""

# PostgreSQL credentials
echo "🐘 Setting up PostgreSQL credentials (PRODUCTION)..."
echo ""

if kubectl get secret postgres-secret -n asocial &> /dev/null; then
    echo "⚠️  PostgreSQL secret already exists in production"
    echo ""
    read -p "Replace it? This will change the database password! (yes/no) " -r
    if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        echo "⏭️  Skipping PostgreSQL secret"
        echo ""
        echo "📋 Current secrets:"
        kubectl get secrets -n asocial
        echo ""
        echo "✅ Setup complete (Firebase only)"
        exit 0
    fi
    kubectl delete secret postgres-secret -n asocial
fi

echo ""
echo "Generating strong random password..."
POSTGRES_PASSWORD=$(openssl rand -base64 32)

echo ""
echo "╔════════════════════════════════════════════════════════════════╗"
echo "║ ⚠️  SAVE THIS PASSWORD TO YOUR PASSWORD MANAGER NOW!         ║"
echo "╚════════════════════════════════════════════════════════════════╝"
echo ""
echo "PostgreSQL Password:"
echo "  $POSTGRES_PASSWORD"
echo ""
echo "Database Connection String:"
echo "  postgres://asocial:$POSTGRES_PASSWORD@postgres:5432/asocial?sslmode=require"
echo ""
read -p "Press Enter after you've saved the password securely..." dummy

kubectl create secret generic postgres-secret \
    --from-literal=POSTGRES_USER=asocial \
    --from-literal=POSTGRES_PASSWORD="$POSTGRES_PASSWORD" \
    --from-literal=POSTGRES_DB=asocial \
    -n asocial

echo ""
echo "✅ PostgreSQL secret created"

echo ""
echo "📋 Verifying secrets..."
kubectl get secrets -n asocial

echo ""
echo "✅ Production secrets setup complete!"
echo ""
echo "⚠️  Important next steps:"
echo "  1. Ensure you've saved the PostgreSQL password securely"
echo "  2. Deploy to production: make remote-deploy"
echo "  3. The init container will run migrations automatically"
echo ""
echo "🔒 Security reminder:"
echo "  - Never commit these credentials to git"
echo "  - Rotate passwords periodically"
echo "  - Use different credentials for dev and prod"
