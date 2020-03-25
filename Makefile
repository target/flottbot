PWD := $(shell pwd)
GOPATH := $(shell go env GOPATH)
GIT_HASH := $(shell git log -1 --pretty=format:"%H")
VERSION := $(shell git describe --tags --always)
GOARCH := $(shell go env GOARCH)
GOOS := $(shell go env GOOS)
BUILD_LDFLAGS := -s -w
BUILD_LDFLAGS += -X github.com/target/flottbot/version.Version=${VERSION}
BUILD_LDFLAGS += -X github.com/target/flottbot/version.GitHash=${GIT_HASH}
GOLANGCI_LINT_VERSION := "v1.23.8"

DOCKER_IMAGE ?= "target/flottbot"
DOCKER_FLAVORS ?= golang ruby python
PLATFORMS ?= linux/amd64 darwin/amd64 windows/amd64

# some helpers for building for each platform
p = $(subst /, ,$@)
os = $(word 1, $(p))
arch = $(word 2, $(p))

.PHONY: all
all: test build

# ┌┬┐┌─┐┌─┐┌┬┐
#  │ ├┤ └─┐ │ 
#  ┴ └─┘└─┘ ┴ 
.PHONY: validate
validate: getdeps fmt vet lint

.PHONY: getdeps
getdeps:
	@mkdir -p ${GOPATH}/bin
	@which golangci-lint 1>/dev/null || \
		(echo "Installing golangci-lint" && \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
		sh -s -- -b $(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION))

.PHONY: lint
lint:
	@echo "Running $@ check"
	@GO111MODULE=on golangci-lint run

.PHONY: fmt
fmt:
	@echo "Running $@ check"
	@GO111MODULE=on go fmt ./...

.PHONY: vet
vet:
	@echo "Running $@ check"
	@GO111MODULE=on go vet ./...

.PHONY: tidy
tidy:
	@echo "Running $@"
	@go mod tidy

.PHONY: test
test:
	@echo "Running unit tests"
	@go test ./...

.PHONY: test-race
test-race:
	@echo "Running unit tests with -race"
	@go test -v -race -coverprofile=coverage.out -coverpkg=./... `go list ./... | grep -v config-example`

.PHONY: clean
clean: validate tidy
	@echo "Running $@ tasks"
	-rm -v ./flottbot*
	-rm -v ./debug

# ┌┐ ┬ ┬┬┬  ┌┬┐
# ├┴┐│ │││   ││
# └─┘└─┘┴┴─┘─┴┘
.PHONY: build
build: clean
	@echo "Building flottbot binary to './flottbot'"
	@GO111MODULE=on go build -a \
		-ldflags '$(BUILD_LDFLAGS)' -o $(PWD)/flottbot ./cmd/flottbot

.PHONY: build-cross
build-cross: clean $(PLATFORMS)

.PHONY: $(PLATFORMS)
$(PLATFORMS):
	@echo "Building for $@"
	@GO111MODULE=on CGO_ENABLED=0 GOOS=$(os) GOARCH=$(arch) go build -a \
		-ldflags '$(BUILD_LDFLAGS)' -o $(PWD)/flottbot-$(os)-$(arch) ./cmd/flottbot
	@echo "Compressing to flottbot-$(os)-$(arch).tgz"
	@tar czf flottbot-$(os)-$(arch).tgz flottbot-$(os)-$(arch)
	@echo "Generating checksum for flottbot-$(os)-$(arch).tgz"
	@shasum -a 256 flottbot-$(os)-$(arch).tgz >> flottbot-$(VERSION)-checksum.txt
	@echo "Removing flottbot-$(os)-$(arch) binary"
	-rm -v flottbot-$(os)-$(arch)

# ┌┬┐┌─┐┌─┐┬┌─┌─┐┬─┐
#  │││ ││  ├┴┐├┤ ├┬┘
# ─┴┘└─┘└─┘┴ ┴└─┘┴└─
.PHONY: docker-base
docker-base:
	@echo "Creating base $@ image"
	@docker build \
		--build-arg "VERSION=$(VERSION)" \
		--build-arg "GIT_HASH=$(GIT_HASH)" \
		-f "./docker/Dockerfile" \
		-t $(DOCKER_IMAGE):$(VERSION) \
		-t $(DOCKER_IMAGE):latest .

.PHONY: docker-flavors
docker-flavors:
	@for flavor in $(DOCKER_FLAVORS); do \
		echo "Creating image for $$flavor"; \
		docker build \
			--build-arg "VERSION=$(VERSION)" \
			--build-arg "GIT_HASH=$(GIT_HASH)" \
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
		docker push $(DOCKER_IMAGE):$$flavor \
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
