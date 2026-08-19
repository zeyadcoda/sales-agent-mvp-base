.PHONY: check backend-test api infra-up infra-down

check:
	./scripts/check-toolchain.sh

backend-test:
	cd backend && GOTOOLCHAIN=local go test ./...

api:
	@test -f .env || (echo "ERROR: .env is missing. Copy .env.example to .env first." && exit 1)
	@set -a; \
	. ./.env; \
	set +a; \
	cd backend && go run ./cmd/api

infra-up:
	docker compose -f infra/compose/docker-compose.yml up -d

infra-down:
	docker compose -f infra/compose/docker-compose.yml down
