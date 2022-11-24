//TODO: This is temporary, change it back to IBM
module github.com/amdabhad/ibm-spectrum-scale-csi/driver

go 1.18

require (
	github.com/container-storage-interface/spec v1.5.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.3.2
	golang.org/x/net v0.0.0-20191028085509-fe3aa8a45271
	golang.org/x/sys v0.0.0-20220722155257-8c9f86f7a55f
	google.golang.org/grpc v1.26.0
	k8s.io/mount-utils v0.25.3
)

require (
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/moby/sys/mountinfo v0.6.0 // indirect
	golang.org/x/text v0.3.0 // indirect
	google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55 // indirect
	k8s.io/klog/v2 v2.70.1 // indirect
	k8s.io/utils v0.0.0-20220728103510-ee6ede2d64ed // indirect
)
