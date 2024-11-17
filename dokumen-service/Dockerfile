# Dockerfile untuk dokumen-service
FROM golang:1.23.2 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o dokumen-service .

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/dokumen-service .

CMD ["./dokumen-service"]