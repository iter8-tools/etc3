[![Go Report Card](https://goreportcard.com/badge/github.com/iter8-tools/etc3)](https://goreportcard.com/report/github.com/kalantar/etc3)
[![Coverage](https://codecov.io/gh/iter8-tools/default-tasks/branch/main/graphs/badge.svg?branch=main)](https://codecov.io/gh/kalantar/etc3)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Reference](https://pkg.go.dev/badge/github.com/iter8-tools/default-tasks.svg)](https://pkg.go.dev/github.com/kalantar/etc3)
# etc3: Extensible Thin Controller with Composable CRD

> The etc3 controller provides core capabilities to orchestrate iter8 experiments across different Kubernetes and Openshift stacks.

## Developers

This section is for iter8 developers and contains documentation on running and testing the etc3 controller locally.

### Install KFServing and iter8-kfserving Domain Package
Pre-requisites: `kubectl` with acccess to a kubernetes cluster.

To install KFServing and the iter8-kfserving domain package, follow Steps 1 and 2 from [here](https://github.com/iter8-tools/iter8-kfserving#quick-start-on-minikube).

### Partial Install of iter8-kfserving Domain
For dev/local-test purposes, it is convenient to run the etc3 locally. Follow the above instructions for iter8-kfserving installation, and then delete the etc3 controller as follows.

```
kubectl delete deployment iter8-controller-manager -n iter8-system
```

### Port-forward iter8-analytics
*In a separate terminal:*

```
kubectl port-forward -n iter8-system svc/iter8-analytics 8080:8080
```

You should now be able to access the iter8-analytics service using the OpenAPI UI at http://localhost:8080/docs

### Run etc3 locally
```
make manager
export ITER8_NAMESPACE=iter8-system
export ITER8_ANALYTICS_ENDPOINT=http://127.0.0.1:8080/v2/analytics_results
export DEFAULTS_DIR=../iter8-kfserving/install/iter8-controller/configmaps/defaults
export HANDLERS_DIR=../iter8-kfserving/install/iter8-controller/configmaps/handlers
bin/manager
``` 

### Test etc3
```
make test
```
