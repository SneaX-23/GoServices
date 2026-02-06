module github.com/SneaX-23/GoServices/email-service

go 1.25.5

require (
	github.com/SneaX-23/GoServices/auth-service v0.0.0-20260205162443-38c82d36d9ac
	github.com/resend/resend-go/v3 v3.0.0
	github.com/segmentio/kafka-go v0.4.49
	google.golang.org/grpc v1.78.0
)

require (
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.39.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

replace github.com/SneaX-23/GoServices/auth-service => ../auth-service
