-- Playback history outlives the media it played. Deleting a media item used to delete
-- its playbacks (and their ratings); now it clears media_id instead. Each playback also
-- keeps the media's kind and display name as they were when it played, so plays of
-- deleted media still say what played.
-- +goose Up

ALTER TABLE "Playback"
  ADD COLUMN "media_kind" text CHECK ("media_kind" IN ('photo', 'video', 'audio')),
  ADD COLUMN "media_display_name" text;

UPDATE "Playback" p
SET "media_kind" = m."kind", "media_display_name" = m."display_name"
FROM "Media" m
WHERE m."id" = p."media_id";

ALTER TABLE "Playback"
  ALTER COLUMN "media_kind" SET NOT NULL,
  ALTER COLUMN "media_display_name" SET NOT NULL,
  ALTER COLUMN "media_id" DROP NOT NULL,
  DROP CONSTRAINT "Playback_media_id_fkey",
  ADD CONSTRAINT "Playback_media_id_fkey"
    FOREIGN KEY ("media_id") REFERENCES "Media"("id") ON DELETE SET NULL;

-- +goose Down
-- Playbacks of deleted media have nothing to point at, so they are removed, as
-- the cascade before this migration would have done.

DELETE FROM "Playback" WHERE "media_id" IS NULL;

ALTER TABLE "Playback"
  DROP CONSTRAINT "Playback_media_id_fkey",
  ADD CONSTRAINT "Playback_media_id_fkey"
    FOREIGN KEY ("media_id") REFERENCES "Media"("id") ON DELETE CASCADE,
  ALTER COLUMN "media_id" SET NOT NULL,
  DROP COLUMN "media_display_name",
  DROP COLUMN "media_kind";
