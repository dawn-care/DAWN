---
description: Generate a PR description for the current branch by inspecting its actual diff against main
---

The user ran `/pr-description`. This command takes no input — everything needed comes from the current branch itself.

$ARGUMENTS

If `$ARGUMENTS` is non-empty, treat it as extra context/emphasis from the user (e.g. a related ticket link, or notes on what to highlight), but it is entirely optional — do not require it.

Before writing the description, inspect the actual code changes on the current branch:

1. Determine the current branch name (`git branch --show-current`) and the base branch to diff against (typically `main`; use `git merge-base HEAD main` if unsure).
2. Run `git diff <base>...HEAD` (and `git log <base>..HEAD` for commit messages) to see everything changed on this branch. If the working tree also has uncommitted changes, include those too (`git status` / `git diff`).
3. Read enough of the actual diff (not just file names) to understand what changed — new files, modified logic, config, CI workflows, dependencies, etc. This diff is the primary source of truth for every section below — do not invent changes that aren't in it.
4. Look for related ticket link(s): use any supplied in `$ARGUMENTS` if present; otherwise check the branch name, recent commit messages, and PR template for ticket references. If none can be found, there are no related tickets.

Then generate a pull request description using exactly this format:

```markdown
# <Concise PR title derived from the description>

**Related Ticket(s):** [<ticket link>](<ticket link>)

## 📝 Summary

- <Bulleted summary of what changed, derived from the description and/or the actual diff>

## 📸 Screenshots/Recordings

_<To be filled in manually by the PR creator>_

## ⚠️ Breaking Changes

<Anything reviewers should prioritize: breaking changes, risky areas, or parts of the diff that need closer attention. Leave empty if none.>

## ✅ Test Plan

- [ ] <Checkbox item for something the PR creator should verify before merging>
```

Formatting rules:

- **Title**: a short, descriptive title summarizing the change, derived from the diff and commit messages.
- **Related Ticket(s)**: if one or more ticket links were found (from `$ARGUMENTS`, the branch name, commits, etc.), render each as a Markdown link using the link itself as the text, comma-separated if there's more than one. If none were found, write `None` (e.g. `**Related Ticket(s):** None`).
- **Summary**: a bulleted list of what changed, grounded in the actual diff inspected above. Call out new/modified files, endpoints, config, and dependencies as relevant.
- **Screenshots/Recordings**: always leave this section as the placeholder text `_<To be filled in manually by the PR creator>_` — this is filled in manually, not generated.
- **Breaking Changes**: derived from the actual diff, call out anything reviewers should prioritize, including but not limited to:
  - Breaking API/schema/config changes
  - New required GitHub Actions/repo/org secrets, environment variables, or other config that must be set up before the change works (e.g. a new webhook URL, API key, or credential referenced in a workflow or `.env`)
  - Risky areas of the diff or parts that need closer review attention
  - Leave empty only if none of the above apply.
- **Test Plan**: a checklist (`- [ ]`) of concrete steps to verify the change works, left unchecked for the PR creator to check off as they complete them.
