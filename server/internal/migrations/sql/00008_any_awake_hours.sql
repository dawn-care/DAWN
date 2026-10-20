-- Awake hours can be any times in one day, as long as bed time is later than
-- wake time. They used to be limited to waking 06:00 to 10:00 and bed 19:00 to
-- 22:00, which a caregiver of someone who stays up late couldn't set.
-- +goose Up

ALTER TABLE "ScreenSettings"
  DROP CONSTRAINT IF EXISTS "ScreenSettings_wake_time_check",
  DROP CONSTRAINT IF EXISTS "ScreenSettings_bed_time_check",
  ADD CONSTRAINT "ScreenSettings_bed_after_wake" CHECK ("bed_time" > "wake_time");

-- +goose Down
-- Hours outside the old ranges are moved to the nearest allowed time first.

ALTER TABLE "ScreenSettings" DROP CONSTRAINT "ScreenSettings_bed_after_wake";

UPDATE "ScreenSettings" SET
  "wake_time" = LEAST(GREATEST("wake_time", '06:00'), '10:00'),
  "bed_time" = LEAST(GREATEST("bed_time", '19:00'), '22:00');

ALTER TABLE "ScreenSettings"
  ADD CONSTRAINT "ScreenSettings_wake_time_check" CHECK ("wake_time" BETWEEN '06:00' AND '10:00'),
  ADD CONSTRAINT "ScreenSettings_bed_time_check" CHECK ("bed_time" BETWEEN '19:00' AND '22:00');
