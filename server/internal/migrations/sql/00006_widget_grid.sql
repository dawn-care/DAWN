-- Widgets move from preset spots to a fixed grid, and layout presets go away.
-- The caregiver places each widget on the tablet's grid by its top-left and
-- bottom-right cells (both inclusive), so a widget can cover several cells. The
-- grid's size and the no-overlap rule are checked in Go, which saves a
-- carepartner's whole layout at once.
-- Existing widgets are deleted: they were placed by spot, which has no position on
-- the grid, and no frontend had saved any yet.
-- +goose Up

DELETE FROM "ScreenWidget";

-- Dropping spot also drops the unique (carepartner_id, spot) constraint
ALTER TABLE "ScreenWidget"
  DROP COLUMN "spot",
  ADD COLUMN "col_start" integer NOT NULL CHECK ("col_start" >= 0),
  ADD COLUMN "row_start" integer NOT NULL CHECK ("row_start" >= 0),
  ADD COLUMN "col_end" integer NOT NULL,
  ADD COLUMN "row_end" integer NOT NULL,
  ADD CHECK ("col_end" >= "col_start" AND "row_end" >= "row_start");

-- The unique constraint used to cover lookups by carepartner
CREATE INDEX "ScreenWidget_carepartner_idx" ON "ScreenWidget" ("carepartner_id");

-- There is one fixed grid, so there is no layout to choose
ALTER TABLE "ScreenSettings" DROP COLUMN "layout_preset";

-- +goose Down
-- Grid positions have no spot to go back to, so widgets are deleted again.

DELETE FROM "ScreenWidget";

DROP INDEX "ScreenWidget_carepartner_idx";

ALTER TABLE "ScreenWidget"
  DROP COLUMN "col_start",
  DROP COLUMN "row_start",
  DROP COLUMN "col_end",
  DROP COLUMN "row_end",
  ADD COLUMN "spot" integer NOT NULL CHECK ("spot" >= 0),
  ADD UNIQUE ("carepartner_id", "spot") DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE "ScreenSettings" ADD COLUMN "layout_preset" text NOT NULL DEFAULT 'balanced';
