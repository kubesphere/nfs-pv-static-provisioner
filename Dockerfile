# Build the manager binary
FROM golang:1.22 as builder

WORKDIR /workspace
ADD . /workspace/

RUN make -f build.mk build-local

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM kubesphere/distroless-static:nonroot
WORKDIR /
COPY --from=builder /workspace/bin/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]