
DOCKER_COMPOSE = docker compose

.PHONY: up, down, build, run

up:
	$(DOCKER_COMPOSE) --env-file .env.docker  up -d

down:
	$(DOCKER_COMPOSE) --env-file .env.docker down

build:
	./build.sh

run:
	go run ./bot

format:
	gofmt -w -s .

run-skip:
	go run ./bot -skip
	
vet:
	go vet ./...

lint:
	golangci-lint run ./...