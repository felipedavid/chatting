include .env
export $(shell sed 's/=.*//' .env)

dump_schema:
	pg_dump -Fp --schema-only $$DATABASE_URL > database/schema.out.sql

new_migration:
	migrate create -ext sql -dir database/migrations -seq new

run_migrations:
	migrate -database $$DATABASE_URL -path database/migrations up

down_migration:
	migrate -database $$DATABASE_URL -path database/migrations down 1

migrate: run_migrations generate_storage

downgrade: down_migration generate_storage

generate_storage:
	sqlc generate