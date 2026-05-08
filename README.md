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

- [Installation](#installation)
  - [Using go](#using-go)
  - [Binaries](#binaries)
- [Docker Images](#docker-images)
- [Helm Chart](#helm-chart)
- [Available remotes](#available-remotes)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [Credits](#credits)
- [Contributors](#contributors)

---

## Installation

### Using go

```sh
go get -u github.com/target/flottbot/cmd/flottbot
```

### Binaries

Binaries for Linux, macOS, and Windows are available as [Github Releases](https://github.com/target/flottbot/releases/latest).

## Docker Images

We currently provide a few Docker images:

[target/flottbot](https://hub.docker.com/r/target/flottbot) - Alpine image and flottbot binary installed

[target/flottbot:ruby](https://hub.docker.com/r/target/flottbot) - Alpine image, flottbot binary, and ruby v3.2 installed

[target/flottbot:golang](https://hub.docker.com/r/target/flottbot) - Alpine image, flottbot binary, and golang v1.23 installed

[target/flottbot:python](https://hub.docker.com/r/target/flottbot) - Alpine image, flottbot binary, and python v3.11 installed

_Note: We highly recommend pinning your image to a version, ie. `target/flottbot:0.10.0` or `target/flottbot:ruby-0.10.0`_

_Note: The images run with the unprivileged `flottbot` user (uid/gid 900) by default_

## Helm Chart

To install using the [Helm](https://helm.sh/) chart located in this repo, clone this repo, create [Kubernetes secrets](https://kubernetes.io/docs/concepts/configuration/secret/) for your Slack Token and Slack App Token in your namespace & install the chart:

```sh
helm install helm/flottbot/
```

## Available remotes

| Remote                                                     | Status | Documentation                                                      |
| ---------------------------------------------------------- | ------ | ------------------------------------------------------------------ |
| [Slack](https://slack.com)                                 | ✔      | [Docs](https://target.github.io/flottbot-docs/basics/slack/)       |
| [Discord](https://discordapp.com)                          | 🚧      | [Docs](https://target.github.io/flottbot-docs/basics/discord/)     |
| [Google Chat](https://workspace.google.com/products/chat/) | 🚧      | [Docs](https://target.github.io/flottbot-docs/basics/google-chat/) |
| [Mattermost](https://mattermost.com/)                      | 🚧      | coming soon                                                        |
| [Telegram](https://telegram.org)                           | 🚧      | coming soon                                                        |

✔ = Done 🚧 = in progress (functional but some features may not work)

## Documentation

For installation and usage, please [visit the flottbot docs](https://target.github.io/flottbot-docs/)

For questions join the [#flottbot](https://gophers.slack.com/messages/flottbot/) channel in the [Gophers Slack](https://invite.slack.golangbridge.org/).

## Contributing

Please do! Check [CONTRIBUTING.md](./.github/CONTRIBUTING.md) for info.

## Credits

Inspired by [Hexbot.io](https://github.com/mmcquillan/hex)

## Contributors

- [List of contributors](https://github.com/target/flottbot/graphs/contributors)

## FAQ

### General Questions

**Q: What is flottbot and why should I use it?**

A: Flottbot is a chatbot framework written in Go that requires no Go knowledge. Configure your bot via YAML files and extend functionality by writing scripts in your favorite language. The philosophy is to create simple, lightweight bots that interact with APIs and scripts housing business logic.

**Q: What does "flott" mean?**

A: "Flott" comes from the German word meaning "quick" or "speedy", reflecting the framework's lightweight and fast design philosophy.

### Installation & Setup

**Q: How do I install flottbot?**

A: Use `go get -u github.com/target/flottbot/cmd/flottbot` or download binaries for Linux, macOS, and Windows from [GitHub Releases](https://github.com/target/flottbot/releases/latest).

**Q: What are the available Docker images?**

A: We provide several Docker images:
- `target/flottbot` - Alpine + flottbot binary
- `target/flottbot:ruby` - Alpine + flottbot + Ruby v3.2
- `target/flottbot:golang` - Alpine + flottbot + Go v1.23
- `target/flottbot:python` - Alpine + flottbot + Python v3.11

**Q: How do I deploy flottbot with Helm?**

A: Clone the repo, create Kubernetes secrets for Slack Token and App Token, then run `helm install helm/flottbot/`.

### Remotes & Integrations

**Q: Which chat platforms does flottbot support?**

A: Supported remotes:
- **Slack** ✔ (fully supported)
- **Discord** 🚧 (in progress)
- **Google Chat** 🚧 (in progress)
- **Mattermost** 🚧 (coming soon)
- **Telegram** 🚧 (coming soon)

**Q: How do I configure Slack integration?**

A: Create Slack Token and App Token as Kubernetes secrets (for Helm deployment) or configure them in your YAML files. See [Slack documentation](https://target.github.io/flottbot-docs/basics/slack/) for details.

### Configuration & Scripting

**Q: Do I need to know Go to use flottbot?**

A: No! Configure your bot entirely via YAML files. Extend functionality by writing scripts in any language (Python, Ruby, Bash, etc.).

**Q: How do I add custom functionality?**

A: Write scripts in your preferred language and reference them in your YAML configuration. Flottbot handles the bot-logic while your scripts handle business logic.

### Troubleshooting

**Q: My bot won't connect to Slack - what should I check?**

A: Verify your Slack Token and App Token are correctly configured. Ensure the bot has proper permissions in your Slack workspace. Check logs for specific error messages.

**Q: Docker container fails to start - what's wrong?**

A: Ensure you're using the correct image version (pin to specific version like `target/flottbot:0.10.0`). The container runs with unprivileged user (uid/gid 900), check file permissions.

### Getting Help

**Q: Where can I find documentation?**

A: Visit [flottbot docs](https://target.github.io/flottbot-docs/) for installation and usage guides.

**Q: Is there a community channel?**

A: Join the [#flottbot](https://gophers.slack.com/messages/flottbot/) channel in [Gophers Slack](https://invite.slack.golangbridge.org/) for questions and discussions.

**Q: How can I contribute?**

A: Check [CONTRIBUTING.md](./.github/CONTRIBUTING.md) for contribution guidelines. Pull requests are welcome!
