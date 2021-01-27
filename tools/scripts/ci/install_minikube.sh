#!/usr/bin/env bash
set -e
set -u

# Install minikube for CI

#. ./ci-env 

# Install kubectl for minikube
curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/${KUBE_VERSION}/bin/linux/amd64/kubectl
chmod +x kubectl
sudo mv kubectl /usr/local/bin/

# Install minikube
curl -Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
chmod +x minikube
sudo mv minikube /usr/local/bin/

# Setup configuration
mkdir -p $HOME/.kube $HOME/.minikube
touch $KUBECONFIG

echo $HOME

# Start minikube and chown for minikub.
sudo chown -R travis: /home/travis/.minikube/
sudo minikube start --driver=none --kubernetes-version=${KUBE_VERSION} --cpus=2
minikube update-context
eval $(minikube docker-env)

