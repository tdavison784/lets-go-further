File: Makefile

help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

confirm:
	@echo 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

## run/api: run the cmd/api application
run/api:
	go run ./cmd/api
## db/migrations/new name=$1: create a new database migration set
db/migrations/new:
	@echo "Creating migration files for ${name}"
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all database migrations
db/migrations/up: confirm
	@echo "Running Migrations..."
	migrate -path ./migrations -database ${postgresConnString} up