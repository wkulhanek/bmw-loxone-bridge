# FROM golang:1.26-alpine AS builder
FROM --platform=$BUILDPLATFORM registry.access.redhat.com/hi/go:1.26-builder AS builder
ARG TARGETARCH

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -ldflags="-s -w" -o /bmw-loxone-bridge ./cmd/bridge

FROM gcr.io/distroless/static:nonroot

COPY --from=builder /bmw-loxone-bridge /bmw-loxone-bridge

USER nonroot:nonroot
EXPOSE 8400
VOLUME /data

ENTRYPOINT ["/bmw-loxone-bridge"]
