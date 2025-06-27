FROM --platform=$BUILDPLATFORM golang:alpine AS builder
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dirty
WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download
COPY . .
RUN apk --update add ca-certificates
RUN --mount=type=cache,target="/root/.cache/go-build" \
    CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    go build -ldflags="-w -s -X 'github.com/nowsecure/nowsecure-ci/cmd/ns/version.version ${VERSION}'" -o ns .

FROM alpine:3
WORKDIR /app
COPY --from=builder /app/ns .
ENTRYPOINT [ "/app/ns" ]
