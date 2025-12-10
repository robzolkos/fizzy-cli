# Fizzy CLI

A command-line interface for the [Fizzy](https://fizzy.do) API.  Read API [docs](https://github.com/basecamp/fizzy/blob/main/docs/API.md).



https://github.com/user-attachments/assets/86b91eae-7e8a-418c-a99e-3493ed7290cc



## Installation

Download the latest release from [GitHub Releases](https://github.com/robzolkos/fizzy-cli/releases) and add it to your PATH.

**From source**
```bash
git clone https://github.com/robzolkos/fizzy-cli.git
cd fizzy-cli
bundle install
./bin/fizzy --help
```

## Quick Start

1. Get your API token from your [Fizzy profile](https://app.fizzy.do/my/profile) under "Personal access tokens"

2. Configure the CLI:

```bash
.bin/fizzy auth login YOUR_TOKEN
```

3. List your accounts:

```bash
fizzy identity show
```

4. Set your default account:

```bash
export FIZZY_ACCOUNT=897362094
```

## Usage

```
fizzy <resource> <action> [options]
```

### Global Options

| Option | Environment Variable | Description |
|--------|---------------------|-------------|
| `--token` | `FIZZY_TOKEN` | API access token |
| `--account` | `FIZZY_ACCOUNT` | Account ID |
| `--format` | | Output format: `json` (default), `text` |
| `--quiet` | | Suppress non-essential output |
| `--verbose` | | Show request/response details |

## Commands

### Boards

```bash
# List all boards
fizzy board list

# Show a board
fizzy board show BOARD_ID

# Create a board
fizzy board create --name "Engineering"

# Update a board
fizzy board update BOARD_ID --name "New Name"

# Delete a board
fizzy board delete BOARD_ID
```

### Cards

```bash
# List cards (with optional filters)
fizzy card list
fizzy card list --board BOARD_ID
fizzy card list --status published
fizzy card list --assignee USER_ID

# Show a card
fizzy card show 42

# Create a card
fizzy card create --board BOARD_ID --title "Fix login bug"
fizzy card create --board BOARD_ID --title "New feature" --description "Details here"

# Update a card
fizzy card update 42 --title "Updated title"

# Delete a card
fizzy card delete 42
```

### Card Actions

```bash
# Close/reopen
fizzy card close 42
fizzy card reopen 42

# Move to "Not Now"
fizzy card postpone 42

# Move into a column
fizzy card column 42 --column COLUMN_ID

# Send back to triage
fizzy card untriage 42

# Assign/unassign (toggles)
fizzy card assign 42 --user USER_ID

# Tag/untag (toggles, creates tag if needed)
fizzy card tag 42 --tag "bug"

# Watch/unwatch
fizzy card watch 42
fizzy card unwatch 42
```

### Columns

```bash
fizzy column list --board BOARD_ID
fizzy column show COLUMN_ID --board BOARD_ID
fizzy column create --board BOARD_ID --name "In Progress"
fizzy column update COLUMN_ID --board BOARD_ID --name "Done"
fizzy column delete COLUMN_ID --board BOARD_ID
```

### Comments

```bash
fizzy comment list --card 42
fizzy comment show COMMENT_ID --card 42
fizzy comment create --card 42 --body "Looks good!"
fizzy comment update COMMENT_ID --card 42 --body "Updated comment"
fizzy comment delete COMMENT_ID --card 42
```

### Users

```bash
fizzy user list
fizzy user show USER_ID
```

### Tags

```bash
fizzy tag list
```

### Notifications

```bash
fizzy notification list
fizzy notification read NOTIFICATION_ID
fizzy notification unread NOTIFICATION_ID
fizzy notification read-all
```

## Output Format

All commands output JSON by default:

```json
{
  "success": true,
  "data": { ... },
  "meta": {
    "timestamp": "2025-12-10T10:00:00Z"
  }
}
```

Errors return a non-zero exit code and structured error info:

```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Card not found",
    "status": 404
  }
}
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments |
| 3 | Authentication failure |
| 4 | Permission denied |
| 5 | Not found |
| 6 | Validation error |
| 7 | Network error |

## Configuration

The CLI looks for configuration in:
- `~/.fizzy/config.yaml`
- `~/.config/fizzy/config.yaml`

```yaml
token: fizzy_abc123...
account: 897362094
api_url: https://app.fizzy.do
```

Environment variables and command-line flags override the config file.

## Pagination

List commands return paginated results. Use `--page` to fetch specific pages or `--all` to fetch everything:

```bash
fizzy card list --page 2
fizzy card list --all
```

## License

MIT
