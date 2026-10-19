-- Media can be sorted for staff by situation and by family member. A carepartner's
-- family members are folders for grouping media, not accounts; a guidance note tells
-- staff what to say (or avoid) when the carepartner asks for that person.
-- Situation names are validated in Go (Situations in models/media), so the list can
-- change without a schema change. "Asking for someone" is not one of them: staff
-- reach that media through the family member instead.
-- Existing media is put under every situation, which is also the default for new
-- uploads, so nothing disappears from the staff app.
-- +goose Up

CREATE TABLE "FamilyMember" (
  "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "carepartner_id" integer NOT NULL REFERENCES "Carepartner"("id") ON DELETE CASCADE,
  "name" text NOT NULL,
  "relationship" text NOT NULL DEFAULT '',
  "guidance_note" text,
  "created_at" timestamptz NOT NULL DEFAULT NOW()
);

CREATE INDEX "FamilyMember_carepartner_idx" ON "FamilyMember" ("carepartner_id");

CREATE TABLE "MediaSituation" (
  "media_id" integer NOT NULL REFERENCES "Media"("id") ON DELETE CASCADE,
  "situation" text NOT NULL,
  PRIMARY KEY ("media_id", "situation")
);

CREATE INDEX "MediaSituation_situation_idx" ON "MediaSituation" ("situation");

-- The family member must belong to the media's carepartner; that is checked in Go
CREATE TABLE "MediaFamilyMember" (
  "media_id" integer NOT NULL REFERENCES "Media"("id") ON DELETE CASCADE,
  "family_member_id" integer NOT NULL REFERENCES "FamilyMember"("id") ON DELETE CASCADE,
  PRIMARY KEY ("media_id", "family_member_id")
);

CREATE INDEX "MediaFamilyMember_family_member_idx" ON "MediaFamilyMember" ("family_member_id");

INSERT INTO "MediaSituation" ("media_id", "situation")
SELECT m."id", s.situation
FROM "Media" m
CROSS JOIN unnest(ARRAY['wants_to_go_home', 'doesnt_know_where', 'anxious_or_upset',
                        'starting_the_day', 'winding_down']) AS s(situation);

-- Demo family for the demo carepartners
INSERT INTO "FamilyMember" ("carepartner_id", "name", "relationship", "guidance_note")
SELECT c."id", f.name, f.relationship, f.guidance_note
FROM "Carepartner" c
CROSS JOIN (VALUES
  (1, 'Elena', 'Daughter', NULL),
  (2, 'Marco', 'Son', NULL),
  (3, 'Sofia', 'Granddaughter', NULL),
  (4, 'Daniel', 'Husband', 'He passed away in 2023. Please don''t tell them again. Reassure them and talk about their garden.')
) AS f(position, name, relationship, guidance_note)
WHERE c."name" IN ('John Doe', 'Jane Doe')
ORDER BY c."id", f.position;

-- +goose Down

DROP TABLE "MediaFamilyMember";
DROP TABLE "MediaSituation";
DROP TABLE "FamilyMember";
