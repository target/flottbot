# Flottbot Rule Examples

The `./rules` directory contains example rule configurations for Flottbot. These examples demonstrate various features and integrations possible in Flottbot.

## File Formats

Rules can be defined in either YAML or JSON format:

- **YAML files (`.yml` or `.yaml`)** - Recommended format with support for comments and better readability
- **JSON files (`.json`)** - Alternative format for programmatic generation (see `joke.json`)
- **TOML** and **HCL** are also supported but not covered in this documentation

## Rule Categories

### Basic Rules
- `hello.yml` - Simple greeting with user name substitution
- `cats.yml` - Random cat facts from API

### HTTP API Integration
- `joke.yml` - Dad jokes from API
- `weather.yml` - Weather forecasts using coordinates
- `number.yml` - Number facts and trivia
- `today.yml` - Historical events for today's date
- `pokemon.yml` - Pokemon stats lookup
- `xkcd.yml` - Random XKCD comics

### Script Execution
- `shell.yml` - Shell script execution
- `python.yml` - Python script execution
- `ruby.yml` - Ruby script execution
- `nodejs.yml` - Node.js script execution
- `golang.yml` - Go program execution

### Advanced Features
- `issue.yml` - GitHub issue creation
- `req.yml` / `req_template.yml` - HTTP request testing with templates
- `gopher.yml` - Random selection with reaction updates
- `dialogflow.yml` - Google Dialogflow integration (catch-all rule)

### Message Patterns
- `hear.yml` - Pattern matching with regex
- `link.yml` - Link detection and response
- `reaction.yml` - Emoji reaction triggers

### Scheduled Rules
- `schedule.yml` - Periodic messages
- `countdown.yml` - Daily countdown script
- `fotd.yml` - Daily fact delivery

### Access Control
- `announce.yml` - User/group permission example

## Configuration Sections

Each rule typically includes:

- **Rule metadata** - Name, active status
- **Trigger configuration** - How the rule is activated (respond, hear, schedule, etc.)
- **Arguments** - Required user inputs
- **Actions** - HTTP requests, script execution, etc.
- **Response configuration** - Output formatting and targeting
- **Help configuration** - Usage information

## Environment Variables

Some rules require environment variables:
- `GHE_TOKEN` - GitHub token for issue creation
- `DIALOG_FLOW_TOKEN` - Dialogflow API token

## Getting Started

1. Copy rules from this example directory to your main `config/rules/` directory
2. Modify the rules to match your needs
3. Set required environment variables
4. Activate rules by setting `active: true`

For more information, see the main Flottbot documentation.
