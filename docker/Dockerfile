FROM golang:1.11-alpine AS build
ARG SOURCE_BRANCH
ARG SOURCE_COMMIT
WORKDIR /go/src/github.com/target/flottbot/
RUN apk add --no-cache git
RUN go get -u github.com/golang/dep/cmd/dep
COPY / .
RUN dep ensure
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags "-X github.com/target/flottbot/version.Version=${SOURCE_BRANCH} -X github.com/target/flottbot/version.GitHash=${SOURCE_COMMIT}" \
    -o flottbot ./cmd/flottbot

FROM alpine:3.8
RUN apk --no-cache add ca-certificates
RUN mkdir config
COPY --from=build /go/src/github.com/target/flottbot/flottbot .
EXPOSE 8080 3000 4000
