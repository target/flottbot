## Contributing

To ensure that all developers follow the same guidelines for development, we have laid out the following procedure.

### Prerequisites

- [Go(lang)](https://golang.org/dl/) - the source code is written in Go.
- Slack API Token
- Slack App Token

### Development Process

- Clone this repository to your Go workspace:

```sh
# Make sure you are running go 1.21 or later

# Clone the project
git clone git@github.com:target/flottbot.git somepath/flottbot
```

- Build the project:

```sh
# Change into the project directory
cd somepath/flottbot

# Install modules
go mod download
# Build project
make build
```

- Write your code and ensure all tests pass.

```sh
# Checkout a branch for your work
git switch -c name_of_your_branch

# Code away!
```

- Build the project and run locally:

```sh
# Export your Slack Token and Slack App Token (the tokens below is redacted)
export SLACK_TOKEN=xoxb-xxxxxxxxxxxx-xxxxxxxxxxxxxxxxxxxxxxxx
export SLACK_APP_TOKEN=xapp-x-xxxxxxxxx-xxxxxxxxxxxxx-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

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

- Satisfied with your contribution? Record your changes in the [changelog](/CHANGELOG.md).

- Submit a PR for your changes.

- After the Github Actions build passes and you have an approved review, we will merge your PR.

- We will tag a release for flottbot when the desired functionality is present and stable.
- Production images of your changes will be published to Docker Hub and new binaries will be built and made available via Github Releases
