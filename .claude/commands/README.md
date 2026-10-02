# Shared Claude Code commands

Slash commands in this directory are checked into the repo, so anyone who
clones StartUp and opens it with Claude Code gets them automatically — no
per-machine setup required.

## Adding a command

Create a Markdown file here, e.g. `.claude/commands/foo.md`:

```markdown
---
description: One-line description shown in the command list
---

Instructions for Claude to follow when /foo is run. $ARGUMENTS is replaced
with whatever the user types after the command name.
```

The file then becomes available as `/foo` in any Claude Code session opened
in this repo.

Personal, machine-specific commands (not meant to be shared with the team)
belong in `~/.claude/commands/` instead, or in `.claude/settings.local.json`,
which is gitignored.
