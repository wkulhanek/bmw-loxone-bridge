# FROM golang:1.26-alpine AS builder
FROM registry.access.redhat.com/hi/go:1.26-builder AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bmw-loxone-bridge ./cmd/bridge

FROM gcr.io/distroless/static:nonroot

COPY --from=builder /bmw-loxone-bridge /bmw-loxone-bridge

USER nonroot:nonroot
EXPOSE 8300
VOLUME /data

ENTRYPOINT ["/bmw-loxone-bridge"]
