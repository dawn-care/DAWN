-- Replaces the module/configuration content model with the i332 screen model:
-- one Media table, ScreenSettings, ScreenWidget (preset slots) and ScheduledItem.
-- Content in the old tables (media rows, widgets, prompts, agenda items) is dropped;
-- caregivers, carepartners, sessions and device pairings are kept. Files already in
-- MinIO are left in place. Every carepartner gets default screen settings.
-- +goose Up

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

-- +goose Down
-- Restores the old tables empty: content created in the new tables is lost.

DROP TABLE IF EXISTS "ScheduledItemSkip", "ScheduledItem", "ScreenWidget", "ScreenSettings", "Media";

ALTER TABLE "Carepartner" DROP CONSTRAINT IF EXISTS "Carepartner_caregiver_id_fkey";
ALTER TABLE "Carepartner" DROP COLUMN IF EXISTS "timezone";
ALTER TABLE "Carepartner" ADD COLUMN "auto_start" time DEFAULT '08:00:00';

CREATE TABLE "Module" (
  "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "module_type" text,
  "caregiver_id" integer,
  "carepartner_id" integer,
  "metadata_id" integer
);

CREATE TABLE "Configuration" (
  "id" integer GENERATED BY DEFAULT AS IDENTITY UNIQUE,
  "carepartner_id" integer,
  "module_id" integer,
  "occurence_type" text,
  "occurence_definition" text[],
  "position" integer,
  "duration" integer,
  "timed" boolean DEFAULT FALSE,
  "time" text,
  PRIMARY KEY ("carepartner_id", "position")
);

CREATE TABLE "Exception" (
  "configuration_id" integer,
  "exception_type" text,
  "exception_definition" text[]
);

CREATE TABLE "PhotoMedia" (
  "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "uri" text,
  "displayed_name" text,
  "filename" text,
  "extension" text
);

CREATE TABLE "VideoMedia" (
  "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "uri" text,
  "displayed_name" text,
  "filename" text,
  "extension" text
);

CREATE TABLE "AudioMedia" (
  "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "uri" text,
  "displayed_name" text,
  "filename" text,
  "extension" text
);

CREATE TABLE "WeatherWidget" (
  "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "zipcode" text
);

CREATE TABLE "NewsWidget" (
  "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "category" text,
  "language" text,
  "country" text
);

CREATE TABLE "CarepartnerLocationPrompt" (
  "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "location" text,
  "message" text,
  "image_uri" text
);

CREATE TABLE "CaregiverLocationPrompt" (
  "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "location" text,
  "message" text,
  "image_uri" text
);

CREATE TABLE "MessagePrompt" (
  "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "message" text,
  "image_uri" text
);

CREATE TABLE "EventPrompt" (
  "list_id" integer,
  "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "time" time,
  "text" text
);

-- +goose StatementBegin
CREATE OR REPLACE PROCEDURE insert_configurations(
    p_carepartner_id INT,
    p_configs JSONB
)
LANGUAGE plpgsql
AS $$
DECLARE
    new_row RECORD;
BEGIN
    CREATE TEMP TABLE tmp_configs AS
    SELECT
        (conf->>'id')::INT AS id,
        (conf->>'module_id')::INT AS module_id,
        conf->>'occurence_type' AS occurence_type,
        ARRAY(SELECT jsonb_array_elements_text(conf->'occurence_definition')) AS occurence_definition,
        (conf->>'position')::INT AS position,
        (conf->>'duration')::INT AS duration,
        COALESCE((conf->>'timed')::BOOLEAN, FALSE) AS timed,
        conf->>'time' AS "time"
    FROM jsonb_array_elements(p_configs) AS conf;

    -- Batch-delete all existing configurations being replaced
    DELETE FROM "Configuration"
    WHERE id IN (SELECT t.id FROM tmp_configs t WHERE t.id IS NOT NULL);

    -- Compact remaining positions to close gaps left by the deletions
    WITH ranked AS (
        SELECT id, ROW_NUMBER() OVER (ORDER BY position) AS new_pos
        FROM "Configuration"
        WHERE carepartner_id = p_carepartner_id
    )
    UPDATE "Configuration" c
    SET position = -ranked.new_pos
    FROM ranked
    WHERE c.id = ranked.id;

    UPDATE "Configuration"
    SET position = -position
    WHERE carepartner_id = p_carepartner_id
      AND position < 0;

    -- Insert all configurations, shifting existing rows as needed
    FOR new_row IN SELECT * FROM tmp_configs ORDER BY position ASC LOOP
        UPDATE "Configuration"
        SET position = -("Configuration".position + 1)
        FROM (
            SELECT id, position
            FROM "Configuration"
            WHERE carepartner_id = p_carepartner_id
              AND position >= new_row.position
            ORDER BY position DESC
        ) desc_positions
        WHERE "Configuration".id = desc_positions.id;

        UPDATE "Configuration"
        SET position = -position
        WHERE carepartner_id = p_carepartner_id
          AND position < 0;

        INSERT INTO "Configuration" (
            carepartner_id,
            module_id,
            occurence_type,
            occurence_definition,
            position,
            duration,
            timed,
            "time"
        ) VALUES (
            p_carepartner_id,
            new_row.module_id,
            new_row.occurence_type,
            new_row.occurence_definition,
            new_row.position,
            new_row.duration,
            new_row.timed,
            new_row."time"
        );
    END LOOP;

    DROP TABLE IF EXISTS tmp_configs;
END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE PROCEDURE delete_configurations(
    p_carepartner_id INT,
    id_list INT[]
)
LANGUAGE plpgsql
AS $$
DECLARE
    current_id INT;
    pos INT;
BEGIN
    FOREACH current_id IN ARRAY id_list
    LOOP
        SELECT position INTO pos FROM "Configuration" WHERE id = current_id;
        DELETE FROM "Configuration" WHERE id = current_id;

        UPDATE "Configuration"
        SET position = -(position - 1)
        WHERE carepartner_id = p_carepartner_id
          AND position > pos;

        UPDATE "Configuration"
        SET position = -position
        WHERE carepartner_id = p_carepartner_id
          AND position < 0;
    END LOOP;
END;
$$;
-- +goose StatementEnd
