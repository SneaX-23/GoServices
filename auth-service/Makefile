MIGRATION_DIR=internal/database/migrations
DB_DNS=

migration:
	goose -dir $(MIGRATION_DIR) create $(name) sql

migration-up:
	goose -dir $(MIGRATION_DIR) postgres $(DB_DNS) up 

migration-down:
	goose -dir $(MIGRATION_DIR) postgres $(DB_DNS) down

build_proto:
	@echo "Creating genproto directory..."
	mkdir -p internal/genproto
	@echo "Generating Go code from proto..."
	protoc --proto_path=proto \
		--go_out=internal/genproto --go_opt=paths=source_relative \
		--go-grpc_out=internal/genproto --go-grpc_opt=paths=source_relative \
		proto/auth.proto
