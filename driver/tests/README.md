
[//]: # (links)
[csi_sanity]: (https://github.com/kubernetes-csi/csi-test/tree/master/pkg/sanity)
[kubernetes e2e storage testsuite]: (https://github.com/kubernetes/kubernetes/tree/master/test/e2e/storage/testsuites)
[KUBECONFIG]: (https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/#the-kubeconfig-environment-variable)
[scale_csi_prereqs]: (../#pre-requisites-for-installing-and-running-the-csi-driver)

## End-to-end
Run [kubernetes e2e storage testsuite][kubernetes e2e storage testsuite] against CSI already deployed in a kubernetes cluster.

#### Prerequisites
* **kubernetes 1.14**+ cluster with [**ibm-spectrum-scale-csi-driver** deployed][scale_csi_prereqs]
* [**KUBECONFIG**][KUBECONFIG] set on machine that orchestrates tests

```bash
go test -v ./e2e -timeout=0 -ginkgo.focus="\[scale-csi\]" -report-dir=$HOME/
```

## Integration
Run [csi-sanity][csi_sanity] against a configured Scale cluster.

#### Prerequisites
* a [configured Scale cluster][scale_csi_prereqs]
* set **SCALE_GUI**, **SCALE_USER**, and **SCALE_PASSWORD** environment variables
* or optionally, set **SCALE_CONFIG** env and have all your secrets in place as files (see configMap.go)
* set **JUNIT_FILE** env to output to a junit xml report
```bash
SCALE_GUI=https://example.scale \
SCALE_USER==**** \
SCALE_PASSWORD=**** \
go test ./integration
```

## Mock Sanity
Run [csi-sanity][csi_sanity] with driver running in mock mode.

* optionally, set **JUNIT_FILE** env to output to a junit xml report
```bash
go test ./sanity
```