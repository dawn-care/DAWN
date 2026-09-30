---
description: Write a ticket description (epic, story, or task) from a GitHub ticket link and short brief
argument-hint: <epic|task|story> <link-to-ticket> <short description of what you want>
---

The user ran `/ticket-description` with the following arguments:

$ARGUMENTS

The arguments have this shape: `<type> <link-to-ticket> <short description>`

- `<type>` is the first word and must be one of `epic`, `story`, or `task` (case-insensitive; accept `user story` / `user-story` as `story`).
- `<link-to-ticket>` is the second token, a GitHub Issue or GitHub Projects item URL.
- Everything after the link is a short description of what the user wants the ticket to cover or look like.

If the type is missing or isn't one of the three, or the link is missing, do not guess. Reply with the usage line `/ticket-description <epic|task|story> <link-to-ticket> <short description>` and stop.

## Steps

1. Fetch the ticket from the link (e.g. `gh issue view <url> --json title,body,labels` for a GitHub Issue, or `gh api` for a Projects item). The ticket is expected to contain only a short description of what the user wants it to cover; use its title and body, together with the description in the arguments, as the source material. If the two conflict, the arguments win.

2. Write the description using the format for the requested type below. Base it on the source material and what it reasonably implies; do not invent unrelated scope. Keep items concrete, short, and testable.

3. Do not include YAML frontmatter or a title. Output only the body sections.

## Formats

### epic

```
### Description
<what the epic is and what it covers>

### Subtasks
- [ ] <short description of a task or user story that belongs under this epic>
- [ ] ...
```

Subtasks are short, one-line items that could each later be broken out into their own user story or task. They must be checkboxes.

### story

```
### User Story
As a <role>, I want <goal>, so that <benefit>.

### Acceptance Criteria
- [ ] <condition that must be done as part of this story>
- [ ] ...

### Definition of Done
<what must be true for this story to be considered complete>

### Notes
<additional notes about the story, or N/A if there is nothing additional>
```

`<role>` should be a real user of the product (e.g. caregiver, carepartner) unless the source clearly says otherwise.

### task

```
### Description
<short description; does not need to be in user story format>

### Acceptance Criteria
- [ ] <condition that must be done as part of this task>
- [ ] ...

### Definition of Done
<what must be true for this task to be considered complete>
```

Tasks are smaller than user stories and are assumed to be linked to a parent story and/or epic, so do not write a user story line or mention the parent.

## Output

Output the result as **raw markdown inside a single fenced code block** (use a ` ```markdown ` fence) so the user can copy it verbatim and paste it into the ticket. Put nothing inside the fence except the description itself. Do not edit the ticket on GitHub. Outside the fence, add at most one short line if something in the source was ambiguous and you made an assumption.
