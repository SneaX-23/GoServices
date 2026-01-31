MIGRATION_DIR=internal/database/migrations
DB_DNS=

migration:
	goose -dir $(MIGRATION_DIR) create $(name) sql

migration-up:
	goose -dir $(MIGRATION_DIR) postgres $(DB_DNS) up 

migration-down:
	goose -dir $(MIGRATION_DIR) postgres $(DB_DNS) down
