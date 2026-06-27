.PHONY: up down restart logs k6-run setup

up:
	docker compose up -d --build

down:
	docker compose down

restart:
	docker compose restart app

logs:
	docker compose logs -f app

k6-run:
	docker compose run --rm k6 run /scripts/scenario.js \
		--out influxdb=http://influxdb:8086/k6

setup:
	cd app && go mod tidy
