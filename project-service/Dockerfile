# Dockerfile untuk dokumen-service
FROM golang:1.23.2 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o project-service .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/project-service .

CMD ["./project-service"]

