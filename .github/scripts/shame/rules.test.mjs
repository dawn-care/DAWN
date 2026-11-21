// node --test .github/scripts/shame/
import { test } from "node:test";
import assert from "node:assert/strict";
import {
  CATEGORIES,
  TYPE_CATEGORY,
  applyShames,
  behemothShame,
  clockTime,
  idleAuthorShames,
  emptyLog,
  isLateNight,
  lateNightShame,
  manualShame,
  ordinal,
  prSize,
  pruneShamed,
  reviewDebts,
  slowReviewShames,
} from "./rules.mjs";

const at = (s) => new Date(s).getTime();
const basePr = (over = {}) => ({
  number: 7,
  author: "alice",
  createdAt: "2026-11-10T12:00:00.000Z",
  readyAt: "2026-11-10T12:00:00.000Z",
  requested: [],
  requestedAt: {},
  reviews: [],
  authorActivity: [],
  ...over,
});

test("size leaves out Swagger docs but counts tests", () => {
  const files = [
    { filename: "server/cmd/docs/swagger.json", additions: 900, deletions: 100 },
    { filename: "server/a_test.go", additions: 10, deletions: 0 },
    { filename: "server/a.go", additions: 5, deletions: 5 },
  ];
  assert.equal(prSize(files), 20);
});

test("behemoth only above 1000 lines", () => {
  assert.equal(behemothShame({ number: 1, author: "a" }, 1000).length, 0);
  assert.equal(behemothShame({ number: 1, author: "a" }, 1001)[0].key, "behemoth:pr1");
});

test("late night is 1am to 5am Eastern, across daylight saving", () => {
  assert.equal(isLateNight(new Date("2026-11-19T06:00:00Z")), true); // 1:00 EST
  assert.equal(isLateNight(new Date("2026-11-19T05:59:00Z")), false); // 12:59 EST
  assert.equal(isLateNight(new Date("2026-11-19T10:00:00Z")), false); // 5:00 EST
  assert.equal(isLateNight(new Date("2026-07-01T05:30:00Z")), true); // 1:30 EDT
  assert.equal(clockTime(new Date("2026-11-19T08:07:00Z")), "3:07 AM");
});

test("late night is one shame per person per night", () => {
  const [a] = lateNightShame("bob", "2026-11-19T07:00:00Z", "pushing to", 1);
  const [b] = lateNightShame("bob", "2026-11-19T09:00:00Z", "merging", 2);
  assert.equal(a.key, b.key);
  assert.equal(lateNightShame("bob", "2026-11-19T15:00:00Z", "merging", 2).length, 0);
});

test("requested reviewer owes a review from when they were asked", () => {
  const pr = basePr({ requested: ["bob"], requestedAt: { bob: "2026-11-11T00:00:00.000Z" } });
  assert.deepEqual(reviewDebts(pr, []), [{ login: "bob", since: "2026-11-11T00:00:00.000Z", rereview: false }]);
  assert.equal(slowReviewShames(pr, [], at("2026-11-11T23:00:00Z")).length, 0);
  const [s] = slowReviewShames(pr, [], at("2026-11-12T01:00:00Z"));
  assert.equal(s.period, 1);
  assert.equal(slowReviewShames(pr, [], at("2026-11-13T01:00:00Z"))[0].period, 2);
});

test("draft time doesn't count", () => {
  const pr = basePr({ readyAt: "2026-11-15T00:00:00.000Z", requested: ["bob"], requestedAt: { bob: "2026-11-11T00:00:00.000Z" } });
  assert.equal(reviewDebts(pr, [])[0].since, "2026-11-15T00:00:00.000Z");
});

test("a push after someone's review means they owe a re-review, approvals included", () => {
  const pr = basePr({
    reviews: [
      { login: "bob", at: "2026-11-11T00:00:00.000Z", state: "APPROVED" },
      { login: "carol", at: "2026-11-11T05:00:00.000Z", state: "CHANGES_REQUESTED" },
    ],
  });
  const pushes = ["2026-11-11T03:00:00.000Z", "2026-11-11T04:00:00.000Z"];
  const debts = reviewDebts(pr, pushes);
  assert.deepEqual(debts, [{ login: "bob", since: "2026-11-11T03:00:00.000Z", rereview: true }]);
  // carol reviewed after both pushes, so she's fine
});

