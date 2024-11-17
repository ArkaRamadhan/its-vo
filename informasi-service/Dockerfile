# Dockerfile untuk dokumen-service
FROM golang:1.23.2 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o informasi-service .

FROM alpine

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/informasi-service .

CMD ["./informasi-service"]

