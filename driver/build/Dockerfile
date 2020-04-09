# Legacy docker build for IBM Spectrum Scale CSI Driver
# usage: docker build -f build/Dockerfile -t my_image_tag .

FROM golang:1.13.1 AS builder
WORKDIR /go/src/github.com/IBM/ibm-spectrum-scale-csi/driver/
COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

COPY . .
ARG GOFLAGS
ARG GIT_COMMIT
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o _output/ibm-spectrum-scale-csi ./cmd/ibm-spectrum-scale-csi
RUN chmod +x _output/ibm-spectrum-scale-csi


FROM registry.access.redhat.com/ubi8-minimal:latest
LABEL name="IBM Spectrum Scale CSI driver" \
      vendor="ibm" \
      version="1.1.0" \
      release="1" \
      run='docker run ibm-spectrum-scale-csi-driver' \
      summary="An implementation of CSI Plugin for the IBM Spectrum Scale product."\
      description="An implementation of CSI Plugin for the IBM Spectrum Scale product."\
      maintainers="IBM Spectrum Scale"\
      git_commit=$GIT_COMMIT

COPY licenses /licenses
COPY --from=builder /go/src/github.com/IBM/ibm-spectrum-scale-csi/driver/_output/ibm-spectrum-scale-csi /ibm-spectrum-scale-csi
ENTRYPOINT ["/ibm-spectrum-scale-csi"]
