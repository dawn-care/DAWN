---
description: Reformat an existing GitHub ticket's description into the repo's epic/user-story template
---

The user ran `/ticket-description` with the following argument:

$ARGUMENTS

This is a link to an existing ticket (a GitHub Issue or GitHub Projects item).

1. Fetch the ticket's existing content from the link (e.g. via `gh issue view <url> --json title,body,labels` if it's a GitHub Issue URL, or `gh api` for a GitHub Projects item URL). Extract its current description/body text — this is the raw description to reformat.

2. Decide whether this ticket is an **Epic** or a **User Story**:
   - **Epic**: a broad body of work or feature area, usually made up of multiple smaller stories (e.g. "device pairing flow", "caregiver media library").
   - **User Story**: a single, specific piece of functionality from one user's perspective (caregiver or carepartner), usually expressible as "As a ___, I want to ___ so that ___".
   - If it cannot be clearly determined from the ticket's existing labels/title/body, **default to User Story**.

3. Read the matching template from this repo before generating anything:
   - Epic → `.github/ISSUE_TEMPLATE/epic-template.md`
   - User Story → `.github/ISSUE_TEMPLATE/user-story-template.md`

4. Reformat the ticket's existing description into that exact template's structure:
   - For an **Epic**, write a `### Description` paragraph summarizing the work, and fill in `### Acceptance criteria` with a checklist (`- [ ]`) of concrete conditions that must be true for the epic to be considered done. Fill `### Notes` with `N/A` unless the original description calls out something worth noting.
   - For a **User Story**, write the `### User Story` line in the "As a ___, I want to ___ so that ___" form, and fill in `### Acceptance Criteria` with a checklist (`- [ ]`) of concrete, testable conditions. Fill `### Notes` with `N/A` unless the original description calls out something worth noting.
   - Base all of this on the ticket's actual existing description — do not invent new scope that wasn't already implied by it.

Do not include the template's YAML frontmatter (`name`/`about`/`title`/`labels`/`assignees`) in the output — only the body content (the `###` sections onward).

Before showing the result, state which type you chose (Epic or User Story) and a one-sentence reason why (or note that it defaulted to User Story because the type couldn't be determined), then show the reformatted template.
