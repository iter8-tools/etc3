#!/bin/bash

# Install iter8 2.0 in one line; both the controller and the analytics engine.
#set -x

install() {

    # Install iter8 2.0 CRDs
    kubectl apply -k https://github.com/iter8-tools/etc3/config/crd/?ref=main

    # Install iter8 2.0 controller (following line is from iter8 1.0.0)
    # kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8-controller/v1.0.0/install/iter8-controller-telemetry-v2-17.yaml
  
    # Install iter8 analytics (following line is from iter8 1.0.0)
    # kubectl apply -f https://raw.githubusercontent.com/iter8-tools/iter8-analytics/master/install/kubernetes/iter8-analytics.yaml
}

install