build:
	go build -v -a -o flottbot cmd/flottbot/main.go

run: build
	./flottbot