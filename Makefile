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

NAME=ibm-spectrum-scale-csi

.PHONY: all $NAME

IMAGE_VERSION=v1.0.0
IMAGE_NAME=$(NAME)

all: $NAME

$NAME:
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o  _output/$(NAME) ./cmd/ibm-spectrum-scale-csi

build-image: $NAME
	docker build --network=host -t $(IMAGE_NAME):$(IMAGE_VERSION) .

save-image: build-image
	docker save $(IMAGE_NAME):$(IMAGE_VERSION) -o _output/$(IMAGE_NAME)_$(IMAGE_VERSION).tar
