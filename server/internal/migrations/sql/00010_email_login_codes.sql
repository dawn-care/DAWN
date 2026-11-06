-- Caregivers can sign in with an emailed 6-digit code as well as with Google (ticket i442).
-- Email becomes the key both login methods share: Caregiver.email is unique ignoring case.
-- It stays nullable because existing Google caregivers have no stored email until their next
-- Google sign-in saves it. google_id stays too, as the extra key only Google-linked accounts have.
-- The demo caregiver's email was seeded into google_id, so it moves to email.
-- LoginCode holds the codes: only a keyed hash (code_hash), never the code itself.
-- LoginCodeRequest is the log the rate limits count (per email and per IP). ip_hash is a keyed
-- hash, never the address. LoginCodeGuess logs wrong code guesses per IP (ip_hash, a keyed hash),
-- for the limit of 20 wrong guesses per IP per hour. Rows of all three tables older than a day
-- are deleted the next time a code request goes through; there is no scheduled cleanup.
-- +goose Up

ALTER TABLE "Caregiver" ADD COLUMN "email" text;

UPDATE "Caregiver"
SET "email" = "google_id", "google_id" = NULL
WHERE "google_id" = 'test.caregiver@example.com';

CREATE UNIQUE INDEX "Caregiver_email_lower_key" ON "Caregiver" (lower("email"));

CREATE TABLE "LoginCode" (
  "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "email" text NOT NULL,
  "code_hash" text NOT NULL,
  "attempts" integer NOT NULL DEFAULT 0,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "expires_at" timestamptz NOT NULL,
  "used_at" timestamptz
);

-- Finding the newest code for an email
CREATE INDEX "LoginCode_email_created_at_idx" ON "LoginCode" (lower("email"), "created_at");

CREATE TABLE "LoginCodeRequest" (
  "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "email" text NOT NULL,
  "ip_hash" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now()
);

-- Counting recent requests per email and per IP
CREATE INDEX "LoginCodeRequest_email_created_at_idx" ON "LoginCodeRequest" (lower("email"), "created_at");
CREATE INDEX "LoginCodeRequest_ip_hash_created_at_idx" ON "LoginCodeRequest" ("ip_hash", "created_at");

CREATE TABLE "LoginCodeGuess" (
  "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "ip_hash" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now()
);

-- Counting recent wrong guesses per IP
CREATE INDEX "LoginCodeGuess_ip_hash_created_at_idx" ON "LoginCodeGuess" ("ip_hash", "created_at");

-- +goose Down
-- Codes and request logs are short-lived, so they are dropped. Only the demo caregiver's
-- email goes back to google_id; other stored emails are lost.

DROP TABLE "LoginCodeGuess";
DROP TABLE "LoginCodeRequest";
DROP TABLE "LoginCode";

DROP INDEX "Caregiver_email_lower_key";

UPDATE "Caregiver"
SET "google_id" = "email"
WHERE "email" = 'test.caregiver@example.com' AND "google_id" IS NULL;

ALTER TABLE "Caregiver" DROP COLUMN "email";
