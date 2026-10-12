-- Replaces the module/configuration content model with the i332 screen model:
-- one Media table, ScreenSettings, ScreenWidget (preset slots) and ScheduledItem.
-- Content in the old tables (media rows, widgets, prompts, agenda items) is dropped;
-- caregivers, carepartners, sessions and device pairings are kept. Files already in
-- MinIO are left in place. Every carepartner gets default screen settings.
-- Playback (migration 2) pointed at modules; it now points at the media that played
-- and, for scheduled plays, the scheduled item. Existing playback history is cleared,
-- since the modules it refers to are dropped.
-- +goose Up

-- Playback rows refer to modules, which are dropped below
DELETE FROM "Playback";
ALTER TABLE "Playback" DROP COLUMN "module_id";

DROP PROCEDURE IF EXISTS insert_configurations(INT, JSONB);
DROP PROCEDURE IF EXISTS delete_configurations(INT, INT[]);

DROP TABLE IF EXISTS
  "Exception", "Configuration", "Module",
  "PhotoMedia", "VideoMedia", "AudioMedia",
  "WeatherWidget", "NewsWidget",
  "CarepartnerLocationPrompt", "CaregiverLocationPrompt", "MessagePrompt", "EventPrompt";

-- wake_time in ScreenSettings replaces auto_start
ALTER TABLE "Carepartner" DROP COLUMN "auto_start";

-- Dates and times elsewhere are stored as local wall time in this zone
ALTER TABLE "Carepartner" ADD COLUMN "timezone" text NOT NULL DEFAULT 'America/New_York';

-- A carepartner is created unpaired; the old code stored 0 instead of NULL, and nothing
-- enforced the link, so clear values that don't point at a caregiver before adding the key
UPDATE "Carepartner" SET "caregiver_id" = NULL
WHERE "caregiver_id" IS NOT NULL
  AND "caregiver_id" NOT IN (SELECT "id" FROM "Caregiver");

ALTER TABLE "Carepartner"
  ADD CONSTRAINT "Carepartner_caregiver_id_fkey"
  FOREIGN KEY ("caregiver_id") REFERENCES "Caregiver"("id") ON DELETE RESTRICT;

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

-- Widgets placed in the slots of a layout preset's regions (e.g. the second of
-- Balanced's three tiles is region 'tiles', slot 1). Region names and sizes are
-- defined per preset in Go. A widget whose slot doesn't exist in the current
-- preset is kept but hidden, and reappears if the caregiver switches back.
CREATE TABLE "ScreenWidget" (
  "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "carepartner_id" integer NOT NULL REFERENCES "Carepartner"("id") ON DELETE CASCADE,
  "type" text NOT NULL CHECK ("type" IN ('photo', 'message', 'reminder', 'weather', 'sports', 'news')),
  "region" text NOT NULL CHECK ("region" <> ''),
  "slot" integer NOT NULL CHECK ("slot" >= 0),
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
  -- One widget per slot; deferred so two widgets can swap slots in one transaction
  UNIQUE ("carepartner_id", "region", "slot") DEFERRABLE INITIALLY DEFERRED
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

-- Every carepartner gets the default settings (Balanced, 07:00 to 21:00)
INSERT INTO "ScreenSettings" ("carepartner_id")
SELECT "id" FROM "Carepartner";

-- A playback records the media that played; scheduled plays also record the scheduled item.
-- NOT NULL is safe because the table was emptied above.
ALTER TABLE "Playback"
  ADD COLUMN "media_id" integer NOT NULL REFERENCES "Media"("id") ON DELETE CASCADE,
  ADD COLUMN "scheduled_item_id" integer REFERENCES "ScheduledItem"("id") ON DELETE SET NULL;

CREATE INDEX "Playback_media_idx" ON "Playback" ("media_id");
CREATE INDEX "Playback_scheduled_item_idx" ON "Playback" ("scheduled_item_id");

-- +goose Down
-- Forward only, like the playback migration: the old content can't be restored.
-- +goose StatementBegin
DO $$
BEGIN
    RAISE EXCEPTION 'The screen, media and schedule migration cannot be rolled back. Restore a database backup instead.';
END;
$$;
-- +goose StatementEnd
