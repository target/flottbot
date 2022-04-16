FROM docker.io/golang:1.18-alpine@sha256:42d35674864fbb577594b60b84ddfba1be52b4d4298c961b46ba95e9fb4712e8 AS build
ARG VERSION

# needed for vcs feature introduced in go 1.18
RUN apk add --no-cache git

WORKDIR /src

# Allow for caching
COPY go.mod go.sum ./
RUN go mod download

COPY / .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -a -ldflags "-s -w -X github.com/target/flottbot/version.Version=${VERSION}" \
  -o flottbot ./cmd/flottbot

FROM docker.io/ruby:3.1-alpine@sha256:0b91e98c4613c2287bb7274a1584ea141b68960fb6b0edd7fe32127dc888d1a0

ENV USERNAME=flottbot
ENV GROUP=flottbot
ENV UID=900
ENV GID=900

RUN apk add --no-cache ca-certificates ruby-dev build-base && mkdir config &&  \
  addgroup -g "$GID" -S "$GROUP" && adduser -S -u "$UID" -G "$GROUP" "$USERNAME"

COPY --from=build /src/flottbot /flottbot

EXPOSE 8080 3000 4000

USER ${USERNAME}
CMD ["/flottbot"]
