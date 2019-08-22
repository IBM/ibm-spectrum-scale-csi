#!/bin/bash

eval $(minikube docker-env)
operator-sdk build csi-scale-operator

cd ../../FSaaS/csi-scale
make build-image
cd -
