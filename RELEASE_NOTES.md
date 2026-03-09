## pager v0.1.0

A terminal-first CLI for managing PagerDuty schedules and overrides without touching the web UI.

### Features

**`pager oncall`**
View who is on-call for any schedule over a custom date range. Shows a table of users, shift start/end times, and durations.

**`pager override`**
Interactively replace a user's on-call shifts with another user. Select a schedule, pick who to replace and who replaces them, confirm, and the overrides are created in bulk via the PagerDuty API.

**`pager cache clear`**
Clear locally cached schedule and user data.

**`pager cache path`**
Print the cache directory path.

### Requirements

- A PagerDuty API token set as `PAGERDUTY_TOKEN` in your environment

### Installation

Download the binary for your platform from the assets below, make it executable, and place it on your `$PATH`:

```sh
chmod +x pager
mv pager /usr/local/bin/pager
```

### Usage

```sh
export PAGERDUTY_TOKEN=your_token_here

pager oncall
pager override
```
