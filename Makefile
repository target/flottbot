PWD := $(shell pwd)
GOPATH := $(shell go env GOPATH)
VERSION := $(shell git describe --tags --always)
GOARCH := $(shell go env GOARCH)
GOOS := $(shell go env GOOS)
BUILD_LDFLAGS := -s -w
BUILD_LDFLAGS += -X github.com/target/flottbot/version.Version=${VERSION}
GOLANGCI_LINT_VERSION := "v1.52.2"
PACKAGES := $(shell go list ./... | grep -v /config-example/)

DOCKER_IMAGE ?= "target/flottbot"
DOCKER_FLAVORS ?= golang ruby python

.PHONY: all
all: test build

# ┌┬┐┌─┐┌─┐┌┬┐
#  │ ├┤ └─┐ │ 
#  ┴ └─┘└─┘ ┴ 
.PHONY: validate
validate: getdeps fmt vet lint tidy

.PHONY: getdeps
getdeps:
	@mkdir -p ${GOPATH}/bin
	@which golangci-lint 1>/dev/null || \
		(echo "Installing golangci-lint" && \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
		sh -s -- -b $(shell go env GOPATH)/bin $(GOLANGCI_LINT_VERSION))

.PHONY: lint
lint:
	@echo "Running $@ check"
	@golangci-lint run

.PHONY: fmt
fmt:
	@echo "Running $@ check"
	@go fmt ./...

.PHONY: vet
vet:
	@echo "Running $@ check"
	@go vet ./...

.PHONY: tidy
tidy:
	@echo "Running $@"
	@go mod tidy

.PHONY: test
test:
	@echo "Running unit tests"
	@go test ./...

.PHONY: test-coverage
test-coverage:
	@echo "Running unit tests with coverage"
	@go test -v -covermode=count -coverpkg=$(PACKAGES) -coverprofile=coverage.out ./...

.PHONY: clean
clean: validate
	@echo "Running $@ tasks"
	-rm -v ./flottbot*
	-rm -v ./debug

# ┌┐ ┬ ┬┬┬  ┌┬┐
# ├┴┐│ │││   ││
# └─┘└─┘┴┴─┘─┴┘
.PHONY: build
build: clean
	@echo "Building flottbot binary to './flottbot'"
	@go build -a \
		-ldflags '$(BUILD_LDFLAGS)' -o $(PWD)/flottbot ./cmd/flottbot

# ┌┬┐┌─┐┌─┐┬┌─┌─┐┬─┐
#  │││ ││  ├┴┐├┤ ├┬┘
# ─┴┘└─┘└─┘┴ ┴└─┘┴└─
.PHONY: docker-base
docker-base:
	@echo "Creating base $@ image"
	@docker build \
		--build-arg "VERSION=$(VERSION)" \
		-f "./docker/Dockerfile" \
		-t $(DOCKER_IMAGE):$(VERSION) \
		-t $(DOCKER_IMAGE):latest .

.PHONY: docker-flavors
docker-flavors:
	@for flavor in $(DOCKER_FLAVORS); do \
		echo "Creating image for $$flavor"; \
		docker build \
			--build-arg "VERSION=$(VERSION)" \
			-f "./docker/Dockerfile.$$flavor" \
			-t $(DOCKER_IMAGE):$$flavor \
			-t $(DOCKER_IMAGE):$$flavor-$(VERSION) .; \
	done

.PHONY: docker-create-all
docker-create-all: docker-base docker-flavors

.PHONY: docker-login
docker-login:
    ifndef DOCKER_USERNAME
		$(error DOCKER_USERNAME not set)
    else ifndef DOCKER_PASSWORD 
		$(error DOCKER_PASSWORD not set)
    endif
	@echo "Logging into docker hub"
	@echo "$$DOCKER_PASSWORD" | docker login -u $$DOCKER_USERNAME --password-stdin

.PHONY: docker-push
docker-push: docker-login
	@echo "Pushing $(DOCKER_IMAGE):$(VERSION) and :latest to docker hub"
	@docker push $(DOCKER_IMAGE):$(VERSION)
	@docker push $(DOCKER_IMAGE):latest
	
	@for flavor in $(DOCKER_FLAVORS); do \
		echo "Pushing $(DOCKER_IMAGE):$$flavor to docker hub"; \
		docker push $(DOCKER_IMAGE):$$flavor; \
		docker push $(DOCKER_IMAGE):$$flavor-$(VERSION); \
	done

.PHONY: docker-push-latest
docker-push-latest: docker-login
	@echo "Pushing to :latest images to docker hub..."
	
	@echo "Pushing $(DOCKER_IMAGE):latest"
	@docker push $(DOCKER_IMAGE):latest
	
	@for flavor in $(DOCKER_FLAVORS); do \
		echo "Pushing $(DOCKER_IMAGE):$$flavor to docker hub"; \
		docker push $(DOCKER_IMAGE):$$flavor; \
	done

# ┬─┐┬ ┬┌┐┌
# ├┬┘│ ││││
# ┴└─└─┘┘└┘
.PHONY: run
run: build
	@echo "Starting flottbot"
	./flottbot

.PHONY: run-docker
run-docker: docker
	@echo "Starting flottbot docker image"
	@docker run -it --rm --name myflottbot -v "$$PWD"/config:/config --env-file .env $(DOCKER_IMAGE):latest /flottbot
