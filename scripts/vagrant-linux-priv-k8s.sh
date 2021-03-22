#!/usr/bin/env bash

# Install minikube
curl -Lo minikube https://storage.googleapis.com/minikube/releases/v1.18.1/minikube-linux-amd64 \
  && chmod +x minikube \
  && mv minikube /usr/local/bin

# Install kubectl
curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/v1.20.2/bin/linux/amd64/kubectl \
  && chmod +x kubectl \
  && mv kubectl /usr/local/bin
