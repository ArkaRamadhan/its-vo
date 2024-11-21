# Dockerfile untuk dokumen-service
FROM golang:1.23.2 AS builder

RUN apt-get update && apt-get install -y tzdata curl
ENV TZ=Asia/Jakarta

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -o weeklyTimeline-service .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/weeklyTimeline-service .

CMD ["./weeklyTimeline-service"]

