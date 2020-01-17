#!/usr/bin/env bash
set -e
set -u

#. ./ci-env

./install_minikube.sh  
./install_operator-sdk.sh
