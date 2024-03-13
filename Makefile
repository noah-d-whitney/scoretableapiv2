run:
	go run ./cmd/api

psql:
	"/Applications/Postgres.app/Contents/Versions/16/bin/psql" -p5432 "scoretable"

up:
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${SCORETABLE_DSN} up