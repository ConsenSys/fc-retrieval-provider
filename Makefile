PORT ?= 9030
VERSION?=dev
IMAGE?=consensys/fc-retrieval-register

.PHONY: build build-dev start start-dev stop

default: clean build tag

build:
	docker build -t ${IMAGE}:${VERSION} .

build-dev:
	go build -v cmd/provider/main.go

push:
	cd scripts; bash push.sh ${VERSION} ${IMAGE}:${VERSION}

tag:
	cd scripts; bash tag.sh ${VERSION} ${IMAGE}:${VERSION}

clean:
	docker rm -f ${IMAGE}:${VERSION} 2> /dev/null || true
	docker rmi -f "${IMAGE}:${VERSION}" || true

start:
	docker-compose up -d

# User `make dev arg=--build` to rebuild
dev:
	docker-compose -f docker-compose.dev.yml up $(arg)

stop:
	docker-compose stop

utest:
	go test ./...

uselocal:
	cd scripts; bash use-local-repos.sh

useremote:
	cd scripts; bash use-remote-repos.sh