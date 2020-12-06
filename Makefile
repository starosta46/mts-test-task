# Options.
#
PROJECT_NAME := mts-test-task
ORG_PATH := github.com
REPO_PATH ?= $(ORG_PATH)/$(PROJECT_NAME)
BINARY_NAME?=sites-data
IMAGE_NAME ?= $(REPO_PATH)/$(BINARY_NAME)
VERSION ?= dev
GOOS ?= linux

build: Dockerfile
	# Building $(PROJECT_NAME)...
	docker build \
		--ulimit nofile=262144:262144 --ulimit nproc=262144:262144 \
		--build-arg "VERSION=$(VERSION)" \
		--build-arg "APP_PKG_NAME=$(REPO_PATH)" \
		--build-arg "GOOS=$(GOOS)" \
		--build-arg "BINARY_NAME=$(BINARY_NAME)" \
		-t $(IMAGE_NAME) .

lint:
	golangci-lint run ./...

