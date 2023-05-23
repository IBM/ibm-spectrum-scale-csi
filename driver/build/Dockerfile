# docker build for IBM Storage Scale CSI Driver
# usage: docker build -f build/Dockerfile -t my_image_tag .
# Multi-arch build for IBM Storage Scale CSI Driver
# usage: docker buildx build -f build/multi-arch.Dockerfile --platform=linux/amd64 -t my_image_tag .

FROM --platform=$BUILDPLATFORM golang:1.20 AS builder
WORKDIR /go/src/github.com/IBM/ibm-spectrum-scale-csi/driver/
COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

COPY . .
ARG TARGETOS
ARG TARGETARCH
ARG GOFLAGS
ARG commit

ENV REVISION $commit

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -a -ldflags "-X 'main.gitCommit=${REVISION}' -extldflags '-static'" -o _output/ibm-spectrum-scale-csi ./cmd/ibm-spectrum-scale-csi
# RUN chmod +x,u+s _output/ibm-spectrum-scale-csi

FROM registry.access.redhat.com/ubi8-minimal:latest
ARG version=2.10.0
ARG commit
ARG build_date

ENV version $version
ENV commit $commit
ENV build_date $build_date

LABEL name="IBM Storage Scale CSI driver" \
      vendor="ibm" \
      version=$version \
      release="1" \
      summary="An implementation of CSI Plugin for the IBM Storage Scale product." \
      description="An implementation of CSI Plugin for the IBM Storage Scale product." \
      maintainers="IBM Storage Scale" \
      build-date=$build_date \
      vcs-ref=$commit \
      vcs-type="git" \
      url="https://www.ibm.com/docs/en/spectrum-scale-csi"

COPY licenses /licenses
COPY --from=builder /go/src/github.com/IBM/ibm-spectrum-scale-csi/driver/_output/ibm-spectrum-scale-csi /ibm-spectrum-scale-csi

## non-root user
# RUN groupadd -g 100006 default
# RUN useradd -u 100006 -g default -r default
# USER default

# RUN microdnf update -y \
#     && microdnf install util-linux

RUN mkdir /chroot
COPY chroot-wrapper.sh /chroot
RUN chmod +x /chroot/chroot-wrapper.sh
RUN ln -s /chroot/chroot-wrapper.sh /chroot/mount
RUN ln -s /chroot/chroot-wrapper.sh /chroot/umount

ENV PATH="/chroot:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

ENTRYPOINT ["/ibm-spectrum-scale-csi"]
