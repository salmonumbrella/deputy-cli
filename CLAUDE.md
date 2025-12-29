# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A Go CLI for the Deputy workforce management API, designed for AI agent automation. Supports dual output modes: tables for humans and JSON for AI agents (`--output json`).

## Build & Development Commands

```bash
make build         # Build to bin/deputy
make install       # Install to $GOPATH/bin
make test          # Run tests with race detection and coverage
make lint          # Run golangci-lint
make fmt           # Format with goimports + gofumpt
make deps          # Download and tidy dependencies
```

Version info is injected via ldflags at build time (Version, CommitSHA, BuildDate).

## Architecture

### Package Structure

- `cmd/deputy/main.go` - Entry point, calls `cmd.Execute()`
- `internal/cmd/` - Cobra commands (one file per resource: auth, employees, timesheets, etc.)
- `internal/api/` - HTTP client and service methods for each Deputy API resource
- `internal/secrets/` - Keychain credential storage (`KeychainStore` + `MockStore` for testing)
- `internal/outfmt/` - Output formatting (table/JSON) with jq filter support via gojq
- `internal/iocontext/` - Context-based stdin/stdout/stderr injection for testability
- `internal/auth/` - Browser-based authentication server

### Key Patterns

**API Client Pattern:** Each resource has a service struct accessed via fluent methods:
```go
client.Employees().List(ctx)
client.Employees().Get(ctx, id)
client.Timesheets().ClockIn(ctx, input)
```

**Context-Based IO:** Commands get IO streams from context, not globals:
```go
io := iocontext.FromContext(cmd.Context())
fmt.Fprintf(io.Out, "message\n")
```

**Output Format Detection:** Check format from context, branch to JSON or table:
```go
format := outfmt.GetFormat(cmd.Context())
if format == "json" {
    return outfmt.New(cmd.Context()).Output(data)
}
// otherwise use table output
```

**Credential Flow:** All API commands use `getClient()` from `internal/cmd/helpers.go` which loads credentials from keychain.

### Deputy API Notes

- Base URL: `https://{install}.{geo}.deputy.com/api/v1`
- Geographic regions: `au`, `uk`, `na`
- Auth: Bearer token in Authorization header
- Credentials stored in macOS Keychain under service name `deputy-cli`

### Global Flags

All commands inherit:
- `--output/-o` (text|json) - Output format
- `--query/-q` - jq filter for JSON output
- `--debug` - Enable debug logging (shows HTTP requests/responses)
- `--no-color` - Disable colored output

List commands also support:
- `--limit` - Maximum number of results (0 = unlimited)
- `--offset` - Number of results to skip

### Available Commands

```
deputy
├── auth          # login, logout, status, add, test
├── completion    # Generate shell completion (bash, zsh, fish, powershell)
├── employees     # list, get, add, update, terminate, invite,
│                 # assign-location, remove-location, reactivate, delete,
│                 # add-unavailability
├── timesheets    # list, get, clock-in, clock-out, start-break, end-break
├── rosters       # list, get, create, copy, publish, discard, swap
├── locations     # list, get, add, update, archive, delete, settings, settings-update
├── leave         # list, get, add, approve, decline
├── departments   # list, get, add, update, delete
├── resource      # list, info, query, get
├── me            # info, timesheets, rosters, leave
├── webhooks      # list, get, add, delete
├── sales         # list, add
├── management    # memo (list, add), journal (list, add)
└── version       # Print version information
```

### API Service Coverage

| Service | Methods |
|---------|---------|
| Employees | List, Get, Create, Update, Terminate, Invite, AssignLocation, RemoveLocation, Reactivate, Delete, AddUnavailability |
| Timesheets | List, Get, ClockIn, ClockOut, StartBreak, EndBreak |
| Rosters | List, Get, Create, Copy, Publish, Discard, GetSwappable |
| Locations | List, Get, Create, Update, Archive, Delete, GetSettings, UpdateSettings |
| Leave | List, Get, Create, Update, Approve, Decline, Query |
| Departments | List, Get, Create, Update, Delete |
| Resource | Info, Query, Get, List |
| Me | Info, Timesheets, Rosters, Leave |
| Webhooks | List, Get, Create, Delete |
| Sales | List, Add, Query |
| Management | CreateMemo, ListMemos, PostJournal, ListJournals |

## Testing Notes

Use `secrets.MockStore` instead of real keychain in tests. IO streams should be injected via `iocontext.WithIO()` for capturing command output.
