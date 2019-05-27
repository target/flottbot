## Contributing

To ensure that all developers follow the same guidelines for development, we have laid out the following procedure.

### Prerequisites

- [Golang](https://golang.org/dl/) - the source code is written in Go.
- [dep](https://github.com/golang/dep) - our Go dependency management tool.
- Slack API token - obtain a Slack API token for development by creating a bot integration.

### Development Process

- Clone this repository to your Go workspace:

```sh
# Make sure you are running go 1.11 or later
# if you plan to clone into your current GOPATH then set the environment variable GO111MODULE=on
# this will tell go to use the new modules support

# Clone the project
git clone git@github.com:target/flottbot.git somepath/src/github.com/target/flottbot
```

- Build the project:

```sh
# Change into the project directory
cd somepath/src/github.com/target/flottbot

# Install modules
go mod download
# Build project
make build
```

- Write your code and ensure all tests pass.

```sh
# Checkout a branch for your work
git checkout -b name_of_your_branch

# Code away!
```

- Build the project and run locally:

```sh
# Export your Slack API token (the token below is redacted)
export SLACK_TOKEN=xoxb-xxxxxxxxxxxx-xxxxxxxxxxxxxxxxxxxxxxxx

# Build the binary and run flottbot
make run
```

- If you followed the above steps correctly, you should see output similar to this:

```sh
DEBU[0000] Registering bot...
DEBU[0000] Bot 'flottbot' registered!
DEBU[0000] Searching for rules directory...
DEBU[0000] Fetching all rule files...
DEBU[0000] Reading and parsing rule files...
DEBU[0000] Successfully created rules
DEBU[0000] Registering flottbot to Slack...
DEBU[0000] Found channels!
DEBU[0000] Registering CLI support for flottbot...
Enter CLI mode: hit <Enter>. <Ctrl-C> to exit.
DEBU[0001] Connection established!
```

- You should now see your bot online in your Slack Workspace where you can manually test your changes.

- Satisfied with your contribution? Record your changes in the [changelog](https://github.com/target/flottbot/blob/master/CHANGELOG.md).

- Submit a PR for your changes.

- After the Travis build passes and you have an approved review, we will merge your PR.

- We will tag a release for flottbot when the desired functionality is present and stable.
  - Production images of your changes will be published to Docker Hub and new binaries will be built and made available via Github Releases
