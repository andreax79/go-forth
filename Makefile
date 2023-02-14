IMAGE_NAME=fcpu
VERSION=`cat VERSION`
OWNER=andreax79

.PHONY: help build test format image push all

help:
	@echo "- make format       Format source code"
	@echo "- make build        Build"
	@echo "- make test         Run tests"
	@echo "- make tag          Add version tag"
	@echo "- make image        Build docker image"
	@echo "- make push         Push docker image"

.DEFAULT_GOAL := help

build:
	@go build -o bin/cpu -ldflags "-X main.version=${VERSION}" ./cmd/cpu

test:
	@go test -v ./...

format:
	@go fmt ./...

tag:
	@git tag -a "v$$(cat VERSION)" -m "version v$$(cat VERSION)"

image:
	@DOCKER_BUILDKIT=1 docker build \
		 --tag ${IMAGE_NAME}:latest \
		 --tag ${IMAGE_NAME}:${VERSION} \
		 .

push:
	@docker tag ${IMAGE_NAME}:${VERSION} ghcr.io/${OWNER}/${IMAGE_NAME}:${VERSION}
	@docker tag ${IMAGE_NAME}:${VERSION} ghcr.io/${OWNER}/${IMAGE_NAME}:latest
	@docker push ghcr.io/${OWNER}/${IMAGE_NAME}:${VERSION}
	@docker push ghcr.io/${OWNER}/${IMAGE_NAME}:latest
	@docker tag ${IMAGE_NAME}:${VERSION} ${OWNER}/${IMAGE_NAME}:${VERSION}
	@docker tag ${IMAGE_NAME}:${VERSION} ${OWNER}/${IMAGE_NAME}:latest
	@docker push ${OWNER}/${IMAGE_NAME}:${VERSION}
	@docker push ${OWNER}/${IMAGE_NAME}:latest

all: build
