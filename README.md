# Deputy CLI — Team scheduling in your terminal.

Deputy in your terminal. Manage employees, timesheets, rosters, leave, locations, departments, and more.

## Features

- **Authentication** - OAuth browser login or manual token entry, stored securely in keychain
- **Employees** - create, update, terminate, invite, assign locations, manage unavailability
- **Timesheets** - clock in/out, start/end breaks, view timesheet history
- **Rosters** - create shifts, copy weeks, publish, swap shifts
- **Leave** - request, approve, decline leave
- **Locations** - manage locations and their settings
- **Departments** - create and manage operational units
- **Management** - post memos and journal entries
- **Sales** - add and query sales data
- **Webhooks** - configure webhook endpoints
- **Resource API** - query any Deputy resource type directly

## Installation

### Homebrew

```bash
brew install salmonumbrella/tap/deputy-cli
```

### From Source

```bash
go install github.com/salmonumbrella/deputy-cli/cmd/deputy@latest
```

## Quick Start

### 1. Authenticate

**Browser (recommended):**
```bash
deputy auth login
# Opens browser for OAuth flow
```

**Manual:**
```bash
deputy auth add
# Prompts for install name, geo (au/uk/na), and access token
```

### 2. Test Authentication

```bash
deputy auth test
# Calls /me endpoint to verify credentials
```

### 3. View Your Info

```bash
deputy me info
```

## Configuration

### Environment Variables

- `DEPUTY_DEBUG` - Enable debug logging
- `DEPUTY_TOKEN` - Deputy API token (bypasses keychain when set)
- `DEPUTY_INSTALL` - Deputy install name (used if `DEPUTY_BASE_URL` not set)
- `DEPUTY_GEO` - Deputy region subdomain (optional; defaults to `install.deputy.com` if omitted)
- `DEPUTY_BASE_URL` - Override API base URL (host or `/api/v1` URL)
- `DEPUTY_AUTH_SCHEME` - Authorization scheme (default `Bearer`, can be `OAuth`)
- `DEPUTY_NO_KEYCHAIN` - Set to `1` to disable keychain credential lookup (env/.env only)
- `DEPUTY_ENV_FILE` - Path to a `.env` file to load (if set, only this file is loaded)
- `DEPUTY_CREDENTIALS_DIR` - Directory for encrypted file-backend credentials (default `~/.config/deputy/credentials`)
- `DEPUTY_KEYRING_PASSWORD` - Password for encrypted file backend (recommended for CI/systemd/headless)
- `DEPUTY_KEYRING_BACKEND` - Optional backend override: `auto` (default), `file`, `keychain`, `secret-service`, `kwallet`, `keyctl`, `pass`, `wincred`
- `NO_COLOR` - Disable colored output (standard convention)

Dotenv loading order when `DEPUTY_ENV_FILE` is not set:
1. `./.env` (current working directory)
2. `~/.openclaw/.env` (OpenClaw/systemd-friendly location)

### Credential Storage

Credentials are stored securely in your system keyring, with Linux headless fallback:
- **macOS**: Keychain Access
- **Linux desktop**: Secret Service (GNOME Keyring, KWallet)
- **Linux headless / no DBUS session**: Encrypted file backend (`~/.config/deputy/credentials` by default)
- **Windows**: Credential Manager

For OpenClaw deployments, set this in `~/.openclaw/.env`:

```bash
DEPUTY_CREDENTIALS_DIR=~/.openclaw/credentials/deputy
DEPUTY_KEYRING_PASSWORD=choose-a-strong-password
```

You can also skip keyring entirely by using env vars or a local `.env` file:

```bash
cat > .env <<'EOF'
DEPUTY_TOKEN=your_token_here
DEPUTY_INSTALL=your_install
DEPUTY_GEO=na
EOF
deputy --no-keychain auth test
```

### Troubleshooting (412 Precondition Failed)

