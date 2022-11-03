<p align="center">
<img alt="flottbot logo" src="https://i.imgur.com/P9NI00w.png" height="160" />

<h3 align="center">Flottbot</h3>
</p>

---

[![GoDoc](https://godoc.org/github.com/target/flottbot?status.svg)](https://godoc.org/github.com/target/flottbot)
[![Build Status](https://github.com/target/flottbot/workflows/release/badge.svg)](https://github.com/target/flottbot/workflows/release)
[![GitHub release](https://img.shields.io/github/release/target/flottbot.svg)](https://github.com/target/flottbot/releases/latest)
[![Coverage Status](https://coveralls.io/repos/target/flottbot/badge.svg)](https://coveralls.io/r/target/flottbot)
[![Go Report Card](https://goreportcard.com/badge/github.com/target/flottbot)](https://goreportcard.com/report/github.com/target/flottbot)

Flottbot is a chatbot framework written in Go. But there's a catch, you don't need to know a lick of Go! Configure your bot via YAML files, extend functionality by writing scripts in your favorite language.

The philosophy behind flottbot is to create very simple, lightweight, "dumb" bots that interact with APIs and scripts which house a bot's business logic. The word **flott** comes from the German word meaning _quick_/_speedy_.

1. [Installation](#installation)
1. [Docker images](#docker-images)
1. [Available remotes](#available-remotes)
1. [Documentation](#documentation)
1. [Contributing](#contributing)

---

## Installation

### Using go

```sh
$ go get -u github.com/target/flottbot/cmd/flottbot
```

### Binaries

Binaries for Linux, macOS, and Windows are available as [Github Releases](https://github.com/target/flottbot/releases/latest).

## Docker Images

We currently provide a few Docker images:

[target/flottbot](https://hub.docker.com/r/target/flottbot) - Alpine image and flottbot binary installed

[target/flottbot:ruby](https://hub.docker.com/r/target/flottbot) - Alpine image, flottbot binary, and ruby v3.1 installed

[target/flottbot:golang](https://hub.docker.com/r/target/flottbot) - Alpine image, flottbot binary, and golang v1.19 installed

[target/flottbot:python](https://hub.docker.com/r/target/flottbot) - Alpine image, flottbot binary, and python v3.11 installed

_Note: We highly recommend pinning your image to a version, ie. `target/flottbot:0.10.0` or `target/flottbot:ruby-0.10.0`_

_Note: The images run with the unprivileged `flottbot` user (uid/gid 900) by default_

## Helm Chart

To install using the [Helm](https://helm.sh/) chart located in this repo, clone this repo, create [Kubernetes secrets](https://kubernetes.io/docs/concepts/configuration/secret/) for your Slack Token and Slack App Token in your namespace & install the chart:

```sh
$ helm install helm/flottbot/
```

## Available remotes

| Remote                              | Status | Documentation                                                      |
| ----------------------------------- | ------ | ------------------------------------------------------------------ |
| [Slack](https://slack.com)          | ✔      | [Docs](https://target.github.io/flottbot-docs/basics/slack/)       |
| [Discord](https://discordapp.com)   | 🚧     | [Docs](https://target.github.io/flottbot-docs/basics/discord/)     |
| [Google Chat](https://telegram.org) | 🚧     | [Docs](https://target.github.io/flottbot-docs/basics/google-chat/) |
| [Telegram](https://telegram.org)    | 🚧     | coming soon                                                        |

✔ = Done 🚧 = in progress (functional but some features may not work)

## Documentation

For installation and usage, please [visit the flottbot docs](https://target.github.io/flottbot-docs/)

For questions join the [#flottbot](https://gophers.slack.com/messages/flottbot/) channel in the [Gophers Slack](https://invite.slack.golangbridge.org/).

## Contributing

Please do! Check [CONTRIBUTING.md](./.github/CONTRIBUTING.md) for info.

## Credits

Inspired by [Hexbot.io](https://github.com/mmcquillan/hex)

## Contributors

* [List of contributors](https://github.com/target/flottbot/graphs/contributors)
