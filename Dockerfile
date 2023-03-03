# kodacd/aws-elasticache-redis-tester:latest
# docker buildx build --platform linux/amd64,linux/arm64 --push -t kodacd/aws-elasticache-redis-tester:latest . 
FROM alpine:latest as certs

RUN apk --update add ca-certificates

FROM golang:1.19 as builder

WORKDIR /app

COPY go.mod .

RUN go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -o /app/main

FROM nicolaka/netshoot

RUN apk update && apk add redis

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY --from=builder /app/main /opt/main

ENTRYPOINT [ "/opt/main" ]

