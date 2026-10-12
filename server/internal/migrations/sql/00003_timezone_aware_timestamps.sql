-- Session and device pairing expiry times become timezone aware (timestamptz), as
-- init/database.sql had them since "make dev database recognize timezones".
-- Existing values are read in the session's time zone (UTC on the database server).
-- +goose Up

ALTER TABLE "Session"
  ALTER COLUMN "created_at" TYPE timestamptz,
  ALTER COLUMN "expires_at" TYPE timestamptz;

ALTER TABLE "DevicePairing"
  ALTER COLUMN "expires_at" TYPE timestamptz;

-- +goose Down

ALTER TABLE "Session"
  ALTER COLUMN "created_at" TYPE timestamp,
  ALTER COLUMN "expires_at" TYPE timestamp;

ALTER TABLE "DevicePairing"
  ALTER COLUMN "expires_at" TYPE timestamp;
