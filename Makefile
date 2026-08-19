.PHONY: check backend-test api infra-up infra-down

check:
	./scripts/check-toolchain.sh

backend-test:
	cd backend && GOTOOLCHAIN=local go test ./...

api:
	cd backend && go run ./cmd/api

infra-up:
	docker compose -f infra/compose/docker-compose.yml up -d

infra-down:
	docker compose -f infra/compose/docker-compose.yml down
