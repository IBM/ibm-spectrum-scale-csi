module github.com/IBM/ibm-spectrum-scale-csi/driver/tests

go 1.13

require (
	github.com/IBM/ibm-spectrum-scale-csi-driver v1.0.0
	github.com/IBM/ibm-spectrum-scale-csi/driver v0.0.0
	github.com/kubernetes-csi/csi-test/v3 v3.1.0
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	golang.org/x/sys v0.0.0-20200413165638-669c56c373c4 // indirect
	golang.org/x/tools v0.0.0-20200318150045-ba25ddc85566 // indirect
	golang.org/x/tools/gopls v0.1.8-0.20200318150045-ba25ddc85566 // indirect
	k8s.io/api v1.17.4
	k8s.io/apimachinery v1.17.4
	k8s.io/kubernetes v1.17.4
)

replace (
	github.com/IBM/ibm-spectrum-scale-csi/driver => ../

	//pin to v1.17.4
	k8s.io/api => k8s.io/kubernetes/staging/src/k8s.io/api v0.0.0-20200312205431-8d8aa3959853
	k8s.io/apiextensions-apiserver => k8s.io/kubernetes/staging/src/k8s.io/apiextensions-apiserver v0.0.0-20200312205431-8d8aa3959853
	k8s.io/apimachinery => k8s.io/kubernetes/staging/src/k8s.io/apimachinery v0.0.0-20200312205431-8d8aa3959853
	k8s.io/apiserver => k8s.io/kubernetes/staging/src/k8s.io/apiserver v0.0.0-20200312205431-8d8aa3959853
	k8s.io/cli-runtime => k8s.io/kubernetes/staging/src/k8s.io/cli-runtime v0.0.0-20200312205431-8d8aa3959853
	k8s.io/client-go => k8s.io/kubernetes/staging/src/k8s.io/client-go v0.0.0-20200312205431-8d8aa3959853
	k8s.io/cloud-provider => k8s.io/kubernetes/staging/src/k8s.io/cloud-provider v0.0.0-20200312205431-8d8aa3959853
	k8s.io/cluster-bootstrap => k8s.io/kubernetes/staging/src/k8s.io/cluster-bootstrap v0.0.0-20200312205431-8d8aa3959853
	k8s.io/code-generator => k8s.io/kubernetes/staging/src/k8s.io/code-generator v0.0.0-20200312205431-8d8aa3959853
	k8s.io/component-base => k8s.io/kubernetes/staging/src/k8s.io/component-base v0.0.0-20200312205431-8d8aa3959853
	k8s.io/cri-api => k8s.io/kubernetes/staging/src/k8s.io/cri-api v0.0.0-20200312205431-8d8aa3959853
	k8s.io/csi-translation-lib => k8s.io/kubernetes/staging/src/k8s.io/csi-translation-lib v0.0.0-20200312205431-8d8aa3959853
	k8s.io/kube-aggregator => k8s.io/kubernetes/staging/src/k8s.io/kube-aggregator v0.0.0-20200312205431-8d8aa3959853
	k8s.io/kube-controller-manager => k8s.io/kubernetes/staging/src/k8s.io/kube-controller-manager v0.0.0-20200312205431-8d8aa3959853
	k8s.io/kube-proxy => k8s.io/kubernetes/staging/src/k8s.io/kube-proxy v0.0.0-20200312205431-8d8aa3959853
	k8s.io/kube-scheduler => k8s.io/kubernetes/staging/src/k8s.io/kube-scheduler v0.0.0-20200312205431-8d8aa3959853
	k8s.io/kubectl => k8s.io/kubernetes/staging/src/k8s.io/kubectl v0.0.0-20200312205431-8d8aa3959853
	k8s.io/kubelet => k8s.io/kubernetes/staging/src/k8s.io/kubelet v0.0.0-20200312205431-8d8aa3959853
	k8s.io/legacy-cloud-providers => k8s.io/kubernetes/staging/src/k8s.io/legacy-cloud-providers v0.0.0-20200312205431-8d8aa3959853
	k8s.io/metrics => k8s.io/kubernetes/staging/src/k8s.io/metrics v0.0.0-20200312205431-8d8aa3959853
	k8s.io/node-api => k8s.io/kubernetes/staging/src/k8s.io/node-api v0.0.0-20200312205431-8d8aa3959853
	k8s.io/sample-apiserver => k8s.io/kubernetes/staging/src/k8s.io/sample-apiserver v0.0.0-20200312205431-8d8aa3959853
	k8s.io/sample-cli-plugin => k8s.io/kubernetes/staging/src/k8s.io/sample-cli-plugin v0.0.0-20200312205431-8d8aa3959853
	k8s.io/sample-controller => k8s.io/kubernetes/staging/src/k8s.io/sample-controller v0.0.0-20200312205431-8d8aa3959853
)
