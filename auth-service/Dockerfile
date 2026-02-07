FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY auth-service/go.mod auth-service/go.sum ./
RUN go mod download

COPY auth-service/ .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o auth-service ./cmd/auth-service/main.go

FROM alpine:latest AS run

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/auth-service /app/auth-service

EXPOSE 8080 
EXPOSE 50051

CMD ["/app/auth-service"]
