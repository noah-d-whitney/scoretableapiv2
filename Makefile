include .envrc

#==================================================================================================#
# HELPERS
#==================================================================================================#

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

#==================================================================================================#
# DEVELOPMENT
#==================================================================================================#

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	@go run ./cmd/api -db-dsn=${SCORETABLE_DSN}

## run/api/dev: run the cmd/api application with live reload
.PHONY: run/api/dev
run/api/dev:
	@echo 'Starting dev server'
	@air -c .air.toml

## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	"/Applications/Postgres.app/Contents/Versions/16/bin/psql" -p5432 "scoretable"

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${SCORETABLE_DSN} up

#==================================================================================================#
# QUALITY CONTROL
#==================================================================================================#

## audit: tidy dependencies and format, vet and test all code
.PHONY: audit
audit: vendor
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

## vendor: tidy and vendor dependencies
.PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vending dependencies...'
	go mod vendor

#==================================================================================================#
# BUILD
#==================================================================================================#

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	@echo ${SCORETABLE_DSN}
	go build -ldflags='-s' -o=./bin/api ./cmd/api