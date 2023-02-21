VERSION ?= `git describe --tags`
BUILDFLAGS := -mod=vendor -ldflags="-s -w" -gcflags=-trimpath=$(CURDIR)
IMAGE_NAME := razchess
IMAGE_REGISTRY ?= ghcr.io/razzie
FULL_IMAGE_NAME := $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(VERSION)

.PHONY: all
all: razchess bot uci

.PHONY: razchess
razchess:
	go build $(BUILDFLAGS) .

.PHONY: bot
bot:
	go build $(BUILDFLAGS) ./tools/bot

.PHONY: uci
uci:
	go build $(BUILDFLAGS) ./tools/uci

.PHONY: docker-build
docker-build:
	docker build . -t $(FULL_IMAGE_NAME)

.PHONY: docker-push
docker-push: docker-build
	docker push $(FULL_IMAGE_NAME)
