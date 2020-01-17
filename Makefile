<<<<<<< HEAD
###############################################################################
# Licensed Materials - Property of IBM.
# Copyright IBM Corporation 2017. All Rights Reserved.
# U.S. Government Users Restricted Rights - Use, duplication or disclosure 
# restricted by GSA ADP Schedule Contract with IBM Corp.
#
# Contributors:
#  IBM Corporation - initial API and implementation
###############################################################################
#
# TODO - Merge - Operator Section Below
#
SHELL = /bin/bash
STABLE_BUILD_DIR = repo/stable
STABLE_REPO_URL ?= https://raw.githubusercontent.com/IBM/charts/master/repo/stable/
STABLE_CHARTS := $(wildcard stable/*)
BUNDLE_HELM_PACKAGE = $(shell find $@/charts -maxdepth 2 -type f -name Chart.yaml | sed 's/Chart.yaml//' | xargs -i helm package {} -d $(STABLE_BUILD_DIR))

.DEFAULT_GOAL=all

$(STABLE_BUILD_DIR):
	@mkdir -p $@

.PHONY: charts charts-stable $(STABLE_CHARTS) 

# Default aliases: charts, repo

charts: charts-stable

repo: repo-stable

charts-stable: $(STABLE_CHARTS)
$(STABLE_CHARTS): $(STABLE_BUILD_DIR)
	cv lint ibmcase-bundle $@
	@echo $(BUNDLE_HELM_PACKAGE)

.PHONY: repo repo-stable repo-incubating 

repo-stable: $(STABLE_CHARTS) $(STABLE_BUILD_DIR)
	helm repo index $(STABLE_BUILD_DIR) --url $(STABLE_REPO_URL)

.PHONY: all
all: repo-stable

#
# TODO - Merge - Driver Section Below
#
NAME=ibm-spectrum-scale-csi

.PHONY: all $NAME

IMAGE_VERSION=v1.0.1
IMAGE_NAME=$(NAME)

all: $NAME

$NAME:
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o  _output/$(NAME) ./cmd/ibm-spectrum-scale-csi

build-image: $NAME
	docker build --network=host -t $(IMAGE_NAME):$(IMAGE_VERSION) .

save-image: build-image
	docker save $(IMAGE_NAME):$(IMAGE_VERSION) -o _output/$(IMAGE_NAME)_$(IMAGE_VERSION).tar
