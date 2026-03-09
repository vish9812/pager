# AGENTS.md

Guidance for AI coding agents working in this repository.

## Project Overview

Go CLI tool (`pager`) for managing PagerDuty schedules and overrides. Uses Cobra for
CLI structure, Charmbracelet huh for interactive prompts, and lipgloss for terminal styling.

- **Module**: `pager`
- **Go version**: 1.25.7
- **Entry point**: `main.go` -> `cmd.Execute()`
- **API reference**: `pager-duty-open-api-v3.json` (OpenAPI v3 spec in repo root)

## Build / Run / Test Commands

```bash
# Build
make build                # or: go build -o pager .

# Run
PAGERDUTY_TOKEN=<token> ./pager override
PAGERDUTY_TOKEN=<token> ./pager oncall
./pager cache clear

# Vet
make vet                  # or: go vet ./...

# Run all tests
make test                 # or: go test ./...

# Run a single test
go test ./pagerduty -run TestFunctionName

# Run tests with verbose output
make test-v               # or: go test -v ./...

# Vet + build together
make check

# Tidy dependencies
make tidy                 # or: go mod tidy

# Remove built binary
make clean
```

No linter config or CI pipeline exists. Use `go vet ./...` as the primary static check.
No test files exist yet -- the test commands above are for when tests are added.

## Project Structure

```
main.go                 Entry point, calls cmd.Execute()
Makefile                Build, test, vet, tidy, clean targets
cmd/
  root.go               Root cobra command, PersistentPreRunE, cache subcommand, Execute()
  oncall.go             "pager oncall" subcommand
  override.go           "pager override" subcommand + override-specific helpers
  shared.go             Shared: data loading, prompts, display, filtering, parsing
cache/
  cache.go              File-based JSON cache (~/.cache/pager/data.json), cleared with 'pager cache clear'
pagerduty/
  client.go             HTTP client with token auth, generic get/post, APIError type
  types.go              All PagerDuty data types and API response wrappers
  users.go              ListUsers method
  schedules.go          ListSchedules, GetSchedule, CreateOverrides methods
```

## Code Style

### Imports

Three groups separated by blank lines, each group sorted alphabetically:

```go
import (
    "fmt"                              // 1. stdlib
    "time"

    "github.com/charmbracelet/huh"     // 2. third-party
    "github.com/spf13/cobra"

    "pager/pagerduty"                  // 3. internal (pager/...)
)
```

Single imports use bare form without parentheses: `import "pager/cmd"`

### Naming

- **Packages**: lowercase single-word (`cmd`, `cache`, `pagerduty`)
- **Exported types/funcs**: PascalCase (`NewClient`, `ListSchedules`, `ScheduleLayerEntry`)
- **Unexported funcs/vars**: camelCase (`runOverride`, `pdClient`)
- **Acronyms**: fully capitalized (`ID`, `HTMLURL`, `APIError`, `baseURL`)
- **Method receivers**: single-letter abbreviation (`c` for `Client`, `e` for `APIError`)
- **Local variables**: short and descriptive (`err`, `resp`, `all`, `params`, `u`, `s`, `e`)
- **JSON tags**: `snake_case` with `omitempty` for optional fields
- **Constants**: unexported camelCase (`baseURL`)

### Types

- All PagerDuty types centralized in `pagerduty/types.go`
- Optional nested objects use pointer types: `FinalSchedule *SubSchedule`
- Response wrappers embed `Pagination`: `type ListUsersResponse struct { Pagination; Users []User }`
- Every struct field has a JSON tag
- Constructor pattern: `func NewXxx(...) *Xxx { return &Xxx{...} }`
- Use `any` (not `interface{}`)

### Error Handling

1. **Wrap with context**: `fmt.Errorf("lowercase context: %w", err)` -- always lowercase,
   always use `%w` for wrapping
2. **Custom error types**: pointer receiver on `Error() string`
3. **Spinner pattern**: separate `xxxErr` for the closure, check spinner err then action err:
   ```go
   var fetchErr error
   err := spinner.New().Title("...").Action(func() {
       result, fetchErr = pdClient.DoThing()
   }).Run()
   if err != nil { return fmt.Errorf("spinner error: %w", err) }
   if fetchErr != nil { return fmt.Errorf("doing thing: %w", fetchErr) }
   ```
4. **Best-effort operations**: `_ = cache.Save(...)` with a comment explaining why
5. **Graceful degradation**: cache `Load()` returns `nil` on any error instead of propagating

### Cobra Commands

- One file per subcommand in `cmd/`
- Self-register in `init()`: `func init() { rootCmd.AddCommand(xxxCmd) }`
- Use `RunE` (not `Run`) to return errors
- Handler naming: `func runXxx(cmd *cobra.Command, args []string) error`
- Always set `Use`, `Short`, `Long`
- Parent commands grouping subcommands (e.g. `cacheCmd`) need only `Use` and `Short`

### Comments

- Doc comments on all exported symbols and non-trivial unexported functions
- Start with the symbol name: `// Client is a PagerDuty API client.`
- Section dividers in longer files: `// --- Section Name ---`
- Step markers in command handlers: `// Step 1: Select schedule`
- Inline comments only for non-obvious logic

### Styling and UI

- `lipgloss` styles as package-level `var` block in `cmd/shared.go`
- Interactive prompts: `charmbracelet/huh` (Select with `.Filtering(true)`, Input, Confirm)
- Loading indicators: `huh/spinner` for API calls
- Shared prompt/display functions live in `cmd/shared.go`, command-specific helpers
  stay in their command file

### HTTP Client

- Custom wrapper in `pagerduty/client.go` with generic `get`/`post` methods
- Auth header: `Authorization: Token token=<value>`
- Accept header: `application/vnd.pagerduty+json;version=2`
- All API methods are on `*Client` receiver, grouped by resource file
- Pagination: loop with `offset`/`limit`, break when `!resp.More`

## Environment

- **Auth**: `PAGERDUTY_TOKEN` env var, validated in `PersistentPreRunE`
- **Cache**: `~/.cache/pager/data.json`, persists until `pager cache clear`; last selected schedule in `~/.cache/pager/preferences.json`
- **No .env files**: token is always from environment
