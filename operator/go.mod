module github.com/IBM/ibm-spectrum-scale-csi/operator

go 1.16

require (
	//github.com/IBM/ibm-spectrum-scale-csi/driver/v2 v2.7.0
	github.com/IBM/ibm-spectrum-scale-csi/driver v2.7.0+incompatible
	github.com/google/uuid v1.3.0
	github.com/imdario/mergo v0.3.12
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.17.0
	github.com/openshift/api v0.0.0-20220222102030-354aa98a475c
	github.com/presslabs/controller-util v0.3.0
	go.uber.org/zap v1.19.1
	k8s.io/api v0.23.4
	k8s.io/apimachinery v0.23.4
	k8s.io/client-go v0.23.4
	sigs.k8s.io/controller-runtime v0.11.1
)

replace github.com/IBM/ibm-spectrum-scale-csi/driver => ../driver
