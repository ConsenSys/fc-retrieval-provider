PORT ?= 9030
VERSION?=dev
IMAGE?=consensys/fc-retrieval-provider

.PHONY: build build-dev start start-dev stop

default: clean build tag

build:
	docker build -t $(IMAGE):$(VERSION) .

build-dev:
	go build -v cmd/provider/main.go

push:
	cd scripts; bash push.sh $(VERSION) $(IMAGE):$(VERSION)

tag:
	cd scripts; bash tag.sh $(VERSION) $(IMAGE):$(VERSION)

clean:
	docker rm -f $(IMAGE):$(VERSION) 2> /dev/null || true
	docker rmi -f "$(IMAGE):$(VERSION)" || true

start:
	docker-compose up -d

# User `make dev arg=--build` to rebuild
dev:
	make -s env-soft
	docker-compose -f docker-compose.dev.yml up $(arg)

stop:
	docker-compose stop

utest:
	go test ./...

uselocal:
	cd scripts; bash use-local-repos.sh

useremote:
	cd scripts; bash use-remote-repos.sh

env-soft:
	make env-compare
	if [ -s .env ]; then \
		echo ".env already exists"; \
	else \
		echo ".env does not exists, create a .env file"; \
		cp .env.example .env; \
	fi

env-compare:
	@cmp -s .env .env.example; \
	RETVAL=$$?; \
	if [ $$RETVAL -eq 0 ]; then \
		echo ".env and .env.example are the same"; \
	else \
		echo ".env and .env.example are NOT the same"; \
	fi