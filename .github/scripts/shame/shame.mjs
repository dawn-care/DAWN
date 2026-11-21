// The shame bot. Run by .github/workflows/discord-shame.yml on PR events
// (late nights, behemoths), on a schedule (overdue reviews, idle authors), and
// when a /shame in Discord gets enough endorsements (sent by the team website
// as a repository_dispatch). Every
// shame is added to the log in metrics/shame.json on the metrics branch and
// posted to Discord. See README.md here for the file formats.
//
// Environment:
//   GH_TOKEN, GITHUB_REPOSITORY, GITHUB_EVENT_NAME, GITHUB_EVENT_PATH  (set by Actions)
//   METRICS_DIR             a checkout of the metrics branch
//   DISCORD_SHAME_WEBHOOK_URL  (falls back to DISCORD_WEBHOOK_URL)
//   DISCORD_USER_MAP        JSON of GitHub login -> Discord user ID
//   DRY_RUN=1               print what would happen; no commit, push or post

import { execFileSync } from "node:child_process";
import { readFileSync, writeFileSync, existsSync, mkdirSync } from "node:fs";
import { join } from "node:path";
import {
  CATEGORIES,
  applyShames,
  behemothShame,
  emptyLog,
  idleAuthorShames,
  lateNightShame,
  manualShame,
  ordinal,
  prSize,
  pruneShamed,
  slowReviewShames,
} from "./rules.mjs";

const env = process.env;
const REPO = env.GITHUB_REPOSITORY;
const [OWNER, NAME] = REPO.split("/");
const DRY_RUN = env.DRY_RUN === "1";
const METRICS_DIR = env.METRICS_DIR;
const WEBHOOK = env.DISCORD_SHAME_WEBHOOK_URL || env.DISCORD_WEBHOOK_URL;
const USER_MAP = JSON.parse(env.DISCORD_USER_MAP || "{}");
const LOG_FILE = join(METRICS_DIR, "metrics", "shame.json");
const STATE_FILE = join(METRICS_DIR, "metrics", "shame-state.json");

const iso = (t) => new Date(t).toISOString();
const TRUSTED = new Set(["OWNER", "MEMBER", "COLLABORATOR"]);

// ---- GitHub -----------------------------------------------------------------

async function github(path, init = {}) {
  const res = await fetch(`https://api.github.com/${path}`, {
    ...init,
    headers: {
      Authorization: `Bearer ${env.GH_TOKEN}`,
      Accept: "application/vnd.github+json",
      "User-Agent": "dawn-shame-bot",
      ...init.headers,
    },
  });
  if (!res.ok) throw new Error(`GitHub ${path}: ${res.status} ${await res.text()}`);
  return res.json();
}

async function githubAll(path) {
  const out = [];
  for (let page = 1; ; page++) {
    const sep = path.includes("?") ? "&" : "?";
    const items = await github(`${path}${sep}per_page=100&page=${page}`);
    out.push(...items);
    if (items.length < 100) return out;
  }
}

async function graphql(query, variables) {
  const body = await github("graphql", { method: "POST", body: JSON.stringify({ query, variables }) });
  if (body.errors) throw new Error(`GraphQL: ${JSON.stringify(body.errors)}`);
  return body.data;
}

const OPEN_PRS = `
query($owner: String!, $name: String!, $cursor: String) {
  repository(owner: $owner, name: $name) {
    pullRequests(states: OPEN, first: 25, after: $cursor) {
      pageInfo { hasNextPage endCursor }
      nodes {
        number title url isDraft isCrossRepository createdAt
        author { login }
        reviewRequests(first: 50) { nodes { requestedReviewer { __typename ... on User { login } } } }
        reviews(last: 100) { nodes { author { __typename login } submittedAt state } }
        comments(last: 100) { nodes { author { login } createdAt } }
        timelineItems(last: 100, itemTypes: [REVIEW_REQUESTED_EVENT, READY_FOR_REVIEW_EVENT]) {
          nodes {
            __typename
            ... on ReviewRequestedEvent { createdAt requestedReviewer { ... on User { login } } }
            ... on ReadyForReviewEvent { createdAt }
          }
        }
      }
    }
  }
}`;

