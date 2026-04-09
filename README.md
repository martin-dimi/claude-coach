# Coach

Your personal wellness coach inside Claude Code. Reminds you to move, hydrate and stretch while you work - without leaving your workflow.

Coach lives inside Claude Code as a plugin. When it's time for a break, Claude reminds you naturally and won't continue until you've done it (or skipped). You never learn CLI commands - you just talk to Claude.

## How it works

1. You're working with Claude normally
2. A hook checks if any activities are due
3. Claude interrupts: "Time for 20 pushups! You've done 60 today, 5 day streak."
4. You say "done" (or "skip")
5. Claude logs it and continues with your work

Setup, stats and config changes all happen through conversation with Claude.

## Install

```bash
# 1. Install the binary
brew install fridge/tap/coach

# 2. Load the plugin
claude --plugin-dir /path/to/coach
```

### Build from source

```bash
git clone https://github.com/fridge/coach
cd coach
go build -o bin/coach .
# then: claude --plugin-dir .
```

## Setup

On first use, Claude will ask you what activities you want. Example activities:

- 20 pushups every 1h
- 20 squats every 2h
- Drink water every 30m
- 2min stretch every 2h
- Eye break every 20m

You can also just tell Claude: "set up coach with pushups and water reminders."

### Permissions

When Claude runs `coach done` or `coach skip` for the first time, you'll get a permission prompt. Select "Always allow" for `coach` commands so it doesn't ask again.

## Stats

Run `! coach stats` in Claude Code to see your 30-day contribution grid:

```
pushups  60 today · 1,240 all time
█ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █

water  3x today · 94x all time
█ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █

Mar 10                                                Apr 9

12 day streak
```

Or ask Claude: "how are my coach stats?"

## Config

Config lives at `~/.config/coach/config.toml`:

```toml
[settings]
active_hours = ["09:00", "18:00"]
skip_cooldown = "10m"

[[activities]]
name = "pushups"
reps = 20
interval = "1h"

[[activities]]
name = "water"
message = "Drink a glass of water"
interval = "30m"

[[activities]]
name = "stretch"
duration = "2m"
interval = "2h"
```

Each activity has its own interval. Three types:

| Type | Fields | Example |
|------|--------|---------|
| Reps | `name`, `reps`, `interval` | pushups, squats |
| Duration | `name`, `duration`, `interval` | stretch, plank |
| Message | `name`, `message`, `interval` | water, eye break |

To change config, just tell Claude: "add eye breaks every 20 minutes to coach."

## CLI

The CLI is mostly a backend for Claude, but you can use it directly:

```bash
coach stats           # contribution grids
coach stats --json    # compact JSON
coach done pushups    # log manually
coach skip water      # skip manually
coach reset           # clear all data
```

## License

MIT
