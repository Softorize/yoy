# YOY - Yahoo Mail CLI

A fast, powerful command-line interface for Yahoo Mail. Read, send, search, and manage your Yahoo Mail directly from the terminal using IMAP/SMTP with app password authentication.

Built with Go. Inspired by [GAC](https://github.com/Softorize/gac) and [GOG CLI](https://github.com/steipete/gogcli).

## Features

- **Full mail operations** - list, search, read, send, reply, forward, delete, move, star, mark read/unread
- **Folder management** - list, create, delete mail folders
- **App Password Auth** - secure app password stored in system keyring
- **Multiple output formats** - table (default), JSON, plain/TSV
- **Shell completions** - bash, zsh, fish
- **Cross-platform** - macOS, Linux, Windows

## Installation

### GitHub Releases (recommended)

Download a prebuilt binary from [GitHub Releases](https://github.com/Softorize/yoy/releases). Works out of the box.

### From Source

```bash
git clone https://github.com/Softorize/yoy.git
cd yoy
make build
sudo cp bin/yoy /usr/local/bin/
```

### Verify Installation

```bash
yoy version
# yoy v0.1.0 (commit: abc1234, built: 2026-02-25T00:00:00Z)
```

## Quick Start

### 1. Generate an App Password

Go to [Yahoo Account Security](https://login.yahoo.com/account/security) and generate an app password for "Other App".

### 2. Authenticate

```bash
yoy auth login --email yourname@yahoo.com --app-password YOUR_APP_PASSWORD
```

```
App password stored. Verifying IMAP connection...
Authentication successful.
```

### 3. List Your Inbox

```bash
yoy mail list
```

Output:
```
UID    Date              From              Subject                                             Flags
45123  2026-02-24 14:32  John Smith        Meeting tomorrow at 3pm                             Seen
45122  2026-02-24 12:01  Amazon            Your order has shipped                              Seen,Flagged
45121  2026-02-24 09:45  Jane Doe          Re: Project proposal
45120  2026-02-23 18:20  GitHub            [yoy] New issue: Add attachment support              Seen
```

Unread messages appear in **bold** in terminals that support it.

### 4. Read a Message

```bash
yoy mail read 45121
```

Output:
```
UID:     45121
Date:    2026-02-24 09:45:30 -0500
From:    Jane Doe <jane@example.com>
To:      yourname@yahoo.com
Subject: Re: Project proposal

Hi,

I've reviewed the proposal and it looks great. Let's discuss
the timeline in our next meeting.

Best,
Jane
```

### 5. Send a Message

```bash
yoy send --to jane@example.com --subject "Re: Project proposal" --body "Thanks Jane, see you at the meeting!"
```

## Commands Reference

### Authentication

| Command | Description |
|---------|-------------|
| `yoy auth login --email EMAIL --app-password PASS` | Authenticate with Yahoo |
| `yoy auth logout` | Remove stored credentials |
| `yoy auth status` | Show current authentication status |

**Examples:**

```bash
# Login
yoy auth login --email yourname@yahoo.com --app-password YOUR_APP_PASSWORD

# Check if you're authenticated
yoy auth status
# Status: authenticated
# Email:  yourname@yahoo.com
# Method: app password

# Logout
yoy auth logout
```

### Mail Operations

| Command | Description |
|---------|-------------|
| `yoy mail list` | List messages in a folder |
| `yoy mail search QUERY` | Search messages by subject or sender |
| `yoy mail read UID` | Read a full message |
| `yoy mail send` | Send a new email |
| `yoy mail reply UID` | Reply to a message |
| `yoy mail forward UID` | Forward a message |
| `yoy mail delete UID` | Delete a message |
| `yoy mail move UID FOLDER` | Move a message to another folder |
| `yoy mail star UID` | Star (flag) a message |
| `yoy mail unstar UID` | Remove star from a message |
| `yoy mail mark-read UID` | Mark a message as read |
| `yoy mail mark-unread UID` | Mark a message as unread |

#### Listing Messages

```bash
# List latest 25 messages in INBOX (default)
yoy mail list

# List latest 10 messages
yoy mail list -n 10

# List messages in a specific folder
yoy mail list -f "Sent"
yoy mail list -f "Draft"
yoy mail list -f "Trash"

# Shorthand alias
yoy ls
yoy ls -n 5
```

#### Searching Messages

```bash
# Search by subject or sender
yoy mail search "invoice"
yoy mail search "john@example.com"

# Search in a specific folder
yoy -f "Sent" mail search "project"

# Shorthand alias
yoy search "meeting"
```

#### Reading Messages

```bash
# Read a message by its UID (shown in mail list output)
yoy mail read 45121

# Read from a specific folder
yoy -f "Sent" mail read 45200
```

#### Sending Email

```bash
# Basic send
yoy mail send --to recipient@example.com --subject "Hello" --body "Hi there!"

# Send to multiple recipients
yoy mail send --to "alice@example.com,bob@example.com" --subject "Team update" --body "Meeting moved to 4pm"

# Send with CC and BCC
yoy mail send \
  --to recipient@example.com \
  --cc "manager@example.com" \
  --bcc "archive@example.com" \
  --subject "Report" \
  --body "Please find the weekly report attached."

# Shorthand alias
yoy send --to friend@example.com --subject "Hey" --body "What's up?"
```

#### Replying to Messages

```bash
# Reply to sender only
yoy mail reply 45121 --body "Thanks for the update!"

# Reply to all recipients
yoy mail reply 45121 --body "Sounds good, see you all there." --all
```

The reply automatically:
- Sets the subject to "Re: <original subject>"
- Sets In-Reply-To and References headers
- Marks the original message as read

#### Forwarding Messages

```bash
# Forward a message
yoy mail forward 45121 --to colleague@example.com

# Forward with additional comments
yoy mail forward 45121 --to colleague@example.com --body "FYI - take a look at this"

# Forward to multiple people
yoy mail forward 45121 --to "alice@example.com,bob@example.com" --body "Sharing this with the team"
```

#### Managing Messages

```bash
# Delete a message (marks as deleted and expunges)
yoy mail delete 45120

# Move to a folder
yoy mail move 45121 "Archive"
yoy mail move 45122 "Work/Projects"

# Star / unstar
yoy mail star 45121
yoy mail unstar 45121

# Mark as read / unread
yoy mail mark-read 45121
yoy mail mark-unread 45121
```

### Folder Management

| Command | Description |
|---------|-------------|
| `yoy folders list` | List all folders with message counts |
| `yoy folders create NAME` | Create a new folder |
| `yoy folders delete NAME` | Delete a folder |

**Examples:**

```bash
# List all folders
yoy folders list
# Name       Messages  Unseen
# Archive    142       0
# Draft      3         3
# INBOX      1247      12
# Sent       891       0
# Trash      45        0

# Create a new folder
yoy folders create "Work"
yoy folders create "Work/Projects"

# Delete a folder
yoy folders delete "OldStuff"
```

### Configuration

| Command | Description |
|---------|-------------|
| `yoy config list` | Show all config values |
| `yoy config get KEY` | Get a specific config value |
| `yoy config set KEY VALUE` | Set a config value |
| `yoy config path` | Print config file location |

**Examples:**

```bash
# View all settings
yoy config list
# Field           Value
# color_mode      auto
# default_folder  INBOX
# mail_limit      25
# output_format   table

# Change default number of messages shown
yoy config set mail_limit 50

# Change default folder
yoy config set default_folder "INBOX"

# Set output to always JSON
yoy config set output_format json

# Disable colors
yoy config set color_mode never

# Show config file path
yoy config path
# /Users/yourname/Library/Application Support/yoy/config.yaml
```

### Version and Completions

```bash
# Print version
yoy version

# Generate shell completions
yoy completion bash >> ~/.bashrc
yoy completion zsh >> ~/.zshrc
yoy completion fish > ~/.config/fish/completions/yoy.fish
```

## Global Flags

These flags work with any command:

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--folder` | `-f` | Mail folder to operate on | `INBOX` |
| `--json` | | Output as JSON | `false` |
| `--plain` | | Output as plain TSV (no colors/borders) | `false` |
| `--color` | | Color mode: `auto`, `always`, `never` | `auto` |
| `--verbose` | `-v` | Enable verbose output | `false` |

**Examples:**

```bash
# List Sent folder as JSON
yoy --json -f "Sent" mail list

# Search with plain output (good for piping)
yoy --plain mail search "invoice" | cut -f3  # extract sender column

# Pipe message UIDs
yoy --plain mail search "newsletter" | tail -n +2 | cut -f1  # get UIDs only
```

## Output Formats

### Table (default)

Human-readable formatted table with optional colors:

```bash
yoy mail list
```

### JSON

Machine-readable JSON, useful for scripting:

```bash
yoy --json mail list
yoy --json mail read 45121
yoy --json folders list
```

### Plain / TSV

Tab-separated values, great for piping to other tools:

```bash
yoy --plain mail list | awk -F'\t' '{print $4}'  # print subjects only
yoy --plain folders list | sort -t$'\t' -k2 -rn   # sort by message count
```

## Configuration File

Location: `~/Library/Application Support/yoy/config.yaml` (macOS)

| Key | Default | Description |
|-----|---------|-------------|
| `output_format` | `table` | Default output format (`table`, `json`, `plain`) |
| `color_mode` | `auto` | Color output (`auto`, `always`, `never`) |
| `default_folder` | `INBOX` | Default mail folder for all operations |
| `mail_limit` | `25` | Default number of messages to show in list |

You can edit this file directly or use `yoy config set`:

```yaml
# ~/Library/Application Support/yoy/config.yaml
output_format: table
color_mode: auto
default_folder: INBOX
mail_limit: 50
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `YOY_FOLDER` | Override default mail folder |
| `YOY_JSON` | Set to `true` for JSON output |
| `YOY_PLAIN` | Set to `true` for plain/TSV output |
| `YOY_COLOR` | Override color mode |
| `YOY_OUTPUT_FORMAT` | Override output format |
| `YOY_CONFIG_DIR` | Custom config directory path |
| `NO_COLOR` | Disable all colors ([no-color.org](https://no-color.org)) |

## Credential Storage

App passwords are stored securely:

1. **System keyring** (preferred) - macOS Keychain, GNOME Keyring, Windows Credential Manager
2. **File fallback** - `~/Library/Application Support/yoy/tokens/credentials.json` (mode 0600)

## Aliases

For convenience, these top-level aliases are available:

| Alias | Equivalent |
|-------|-----------|
| `yoy send` | `yoy mail send` |
| `yoy ls` | `yoy mail list` |
| `yoy search` | `yoy mail search` |

## Building from Source

```bash
git clone https://github.com/Softorize/yoy.git
cd yoy
make build       # Build to bin/yoy
make install     # Install to $GOPATH/bin
make test        # Run tests
make vet         # Run go vet
make fmt         # Format code
make clean       # Remove build artifacts
make all         # Format + vet + test + build
```

## Technical Details

- **Protocol**: IMAP (read) + SMTP (send) over TLS
- **IMAP Server**: `imap.mail.yahoo.com:993`
- **SMTP Server**: `smtp.mail.yahoo.com:465`
- **Auth Mechanism**: LOGIN/PLAIN with app password
- **CLI Framework**: [Kong](https://github.com/alecthomas/kong)
- **Credential Storage**: [99designs/keyring](https://github.com/99designs/keyring)

## License

MIT - see [LICENSE](LICENSE)
