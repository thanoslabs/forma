DATABASE_URL ?= postgres://forma:forma_dev_only@localhost:55434/forma?sslmode=disable

.PHONY: dev-up dev-down run test migrate-up
dev-up:
	docker compose up -d postgres
dev-down:
	docker compose down
run:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/forma
test:
	go test ./...
migrate-up:
	@for f in db/migrations/*.up.sql; do psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -f "$$f" || exit 1; done
