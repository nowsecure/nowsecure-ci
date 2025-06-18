FROM --platform=$BUILDPLATFORM golang:alpine AS builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download
COPY . .
RUN --mount=type=cache,target="/root/.cache/go-build" CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-w -s" -o ns .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/ns .
ENTRYPOINT [ "/app/ns" ]