If `deputy auth test` (or `deputy employees list`) returns `API error 412: request failed`, run with debug enabled so you can see the exact URL being called:

```bash
deputy --no-keychain --debug auth test
```

Common fixes:

```bash
# 1) Your tenant might not use install.<geo>.deputy.com
# Provide a full base URL (host or /api/v1 URL is fine).
export DEPUTY_BASE_URL="https://YOUR_INSTALL.deputy.com/api/v1"
export DEPUTY_TOKEN="..."
deputy --no-keychain --debug auth test

# 2) Some tokens require a different Authorization scheme.
export DEPUTY_AUTH_SCHEME="OAuth"
deputy --no-keychain --debug auth test
```

## Error Handling

### Text Mode (default)

Errors are printed to stderr with human-readable hints:

```bash
$ deputy employees get 999999
Error: API error 404: not found
Hint: Resource not found. Try 'deputy resource list' to verify names.
```

### JSON Mode

When `--output json` is set, errors are structured JSON on stdout:

```bash
$ deputy employees get 999999 --output json
{"error":{"code":"NOT_FOUND","status":404,"message":"not found","retryable":false,"hint":"Resource not found"}}
$ echo $?
1
```

Error codes for programmatic handling:

| Code | Status | Retryable | Description |
|------|--------|-----------|-------------|
| `AUTH_REQUIRED` | 401 | No | Not authenticated |
| `AUTH_FORBIDDEN` | 403 | No | No permission |
| `NOT_FOUND` | 404 | No | Resource not found |
| `CONFLICT` | 409 | No | Resource conflict |
| `VALIDATION_FAILED` | 422 | No | Invalid input |
| `RATE_LIMITED` | 429 | Yes | Too many requests |
| `SERVER_ERROR` | 5xx | Yes | Server error |
| `NETWORK_ERROR` | - | Yes | Connection failed |
| `TIMEOUT` | - | Yes | Request timed out |
| `INVALID_FLAG` | - | No | Bad CLI argument |

## Commands

### Authentication

```bash
deputy auth login                  # Authenticate via browser (OAuth)
deputy auth add                    # Add credentials manually
deputy auth status                 # Show current authentication status
deputy auth test                   # Test credentials
deputy auth logout                 # Remove stored credentials
```

### Employees

```bash
deputy employees list                                    # List all employees
deputy employees get <id>                                # Get employee details
deputy employees add --first-name John --last-name Doe --company 1
deputy employees update <id> --email new@example.com
deputy employees terminate <id> --date 2024-12-31
deputy employees invite <id>                             # Send invitation email
deputy employees assign-location <id> --location 1
deputy employees remove-location <id> --location 1
deputy employees reactivate <id>
deputy employees delete <id>
deputy employees status <id>                             # Check if clocked in/out
deputy employees unavailability <id>                     # Get unavailability
deputy employees add-unavailability <id> --start "2024-01-01" --end "2024-01-07"
deputy employees agreed-hours <id>
```

### Timesheets

```bash
deputy timesheets list [--from <date>] [--to <date>] [--employee <id>]    # List timesheets
deputy timesheets get <id>                               # Get timesheet details
deputy timesheets update <id> --cost <amount>            # Update timesheet cost
deputy timesheets clock-in --employee <id> [--location <id>]
deputy timesheets clock-out --timesheet <id>             # End timesheet by ID (preferred)
deputy timesheets clock-out --employee <id>              # End timesheet by employee
deputy timesheets start-break --timesheet <id>           # Start break
deputy timesheets end-break --timesheet <id>             # End break

# Pay Rules
deputy timesheets list-pay-rules                         # List all pay rules
deputy timesheets list-pay-rules --hourly-rate 190       # Filter by hourly rate
deputy timesheets select-pay-rule <id> --pay-rule <id>   # Assign pay rule to timesheet
```

**Example: Set pay rate for approved timesheets**
```bash
# 1. Find pay rules with $190/hr rate
deputy timesheets list-pay-rules --hourly-rate 190

# 2. Assign pay rule 304 to timesheet 19379
deputy timesheets select-pay-rule 19379 --pay-rule 304
```

