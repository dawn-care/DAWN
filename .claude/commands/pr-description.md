---
description: Generate a PR description from an inline summary of changes, with an optional linked GitHub Projects ticket
---

The user ran `/pr-description` with the following arguments:

$ARGUMENTS

Arguments are provided as: `[<link-to-related-ticket>] <description-of-changes>`

- The ticket link is optional and, if present, is a URL (e.g. a GitHub Projects/Issues link) appearing first.
- Everything else is the inline description of the changes to describe.

Parse the arguments to separate the optional ticket link from the description. Then generate a pull request description using exactly this format:

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
- **Summary**: a bulleted list of what changed. Base this on the provided description, and on the actual diff/commits if available, rather than just restating the description verbatim.
- **Screenshots/Recordings**: always leave this section as the placeholder text `_<To be filled in manually by the PR creator>_` — this is filled in manually, not generated.
- **Breaking Changes**: call out anything reviewers should prioritize, such as breaking API/schema changes, risky areas of the diff, or parts that need closer review attention. Leave empty if there are none.
- **Test Plan**: a checklist (`- [ ]`) of concrete steps to verify the change works, left unchecked for the PR creator to check off as they complete them.
