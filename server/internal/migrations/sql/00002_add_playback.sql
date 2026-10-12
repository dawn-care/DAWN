-- +goose Up
CREATE TABLE "Playback" (
  "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "carepartner_id" integer NOT NULL REFERENCES "Carepartner"("id"),
  "module_id" integer NOT NULL REFERENCES "Module"("id"),
  "source" text NOT NULL CHECK ("source" IN ('scheduled', 'staff', 'tap')),
  "status" text NOT NULL DEFAULT 'requested' CHECK (
    "status" IN ('requested', 'acknowledged', 'started', 'finished', 'stopped', 'failed', 'expired')
  ),
  "requested_at" TIMESTAMP NOT NULL DEFAULT NOW(),
  "acknowledged_at" TIMESTAMP,
  "started_at" TIMESTAMP,
  -- Terminal status identifies whether playback finished, stopped, failed or expired.
  "ended_at" TIMESTAMP,
  "expires_at" TIMESTAMP NOT NULL,
  -- The rating scale is not specified yet; do not assume a range.
  "rating" integer
);

-- +goose Down
-- Preserve playback history, following the initial migration's forward-only policy.
-- +goose StatementBegin
DO $$
BEGIN
    RAISE EXCEPTION 'The playback migration cannot be rolled back. Restore a database backup instead.';
END;
$$;
-- +goose StatementEnd
