# Build the manager binary
FROM golang:1.17 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

#ENV CR_YAML=/usr/local/etc/csi.ibm.com_csiscaleoperators_cr.yaml

# Copy the go source
COPY main.go main.go
COPY hacks/health_check.go health_check.go
COPY api/ api/
COPY controllers/ controllers/
#COPY ./config/crd/bases/csi.ibm.com_csiscaleoperators_cr.yaml ${CR_YAML}
#RUN chmod 0444 ${CR_YAML}

# Build dummy health checker
RUN CGO_ENABLED=0 GOOS=linux go build -a -o health_check.sh health_check.go
# Build CSI Operator
RUN CGO_ENABLED=0 GOOS=linux go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder --chmod=0771 /workspace/health_check.sh .
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]

