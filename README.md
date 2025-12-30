# Fizzy CLI

A command-line interface for the [Fizzy](https://fizzy.do) API. See the official API docs at https://github.com/basecamp/fizzy/blob/main/docs/API.md.

https://github.com/user-attachments/assets/86b91eae-7e8a-418c-a99e-3493ed7290cc

## Installation

Download the latest release from [GitHub Releases](https://github.com/robzolkos/fizzy-cli/releases) and add it to your PATH.

**With Go**
```bash
go install github.com/robzolkos/fizzy-cli/cmd/fizzy@latest
```

**From source**
```bash
git clone https://github.com/robzolkos/fizzy-cli.git
cd fizzy-cli
go build -o fizzy ./cmd/fizzy
./fizzy --help
```

## Quick Start

1. Get your API token from your [Fizzy profile](https://app.fizzy.do/my/profile) under "Personal access tokens"

2. Configure the CLI:

```bash
fizzy auth login YOUR_TOKEN
```

3. List your accounts:

```bash
fizzy identity show
```

4. Set your default account (use the numeric slug without the leading slash):

```bash
# From identity show: "slug": "/897362094" → use 897362094
export FIZZY_ACCOUNT=897362094
```

5. (Optional) Set your default board:

```bash
export FIZZY_BOARD=BOARD_ID
```

**Self-hosted domains:**
If you're using a self-hosted domain instead of the default `https://app.fizzy.do`, specify it using one of these methods:

**Option 1: Per-command flag**
```bash
fizzy identity show --api-url=https://your_domain
```

**Option 2: Configuration file**
Add to your `~/.config/fizzy/config.yaml`:

```yaml
token: your_token
account: your_account_id
api_url: https://your_domain
board: your_default_board_id
```

## Usage

```
fizzy <resource> <action> [options]
```

```bash
fizzy version
```

### Global Options

| Option | Environment Variable | Description |
|--------|---------------------|-------------|
| `--token` | `FIZZY_TOKEN` | API access token |
| `--account` | `FIZZY_ACCOUNT` | Account slug (from `fizzy identity show`) |
| `--api-url` | `FIZZY_API_URL` | API base URL (default: https://app.fizzy.do) |
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
fizzy card list --tag TAG_ID
fizzy card list --status published
fizzy card list --assignee USER_ID

# Show a card
fizzy card show 42

# Create a card
fizzy card create --board BOARD_ID --title "Fix login bug"
fizzy card create --board BOARD_ID --title "New feature" --description "Details here"
fizzy card create --board BOARD_ID --title "Card" --tag-ids "TAG_ID1,TAG_ID2"
fizzy card create --board BOARD_ID --title "Card" --image /path/to/header.png

# Create with custom timestamp (for data imports)
fizzy card create --board BOARD_ID --title "Old card" --created-at "2020-01-15T10:30:00Z"

# Update a card
fizzy card update 42 --title "Updated title"
fizzy card update 42 --created-at "2019-01-01T00:00:00Z"

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
fizzy comment create --card 42 --body-file /path/to/comment.html

# Create with custom timestamp (for data imports)
fizzy comment create --card 42 --body "Old comment" --created-at "2020-01-15T10:30:00Z"

fizzy comment update COMMENT_ID --card 42 --body "Updated comment"
fizzy comment delete COMMENT_ID --card 42
```

### Steps (To-Do Items)

```bash
# Show a step
fizzy step show STEP_ID --card 42

# Create a step
fizzy step create --card 42 --content "Review PR"
fizzy step create --card 42 --content "Already done" --completed

# Update a step
fizzy step update STEP_ID --card 42 --completed
fizzy step update STEP_ID --card 42 --not-completed
fizzy step update STEP_ID --card 42 --content "New content"

# Delete a step
fizzy step delete STEP_ID --card 42
```

### Reactions

```bash
# List reactions on a comment
fizzy reaction list --card 42 --comment COMMENT_ID

# Add a reaction (emoji, max 16 chars)
fizzy reaction create --card 42 --comment COMMENT_ID --content "👍"

# Remove a reaction
fizzy reaction delete REACTION_ID --card 42 --comment COMMENT_ID
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

### File Uploads

Upload files for use in rich text fields (card descriptions, comment bodies).

```bash
# Upload a file and get a signed_id
fizzy upload file /path/to/image.png

# Use the signed_id in a card description
fizzy card create --board BOARD_ID --title "Card" \
  --description '<p>See image:</p><action-text-attachment sgid="SIGNED_ID"></action-text-attachment>'
```

### Identity

```bash
# Show your identity and all accessible accounts
fizzy identity show
```

## Output Format

Command results output JSON. (`--help` and `--version` output plain text.)

```json
{
  "success": true,
  "data": { ... },
  "meta": {
    "timestamp": "2025-12-10T10:00:00Z"
  }
}
```

When creating resources, the CLI automatically follows the `Location` header to fetch the complete resource data:

```json
{
  "success": true,
  "data": {
    "id": "abc123",
    "number": 42,
    "title": "New Card",
    "status": "published"
  },
  "location": "https://app.fizzy.do/account/cards/42",
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

The CLI looks for configuration in multiple locations:

### Global Configuration

Global config is stored in one of these locations:
- `~/.config/fizzy/config.yaml` (preferred)
- `~/.fizzy/config.yaml`

```yaml
token: fizzy_abc123...
account: 897362094
api_url: https://app.fizzy.do
board: 123456
```

### Local Project Configuration

You can also create a `.fizzy.yaml` file in your project directory. The CLI walks up the directory tree to find it, so you can run commands from any subdirectory.

```yaml
# .fizzy.yaml - project-specific settings
account: 123456789
api_url: https://self-hosted.example.com
board: 123456
```

Local config values merge with global config:
- Values in local config override global config
- Empty values in local config do not override global values
- This allows you to keep your token in global config while overriding account per project

**Example:** Global config has your token, local config specifies which account to use for this project:

```yaml
# ~/.config/fizzy/config.yaml (global)
token: fizzy_abc123...

# /path/to/project/.fizzy.yaml (local)
account: 123456789
```

### Priority Order

Configuration priority (highest to lowest):
1. Command-line flags (`--token`, `--account`, `--api-url`)
2. Environment variables (`FIZZY_TOKEN`, `FIZZY_ACCOUNT`, `FIZZY_API_URL`, `FIZZY_BOARD`)
3. Local project config (`.fizzy.yaml` in current or parent directories)
4. Global config (`~/.config/fizzy/config.yaml` or `~/.fizzy/config.yaml`)
5. Defaults

## Pagination

List commands return paginated results. Use `--page` to fetch specific pages or `--all` to fetch everything:

```bash
fizzy card list --page 2
fizzy card list --all
```

## Development

### Building

```bash
go build -o bin/fizzy ./cmd/fizzy
```

### Running Tests

**Unit tests** (no API credentials required):

```bash
make test-unit
```

**E2E tests** (requires live API credentials):

```bash
# Set required environment variables
export FIZZY_TEST_TOKEN=your-api-token
export FIZZY_TEST_ACCOUNT=your-account-slug

# Build and run e2e tests
make test-e2e
```

Run a specific e2e test:

```bash
make test-run NAME=TestBoardCRUD
```

## License

MIT
