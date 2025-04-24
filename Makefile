run:
	docker compose down
	docker compose up -d --build --remove-orphans

stop:
	docker compose down

logs:
	docker compose logs -f --tail=100

enter-app:
	docker compose exec app sh

format:
	docker compose exec app sh -c "go fmt ./... && go vet ./..."

.PHONY: run stop logs enter-app format
