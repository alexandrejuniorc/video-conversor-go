# Description: Makefile for workerpool

## Docker
docker/dev/start:
	@docker compose up -d

docker/dev/stop:
	@docker compose down

docker/dev/restart:
	@docker compose down
	@docker compose up -d

docker/dev/clean:
	@docker compose down --rmi all --volumes

## Development
start/dev:
	@go run main.go