FROM docker.io/golang:1.18-alpine@sha256:7cc62574fcf9c5fb87ad42a9789d5539a6a085971d58ee75dd2ee146cb8a8695 AS build
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

FROM docker.io/alpine:3.16@sha256:686d8c9dfa6f3ccfc8230bc3178d23f84eeaf7e457f36f271ab1acc53015037c

ENV USERNAME=flottbot
ENV GROUP=flottbot
ENV UID=900
ENV GID=900

RUN apk --no-cache add ca-certificates && mkdir config && \
  addgroup -g "$GID" -S "$GROUP" && adduser -S -u "$UID" -G "$GROUP" "$USERNAME"

COPY --from=build /src/flottbot /flottbot

EXPOSE 3000 4000 8080

USER ${USERNAME}
CMD [ "/flottbot" ]
