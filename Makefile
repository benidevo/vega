setup-hooks:
	git config core.hooksPath .githooks
	chmod +x .githooks/pre-commit

test:
	docker compose exec app sh -c "go test ./... -coverprofile=coverage.out -coverpkg=github.com/benidevo/prospector/internal/... && go tool cover -func=coverage.out | grep total:"

test-verbose:
	docker compose exec app sh -c "go test ./... -coverprofile=coverage.out -coverpkg=github.com/benidevo/prospector/internal/... && go tool cover -func=coverage.out"

run:
	docker compose down
	docker compose up -d --build --remove-orphans

run-it:
	docker compose down
	docker compose up --build --remove-orphans

stop:
	docker compose down

logs:
	docker compose logs -f --tail=100

enter-app:
	docker compose exec app sh

format:
	docker compose exec app sh -c "go fmt ./... && go vet ./..."

.PHONY: run stop logs enter-app format test test-verbose setup-hooks
