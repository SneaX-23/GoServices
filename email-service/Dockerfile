FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY auth-service /auth-service

COPY email-service/go.mod email-service/go.sum ./
RUN go mod download

COPY email-service/ .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o email-service ./cmd/email-service/main.go

FROM alpine:latest AS run

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/email-service /app/email-service

COPY --from=builder /app/internal/templates/emails ./internal/templates/emails

EXPOSE 8080

CMD ["/app/email-service"]
