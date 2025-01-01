ifneq (,$(wildcard ./.env))
    include .env
    export
endif

build:
	go build -o build/homecloud ./cmd/homecloud.go

run:
	docker run -it --rm \
		-w "/opt/homecloud" \
		-e "air_wd=/opt/homecloud" \
		-v $(shell pwd):/opt/homecloud \
		-v $(HOME)/go/pkg/mod:/go/pkg/mod \
		-v $(HOME)/go/pkg/cache:/root/.cache/go-build \
		-v $(HOST_DOCKER_SOCKET):/var/run/docker.sock \
		-p 1323:1323 \
		cosmtrek/air \
		-c ".air.toml"

test:
	docker run -it --rm \
		-w "/opt/homecloud" \
		-e "TEST_CONTAINER_NAME=homecloud-tester" \
		-v $(shell pwd):/opt/homecloud \
		-v $(HOME)/go/pkg/mod:/go/pkg/mod \
		-v $(HOME)/go/pkg/cache:/root/.cache/go-build \
		-v $(HOST_DOCKER_SOCKET):/var/run/docker.sock \
		--name "homecloud-tester" \
		--entrypoint go \
		golang:1.23 \
		test ./...
