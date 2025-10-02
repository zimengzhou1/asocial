#!/bin/bash
# Alternative: Remote server setup with Traefik (k3s default)
# This version uses Traefik instead of NGINX Ingress Controller

set -e

echo "🚀 Starting remote server setup for k3s (with Traefik)..."

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

echo "☸️  Installing k3s with Traefik..."
if ! command -v k3s &> /dev/null; then
    # Install k3s WITH Traefik (don't disable it)
    curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="--tls-san 100.124.233.93" sh -

    # Wait for k3s to be ready
    echo "⏳ Waiting for k3s to be ready..."
    sleep 10

    # Make kubectl accessible without sudo (optional)
    mkdir -p $HOME/.kube
    cp /etc/rancher/k3s/k3s.yaml $HOME/.kube/config
    chmod 600 $HOME/.kube/config

    echo "✅ k3s installed with Traefik"
else
    echo "✅ k3s already installed"
fi

echo "📋 Checking k3s status..."
systemctl status k3s --no-pager | head -20

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
echo "3. Update k8s/ingress.yaml to use Traefik:"
echo "   Change 'ingressClassName: nginx' to 'ingressClassName: traefik'"
echo ""
echo "4. Deploy the application:"
echo "   make remote-deploy"
