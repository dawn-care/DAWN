-- Staff rate a play from the staff app: 1 = helped, 0 = no change, -1 = made it worse
-- (Rating* in models/playback). The ranking weighs these in Go, so changing how much
-- a rating counts needs no migration.
-- Nothing wrote ratings before this, so any other value is cleared rather than
-- guessed at.
-- +goose Up

UPDATE "Playback" SET "rating" = NULL WHERE "rating" NOT IN (-1, 0, 1);

ALTER TABLE "Playback"
  ADD CONSTRAINT "Playback_rating_scale" CHECK ("rating" IN (-1, 0, 1));

-- +goose Down

ALTER TABLE "Playback" DROP CONSTRAINT "Playback_rating_scale";
