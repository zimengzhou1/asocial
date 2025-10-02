#!/bin/bash
# Setup Cloudflare Tunnel on remote server
# Run this script on your Proxmox VM via SSH

set -e

echo "🌐 Setting up Cloudflare Tunnel..."

# Check if running as root or with sudo
if [ "$EUID" -ne 0 ]; then
    echo "⚠️  Please run as root or with sudo"
    exit 1
fi

echo "📦 Installing cloudflared..."
if ! command -v cloudflared &> /dev/null; then
    # Download and install cloudflared
    wget -q https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
    dpkg -i cloudflared-linux-amd64.deb
    rm cloudflared-linux-amd64.deb
    echo "✅ cloudflared installed"
else
    echo "✅ cloudflared already installed"
fi

echo ""
echo "🔐 Please authenticate with Cloudflare..."
echo "   This will open a browser window. If running over SSH, you may need to:"
echo "   1. Run this locally: cloudflared tunnel login"
echo "   2. Copy ~/.cloudflared/cert.pem to the server"
echo ""
read -p "Press Enter to continue with authentication..."

cloudflared tunnel login

if [ ! -f ~/.cloudflared/cert.pem ]; then
    echo "❌ Authentication failed or cert.pem not found"
    exit 1
fi

echo ""
read -p "Enter a name for your tunnel (e.g., asocial-tunnel): " TUNNEL_NAME

if [ -z "$TUNNEL_NAME" ]; then
    echo "❌ Tunnel name cannot be empty"
    exit 1
fi

echo "🔧 Creating tunnel: $TUNNEL_NAME..."
cloudflared tunnel create "$TUNNEL_NAME"

# Get tunnel UUID
TUNNEL_UUID=$(cloudflared tunnel list | grep "$TUNNEL_NAME" | awk '{print $1}')

if [ -z "$TUNNEL_UUID" ]; then
    echo "❌ Failed to get tunnel UUID"
    exit 1
fi

echo "✅ Tunnel created with UUID: $TUNNEL_UUID"

echo ""
read -p "Enter your domain name (e.g., asocial.yourdomain.com): " DOMAIN_NAME

if [ -z "$DOMAIN_NAME" ]; then
    echo "❌ Domain name cannot be empty"
    exit 1
fi

echo "🌐 Setting up DNS routing for $DOMAIN_NAME..."
cloudflared tunnel route dns "$TUNNEL_NAME" "$DOMAIN_NAME"

echo "📝 Creating tunnel configuration..."
mkdir -p ~/.cloudflared

cat > ~/.cloudflared/config.yml <<EOF
tunnel: $TUNNEL_UUID
credentials-file: /root/.cloudflared/$TUNNEL_UUID.json

ingress:
  - hostname: $DOMAIN_NAME
    service: http://localhost:80
  - service: http_status:404
EOF

echo "✅ Configuration created at ~/.cloudflared/config.yml"

echo "🔧 Installing tunnel as a system service..."
cloudflared service install

echo "🚀 Starting tunnel service..."
systemctl start cloudflared
systemctl enable cloudflared

echo ""
echo "✅ Cloudflare Tunnel setup complete!"
echo ""
echo "📊 Tunnel status:"
systemctl status cloudflared --no-pager | head -15

echo ""
echo "🌐 Your application should now be accessible at:"
echo "   https://$DOMAIN_NAME"
echo ""
echo "📝 Useful commands:"
echo "   systemctl status cloudflared    # Check tunnel status"
echo "   systemctl restart cloudflared   # Restart tunnel"
echo "   journalctl -u cloudflared -f    # View tunnel logs"
echo "   cloudflared tunnel list         # List all tunnels"