test("author is idle 48 hours after a review, unless they acted or a review is owed", () => {
  const review = { login: "bob", at: "2026-11-11T00:00:00.000Z", state: "CHANGES_REQUESTED" };
  const pr = basePr({ reviews: [review] });
  assert.equal(idleAuthorShames(pr, [], at("2026-11-12T23:00:00Z")).length, 0);
  const [s] = idleAuthorShames(pr, [], at("2026-11-13T01:00:00Z"));
  assert.equal(s.login, "alice");
  assert.equal(s.approved, false);
  // replied to a thread
  assert.equal(idleAuthorShames(basePr({ reviews: [review], authorActivity: ["2026-11-11T01:00:00.000Z"] }), [], at("2026-11-14T00:00:00Z")).length, 0);
  // pushed: now bob owes a re-review instead
  assert.equal(idleAuthorShames(pr, ["2026-11-11T02:00:00.000Z"], at("2026-11-14T00:00:00Z")).length, 0);
  // someone else still hasn't reviewed
  assert.equal(idleAuthorShames(basePr({ reviews: [review], requested: ["carol"] }), [], at("2026-11-14T00:00:00Z")).length, 0);
});

test("approved and not merged counts as idle", () => {
  const pr = basePr({ reviews: [{ login: "bob", at: "2026-11-11T00:00:00.000Z", state: "APPROVED" }] });
  assert.equal(idleAuthorShames(pr, [], at("2026-11-13T01:00:00Z"))[0].approved, true);
});

test("each shame is logged once under its category, and old keys are pruned", () => {
  const state = { shamed: {}, pushes: {} };
  const log = emptyLog();
  const c = { key: "behemoth:pr1", type: "behemoth", login: "alice", pr: 1, lines: 2000 };
  assert.equal(applyShames(state, log, [c], "2026-11-01T00:00:00.000Z")[0].total, 1);
  assert.equal(applyShames(state, log, [c], "2026-11-01T00:00:00.000Z").length, 0);
  assert.deepEqual(log.shames, [{ at: "2026-11-01T00:00:00.000Z", login: "alice", category: "pr_size", type: "behemoth", pr: 1 }]);
  const late = { key: "late_night:alice:2026-11-02", type: "late_night", login: "alice", pr: 2 };
  assert.equal(applyShames(state, log, [late], "2026-11-02T07:00:00.000Z")[0].total, 2);
  pruneShamed(state, at("2027-02-01T00:00:00Z"));
  assert.deepEqual(state.shamed, {});
});

test("every automatic shame has a category", () => {
  for (const category of Object.values(TYPE_CATEGORY)) assert.ok(CATEGORIES[category], category);
});

const team = { alice: "1", bob: "2", carol: "3", dave: "4" };
const endorsed = (over = {}) => ({
  messageId: "1234567890",
  target: "2",
  by: "1",
  endorsedBy: ["3", "4"],
  category: "meetings",
  reason: "  late to standup  ",
  ...over,
});

test("a /shame with 2 endorsements becomes a manual shame", () => {
  const shame = manualShame(endorsed(), team);
  assert.deepEqual(shame, {
    key: "manual:1234567890",
    type: "manual",
    category: "meetings",
    login: "bob",
    by: "alice",
    endorsedBy: ["carol", "dave"],
    reason: "late to standup",
  });
  const log = emptyLog();
  applyShames({ shamed: {}, pushes: {} }, log, [shame], "2026-11-20T15:00:00.000Z");
  assert.equal(log.shames[0].reason, "late to standup");
  assert.deepEqual(log.shames[0].endorsedBy, ["carol", "dave"]);
});

test("the shamer, the shamed and repeats don't count as endorsements", () => {
  assert.throws(() => manualShame(endorsed({ endorsedBy: ["1", "2", "3", "3"] }), team), /Needs 2 endorsements.*got 1/);
});

test("a /shame from outside the team, or a bad one, is refused", () => {
  assert.throws(() => manualShame(endorsed({ target: "99" }), team), /isn't in DISCORD_USER_MAP/);
  assert.throws(() => manualShame(endorsed({ endorsedBy: ["3", "99"] }), team), /isn't in DISCORD_USER_MAP/);
  assert.throws(() => manualShame(endorsed({ category: "vibes" }), team), /Unknown category/);
  assert.throws(() => manualShame(endorsed({ reason: "   " }), team), /needs a reason/);
  assert.throws(() => manualShame(endorsed({ reason: "x".repeat(301) }), team), /over 300/);
  assert.throws(() => manualShame(endorsed({ messageId: "abc" }), team), /messageId/);
});

test("ordinals", () => {
  assert.deepEqual([1, 2, 3, 4, 11, 12, 13, 21, 22, 112].map(ordinal), ["1st", "2nd", "3rd", "4th", "11th", "12th", "13th", "21st", "22nd", "112th"]);
});
