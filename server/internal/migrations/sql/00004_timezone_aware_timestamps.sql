-- Session, device pairing and playback times become timezone aware (timestamptz).
-- Session and DevicePairing had this in init/database.sql since "make dev database
-- recognize timezones"; Playback (migration 2) already stores UTC values.
-- Existing values are read in the session's time zone (UTC on the database server).
-- +goose Up

ALTER TABLE "Session"
  ALTER COLUMN "created_at" TYPE timestamptz,
  ALTER COLUMN "expires_at" TYPE timestamptz;

ALTER TABLE "DevicePairing"
  ALTER COLUMN "expires_at" TYPE timestamptz;

ALTER TABLE "Playback"
  ALTER COLUMN "requested_at" TYPE timestamptz,
  ALTER COLUMN "acknowledged_at" TYPE timestamptz,
  ALTER COLUMN "started_at" TYPE timestamptz,
  ALTER COLUMN "ended_at" TYPE timestamptz,
  ALTER COLUMN "expires_at" TYPE timestamptz;

-- +goose Down

ALTER TABLE "Session"
  ALTER COLUMN "created_at" TYPE timestamp,
  ALTER COLUMN "expires_at" TYPE timestamp;

ALTER TABLE "DevicePairing"
  ALTER COLUMN "expires_at" TYPE timestamp;

ALTER TABLE "Playback"
  ALTER COLUMN "requested_at" TYPE timestamp,
  ALTER COLUMN "acknowledged_at" TYPE timestamp,
  ALTER COLUMN "started_at" TYPE timestamp,
  ALTER COLUMN "ended_at" TYPE timestamp,
  ALTER COLUMN "expires_at" TYPE timestamp;
