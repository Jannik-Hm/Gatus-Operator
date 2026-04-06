#!/bin/bash

echo "Bootstrapping Cluster"

context=$1
echo "Using kubecontext \"$context\""

# Install cert manager
echo "Installing Cert Manager"
kubectl apply --context $context -f https://github.com/cert-manager/cert-manager/releases/download/v1.20.0/cert-manager.yaml
