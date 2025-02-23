ARG GOLANG_VERSION=1.23

# Build container should always be on the native build platform.
FROM --platform=$BUILDPLATFORM golang:${GOLANG_VERSION} AS builder

WORKDIR /src

ARG TARGETOS TARGETARCH

RUN --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x

# Mount the source code into the build container and build the binary.
RUN --mount=type=bind,target=. \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=$TARGETARCH \
    go build -v -o /go/bin/manager cmd/main.go

FROM scratch

USER 65532:65532

WORKDIR /workspace

COPY --from=builder /go/bin/manager /workspace/manager

ENTRYPOINT ["/workspace/manager"]
