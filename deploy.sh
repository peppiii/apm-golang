#!/usr/bin/env bash

set -e
set -o pipefail

if [ $# -eq 0 ]
  then
    echo "Usage: deploy.sh [version]"
    exit 1
fi

cat k8s/deployment.yaml | sed 's/\$BUILD_NUMBER'"/$1/g" | sed 's/\$NAMESPACE'"/$2/g" | kubectl apply -n $3 -f - --kubeconfig=kubeconfig.conf
kubectl apply -f k8s/service.yaml --kubeconfig=kubeconfig.conf
