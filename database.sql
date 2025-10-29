CREATE TABLE "Caregiver" (
  "id" integer PRIMARY KEY,
  "carepartner_id" integer,
  "name" varchar,
  "role" varchar
);

CREATE TABLE "Carepartner" (
  "id" integer PRIMARY KEY,
  "name" varchar
);

CREATE TABLE "Module" (
  "id" integer PRIMARY KEY,
  "module_type" varchar,
  "caregiver_id" integer,
  "carepartner_id" integer,
  "metadata" integer
);

CREATE TABLE "Configuration" (
  "id" integer PRIMARY KEY,
  "carepartner_id" integer,
  "module_id" varchar,
  "occurence_type" varchar,
  "occurence_definition" varchar[],
  "position" integer
);

CREATE TABLE "Exception" (
  "configuration_id" integer,
  "exception_type" varchar,
  "exception_definition" varchar[]
);

CREATE TABLE "PhotoMedia" (
  "id" integer PRIMARY KEY,
  "uri" varchar,
  "filename" varchar,
  "extension" integer
);

CREATE TABLE "VideoMedia" (
  "id" integer PRIMARY KEY,
  "uri" varchar,
  "filename" varchar,
  "extension" integer
);

CREATE TABLE "AudioMedia" (
  "id" integer PRIMARY KEY,
  "uri" varchar,
  "filename" varchar,
  "extension" integer
);

CREATE TABLE "WeatherWidget" (
  "id" integer PRIMARY KEY,
  "location" varchar
);

CREATE TABLE "NewsWidget" (
  "id" integer PRIMARY KEY,
  "type" varchar,
  "configuration" json
);

CREATE TABLE "CarepartnerLocationPrompt" (
  "id" integer PRIMARY KEY,
  "location" varchar,
  "message" varchar,
  "image_uri" varchar
);

CREATE TABLE "CaregiverLocationPrompt" (
  "id" integer PRIMARY KEY,
  "location" varchar,
  "message" varchar,
  "image_uri" varchar
);

CREATE TABLE "MessagePrompt" (
  "id" integer PRIMARY KEY,
  "message" varchar,
  "image_uri" varchar
);

CREATE TABLE "EventPrompt" (
  "list_id" integer,
  "id" integer PRIMARY KEY,
  "time" datetime,
  "text" varchar
);

ALTER TABLE "Caregiver" ADD FOREIGN KEY ("carepartner_id") REFERENCES "Carepartner" ("id");

ALTER TABLE "Caregiver" ADD FOREIGN KEY ("id") REFERENCES "Module" ("caregiver_id");

ALTER TABLE "Carepartner" ADD FOREIGN KEY ("id") REFERENCES "Module" ("carepartner_id");

ALTER TABLE "Configuration" ADD FOREIGN KEY ("module_id") REFERENCES "Module" ("id");

ALTER TABLE "Module" ADD FOREIGN KEY ("metadata") REFERENCES "PhotoMedia" ("id");

ALTER TABLE "Module" ADD FOREIGN KEY ("metadata") REFERENCES "VideoMedia" ("id");

ALTER TABLE "Module" ADD FOREIGN KEY ("metadata") REFERENCES "AudioMedia" ("id");

ALTER TABLE "Module" ADD FOREIGN KEY ("metadata") REFERENCES "WeatherWidget" ("id");

ALTER TABLE "Module" ADD FOREIGN KEY ("metadata") REFERENCES "NewsWidget" ("id");

ALTER TABLE "Module" ADD FOREIGN KEY ("metadata") REFERENCES "CarepartnerLocationPrompt" ("id");

ALTER TABLE "Module" ADD FOREIGN KEY ("metadata") REFERENCES "CaregiverLocationPrompt" ("id");

ALTER TABLE "Module" ADD FOREIGN KEY ("metadata") REFERENCES "MessagePrompt" ("id");

ALTER TABLE "EventPrompt" ADD FOREIGN KEY ("list_id") REFERENCES "Module" ("metadata");

ALTER TABLE "Exception" ADD FOREIGN KEY ("configuration_id") REFERENCES "Configuration" ("id");

ALTER TABLE "Carepartner" ADD FOREIGN KEY ("id") REFERENCES "Configuration" ("carepartner_id");