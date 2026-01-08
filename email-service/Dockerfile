FROM golang:1.23-alpine AS build

WORKDIR /app

# RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /myapp .

FROM alpine:latest AS run

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=build /myapp /myapp

COPY --from=build /app/internal/templates/emails ./internal/templates/emails

EXPOSE 8080

CMD ["/myapp"]