FROM golang:1.12-alpine AS build
ARG SOURCE_BRANCH
ARG SOURCE_COMMIT
WORKDIR /go/src/github.com/target/flottbot/
RUN apk add --no-cache git
ENV GO111MODULE=on
COPY / .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags "-X github.com/target/flottbot/version.Version=${SOURCE_BRANCH} -X github.com/target/flottbot/version.GitHash=${SOURCE_COMMIT}" \
    -o flottbot ./cmd/flottbot

FROM alpine:3.9
RUN apk --no-cache add ca-certificates && mkdir config
COPY --from=build /go/src/github.com/target/flottbot/flottbot .
EXPOSE 8080 3000 4000

CMD ["/flottbot"]
