#!/bin/bash

echo "📝 Tailing logs from all asocial pods..."
echo "   (Press Ctrl+C to stop)"
echo ""

kubectl logs -n asocial -l app=backend --all-containers=true --follow --tail=50 --max-log-requests=10 &
kubectl logs -n asocial -l app=frontend --all-containers=true --follow --tail=50 --max-log-requests=10 &
kubectl logs -n asocial -l app=redis --all-containers=true --follow --tail=50 --max-log-requests=10 &
kubectl logs -n asocial -l app=postgres --all-containers=true --follow --tail=50 --max-log-requests=10 &

wait