### Pay Rates

```bash
# Award library
deputy pay awards list
deputy pay awards get <award-code>
deputy pay awards set <employee-id> --award <code> --country <code> [--override <payRuleId:hourlyRate>]

# Employee agreements (base rate + area config)
deputy pay agreements list --employee <id> [--active-only]
deputy pay agreements get <agreement-id>
deputy pay agreements update <agreement-id> --base-rate 23
deputy pay agreements update <agreement-id> --config '{"DepartmentalPay": []}'
deputy pay agreements update <agreement-id> --config-file ./agreement-config.json
```

Notes:
- Area-specific rates live in the employee agreement `Config` payload (schema varies by tenant). Start with `deputy pay agreements get <agreement-id> -o json` and edit only the relevant keys.
- Award library commands require the Deputy pay-rate library to be enabled for your install.

### Rosters / Shifts

```bash
deputy rosters list                                      # List rosters (12h past + 36h future)
deputy rosters get <id>                                  # Get roster details
deputy rosters create --employee <id> --location <id> --start "2024-01-15T09:00:00" --end "2024-01-15T17:00:00"
deputy rosters copy --from-date 2024-01-08 --to-date 2024-01-15 --location <id>
deputy rosters publish --start 2024-01-15 --end 2024-01-21 --location <id>
deputy rosters discard --start 2024-01-15 --end 2024-01-21 --location <id>
deputy rosters swap <id>                                 # List swap candidates
```

### Leave

```bash
deputy leave list [--status pending|approved|declined]   # List leave requests
deputy leave get <id>                                    # Get leave details
deputy leave add --employee <id> --start "2024-01-15" --end "2024-01-20" --type 1
deputy leave approve <id> [--comment "Approved"]
deputy leave decline <id> [--comment "Insufficient notice"]
```

### Locations

```bash
deputy locations list                                    # List all locations
deputy locations get <id>                                # Get location details
deputy locations add --name "Sydney Office" --code SYD --timezone "Australia/Sydney"
deputy locations update <id> --name "Melbourne Office"
deputy locations archive <id>
deputy locations delete <id>
deputy locations settings <id>                           # Get location settings
deputy locations settings-update <id> --settings '{"key": "value"}'
```

### Departments

```bash
deputy departments list                                  # List all departments
deputy departments get <id>                              # Get department details
deputy departments add --name "Kitchen" --company <id>
deputy departments update <id> --name "Back of House"
deputy departments delete <id>
```

### Current User

```bash
deputy me info                                           # Show your user info
deputy me timesheets                                     # List your timesheets
deputy me rosters                                        # List your rosters
deputy me leave                                          # List your leave requests
```

### Management

```bash
deputy management memo list --company <id>               # List memos
deputy management memo add --company <id> --content "Team meeting" --location <id>
deputy management memo add --company <id> --content "Note" --employee <id>  # Target specific employee
deputy management journal list --employee <id>           # List journal entries
deputy management journal add --employee <id> --company <id> --comment "Performance review notes"
```

### Sales

```bash
deputy sales list --location <id> --from 2024-01-01 --to 2024-01-31
deputy sales add --location <id> --date 2024-01-15 --amount 1500.00
```

### Webhooks

```bash
deputy webhooks list                                     # List all webhooks
deputy webhooks get <id>                                 # Get webhook details
deputy webhooks add --topic Timesheet.Insert --url https://example.com/hook
deputy webhooks delete <id>

# Valid topic formats: {Resource}.{Action}
# Actions: Insert, Update, Save, Delete
# Resources: Employee, Timesheet, Roster, Leave, Comment, Memo, Task, etc.
# Special: User.Login, Roster.Publish, TimesheetExport.Begin/End
```

### Resource API

Query any Deputy resource type directly:

