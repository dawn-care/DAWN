-- Touches the resident makes on the tablet (ticket #381), sent in batches by the
-- tablet's touch telemetry (carepartner/src/telemetry/touchLog.ts). Nothing on the
-- tablet needs a tap, so these show where residents expect to interact.
-- One row is one event: quick repeats of the same gesture on the same thing are
-- grouped by the tablet into one event with a count. There are no positions and no
-- text, only these structured fields.
-- widget_type and widget_id are kept as sent, with no foreign key, so a touch still
-- says what was touched after the widget is deleted.
-- The unique key makes a resent batch harmless: the tablet retries after a server
-- error, and a batch that was saved but whose response was lost comes back again.
-- Two different events can't share a screen, gesture and first touch to the
-- millisecond, so the key never merges real events.
-- +goose Up

CREATE TABLE "TabletTouch" (
  "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "carepartner_id" integer NOT NULL REFERENCES "Carepartner"("id") ON DELETE CASCADE,
  "screen" text NOT NULL CHECK ("screen" IN ('always_on', 'default', 'night', 'playing')),
  -- NULL when the screen has nothing to name under the finger, such as the night screen
  "target_type" text CHECK ("target_type" IN ('widget', 'empty', 'media')),
  "widget_type" text CHECK ("widget_type" IN ('photo', 'message', 'reminder', 'weather', 'sports', 'news')),
  -- The ScreenWidget id as the tablet sent it
  "widget_id" text CHECK ("widget_id" ~ '^[1-9][0-9]{0,9}$'),
  "media_kind" text CHECK ("media_kind" IN ('video', 'audio', 'photo')),
  "gesture" text NOT NULL CHECK ("gesture" IN ('tap', 'long_press', 'swipe')),
  "count" integer NOT NULL CHECK ("count" >= 1),
  "first_at" timestamptz NOT NULL,
  "last_at" timestamptz NOT NULL,
  "received_at" timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT "TabletTouch_times" CHECK ("first_at" <= "last_at"),
  -- Widget fields only for a widget, media_kind only for media. COALESCE keeps a NULL
  -- target_type from making the check NULL, which would let any row through.
  CONSTRAINT "TabletTouch_target_fields" CHECK (
    CASE COALESCE("target_type", 'none')
      WHEN 'widget' THEN "widget_type" IS NOT NULL AND "widget_id" IS NOT NULL AND "media_kind" IS NULL
      WHEN 'media' THEN "media_kind" IS NOT NULL AND "widget_type" IS NULL AND "widget_id" IS NULL
      ELSE "widget_type" IS NULL AND "widget_id" IS NULL AND "media_kind" IS NULL
    END
  ),
  -- Also serves reading one carepartner's touches by time
  CONSTRAINT "TabletTouch_event_key" UNIQUE ("carepartner_id", "first_at", "screen", "gesture")
);

-- +goose Down
-- Touches are design feedback, not care history, so rolling back drops them.

DROP TABLE "TabletTouch";
