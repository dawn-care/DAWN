-- Caregiver sessions become long-lived and revocable.
-- The session token is now stored as a SHA-256 hash (token_hash, 64 hex characters), so a
-- leaked database or backup no longer holds usable logins; only the cookie holds the token.
-- last_seen_at is when the session last renewed its 90 day window, which the server only
-- updates about once a day. A session also ends 1 year after created_at, however active.
-- user_id now references Caregiver, so deleting a caregiver deletes their sessions.
-- Existing sessions are deleted: their stored value is the raw token, which would never
-- match its hash, so those caregivers sign in once more.
-- +goose Up

DELETE FROM "Session";

ALTER TABLE "Session" RENAME COLUMN "session_id" TO "token_hash";
ALTER TABLE "Session" RENAME CONSTRAINT "Session_session_id_key" TO "Session_token_hash_key";

ALTER TABLE "Session"
  ALTER COLUMN "created_at" SET NOT NULL,
  ADD COLUMN "last_seen_at" timestamptz NOT NULL DEFAULT now(),
  ADD CONSTRAINT "Session_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "Caregiver" ("id") ON DELETE CASCADE;

-- Logging out and deleting an account remove every session of one caregiver
CREATE INDEX "Session_user_id_idx" ON "Session" ("user_id");

-- +goose Down
-- A hash cannot be turned back into the token it came from, so sessions are deleted again.

DELETE FROM "Session";

DROP INDEX "Session_user_id_idx";

ALTER TABLE "Session"
  DROP CONSTRAINT "Session_user_id_fkey",
  DROP COLUMN "last_seen_at",
  ALTER COLUMN "created_at" DROP NOT NULL;

ALTER TABLE "Session" RENAME CONSTRAINT "Session_token_hash_key" TO "Session_session_id_key";
ALTER TABLE "Session" RENAME COLUMN "token_hash" TO "session_id";
