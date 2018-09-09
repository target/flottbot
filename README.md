<p align="center">
  <img alt="flottbot logo" src="https://i.imgur.com/P9NI00w.png" height="160" />

  <h3 align="center">Flottbot</h3>
</p>

--------

[![GoDoc](https://godoc.org/github.com/target/flottbot?status.svg)](https://godoc.org/github.com/target/flottbot)
[![Build Status](https://travis-ci.org/target/flottbot.svg)](https://travis-ci.org/target/flottbot)
[![GitHub release](https://img.shields.io/github/release/target/flottbot.svg)](https://github.com/target/flottbot/releases/latest)
[![Coverage Status](https://coveralls.io/repos/target/flottbot/badge.svg?branch=master)](https://coveralls.io/r/target/flottbot?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/target/flottbot)](https://goreportcard.com/report/github.com/target/flottbot)
[![docker](https://img.shields.io/docker/automated/target/flottbot.svg)](https://hub.docker.com/r/target/flottbot)

Flottbot is a chatbot framework written in Go. But there's a catch, you don't need to know a lick of Go! Configure your bot via YAML files, extend functionality by writing scripts in your favorite language.

The philosophy behind flottbot is to create very simple, lightweight, "dumb" bots that interact with APIs and scripts which house a bot's business logic. The word **flott** comes from the German word meaning _quick_/_speedy_.

1. [Installation](#installation)
1. [Docker images](#docker-images)
1. [Available remotes](#available-remotes)
1. [Documentation](#documentation)
1. [Contributing](#contributing)

-------------------

## Installation

### Using go

```bash
go get -u github.com/target/flottbot
```

### Binaries

Binaries for Linux, macOS, and Windows are available as [Github Releases](https://github.com/target/flottbot/releases/latest).

## Docker Images

We currently provide a few Docker images:

[target/flottbot](https://hub.docker.com/r/target/flottbot) - Alpine image and flottbot binary installed

[target/flottbot:ruby](https://hub.docker.com/r/target/flottbot) - Alpine image, flottbot binary, and ruby v2.5 installed

[target/flottbot:golang](https://hub.docker.com/r/target/flottbot) - Alpine image, flottbot binary, and golang v1.11 installed

## Available remotes

| Remote                | Status | Documentation |
| --------------------- | -------| ------------- |
| [Slack](https://slack.com) | ✔ | - |
| [Discord](https://discordapp.com)  | 🚧 | - |

✔ = Done 🚧 = in progress

## Documentation

For installation and usage, please [visit the flottbot docs](https://pages.github.com/target/flottbot-docs)

For questions join the [#flottbot](https://gophers.slack.com/messages/flottbot/) channel in the [Gophers Slack](https://invite.slack.golangbridge.org/).

## Contributing

Please do! Check [CONTRIBUTING.md](./.github/CONTRIBUTING.md) for info.

## Credits

Inspired by [Hexbot.io](https://github.com/mmcquillan/hex)

## Authors

- [David May](https://github.com/wass3r)
- [Sean Quinn](https://github.com/sjqnn)
- [Raphael Santo Domingo](https://github.com/pa3ng)
- [Jordan Sussman](https://github.com/JordanSussman)
