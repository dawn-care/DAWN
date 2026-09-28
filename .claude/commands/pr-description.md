---
description: Generate a PR description from an inline summary of changes, with an optional linked GitHub Projects ticket
---

The user ran `/pr-description` with the following arguments:

$ARGUMENTS

Arguments are provided as: `[<link-to-related-ticket>] <description-of-changes>`

- The ticket link is optional and, if present, is a URL (e.g. a GitHub Projects/Issues link) appearing first.
- Everything else is the inline description of the changes to describe.

Parse the arguments to separate the optional ticket link from the description.

Before writing the description, inspect the actual code changes on the current branch:

1. Determine the base branch to diff against (typically `main`; use `git merge-base HEAD main` if unsure).
2. Run `git diff <base>...HEAD` (and `git log <base>..HEAD` for commit messages) to see everything changed on this branch. If the working tree also has uncommitted changes, include those too (`git status` / `git diff`).
3. Read enough of the actual diff (not just file names) to understand what changed — new files, modified logic, config, CI workflows, dependencies, etc. Use this, together with the user's inline description, to ground every section below in what was really changed rather than only in the description text.

Then generate a pull request description using exactly this format:

```markdown
# <Concise PR title derived from the description>

**Related Ticket:** [<ticket link>](<ticket link>)

## Summary

- <Bulleted summary of what changed, derived from the description and/or the actual diff>

## Screenshots/Recordings

_<To be filled in manually by the PR creator>_

## Breaking Changes

<Anything reviewers should prioritize: breaking changes, risky areas, or parts of the diff that need closer attention. Leave empty if none.>

## Test Plan

- [ ] <Checkbox item for something the PR creator should verify before merging>
```

Formatting rules:

- **Title**: a short, descriptive title summarizing the change, derived from the provided description.
- **Related Ticket**: if a ticket link was provided, render it as a Markdown link using the link itself as the text. If no ticket link was provided, leave this line empty (e.g. `**Related Ticket:**`).
- **Summary**: a bulleted list of what changed, grounded in the actual diff inspected above (not just the inline description). Call out new/modified files, endpoints, config, and dependencies as relevant.
- **Screenshots/Recordings**: always leave this section as the placeholder text `_<To be filled in manually by the PR creator>_` — this is filled in manually, not generated.
- **Breaking Changes**: derived from the actual diff, call out anything reviewers should prioritize, including but not limited to:
  - Breaking API/schema/config changes
  - New required GitHub Actions/repo/org secrets, environment variables, or other config that must be set up before the change works (e.g. a new webhook URL, API key, or credential referenced in a workflow or `.env`)
  - Risky areas of the diff or parts that need closer review attention
  - Leave empty only if none of the above apply.
- **Test Plan**: a checklist (`- [ ]`) of concrete steps to verify the change works, left unchecked for the PR creator to check off as they complete them.
