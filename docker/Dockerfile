FROM golang:1.15-alpine AS build
ARG VERSION
ARG GIT_HASH
ENV GO111MODULE=on

WORKDIR /src

# Allow for caching
COPY go.mod go.sum ./
RUN go mod download

COPY / .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -a -ldflags "-s -w -X github.com/target/flottbot/version.Version=${VERSION} -X github.com/target/flottbot/version.GitHash=${GIT_HASH}" \
  -o flottbot ./cmd/flottbot

FROM alpine:3.11
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