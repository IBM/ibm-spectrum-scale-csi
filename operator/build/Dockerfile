# docker build for IBM Storage Scale CSI Operator
FROM --platform=$BUILDPLATFORM golang:1.20 as builder

ARG TARGETOS
ARG TARGETARCH
ARG commit

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY hacks/health_check.go health_check.go
COPY api/ api/
COPY controllers/ controllers/

# Set GOENV to target OS and ARCH
ENV REVISION $commit
ENV GOOS $TARGETOS
ENV GOARCH $TARGETARCH

# Build dummy health checker
RUN CGO_ENABLED=0 go build -ldflags="-X 'main.gitCommit=${REVISION}'" -a -o health_check.sh health_check.go
# Build CSI Operator
RUN CGO_ENABLED=0 go build -ldflags="-X 'main.gitCommit=${REVISION}'" -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
ARG version=2.10.0
ARG commit
ARG build_date

ENV version $version
ENV commit $commit
ENV build_date $build_date

LABEL name="IBM Storage Scale CSI Operator" \
      vendor="ibm" \
      version=$version \
      release="1" \
      summary="A Go based operator to run and manage the deployment of the IBM Storage Scale CSI Driver." \
      description="A Go based operator to run and manage the deployment of the IBM Storage Scale CSI Driver." \
      maintainers="IBM Storage Scale" \
      build-date=$build_date \
      vcs-ref=$commit \
      vcs-type="git" \
      url="https://www.ibm.com/docs/en/spectrum-scale-csi"

COPY licenses /licenses
COPY --from=builder --chmod=0771 /workspace/health_check.sh .
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]

