# pager

CLI tool for managing PagerDuty schedules and overrides.

## Setup

Requires Go 1.25+.

```bash
go build -o pager .
```

Set your PagerDuty API token:

```bash
export PAGERDUTY_TOKEN="your-api-token"
```

To create a User API Token:

1. Click your profile icon in the top-right corner of the PagerDuty web app.
2. Select **My Profile**, then navigate to the **User Settings** tab.
3. In the **API Access** section, click **Create API User Token**.
4. Enter a descriptive name (e.g., `pager CLI`) in the Description field.
5. Click **Create Token**.
6. Copy and securely store the token immediately — it will not be shown again.

## Commands

### `pager oncall`

Check who is on-call for a schedule in a given date range.

```bash
./pager oncall
```

Interactive flow:
1. Select a schedule (filterable list, defaults to last selected)
2. Enter start date (pre-filled with today)
3. Enter end date (pre-filled with start + 7 days)
4. View on-call table showing all users and their shifts

### `pager override`

Override a user's on-call shifts with another user.

```bash
./pager override
```

Interactive flow:
1. Select a schedule
2. Select the user to replace (filterable type-ahead by name/email)
3. Enter start and end dates
4. If the user has no shifts in that range, you'll be informed
5. If found, their current shifts are displayed
6. Select the replacement user
7. Confirm the override
8. Overrides are created and the updated schedule is displayed

### `pager cache clear`

Clear the local cache of schedules and users. The next command will fetch
fresh data from PagerDuty.

```bash
./pager cache clear
```

## Caching

Schedules and users are cached locally at `~/.cache/pager/data.json` since
the data changes infrequently. The first run fetches from the API; subsequent
runs load instantly from cache. The last selected schedule is saved to
`~/.cache/pager/preferences.json` and pre-selected on the next run.

Run `pager cache clear` to force a refresh and reset preferences.

## Development

```bash
make build       # Build the binary
make test        # Run all tests
make test-v      # Run tests with verbose output
make vet         # Run go vet
make tidy        # Tidy module dependencies
make check       # Vet + build
make clean       # Remove built binary
```

## Project Structure

```
main.go              Entry point
cmd/
  root.go            Root command, token validation, cache subcommand
  oncall.go          pager oncall
  override.go        pager override
  shared.go          Shared prompts, display helpers, data loading
cache/
  cache.go           File-based JSON cache with 90-day TTL
pagerduty/
  client.go          HTTP client with token auth
  types.go           API data types
  users.go           User API methods
  schedules.go       Schedule and override API methods
```
