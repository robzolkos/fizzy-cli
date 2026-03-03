---
name: context-linker
description: |
  Automatically link code changes to Fizzy cards.
  Use when: committing code, creating PRs, resolving issues.
  Detects card numbers from branch names, commit messages, and PR descriptions.
---

# Context Linker Agent

Connect code changes to Fizzy cards and discussions.

## Detection Patterns

Look for Fizzy references in:

1. **Branch names**:
   - `feature/card-42-description` -> card 42
   - `fix/FIZZY-42-auth-bug` -> card 42
   - `42-feature-name` -> card 42

2. **Commit messages**:
   - `[FIZZY-42] Fix authentication` -> card 42
   - `[card:42] Update docs` -> card 42

3. **PR descriptions**:
   - `Closes FIZZY-42` -> card 42
   - `Related: card #42` -> card 42

## Commands

```bash
# Link current commit to a card
fizzy comment create --card 42 --body "Commit $(git rev-parse --short HEAD): $(git log -1 --format=%s)"

# Link a PR
fizzy comment create --card 42 --body "PR: <pr_url>"

# Close a card
fizzy card close 42
```

## Workflow: On Commit

When user commits code:

1. Extract card number from branch name:
   ```bash
   BRANCH=$(git branch --show-current)
   CARD_NUM=$(echo "$BRANCH" | grep -oE '(FIZZY-|card-?|^)[0-9]+' | grep -oE '[0-9]+' | head -1)
   ```

2. If found, offer to link:
   ```bash
   COMMIT=$(git rev-parse --short HEAD)
   MSG=$(git log -1 --format=%s)
   fizzy comment create --card $CARD_NUM --body "Commit $COMMIT: $MSG"
   ```

## Workflow: On PR Creation

When user creates a PR:

1. Check branch name and PR description for card references
2. For each referenced card, add PR link as comment
3. Offer to close cards if PR is merged

## Workflow: On Merge

When a PR is merged:

1. Find all referenced cards
2. Offer to close them:
   ```bash
   fizzy card close 42
   ```

## Project Context

Check for `.fizzy.yaml` to get default board:
```bash
# If .fizzy.yaml exists with a board setting, card commands
# will automatically scope to that board
cat .fizzy.yaml
```

This enables card operations without explicit `--board` flags.

## Example Session

```
User: I just committed the auth fix

Agent: I see you're on branch `fix/FIZZY-42-auth-bug`.
       The commit is: abc1234 "Fix OAuth token refresh"

       Would you like me to link this commit to card #42?

User: yes

Agent: [runs: fizzy comment create --card 42 --body "Commit abc1234: Fix OAuth token refresh"]
       Done! Comment added to card #42.

       Should I close this card?
```
