-- Schema for ticket i332 (approved 2026-10-08). See i332plan.md for the decisions behind it.

CREATE TABLE "Caregiver" (
  "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "name" text,
  "role" text,
  "google_id" text UNIQUE
);

CREATE TABLE "Carepartner" (
  "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "caregiver_id" integer REFERENCES "Caregiver"("id") ON DELETE RESTRICT,
  "name" text,
  -- Dates and times elsewhere are stored as local wall time in this zone
  "timezone" text NOT NULL DEFAULT 'America/New_York'
);

CREATE TABLE "Session" (
    id SERIAL PRIMARY KEY,
    session_id TEXT UNIQUE NOT NULL,
    user_id INT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE "DevicePairing" (
  "id" SERIAL PRIMARY KEY,
  "carepartner_id" INTEGER NOT NULL REFERENCES "Carepartner"("id"),
  "pairing_code" VARCHAR(10) NOT NULL UNIQUE,
  "expires_at" TIMESTAMPTZ NOT NULL,
  "paired" BOOLEAN DEFAULT FALSE
);

-- Photos, videos and audio. The file itself lives in MinIO under object_key.
-- Usage status (on screen, audio background, scheduled) is derived at read time, not stored.
CREATE TABLE "Media" (
  "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "carepartner_id" integer NOT NULL REFERENCES "Carepartner"("id") ON DELETE CASCADE,
  "kind" text NOT NULL CHECK ("kind" IN ('photo', 'video', 'audio')),
  "display_name" text NOT NULL,
  "object_key" text NOT NULL UNIQUE,
  "mime_type" text NOT NULL,
  "size_bytes" bigint NOT NULL CHECK ("size_bytes" >= 0),
  -- Read with ffprobe on upload; photos have none, video and audio are capped at 10 minutes
  "duration_seconds" integer,
  "created_at" timestamptz NOT NULL DEFAULT NOW(),
  CHECK (
    ("kind" = 'photo' AND "duration_seconds" IS NULL)
    OR ("kind" IN ('video', 'audio') AND "duration_seconds" IS NOT NULL AND "duration_seconds" BETWEEN 0 AND 600)
  )
);

CREATE INDEX "Media_carepartner_kind_idx" ON "Media" ("carepartner_id", "kind");

-- One row per carepartner, created in the same transaction as the carepartner.
-- layout_preset values are validated in Go so new presets need no schema change.
CREATE TABLE "ScreenSettings" (
  "carepartner_id" integer PRIMARY KEY REFERENCES "Carepartner"("id") ON DELETE CASCADE,
  "layout_preset" text NOT NULL DEFAULT 'balanced',
  "wake_time" time NOT NULL DEFAULT '07:00' CHECK ("wake_time" BETWEEN '06:00' AND '10:00'),
  "bed_time" time NOT NULL DEFAULT '21:00' CHECK ("bed_time" BETWEEN '19:00' AND '22:00')
);

-- Widgets on the carepartner's screen, in order. Widgets that don't fit the
-- current layout preset are kept; the /screen endpoint reports them as overflow.
CREATE TABLE "ScreenWidget" (
  "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "carepartner_id" integer NOT NULL REFERENCES "Carepartner"("id") ON DELETE CASCADE,
  "type" text NOT NULL CHECK ("type" IN ('photo', 'message', 'reminder', 'weather', 'sports', 'news')),
  -- Gaps are allowed; readers order by position
  "position" integer NOT NULL CHECK ("position" >= 0),
  -- Validated per type in Go
  "config" jsonb NOT NULL DEFAULT '{}' CHECK (jsonb_typeof("config") = 'object'),
  -- SET NULL clears a message widget's photo. Photo widgets must be deleted by the
  -- media-delete transaction first; otherwise the CHECK below blocks the delete.
  "photo_media_id" integer REFERENCES "Media"("id") ON DELETE SET NULL,
  -- Required for photo widgets, optional for message widgets, not allowed otherwise
  CHECK (CASE "type"
    WHEN 'photo' THEN "photo_media_id" IS NOT NULL
    WHEN 'message' THEN TRUE
    ELSE "photo_media_id" IS NULL
  END),
  -- Deferred so a reorder can swap positions inside one transaction
  UNIQUE ("carepartner_id", "position") DEFERRABLE INITIALLY DEFERRED
);

CREATE INDEX "ScreenWidget_photo_media_idx" ON "ScreenWidget" ("photo_media_id");

-- Media played at a set time. Allowed shapes (media kind is checked in Go):
--   video, no background
--   audio, no background
--   audio with background_photo_id
--   photo with display_seconds
CREATE TABLE "ScheduledItem" (
  "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "carepartner_id" integer NOT NULL REFERENCES "Carepartner"("id") ON DELETE CASCADE,
  "media_id" integer NOT NULL REFERENCES "Media"("id") ON DELETE CASCADE,
  "background_photo_id" integer REFERENCES "Media"("id") ON DELETE SET NULL,
  "display_seconds" integer CHECK ("display_seconds" BETWEEN 1 AND 600),
  "start_date" date NOT NULL,
  -- 15 minute steps; awake hours are checked in Go so changing them never invalidates rows
  "start_time" time NOT NULL CHECK (
    EXTRACT(MINUTE FROM "start_time") IN (0, 15, 30, 45)
    AND EXTRACT(SECOND FROM "start_time") = 0
  ),
  "repeat" text NOT NULL CHECK ("repeat" IN ('once', 'daily', 'weekly', 'monthly')),
  "end_date" date CHECK ("end_date" IS NULL OR "end_date" >= "start_date"),
  "created_at" timestamptz NOT NULL DEFAULT NOW(),
  CHECK ("background_photo_id" IS NULL OR "display_seconds" IS NULL)
);

CREATE INDEX "ScheduledItem_carepartner_idx" ON "ScheduledItem" ("carepartner_id");
CREATE INDEX "ScheduledItem_media_idx" ON "ScheduledItem" ("media_id");
CREATE INDEX "ScheduledItem_background_photo_idx" ON "ScheduledItem" ("background_photo_id");

-- Days removed with "delete just this day"
CREATE TABLE "ScheduledItemSkip" (
  "scheduled_item_id" integer NOT NULL REFERENCES "ScheduledItem"("id") ON DELETE CASCADE,
  "skip_date" date NOT NULL,
  PRIMARY KEY ("scheduled_item_id", "skip_date")
);

-- Dev seed: one caregiver, one carepartner with default screen settings, no widgets

WITH caregiver_insert AS (
  INSERT INTO "Caregiver" (name, role, google_id)
  VALUES ('Jane Smith', 'ADMIN', 'test.caregiver@example.com')
  RETURNING id AS caregiver_id
),
carepartner_insert AS (
  INSERT INTO "Carepartner" (name, caregiver_id)
  SELECT 'John Doe', caregiver_id
  FROM caregiver_insert
  RETURNING id AS carepartner_id
)
INSERT INTO "ScreenSettings" (carepartner_id)
SELECT carepartner_id FROM carepartner_insert;
