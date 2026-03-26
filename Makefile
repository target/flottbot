PWD := $(shell pwd)
GOPATH := $(shell go env GOPATH)
VERSION := $(shell git describe --tags --always)
GOARCH := $(shell go env GOARCH)
GOOS := $(shell go env GOOS)
BUILD_LDFLAGS := -s -w
BUILD_LDFLAGS += -X github.com/target/flottbot/version.Version=${VERSION}
GOLANGCI_LINT_VERSION := "v2.11.4"
PACKAGES := $(shell go list ./... | grep -v /config-example/)
PLATFORM := "linux/amd64,linux/arm64"

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
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/main/install.sh | \
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
	-rm -v ./debug #Not created

# ┌┐ ┬ ┬┬┬  ┌┬┐
# ├┴┐│ │││   ││
# └─┘└─┘┴┴─┘─┴┘
.PHONY: build
build: clean
	@echo "Building flottbot binary to './flottbot'"
	@go build -a -ldflags '$(BUILD_LDFLAGS)' -o $(PWD)/flottbot ./cmd/flottbot

# ┌┬┐┌─┐┌─┐┬┌─┌─┐┬─┐
#  │││ ││  ├┴┐├┤ ├┬┘
# ─┴┘└─┘└─┘┴ ┴└─┘┴└─
.PHONY: docker-login
docker-login:
    ifndef DOCKER_USERNAME
		$(error DOCKER_USERNAME not set)
    else ifndef DOCKER_PASSWORD 
		$(error DOCKER_PASSWORD not set)
    endif
	@echo "Logging into docker hub"
	@echo "$$DOCKER_PASSWORD" | docker login -u $$DOCKER_USERNAME --password-stdin

.PHONY: docker-build-push-latest
docker-build-push-latest: docker-login
	@echo "Building and pushing latest to docker hub..."
	@echo "Building and pushing $(DOCKER_IMAGE):latest"
	@docker buildx build \
		--progress=plain \
		--build-arg "VERSION=$(VERSION)" \
		--platform $(PLATFORM) \
		--file "./docker/Dockerfile" \
		--tag $(DOCKER_IMAGE):latest \
		--push .
	@for flavor in $(DOCKER_FLAVORS); do \
		echo "Building and pushing $(DOCKER_IMAGE):$$flavor"; \
		docker buildx build \
			--progress=plain \
		  --build-arg "VERSION=$(VERSION)" \
			--platform $(PLATFORM) \
			--file "./docker/Dockerfile.$$flavor" \
			--tag $(DOCKER_IMAGE):$$flavor \
			--push .; \
	done

.PHONY: docker-build-push
docker-build-push: docker-login
	@echo "Building and pushing $(VERSION) to docker hub..."
	@echo "Building and pushing $(DOCKER_IMAGE):$(VERSION)"
	@docker buildx build \
		--progress=plain \
		--build-arg "VERSION=$(VERSION)" \
		--platform $(PLATFORM) \
		--file "./docker/Dockerfile" \
		--tag $(DOCKER_IMAGE):$(VERSION) \
		--tag $(DOCKER_IMAGE):latest \
		--push .
	@for flavor in $(DOCKER_FLAVORS); do \
		echo "Building and pushing $(DOCKER_IMAGE):$$flavor-$(VERSION)"; \
		docker buildx build \
			--progress=plain \
		  --build-arg "VERSION=$(VERSION)" \
			--platform $(PLATFORM) \
			--file "./docker/Dockerfile.$$flavor" \
			--tag $(DOCKER_IMAGE):$$flavor-$(VERSION) \
			--tag $(DOCKER_IMAGE):$$flavor \
			--push .; \
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
