# kodacd/aws-elasticache-redis-tester:latest
# docker buildx build . -t kodacd/aws-elasticache-redis-tester:latest

FROM golang:1.19 as builder

WORKDIR /app

COPY go.mod .

RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -o /app/main

# FROM scratch

# COPY --from=builder /app/main /main

ENTRYPOINT [ "/app/main" ]

