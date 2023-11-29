File: Makefile

confirm:
	@echo 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

run/api:
	go run ./cmd/api

db/migrations/new:
	@echo "Creating migration files for ${name}"
	migrate create -seq -ext=.sql -dir=./migrations ${name}

db/migrations/up: confirm
	@echo "Running Migrations..."
	migrate -path ./migrations -database ${postgresConnString} up