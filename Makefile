# Load .env file
include .env
export

migrationPath=./migrations

db-status:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DATABASE_URL) goose -dir=$(migrationPath) status

up:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DATABASE_URL) goose -dir=$(migrationPath) up

down:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DATABASE_URL) goose -dir=$(migrationPath) down

reset:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DATABASE_URL) goose -dir=$(migrationPath) reset

create:
	@read -p "Enter migration name: " name; \
	GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DATABASE_URL) goose -dir=$(migrationPath) create $$name sql