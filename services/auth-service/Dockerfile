# Build from repo root so "replace ecommerce-backend-microservice => ../.." resolves.
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy full repo (context must be repo root)
COPY . .

WORKDIR /app/services/auth-service
RUN go mod download
# Static binary so it runs on Alpine (no glibc/musl mismatch)
ENV CGO_ENABLED=0
RUN go build -o /auth-service ./cmd

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /auth-service /app/auth-service

EXPOSE 8080

CMD ["/app/auth-service"]
