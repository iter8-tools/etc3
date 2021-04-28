[![Go Report Card](https://goreportcard.com/badge/github.com/iter8-tools/etc3)](https://goreportcard.com/report/github.com/iter8-tools/etc3)
[![Coverage](https://codecov.io/gh/iter8-tools/etc3/branch/main/graphs/badge.svg?branch=main)](https://codecov.io/gh/iter8-tools/etc3)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Reference](https://pkg.go.dev/badge/github.com/iter8-tools/default-tasks.svg)](https://pkg.go.dev/github.com/iter8-tools/etc3)

# etc3: Extensible Thin Controller with Composable CRD

> The etc3 controller provides core capabilities to orchestrate iter8 experiments across different Kubernetes and Openshift stacks.

## Developers

This section is for iter8 developers and contains documentation on running and testing the etc3 controller locally.

### Install Iter8

To install Iter8 see [here](https://iter8.tools).

### Delete etc3 from Cluster
After installing Iter8, delete the etc3 controller:

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
export HANDLERS_DIR=../iter8-install/core/iter8-handler
bin/manager
``` 

### Testing etc3

```
make test
```
