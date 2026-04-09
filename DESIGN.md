# Coach - Design Document

Your personal wellness coach inside Claude Code. Reminds you to move, hydrate and stretch while you work - without leaving your workflow.

## Philosophy

Coach is not a habit tracker. It's a Claude-native experience. The user never learns CLI commands - they just talk to Claude. The CLI is a silent backend engine. Claude is the interface.

## Architecture

```
┌─────────────────────────────────────────────────────┐
│  Claude Code                                        │
│                                                     │
│  ┌─────────────┐    ┌──────────┐    ┌────────────┐  │
│  │    Hook      │───>│  coach   │───>│  Claude    │  │
│  │ (trigger)    │    │  check   │    │  (brain)   │  │
│  │              │    │          │    │            │  │
│  │ UserPrompt   │    │ Returns  │    │ Reads      │  │
│  │ Submit       │    │ JSON     │    │ SKILL.md   │  │
│  └─────────────┘    └──────────┘    └────────────┘  │
│                                                     │
│  User just talks to Claude. Claude handles the rest │
└─────────────────────────────────────────────────────┘
```

Three layers:

| Layer | Role | What it does |
|-------|------|--------------|
| **Hook** (UserPromptSubmit) | Trigger | Fires on every user prompt. Calls `coach check`. Injects context into Claude when a break is due or setup is needed. |
| **Skill** (SKILL.md) | Brain | Tells Claude how to be a good coach - tone, behavior, how to handle done/skip/setup/stats. Loaded by Claude when it sees coach-related context. |
| **Go binary** (coach) | Engine | Manages data (SQLite), config (TOML), timers and stats. Pure backend. |

## Installation

### Path A: Claude Code plugin (preferred)

```
claude plugin install coach
```

Everything just works. The plugin bundles the binary (`bin/coach`), hook (`hooks/hooks.json`), and skill (`skills/coach/SKILL.md`). No further setup.

Plugin structure:
```
coach/
├── .claude-plugin/
│   └── plugin.json
├── skills/
│   └── coach/
│       └── SKILL.md
├── hooks/
│   └── hooks.json
├── bin/
│   └── coach
└── scripts/
    └── check.sh
```

### Path B: Standalone binary

```
go install github.com/fridge/coach@latest
coach install
```

`coach install` writes the hook to `~/.claude/settings.json`. Also works without Claude Code - just the CLI.

## How it works

### The hook: `UserPromptSubmit`

The hook fires every time the user sends a prompt. It calls `coach check`, which returns one of three responses:

**1. No break due (most common):**
```
exit 0, no output
```
Zero context added by the hook. Claude still knows coach exists via the skill metadata (~50-100 tokens, always present) - so the user can ask about coach anytime - but the hook adds nothing.

**2. Break is due:**
```json
exit 0
{
  "hookSpecificOutput": {
    "hookEventName": "UserPromptSubmit",
    "additionalContext": "[COACH] Break reminder:\nDue: 20 pushups, Drink a glass of water\nToday: 40 pushups, 5x water\nStreak: 12 days\nLog with: coach done <activity> --json | Skip with: coach skip <activity> --json"
  }
}
```
Claude receives this context alongside the user's prompt. Claude reads the skill instructions and naturally reminds the user.

**3. No config (first run):**
```json
exit 0
{
  "hookSpecificOutput": {
    "hookEventName": "UserPromptSubmit",
    "additionalContext": "[COACH] Not configured. Guide the user through setup. Write config to ~/.config/coach/config.toml. See skill for config schema and presets."
  }
}
```
Claude guides the user through setup conversationally.

### The skill: `SKILL.md`

Loaded by Claude when it detects coach-related context (`[COACH]` prefix from hooks, or user asking about coach). Contains:

- How to handle break reminders (tone, brevity)
- How to handle "done" (run `coach done <activity> --json`, celebrate briefly)
- How to handle "skip" (run `coach skip <activity> --json`, no guilt)
- How to set up config (schema, presets, examples)
- How to show stats (run `coach stats`, interpret the output)
- The config TOML schema so Claude can write it directly

The skill includes setup presets:

**Desk Athlete** - movement focused
- pushups: 20 reps, every 1h
- squats: 20 reps, every 2h
- stretch: 2min, every 2h
- water: every 30m

**Hydration** - minimal
- water: every 30m

**Balanced** - movement + wellness
- pushups: 15 reps, every 1h
- stretch: 2min, every 2h
- water: every 30m
- eye break: every 20m

Claude suggests these during setup and adapts to the user's comfort level.

### The Go binary: `coach`

Pure backend. Two output modes:

- **Human mode** (default): colored contribution grids, formatted text. For `! coach stats`.
- **JSON mode** (`--json`): compact JSON. For Claude to consume with minimal context.

The SKILL.md instructs Claude to always use `--json`.

Commands:

