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

NAME=csi-gpfs

.PHONY: all $NAME

IMAGE_NAME=faas-registry.sl.cloud9.ibm.com:5000/$(NAME)
IMAGE_VERSION=v0.1

all: $NAME

$NAME:
	if [ ! -d ./vendor ]; then dep ensure; fi
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o  _output/$(NAME) ./cmd/csi-gpfs

build-image: $NAME
	docker build -t $(IMAGE_NAME):$(IMAGE_VERSION) .

push-image: build-image
	docker push $(IMAGE_NAME):$(IMAGE_VERSION)
