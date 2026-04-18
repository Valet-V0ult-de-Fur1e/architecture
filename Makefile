COMPOSE_BASE := docker compose -f docker/docker-compose.yml
COMPOSE_DEBUG := $(COMPOSE_BASE) -f docker/docker-compose.debug.yml
COMPOSE_PROD := $(COMPOSE_BASE) -f docker/docker-compose.prod.yml

.PHONY: up up-debug up-prod down down-debug down-prod restart restart-debug restart-prod logs logs-debug logs-prod ps ps-debug ps-prod config config-debug config-prod pull pull-debug pull-prod build build-debug build-prod test-api redis-cli redis-lab-ttl-demo gotest

up: up-debug

up-debug:
	$(COMPOSE_DEBUG) up --build -d

up-prod:
	$(COMPOSE_PROD) up --build -d

down: down-debug

down-debug:
	$(COMPOSE_DEBUG) down

down-prod:
	$(COMPOSE_PROD) down

restart: restart-debug

restart-debug:
	$(COMPOSE_DEBUG) down
	$(COMPOSE_DEBUG) up --build -d

restart-prod:
	$(COMPOSE_PROD) down
	$(COMPOSE_PROD) up --build -d

logs: logs-debug

logs-debug:
	$(COMPOSE_DEBUG) logs -f --tail=200

logs-prod:
	$(COMPOSE_PROD) logs -f --tail=200

ps: ps-debug

ps-debug:
	$(COMPOSE_DEBUG) ps

ps-prod:
	$(COMPOSE_PROD) ps

config: config-debug

config-debug:
	$(COMPOSE_DEBUG) config

config-prod:
	$(COMPOSE_PROD) config

pull: pull-debug

pull-debug:
	$(COMPOSE_DEBUG) pull

pull-prod:
	$(COMPOSE_PROD) pull

build: build-debug

build-debug:
	$(COMPOSE_DEBUG) build

build-prod:
	$(COMPOSE_PROD) build

test-api:
	python tests/run_e2e.py

redis-cli:
	$(COMPOSE_DEBUG) exec redis redis-cli


gotest:
	cd backend && go test ./...

redis-lab-ttl-demo:
	$(COMPOSE_DEBUG) exec -T redis redis-cli SET lab:ttl:key demo
	$(COMPOSE_DEBUG) exec -T redis redis-cli EXPIRE lab:ttl:key 120
	$(COMPOSE_DEBUG) exec -T redis redis-cli TTL lab:ttl:key
	$(COMPOSE_DEBUG) exec -T redis redis-cli PERSIST lab:ttl:key
	$(COMPOSE_DEBUG) exec -T redis redis-cli TTL lab:ttl:key

# Демонстрация RabbitMQ routing (fanout/direct/topic/headers)
rabbitmq-demo:
	cd backend && go run internal/platform/rabbitmq/demo.go

