# Shame bot

Posts shame to Discord and logs it on the `metrics` branch, where the team
website (dawn-care/Internal-Website) reads it. Run by
[`discord-shame.yml`](../../workflows/discord-shame.yml). The rules are in
`rules.mjs`, and everything else (GitHub, the log, Discord) is in `shame.mjs`.

```bash
node --test .github/scripts/shame/rules.test.mjs
```

## `metrics/shame.json`

Every shame that has happened, one entry each, oldest first. It only grows:
changing a rule never changes past shames. Like `pull-requests.json`, it holds
facts, and the website does the counting (by person, by category, over time).
Read it from
`https://raw.githubusercontent.com/dawn-care/DAWN/metrics/metrics/shame.json`.

```json
{
  "categories": { "reviews": "Reviews", "meetings": "Meetings", "...": "..." },
  "shames": [
    { "at": "2026-11-19T13:17:02.000Z", "login": "bob", "category": "reviews", "type": "slow_review", "pr": 487 },
    {
      "at": "2026-11-20T15:00:00.000Z", "login": "bob", "category": "meetings", "type": "manual",
      "reason": "late to standup", "by": "alice", "endorsedBy": ["carol", "dave"]
    }
  ],
  "updatedAt": "2026-11-20T15:00:00.000Z"
}
```

- `categories`: category id to its label, from `CATEGORIES` in `rules.mjs`.
  Show the label and group by the id; ids never get renamed.
- `shames[].at`: when the bot recorded it (UTC).
- `shames[].login`: the GitHub login of the person shamed.
- `shames[].category`: one of the `categories` ids.
- `shames[].type`: what caught it. `behemoth` (PR size), `late_night` (sleep),
  `slow_review` (reviews), `idle_author` (follow-through), or
  `manual` (a `/shame` in Discord, any category).
- `shames[].pr`: the PR number, for shames about a PR.
- `reason`, `by`, `endorsedBy`: only on `manual` shames. The reason is what the
  shamer typed; `by` and `endorsedBy` are GitHub logins.
- `updatedAt`: when the last shame was added.

`metrics/shame-state.json` is the bot's own bookkeeping (which shames it has
already logged, and when each open PR was pushed to). It isn't a metric, and
its format can change at any time.

## `/shame` (sent by the team website)

`/shame` and its Endorse button run in the team website's `/api`, since
Discord sends commands and button clicks to a web address, and the webhook this
bot posts with can't receive anything. Once a `/shame` has enough
endorsements, the website sends it here:

```http
POST https://api.github.com/repos/dawn-care/DAWN/dispatches
Authorization: Bearer <token with Contents: write on dawn-care/DAWN>

{
  "event_type": "manual-shame",
  "client_payload": {
    "messageId": "<the /shame message's Discord ID>",
    "target": "<Discord user ID of the person shamed>",
    "by": "<Discord user ID of the shamer>",
    "endorsedBy": ["<Discord user ID>", "..."],
    "category": "<a categories id>",
    "reason": "<up to 300 characters>"
  }
}
```

The bot checks it again (`manualShame` in `rules.mjs`) and refuses it, failing
the run, unless:

- everyone in it is in the `DISCORD_USER_MAP` secret (so only the team can
  shame, be shamed or endorse);
- at least 2 endorsers are neither the shamer nor the person shamed;
- the category is one of `categories`, and the reason isn't empty or over 300
  characters.

Each `messageId` is logged once, so sending the same one again (say, two
Endorse clicks at the same moment) does nothing. The bot posts the shame to
Discord, mentioning only the person shamed, so anything typed in the reason
can't ping anyone.
