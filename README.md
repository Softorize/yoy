# yoy - Yahoo Mail CLI

A command-line interface for Yahoo Mail using IMAP/SMTP with OAuth2 XOAUTH2 authentication.

## Installation

```bash
go install github.com/Softorize/yoy@latest
```

Or build from source:

```bash
git clone https://github.com/Softorize/yoy.git
cd yoy
make build
```

## Setup

### OAuth2 Credentials

Set your Yahoo OAuth2 credentials:

```bash
export YOY_CLIENT_ID=your-client-id
export YOY_CLIENT_SECRET=your-client-secret
```

### Authentication

```bash
yoy auth login --email your@yahoo.com
```

This opens your browser for Yahoo OAuth2 authentication.

## Usage

### Mail Operations

```bash
# List messages
yoy mail list
yoy mail list -f "Sent" -n 10
yoy ls

# Search messages
yoy mail search "invoice"
yoy search "from:john"

# Read a message
yoy mail read 12345

# Send an email
yoy mail send --to recipient@example.com --subject "Hello" --body "Hi there"
yoy send --to recipient@example.com --subject "Test" --body "Hello"

# Reply to a message
yoy mail reply 12345 --body "Thanks!"
yoy mail reply 12345 --body "Thanks!" --all

# Forward a message
yoy mail forward 12345 --to friend@example.com --body "FYI"

# Delete a message
yoy mail delete 12345

# Move a message
yoy mail move 12345 "Archive"

# Star/unstar
yoy mail star 12345
yoy mail unstar 12345

# Mark read/unread
yoy mail mark-read 12345
yoy mail mark-unread 12345
```

### Folder Management

```bash
yoy folders list
yoy folders create "Projects"
yoy folders delete "OldFolder"
```

### Configuration

```bash
yoy config list
yoy config get default_folder
yoy config set mail_limit 50
yoy config path
```

### Output Formats

```bash
yoy mail list --json          # JSON output
yoy mail list --plain         # Tab-separated values
yoy mail list --color=never   # Disable colors
```

## Configuration

Config file: `~/Library/Application Support/yoy/config.yaml`

| Key | Default | Description |
|-----|---------|-------------|
| `output_format` | `table` | Default output format (table, json, plain) |
| `color_mode` | `auto` | Color mode (auto, always, never) |
| `oauth_port` | `8086` | Local OAuth callback port |
| `default_folder` | `INBOX` | Default mail folder |
| `mail_limit` | `25` | Default number of messages to list |

## License

MIT
