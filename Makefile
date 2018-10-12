# Copyright 2018 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

.PHONY: all gpfsplugin

GPFS_IMAGE_NAME=faas-registry.sl.cloud9.ibm.com:5000/gpfs-csi/gpfsplugin
GPFS_IMAGE_VERSION=v0.1

all: gpfsplugin

gpfsplugin:
	if [ ! -d ./vendor ]; then dep ensure; fi
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o  _output/gpfsplugin ./gpfs

image-gpfsplugin: gpfsplugin
	cp _output/gpfsplugin  deploy/gpfs/docker
	docker build -t $(GPFS_IMAGE_NAME):$(GPFS_IMAGE_VERSION) deploy/gpfs/docker

image-push-gpfsplugin: gpfsplugin image-gpfsplugin
	docker push $(GPFS_IMAGE_NAME):$(GPFS_IMAGE_VERSION)