// Shapes a GraphQL pull request into what rules.mjs expects.
function normalizePr(node) {
  const author = node.author?.login ?? "ghost";
  const requested = node.reviewRequests.nodes
    .map((n) => n.requestedReviewer)
    .filter((r) => r?.__typename === "User")
    .map((r) => r.login);
  const requestedAt = {};
  let readyAt = iso(node.createdAt);
  for (const item of node.timelineItems.nodes) {
    const at = iso(item.createdAt);
    if (item.__typename === "ReadyForReviewEvent" && at > readyAt) readyAt = at;
    const login = item.requestedReviewer?.login;
    if (item.__typename === "ReviewRequestedEvent" && login && !(requestedAt[login] > at)) requestedAt[login] = at;
  }
  const allReviews = node.reviews.nodes.filter((r) => r.submittedAt && r.author);
  const reviews = allReviews
    .filter((r) => r.author.__typename === "User" && r.author.login !== author)
    .map((r) => ({ login: r.author.login, at: iso(r.submittedAt), state: r.state }));
  const authorActivity = [
    ...allReviews.filter((r) => r.author.login === author).map((r) => iso(r.submittedAt)),
    ...node.comments.nodes.filter((c) => c.author?.login === author).map((c) => iso(c.createdAt)),
  ];
  return {
    number: node.number,
    title: node.title,
    url: node.url,
    isDraft: node.isDraft,
    isCrossRepository: node.isCrossRepository,
    author,
    createdAt: iso(node.createdAt),
    readyAt,
    requested,
    requestedAt,
    reviews,
    authorActivity,
  };
}

async function openPrs() {
  const prs = [];
  let cursor = null;
  do {
    const data = await graphql(OPEN_PRS, { owner: OWNER, name: NAME, cursor });
    const page = data.repository.pullRequests;
    prs.push(...page.nodes.map(normalizePr));
    cursor = page.pageInfo.hasNextPage ? page.pageInfo.endCursor : null;
  } while (cursor);
  return prs;
}

// ---- What happened ----------------------------------------------------------