| Command | Purpose | Called by |
|---------|---------|-----------|
| `coach check` | Check which activities are due. Returns JSON. | Hook (automatic) |
| `coach done <activity>` | Log activity completion. Returns summary. | Claude (via Bash) |
| `coach skip <activity>` | Log skip with cooldown. Returns status. | Claude (via Bash) |
| `coach stats` | Show contribution grids + stats. | User (`! coach stats`) or Claude (`--json`) |
| `coach install` | Set up the hook in Claude Code settings. | User (standalone install only) |
| `coach uninstall` | Remove the hook. | User |

## Context overhead

**Baseline (always):** ~50-100 tokens for the skill metadata (name + description). This lets the user ask about coach anytime, even without the hook firing.

**When no break is due:** Baseline only. Hook exits 0 with no output.

**When a break is due:** +~100-150 tokens of additionalContext, injected on that one prompt. The full skill body loads on-demand. After "done", timer resets and subsequent prompts go back to baseline.

**After the break:** Back to baseline until the next interval.

## User flows

### First time setup

```
User: "hey, I want coach to remind me to do pushups and drink water"

[Hook fires: no config detected]
[additionalContext: "[COACH] Not configured. Guide setup..."]
[Claude loads skill, sees config schema + presets]

Claude: "Let's set that up! How about pushups every hour and water
every 30 minutes? What hours are you usually working?"

User: "yeah that works, 9 to 6"

Claude: *writes ~/.config/coach/config.toml*
"Done! I'll remind you about pushups every hour and water every
30 minutes, between 9-6. Just tell me if you want to change anything."
```

### Break reminder

```
User: "refactor the payment module"

[Hook fires: pushups due (1h since last), water due (30m since last)]
[additionalContext: "[COACH] Due: 20 pushups, water..."]
[Claude loads skill]

Claude: "On it - but first, time for 20 pushups and some water.
You've done 40 pushups today, 12 day streak."

User: "done"

Claude: *runs `coach done pushups --json` and `coach done water --json`*
"60 pushups today, nice. Now let me look at the payment module..."
*continues with original task*
```

### Checking stats

```
User: "how am I doing with coach?"

[Claude recognizes coach context, loads skill]
[Claude runs `coach stats --json`]

Claude: "Solid week. 340 pushups, water 28 times, stretched 12 times.
12-day streak, your best is 18. You tend to skip stretching -
want me to make it less frequent?"
```

### Modifying config

```
User: "add eye breaks every 2 hours to coach"

[Claude loads skill, reads config schema]
[Claude reads current config, adds activity, writes config]

Claude: "Added eye breaks every 2 hours. Your activities now:
pushups (1h), water (30m), stretch (2h), eye breaks (2h)."
```

## Data model

### Config (`~/.config/coach/config.toml`)

```toml
[settings]
active_hours = ["09:00", "18:00"]
skip_cooldown = "10m"

[[activities]]
name = "pushups"
reps = 20
interval = "1h"

[[activities]]
name = "squats"
reps = 20
interval = "2h"

[[activities]]
name = "water"
message = "Drink a glass of water"
interval = "30m"

[[activities]]
name = "stretch"
duration = "2m"
interval = "2h"

[[activities]]
name = "eye break"
message = "Look at something 20 feet away for 20 seconds"
interval = "20m"
```

Each activity has its own `interval`. Independent timers. When multiple are due at once, all are returned - Claude presents them together naturally.

### Activity types

| Type | Fields | Example |
|------|--------|---------|
| Reps-based | `name`, `reps`, `interval` | pushups, squats |
| Duration-based | `name`, `duration`, `interval` | stretch, plank |
| Message-based | `name`, `message`, `interval` | water, eye break, stand up |

### Database (`~/.config/coach/coach.db`)

SQLite. Two tables:

**activity_log:** every done/skip event
- id, activity, reps, duration, action (done/skip), created_at

**state:** key-value for per-activity timers and flags
- last_done:{activity}, last_skip:{activity}, milestones

### Stats output (`coach stats`)

Human mode - 30-day contribution grids per activity with colored blocks. Each activity has its own color. Intensity = how much you did that day.

```
pushups                                    60 today · 1,240 all time
█ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █
Mar 10                                                    Apr 9

water                                       3x today · 94x all time
█ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █ █

12 day streak · best: 18 days
```

JSON mode (`--json`) - compact data for Claude to interpret.

## Milestones

Tracked in state table. Announced by Claude (via `coach done --json` output):
- 100, 250, 500, 1000, 2500, 5000, 10000 lifetime reps
- 7, 30, 100, 365 day streaks

## What coach does NOT do (v1)

- No progression (auto-increasing reps) - added later
- No team/social features
- No AI-adaptive coaching (but the architecture supports it)
- No mobile/desktop notifications
- No data sync

## Open questions

1. **Strict mode:** Should there be a config option for strict mode (exit 2, actually blocks the prompt) vs gentle mode (exit 0, context injection)? Some users may want hard enforcement.

2. **Plugin vs standalone first:** Build as a Claude Code plugin from the start, or build the standalone binary first and wrap it as a plugin later?
