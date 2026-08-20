.PHONY: check backend-test backend-vet frontend-install frontend-typecheck frontend-test frontend-build api web migrate-up migrate-status migrate-down bootstrap-super-admin infra-up infra-down mailpit-up mailpit-logs mailpit-test

check:
	./scripts/check-toolchain.sh

backend-test:
	cd backend && GOTOOLCHAIN=local go test ./...

backend-vet:
	cd backend && GOTOOLCHAIN=local go vet ./...

frontend-install:
	pnpm install

frontend-typecheck:
	pnpm --filter @sales-agent/admin-web typecheck

frontend-test:
	pnpm --filter @sales-agent/admin-web test

frontend-build:
	pnpm --filter @sales-agent/admin-web build

api:
	@test -f .env || (echo "ERROR: .env is missing. Copy .env.example to .env first." && exit 1)
	@set -a; \
	. ./.env; \
	set +a; \
	cd backend && go run ./cmd/api

web:
	@test -f .env || (echo "ERROR: .env is missing. Copy .env.example to .env first." && exit 1)
	@set -a; \
	. ./.env; \
	set +a; \
	pnpm --filter @sales-agent/admin-web dev

migrate-up:
	@test -f .env || (echo "ERROR: .env is missing. Copy .env.example to .env first." && exit 1)
	@set -a; \
	. ./.env; \
	set +a; \
	cd backend && go run ./cmd/migrate up

migrate-status:
	@test -f .env || (echo "ERROR: .env is missing. Copy .env.example to .env first." && exit 1)
	@set -a; \
	. ./.env; \
	set +a; \
	cd backend && go run ./cmd/migrate status

# Local rollback aid only. The migration command also enforces APP_ENV=local
# and a loopback PostgreSQL host before it opens a database connection.
migrate-down:
	@test -f .env || (echo "ERROR: .env is missing. Copy .env.example to .env first." && exit 1)
	@set -a; \
	. ./.env; \
	set +a; \
	cd backend && go run ./cmd/migrate down

bootstrap-super-admin:
	@test -f .env || (echo "ERROR: .env is missing. Copy .env.example to .env first." && exit 1)
	@test -n "$(EMAIL)" || (echo "ERROR: EMAIL is required." && exit 1)
	@set -a; \
	. ./.env; \
	set +a; \
	cd backend && go run ./cmd/bootstrap-super-admin --email "$(EMAIL)" --name "$(if $(NAME),$(NAME),Super Admin)"

infra-up:
	docker compose -f infra/compose/docker-compose.yml up -d

infra-down:
	docker compose -f infra/compose/docker-compose.yml down

mailpit-up:
	docker compose -f infra/compose/docker-compose.yml up -d mailpit

mailpit-logs:
	docker compose -f infra/compose/docker-compose.yml logs -f mailpit

mailpit-test:
	cd backend && TEST_MAILPIT=1 GOTOOLCHAIN=local go test ./internal/notification/email -run TestMailpitIntegration -count=1 -v