// Reads the webhook event. Returns the PRs a shame can be about (by number),
// a function from the current state to shame candidates, and the changes to
// make to the state either way (recording a push, forgetting a closed PR).
async function fromEvent(name, event) {
  const prs = {};
  const sameRepo = (pr) => pr.head?.repo?.full_name === REPO;

  if (name === "schedule" || name === "workflow_dispatch") {
    const open = (await openPrs()).filter((pr) => !pr.isDraft && !pr.isCrossRepository);
    for (const pr of open) prs[pr.number] = pr;
    const openNumbers = new Set(open.map((pr) => String(pr.number)));
    return {
      prs,
      update: (state) => {
        // Forget pushes to PRs that are no longer open (or are drafts again).
        for (const n of Object.keys(state.pushes)) if (!openNumbers.has(n)) delete state.pushes[n];
      },
      candidates: (state, now) =>
        open.flatMap((pr) => {
          const pushes = state.pushes[pr.number] ?? [];
          return [
            ...slowReviewShames(pr, pushes, now),
            ...idleAuthorShames(pr, pushes, now),
          ];
        }),
    };
  }

  // An endorsed /shame from the team website.
  if (name === "repository_dispatch") {
    if (event.action !== "manual-shame") throw new Error(`Unknown repository_dispatch type "${event.action}"`);
    const shame = manualShame(event.client_payload ?? {}, USER_MAP);
    return { prs, update: () => {}, candidates: () => [shame] };
  }

  const candidates = [];
  let update = () => {};
  const sender = event.sender;
  const human = sender && sender.type !== "Bot";

  if (name === "pull_request") {
    const pr = event.pull_request;
    if (!sameRepo(pr)) return { prs, candidates: () => [], update };
    prs[pr.number] = { number: pr.number, title: pr.title, url: pr.html_url };
    const action = event.action;

    let at = pr.updated_at;
    let verb = { opened: "opening", reopened: "reopening", synchronize: "pushing to", ready_for_review: "marking ready", review_requested: "requesting a review on" }[action];
    if (action === "closed") {
      at = pr.merged_at || pr.closed_at;
      verb = pr.merged ? "merging" : "closing";
    }
    if (action === "opened") at = pr.created_at;
    if (human && verb) candidates.push(...lateNightShame(sender.login, at, verb, pr.number));

    if (["opened", "reopened", "synchronize", "ready_for_review"].includes(action)) {
      const files = await githubAll(`repos/${REPO}/pulls/${pr.number}/files`);
      candidates.push(...behemothShame({ number: pr.number, author: pr.user.login }, prSize(files)));
    }

    if (action === "synchronize") {
      const pushedAt = iso(pr.updated_at);
      update = (state) => {
        const pushes = (state.pushes[pr.number] ??= []);
        if (!pushes.includes(pushedAt)) pushes.push(pushedAt);
        state.pushes[pr.number] = pushes.slice(-100);
      };
    } else if (action === "closed") {
      update = (state) => delete state.pushes[pr.number];
    }
  }

  if (name === "pull_request_review") {
    const { review, pull_request: pr } = event;
    prs[pr.number] = { number: pr.number, title: pr.title, url: pr.html_url };
    if (review.user.type !== "Bot" && TRUSTED.has(review.author_association)) {
      const verb = { approved: "approving", changes_requested: "reviewing" }[review.state] ?? "commenting on";
      candidates.push(...lateNightShame(review.user.login, review.submitted_at, verb, pr.number));
    }
  }

  if (name === "issue_comment") {
    const { comment, issue } = event;
    if (issue.pull_request && comment.user.type !== "Bot" && TRUSTED.has(comment.author_association)) {
      prs[issue.number] = { number: issue.number, title: issue.title, url: issue.html_url };
      candidates.push(...lateNightShame(comment.user.login, comment.created_at, "commenting on", issue.number));
    }
  }

  return { prs, update, candidates: () => candidates };
}

// ---- Messages ---------------------------------------------------------------

const mention = (login) => (USER_MAP[login] ? `<@${USER_MAP[login]}>` : login);
const pick = (options) => options[Math.floor(Math.random() * options.length)];
// Escape [ and ] so a title can't break the markdown link
const link = (pr) => `[${pr.title.replace(/[[\]]/g, "\\$&")}](${pr.url})`;

function messageFor(shame, pr) {
  const who = mention(shame.login);
  const tally = `(${shame.login}'s ${ordinal(shame.total)} shame)`;
  let text;
  switch (shame.type) {
    case "behemoth":
      text = `${who}...you know what you did 🙂‍↔️ (${shame.lines.toLocaleString("en-US")} lines)`;
      break;
    case "late_night":
      text = `${who} needs to go to sleep. Go to sleep ${who} 🫵😐`;
      break;
    case "slow_review": {
      const what = shame.rereview ? "re-review" : "review it";
      if (shame.period === 1) {
        text = `😬 ${link(pr)} ${pick(["looks a little moldy, ew", "is getting kinda crusty", "is starting to smell bad 🤢"])}: ${who} hasn't gotten to ${what} in 24+ hours ☹️`;
      } else if (shame.period === 2) {
        text = `😞 ${link(pr)} ${pick(["has gone crusty and moldy", "has shriveled up", "is sad and lonely"])} after ${shame.days} days waiting for ${who} to ${what} 😰`;
      } else {
        text = `👺 ${pick(["HOLY MOLY", "CODE RED", "THIS IS A TRAVESTY"])}: ${link(pr)} HAS WAITED ${shame.days} DAYS FOR ${who} TO ${what.toUpperCase()}(point and laugh at ${who}) 🫵😐`;
      }
      break;
    }
    case "idle_author":
      text = shame.approved
        ? `🦥 ${link(pr)} was approved ${shame.hours} hours ago and ${who} still hasn't merged it`
        : `🦥 ${who} hasn't touched ${link(pr)} in the ${shame.hours} hours since it was reviewed`;
      break;
    case "manual": {
      const endorsers = shame.endorsedBy.join(" and ");
      text = `🔔 ${who} has been officially shamed for ${CATEGORIES[shame.category]}: "${shame.reason}" (shame on you, ${who} 🙂‍↔️)`;
      break;
    }
  }
  return `${text} ${tally}`;
}

