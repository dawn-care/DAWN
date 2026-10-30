-- Staff rate a play from the staff app on five steps: 2 = helped a lot, 1 = helped a
-- little, 0 = no change, -1 = made it a little worse, -2 = made it much worse
-- (Rating* in models/playback). The ranking weighs these in Go, so changing how much
-- a rating counts needs no migration.
-- Nothing wrote ratings before this, so any other value is cleared rather than
-- guessed at.
-- +goose Up

UPDATE "Playback" SET "rating" = NULL WHERE "rating" NOT BETWEEN -2 AND 2;

ALTER TABLE "Playback"
  ADD CONSTRAINT "Playback_rating_scale" CHECK ("rating" BETWEEN -2 AND 2);

-- +goose Down

ALTER TABLE "Playback" DROP CONSTRAINT "Playback_rating_scale";
