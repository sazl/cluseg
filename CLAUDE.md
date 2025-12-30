# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Status

This is a newly initialized repository. No application code, build system, or tests are configured yet.

## Issue Tracking with Beads

This project uses **bd** (beads) for AI-native issue tracking. Issues are stored in `.beads/issues.jsonl` and sync with git.

### Key Commands

```bash
bd ready                    # Find available work (no blockers)
bd show <id>                # View issue details
bd update <id> --status in_progress  # Claim work
bd close <id>               # Complete work
bd create --title="..." --type=task|bug|feature --priority=2  # Create issue
bd sync                     # Sync with git
bd list                     # View all issues
```

### Priority Values

Use numeric priorities 0-4 (not "high"/"medium"/"low"):
- 0 = critical
- 2 = medium (default)
- 4 = backlog

## Session Completion Workflow

When ending a work session, complete ALL steps:

1. File issues for remaining work
2. Run quality gates if code changed (tests, linters, builds)
3. Update issue status - close finished work
4. Push to remote:
   ```bash
   git pull --rebase
   bd sync
   git push
   ```
5. Verify `git status` shows "up to date with origin"

Work is NOT complete until `git push` succeeds.
