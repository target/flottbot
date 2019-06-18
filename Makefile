SOURCE_COMMIT := $(shell git log -1 --pretty=format:"%H")
SOURCE_BRANCH := $(shell git branch | grep \* | cut -d ' ' -f2)
TAG := target/flottbot

.PHONY: all docker

build:
	go build -v -a -o flottbot cmd/flottbot/main.go

run: build
	./flottbot

docker:
	docker build \
		--build-arg "SOURCE_BRANCH=$(SOURCE_BRANCH)" \
		--build-arg "SOURCE_COMMIT=$(SOURCE_COMMIT)" \
		-f "./docker/Dockerfile.ruby" \
		-t $(TAG):ruby .
	
	docker build \
		--build-arg "SOURCE_BRANCH=$(SOURCE_BRANCH)" \
		--build-arg "SOURCE_COMMIT=$(SOURCE_COMMIT)" \
		-f "./docker/Dockerfile.python" \
		-t $(TAG):python .

	docker build \
		--build-arg "SOURCE_BRANCH=$(SOURCE_BRANCH)" \
		--build-arg "SOURCE_COMMIT=$(SOURCE_COMMIT)" \
		-f "./docker/Dockerfile.golang" \
		-t $(TAG):golang .
