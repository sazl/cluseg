# cluseg - Claude Usage Monitor

A minimal TUI for monitoring Claude Code usage, designed to run in a tmux pane.

## Overview

**Purpose:** Track Claude Code CLI usage with a glanceable status bar showing usage percentage, reset countdown, and warning indicators.

**Key features:**
- Reads local `~/.claude/stats-cache.json` (no authentication required)
- Configurable warning thresholds (default 80%, 95%)
- Manual limit configuration with adaptive learning
- Compact display optimized for small terminal panes
- 30-second refresh interval (configurable)

## Architecture

```
┌─────────────────────────────────────────────────┐
│                    cluseg                       │
├─────────────────────────────────────────────────┤
│  config/          - YAML config + CLI parsing   │
│  stats/           - Read ~/.claude/stats-cache  │
│  limits/          - Limit tracking & learning   │
│  ui/              - Minimal TUI rendering       │
│  main.go          - Entry point, refresh loop   │
└─────────────────────────────────────────────────┘
```

**Data flow:**
1. On startup, load config from `~/.config/cluseg/config.yaml`
2. Every 30s (configurable), read `~/.claude/stats-cache.json`
3. Compare today's usage against configured/learned limits
4. Render compact status line with usage %, countdown, warning state
5. If rate limit detected, update learned limit

**Dependencies:**
- `bubbletea` - Minimal TUI framework
- `lipgloss` - Terminal styling
- `viper` - Config file + CLI flag merging
- Standard library for file watching and JSON parsing

## Display Format

**Compact status line (~60 chars):**

```
Normal state:
 ██████████░░░░░░ 67% │ resets in 4h 32m

Warning state (≥80%):
 ████████████████ 85% │ resets in 4h 32m ⚠

Critical state (≥95%):
 ████████████████ 98% │ resets in 2h 15m 🔴
```

**Color scheme:**
- Normal (0-79%): Green bar, white text
- Warning (80-94%): Yellow bar, yellow text
- Critical (95%+): Red bar, red text

**Progress bar:** 16 characters using block characters (█░).

**Reset countdown:** Time until configured reset (default midnight UTC). Shows `Xh Ym` or `Xm` if under an hour.

**Ultra-compact mode** for very small panes:
```
67% 4h32m
85% 4h32m ⚠
```

## Configuration

**Location:** `~/.config/cluseg/config.yaml`

```yaml
# Usage limits (manual configuration)
limits:
  daily_tokens: 500000        # Your estimated daily token limit
  reset_time: "00:00"         # When limits reset (24h format, UTC)

# Warning thresholds
thresholds:
  warning: 80                 # Yellow warning at 80%
  critical: 95                # Red alert at 95%

# Display settings
display:
  refresh_seconds: 30         # How often to poll stats
  compact: false              # Ultra-minimal mode

# Data source
source:
  stats_file: "~/.claude/stats-cache.json"

# Learned limits (auto-updated by cluseg)
learned:
  last_observed_limit: null   # Updated when rate limit hit
  confidence: 0               # How many times limit observed
```

**CLI overrides:**
```bash
cluseg                           # Run with defaults
cluseg --warn-at=70              # Override warning threshold
cluseg --refresh=10s             # Faster refresh
cluseg --compact                 # Ultra-minimal mode
cluseg --limit=300000            # Override daily limit
```

## Limit Learning

**Detection methods:**
1. Watch for usage plateaus (daily tokens stop increasing despite activity)
2. Manual reporting via `cluseg hit-limit` command

**Learning algorithm:**
- Record current daily usage when limit detected
- After 3+ observations, set learned limit to minimum observed value
- Use learned limit only if confidence >= 3 and no manual limit configured

## Error Handling

**Stats file issues:**
- File doesn't exist: Show "No Claude Code data found"
- File stale (>5 min old): Show stale indicator `67% ⏸ 2h ago`
- Invalid JSON: Log error, retry next cycle

**No limit configured:**
- Show raw token count: `124k tokens today │ resets in 4h 32m │ no limit set`

**Terminal:**
- Graceful degradation without color support
- Handle resize (SIGWINCH)
- Clean exit on SIGINT/SIGTERM

## File Structure

```
cluseg/
├── main.go                 # Entry point, refresh loop
├── config/
│   └── config.go           # Viper setup, CLI flags, defaults
├── stats/
│   └── reader.go           # Parse stats-cache.json
├── limits/
│   └── tracker.go          # Limit comparison, learning logic
├── ui/
│   ├── display.go          # Bubbletea model, render logic
│   └── styles.go           # Lipgloss color definitions
├── go.mod
├── go.sum
└── README.md
```

## Build & Run

```bash
go build -o cluseg .
./cluseg                    # Run monitor
./cluseg --help             # Show options
./cluseg hit-limit          # Record rate limit observation
```

## Out of Scope

- Historical graphs/charts
- Model-by-model breakdown in display
- Notifications beyond visual indicators
- API server or remote access
- Session-level detail view
