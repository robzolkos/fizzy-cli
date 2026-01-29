# Fizzy CLI

A command-line interface for the [Fizzy](https://fizzy.do) API. See the official [API docs](https://github.com/basecamp/fizzy/blob/main/docs/API.md).

## Installation

**Arch Linux (AUR)**
```bash
yay -S fizzy-cli
```

**macOS (Homebrew)**
```bash
brew install robzolkos/fizzy-cli/fizzy-cli
```

**Debian/Ubuntu**
```bash
# Download the .deb for your architecture (amd64 or arm64)
curl -LO https://github.com/robzolkos/fizzy-cli/releases/latest/download/fizzy-cli_VERSION_amd64.deb
sudo dpkg -i fizzy-cli_VERSION_amd64.deb
```

**Fedora/RHEL**
```bash
# Download the .rpm for your architecture (x86_64 or aarch64)
curl -LO https://github.com/robzolkos/fizzy-cli/releases/latest/download/fizzy-cli-VERSION-1.x86_64.rpm
sudo rpm -i fizzy-cli-VERSION-1.x86_64.rpm
```

**Windows**

Download `fizzy-windows-amd64.exe` from [GitHub Releases](https://github.com/robzolkos/fizzy-cli/releases), rename it to `fizzy.exe`, and add it to your PATH.

**With Go**
```bash
go install github.com/robzolkos/fizzy-cli/cmd/fizzy@latest
```

**From binary**

Download the latest release for your platform from [GitHub Releases](https://github.com/robzolkos/fizzy-cli/releases) and add it to your PATH.

**From source**
```bash
git clone https://github.com/robzolkos/fizzy-cli.git
cd fizzy-cli
go build -o fizzy ./cmd/fizzy
./fizzy --help
```

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

## Quick Start

1. Get your API token from My Profile → Personal Access Tokens (see [instructions](https://github.com/basecamp/fizzy/blob/main/docs/API.md#personal-access-tokens))

2. Run the interactive setup wizard:

```bash
fizzy setup
```

The wizard will guide you through configuring your token, selecting your account, and optionally setting a default board.

That's it! Try `fizzy board list` to verify everything is working.

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
fizzy card list --column COLUMN_ID
fizzy card list --column maybe
fizzy card list --column done
fizzy card list --tag TAG_ID
fizzy card list --indexed-by not_now
fizzy card list --assignee USER_ID
fizzy card list --sort newest    # oldest cards first (by created_at)
fizzy card list --sort oldest    # newest cards first (by created_at)
fizzy card list --sort latest    # most recently updated (default)

# Additional filters
fizzy card list --search "bug"           # Search by text
fizzy card list --sort newest            # Sort: newest, oldest, latest (default)
fizzy card list --creator USER_ID        # Filter by creator
fizzy card list --closer USER_ID         # Filter by who closed
fizzy card list --unassigned             # Only unassigned cards
fizzy card list --created thisweek       # Created: today, yesterday, thisweek, lastweek, thismonth, lastmonth
fizzy card list --closed thisweek        # Closed: today, yesterday, thisweek, lastweek, thismonth, lastmonth

# Tip: if you set a default `board` in config (or `FIZZY_BOARD`), `fizzy card list` automatically filters to that board unless you pass `--board`.

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
fizzy card update 42 --image SIGNED_ID
fizzy card update 42 --created-at "2019-01-01T00:00:00Z"

# Delete a card
fizzy card delete 42
```

### Card Statuses

Cards in Fizzy exist in different states. By default, `fizzy card list` returns **open cards only** (cards in triage or columns). To fetch cards in other states, use the `--indexed-by` or `--column` flags:

| Status | How to fetch | Description |
|--------|--------------|-------------|
| Open (default) | `fizzy card list` | Cards in triage ("Maybe?") or any column |
| Closed/Done | `fizzy card list --indexed-by closed` | Completed cards |
| Not Now | `fizzy card list --indexed-by not_now` | Postponed cards |
| Golden | `fizzy card list --indexed-by golden` | Starred/important cards |
| Stalled | `fizzy card list --indexed-by stalled` | Cards with no recent activity |

You can also use pseudo-columns:

```bash
fizzy card list --column done --all     # Same as --indexed-by closed
fizzy card list --column not-now --all  # Same as --indexed-by not_now
fizzy card list --column maybe --all    # Cards in triage (no column assigned)
```

**Fetching all cards on a board:**

To get all cards regardless of status (for example, to build a complete board view), you need to make separate queries and combine the results:

```bash
# Open cards (triage + columns)
fizzy card list --board BOARD_ID --all

# Closed/Done cards
fizzy card list --board BOARD_ID --indexed-by closed --all

# Optionally, Not Now cards
fizzy card list --board BOARD_ID --indexed-by not_now --all
```

### Card Actions

```bash
# Close/reopen
fizzy card close 42
fizzy card reopen 42

# Move to a different board
fizzy card move 42 --to BOARD_ID
fizzy card move 42 -t BOARD_ID

# Move to "Not Now"
fizzy card postpone 42

# Move into a column
fizzy card column 42 --column COLUMN_ID

# Move into UI lanes (pseudo columns)
fizzy card column 42 --column not-now
fizzy card column 42 --column maybe
fizzy card column 42 --column done

# Send back to triage
fizzy card untriage 42

# Assign/unassign (toggles)
fizzy card assign 42 --user USER_ID

# Tag/untag (toggles, creates tag if needed)
fizzy card tag 42 --tag "bug"

# Watch/unwatch
fizzy card watch 42
fizzy card unwatch 42

# Remove card header image
fizzy card image-remove 42

# Pin/unpin a card
fizzy card pin 42
fizzy card unpin 42

# Mark/unmark card as golden
fizzy card golden 42
fizzy card ungolden 42
```

### Card Attachments

```bash
# List attachments on a card
fizzy card attachments show 42

# Download all attachments from a card
fizzy card attachments download 42

# Download a specific attachment by index (1-based)
fizzy card attachments download 42 1

# Download with a custom filename
fizzy card attachments download 42 1 -o my-file.png
```

### Columns

```bash
fizzy column list --board BOARD_ID
fizzy column show COLUMN_ID --board BOARD_ID
fizzy column create --board BOARD_ID --name "In Progress"
fizzy column update COLUMN_ID --board BOARD_ID --name "Done"
fizzy column delete COLUMN_ID --board BOARD_ID
```

`fizzy column list` also includes the UI's built-in lanes as pseudo columns in this order:
- `not-now` (Not Now)
- `maybe` (Maybe?)
- your real columns…
- `done` (Done)

When filtering cards by `--column maybe` (triage) or a real column ID, the CLI filters client-side; use `--all` to fetch all pages before filtering.

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
# List reactions on a card
fizzy reaction list --card 42

# List reactions on a comment
fizzy reaction list --card 42 --comment COMMENT_ID

# Add a reaction to a card (max 16 chars)
fizzy reaction create --card 42 --content "👍"

# Add a reaction to a comment (max 16 chars)
fizzy reaction create --card 42 --comment COMMENT_ID --content "👍"

# Remove a reaction from a card
fizzy reaction delete REACTION_ID --card 42

# Remove a reaction from a comment
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

### Pins

```bash
# List your pinned cards
fizzy pin list
```

### Search

```bash
# Search cards by text
fizzy search "bug"
fizzy search "login error"              # Multiple terms (AND)

# Combine with filters
fizzy search "bug" --board BOARD_ID
fizzy search "bug" --tag TAG_ID
fizzy search "bug" --assignee USER_ID
fizzy search "bug" --indexed-by closed  # Include closed cards
fizzy search "bug" --sort newest        # Sort by created_at desc
```

### Notifications

```bash
fizzy notification list
fizzy notification read NOTIFICATION_ID
fizzy notification unread NOTIFICATION_ID
fizzy notification read-all
```

### File Uploads

Upload files for use in rich text fields (card descriptions, comment bodies) or as card header images.

```bash
# Upload a file
fizzy upload file /path/to/image.png
# Returns: { "signed_id": "...", "attachable_sgid": "..." }
```

The upload returns two IDs for different purposes:

| ID | Use Case |
|----|----------|
| `signed_id` | Card header images (`--image` flag) |
| `attachable_sgid` | Inline images in rich text (`<action-text-attachment>`) |

**Header image:**
```bash
SIGNED_ID=$(fizzy upload file header.png | jq -r '.data.signed_id')
fizzy card create --board BOARD_ID --title "Card" --image "$SIGNED_ID"
```

**Inline image in description:**
```bash
SGID=$(fizzy upload file image.png | jq -r '.data.attachable_sgid')
cat > description.html << EOF
<p>See image:</p>
<action-text-attachment sgid="$SGID"></action-text-attachment>
EOF
fizzy card create --board BOARD_ID --title "Card" --description_file description.html
```

> **Note:** Each `attachable_sgid` can only be used once. Upload the file again if you need to attach it to multiple cards.

### Identity

```bash
# Show your identity and all accessible accounts
fizzy identity show
```

### Board Migration

Migrate boards and cards between accounts. Useful when you've created a board in your personal account and want to transfer it to a team account.

```bash
# Preview what would be migrated (dry run)
fizzy migrate board BOARD_ID --from SOURCE_ACCOUNT --to TARGET_ACCOUNT --dry-run

# Migrate a board with all cards
fizzy migrate board BOARD_ID --from SOURCE_ACCOUNT --to TARGET_ACCOUNT

# Include card header images
fizzy migrate board BOARD_ID --from SOURCE_ACCOUNT --to TARGET_ACCOUNT --include-images

# Include comments and inline attachments
fizzy migrate board BOARD_ID --from SOURCE_ACCOUNT --to TARGET_ACCOUNT --include-images --include-comments

# Include steps (to-do items)
fizzy migrate board BOARD_ID --from SOURCE_ACCOUNT --to TARGET_ACCOUNT --include-steps
```

| Flag | Description |
|------|-------------|
| `--from` | Source account slug (required) |
| `--to` | Target account slug (required) |
| `--include-images` | Migrate card header images and inline attachments |
| `--include-comments` | Migrate card comments |
| `--include-steps` | Migrate card steps (to-do items) |
| `--dry-run` | Preview migration without making changes |

**What gets migrated:**
- Board with the same name
- All columns (preserving order and colors)
- All cards with titles, descriptions, timestamps, and tags
- Card states (closed, golden, column placement)
- Optionally: header images, inline attachments, comments, and steps

**What cannot be migrated:**
- Card creators (become the migrating user)
- Card numbers (new sequential numbers in target)
- Comment authors (become the migrating user)
- User assignments (team will need to reassign)

> **Note:** You must have access to both source and target accounts with your API token. Use `fizzy identity show` to see your accessible accounts.

### Skill Installation

Install the Fizzy skill file ([SKILL.md](skills/fizzy/SKILL.md)) for use with AI coding assistants like Codex, Claude Code, or OpenCode.

```bash
fizzy skill
```

This interactive command downloads the latest `SKILL.md` from this repository and lets you choose where to install it:

| Location | Path |
|----------|------|
| Claude Code (Global) | `~/.claude/skills/fizzy/SKILL.md` |
| Claude Code (Project) | `.claude/skills/fizzy/SKILL.md` |
| OpenCode (Global) | `~/.config/opencode/skill/fizzy/SKILL.md` |
| OpenCode (Project) | `.opencode/skill/fizzy/SKILL.md` |
| Codex (Global) | `~/.codex/skills/fizzy/SKILL.md` (or `$CODEX_HOME/skills/fizzy/SKILL.md`) |
| Other | Custom path of your choice |

The skill file enables AI assistants to understand and use Fizzy CLI commands effectively.

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

### Summary Field

List and show commands include a `summary` field with a human-readable description of the response. This is useful for quick feedback in scripts or when piping output.

```bash
fizzy board list | jq -r '.summary'
# "5 boards"

fizzy card list --board ABC --all | jq -r '.summary'
# "42 cards (all)"

fizzy search "bug" | jq -r '.summary'
# "7 results for \"bug\""
```

The summary adapts to pagination flags (`--page N` or `--all`) and includes contextual details like unread counts for notifications.

### Breadcrumbs

Command responses include a `breadcrumbs` array with suggested next actions. This is designed for AI agents (and humans) to discover contextual workflows without needing to know the full CLI.

```bash
fizzy card show 42 | jq '.breadcrumbs'
```

```json
[
  {"action": "comment", "cmd": "fizzy comment create --card 42 --body \"text\"", "description": "Add comment"},
  {"action": "triage", "cmd": "fizzy card column 42 --column <column_id>", "description": "Move to column"},
  {"action": "close", "cmd": "fizzy card close 42", "description": "Close card"},
  {"action": "assign", "cmd": "fizzy card assign 42 --user <user_id>", "description": "Assign user"}
]
```

Each breadcrumb contains:
- `action`: A short identifier for the action type
- `cmd`: The complete CLI command to execute
- `description`: Human-readable description

Breadcrumbs are included by default in all responses. They are contextual - after creating a card you'll see suggestions to view, triage, or comment on it; after listing cards you'll see suggestions to show a specific card, create a new one, or search.

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

## Pagination

List commands return paginated results. Use `--page` to fetch specific pages or `--all` to fetch all pages:

```bash
fizzy card list --page 2
fizzy card list --all      # Fetches all pages of the current filter
```

> **Note:** The `--all` flag controls pagination only - it fetches all pages of results for your current filter. It does not change which cards are included. See [Card Statuses](#card-statuses) for how to fetch closed or postponed cards.

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
