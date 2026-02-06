FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod ./
# COPY go.sum ./ # Not created yet, will be created by 'go mod tidy'
RUN go mod download || true

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /sing-box-exporter ./cmd/exporter

FROM scratch
COPY --from=builder /sing-box-exporter /sing-box-exporter
ENTRYPOINT ["/sing-box-exporter"]