// Only the person being shamed is pinged, whatever a /shame reason contains.
async function postToDiscord(content, login) {
  const res = await fetch(WEBHOOK, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      content,
      username: "Shame Bot",
      flags: 4,
      allowed_mentions: { parse: [], users: USER_MAP[login] ? [String(USER_MAP[login])] : [] },
    }),
  });
  if (!res.ok) throw new Error(`Discord: ${res.status} ${await res.text()}`);
}

// ---- The counter ------------------------------------------------------------

const git = (...args) => execFileSync("git", ["-C", METRICS_DIR, ...args], { encoding: "utf8" }).trim();
const readJson = (file, fallback) => (existsSync(file) ? JSON.parse(readFileSync(file, "utf8")) : fallback);
const writeJson = (file, value) => writeFileSync(file, `${JSON.stringify(value, null, 2)}\n`);

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

async function main() {
  const event = env.GITHUB_EVENT_PATH ? JSON.parse(readFileSync(env.GITHUB_EVENT_PATH, "utf8")) : {};
  const happened = await fromEvent(env.GITHUB_EVENT_NAME, event);

  // Two runs can save at once; the second push is rejected, so it starts
  // over from the newer counts. Discord only hears about shames that saved.
  for (let attempt = 1; attempt <= 5; attempt++) {
    if (!DRY_RUN) {
      git("fetch", "--quiet", "--depth=1", "origin", "+refs/heads/metrics:refs/remotes/origin/metrics");
      git("reset", "--quiet", "--hard", "origin/metrics");
    }
    mkdirSync(join(METRICS_DIR, "metrics"), { recursive: true });

    const nowDate = new Date();
    const now = nowDate.getTime();
    const log = readJson(LOG_FILE, emptyLog());
    log.categories = CATEGORIES; // so the labels on the website follow rules.mjs
    const state = readJson(STATE_FILE, { shamed: {}, pushes: {} });

    happened.update(state);
    pruneShamed(state, now);
    const added = applyShames(state, log, happened.candidates(state, now), iso(nowDate));
    const messages = added.map((s) => ({ text: messageFor(s, happened.prs[s.pr]), login: s.login }));

    if (added.length) log.updatedAt = iso(nowDate);
    writeJson(LOG_FILE, log);
    writeJson(STATE_FILE, state);

    if (DRY_RUN) {
      console.log(messages.length ? messages.map((m) => m.text).join("\n") : "No new shame.");
      console.log(JSON.stringify(log, null, 2));
      return;
    }

    if (git("status", "--porcelain").length === 0) {
      console.log("Nothing changed.");
      return;
    }
    git("add", "metrics/shame.json", "metrics/shame-state.json");
    git("commit", "--quiet", "-m", added.length ? `Shame: ${added.map((s) => `${s.type} ${s.login}`).join(", ")}` : "Update shame bot state");

    try {
      git("push", "--quiet", "origin", "HEAD:metrics");
    } catch (err) {
      console.log(`Push rejected (attempt ${attempt}), retrying: ${err.message}`);
      await sleep(1000 + Math.random() * 4000);
      continue;
    }

    for (const { text, login } of messages) {
      console.log(text);
      await postToDiscord(text, login);
    }
    console.log(`${added.length} new shame(s).`);
    return;
  }
  throw new Error("Couldn't save the shame counts after 5 tries.");
}

await main();
