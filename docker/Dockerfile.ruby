FROM golang:1.18-alpine AS build
ARG VERSION

WORKDIR /src

# Allow for caching
COPY go.mod go.sum ./
RUN go mod download

COPY / .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -a -ldflags "-s -w -X github.com/target/flottbot/version.Version=${VERSION}" \
  -o flottbot ./cmd/flottbot

FROM ruby:3.1-alpine
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
