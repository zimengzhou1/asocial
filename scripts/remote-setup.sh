#!/bin/bash
# Remote server initial setup script
# Run this on your Proxmox VM via SSH

set -e

echo "🚀 Starting remote server setup for k3s..."

# Check if running as root or with sudo
if [ "$EUID" -ne 0 ]; then
    echo "⚠️  Please run as root or with sudo"
    exit 1
fi

echo "📦 Updating package lists..."
apt update

echo "📦 Installing required packages..."
apt install -y curl wget git

echo "🐳 Installing Docker (for pulling images from GHCR)..."
if ! command -v docker &> /dev/null; then
    curl -fsSL https://get.docker.com | sh
    systemctl enable docker
    systemctl start docker
    echo "✅ Docker installed"
else
    echo "✅ Docker already installed"
fi

echo "☸️  Installing k3s..."
if ! command -v k3s &> /dev/null; then
    # Install k3s with Traefik disabled (we use NGINX Ingress)
    # curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="--disable traefik" sh -
    curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="--disable traefik --tls-san 100.124.233.93" sh -

    # Wait for k3s to be ready
    echo "⏳ Waiting for k3s to be ready..."
    sleep 10

    # Make kubectl accessible without sudo (optional)
    mkdir -p $HOME/.kube
    cp /etc/rancher/k3s/k3s.yaml $HOME/.kube/config
    chmod 600 $HOME/.kube/config

    echo "✅ k3s installed"
else
    echo "✅ k3s already installed"
fi

echo "📋 Checking k3s status..."
systemctl status k3s --no-pager | head -20

echo "🌐 Installing NGINX Ingress Controller..."
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.1/deploy/static/provider/cloud/deploy.yaml

echo "⏳ Waiting for NGINX Ingress Controller to be ready..."
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=180s || echo "⚠️  Warning: Ingress controller may not be ready yet"

echo "✅ NGINX Ingress Controller installed"

echo ""
echo "✅ Remote server setup complete!"
echo ""
echo "📝 Next steps:"
echo "1. Copy kubeconfig to your local Mac:"
echo "   scp root@<server-ip>:/etc/rancher/k3s/k3s.yaml ~/.kube/k3s-config"
echo "   # Edit ~/.kube/k3s-config and replace 127.0.0.1 with your server IP"
echo ""
echo "2. Use the config on your Mac:"
echo "   export KUBECONFIG=~/.kube/k3s-config"
echo "   kubectl get nodes"
echo ""
echo "3. Deploy the application:"
echo "   make remote-deploy"
