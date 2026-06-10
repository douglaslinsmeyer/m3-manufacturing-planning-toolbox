FROM golang:1.26-alpine AS builder

WORKDIR /app

# Cache dependencies layer
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with size optimizations
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /app/bin/server \
    ./cmd/server

# Use distroless for minimal image size; :nonroot runs as uid 65532
# (matches the pod securityContext in k8s/backend.yaml)
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/migrations /app/migrations

EXPOSE 8080

CMD ["/app/server"]