```bash
deputy resource list                                     # List known resource types
deputy resource info Employee                            # Get schema for a resource
deputy resource get Employee 123                         # Get specific resource by ID
deputy resource query Employee --filter "Active=1"       # Query with filters
```

## Output Formats

### Text (default)

Human-readable tables with colors:

```bash
$ deputy employees list
ID    NAME          EMAIL                 ACTIVE
1     John Doe      john@example.com      Yes
2     Jane Smith    jane@example.com      Yes

$ deputy me info
ID:       12345
Name:     John Doe
Email:    john@example.com
Company:  Acme Corp
```

### JSON

Machine-readable output for automation:

```bash
$ deputy employees list --output json
{
  "items": [
    {"Id": 1, "FirstName": "John", "LastName": "Doe", "Email": "john@example.com", "Active": true},
    {"Id": 2, "FirstName": "Jane", "LastName": "Smith", "Email": "jane@example.com", "Active": true}
  ],
  "meta": {
    "count": 2,
    "limit": 0,
    "offset": 0
  }
}
```

### JQ Filtering

Filter JSON output with JQ expressions:

```bash
# Get just employee IDs
deputy employees list --output json --query '[.items[].Id]'

# Filter active employees only
deputy employees list --output json --query '[.items[] | select(.Active == true)]'

# Get employee emails
deputy employees list --output json --query '[.items[].Email]'
```

## Examples

### Clock in/out an employee

```bash
# Get employee and location IDs
deputy employees list
deputy locations list

# Clock in
deputy timesheets clock-in --employee 123 --location 1

# Clock out (use the timesheet ID returned from clock-in)
deputy timesheets clock-out --timesheet 456
```

### Create a weekly roster

```bash
# Create shifts for the week
deputy rosters create --employee 123 --location 1 \
  --start "2024-01-15T09:00:00" --end "2024-01-15T17:00:00"

# Publish the roster
deputy rosters publish --start 2024-01-15 --end 2024-01-21 --location 1
```

### Process leave requests

```bash
# List pending leave requests
deputy leave list --status pending

# Approve a request
deputy leave approve 456 --comment "Enjoy your vacation!"
```

### Automation with JSON

```bash
# Get all active employee IDs
deputy employees list --output json --query '[.items[] | select(.Active) | .Id]'

# Pipeline: get timesheet hours for each employee
for id in $(deputy employees list -o json -q '.items[].Id'); do
  echo "Employee $id:"
  deputy timesheets list --employee $id -o json -q '[.items[].TotalTime] | add'
done
```

### Debug Mode

Enable verbose output for troubleshooting:

```bash
deputy --debug employees list
# Shows: API request/response details
```

## Global Flags

All commands support these flags:

- `--output, -o <format>` - Output format: `text` or `json` (default: text)
- `--query, -q <expr>` - JQ filter expression for JSON output
- `--raw` - Output JSON Lines (one object per line). Implies JSON output if `--output text` is set.
- `--debug` - Enable debug output (shows API requests/responses)
- `--no-color` - Disable colored output
- `--help, -h` - Show help for any command

## Shell Completions

Generate shell completions for your preferred shell:

### Bash

```bash
deputy completion bash > /etc/bash_completion.d/deputy
# Or for macOS with Homebrew:
deputy completion bash > $(brew --prefix)/etc/bash_completion.d/deputy
```

### Zsh

```zsh
deputy completion zsh > "${fpath[1]}/_deputy"
```

### Fish

```fish
deputy completion fish > ~/.config/fish/completions/deputy.fish
```

### PowerShell

```powershell
deputy completion powershell | Out-String | Invoke-Expression
```

## Development

```bash
make build         # Build to bin/deputy
make install       # Install to $GOPATH/bin
make test          # Run tests
make lint          # Run linter
make fmt           # Format code
```

After cloning, install git hooks:

```bash
lefthook install
```

## License

MIT

## Links

- [Deputy API Documentation](https://developer.deputy.com/)
