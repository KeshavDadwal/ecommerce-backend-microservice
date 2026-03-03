FROM golang:1.22-alpine AS builder

WORKDIR /app

# Only go.mod initially to leverage Docker cache
COPY go.mod ./
RUN go mod download

# Copy the rest of the source
COPY . .

WORKDIR /app/cmd
RUN go build -o /auth-service

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /auth-service /app/auth-service

EXPOSE 8080

CMD ["/app/auth-service"]
