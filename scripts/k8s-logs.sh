#!/bin/bash

echo "📝 Tailing logs from all asocial pods..."
echo "   (Press Ctrl+C to stop)"
echo ""

kubectl logs -n asocial -l app=backend --all-containers=true --follow --tail=50 &
kubectl logs -n asocial -l app=frontend --all-containers=true --follow --tail=50 &
kubectl logs -n asocial -l app=redis --all-containers=true --follow --tail=50 &

wait
