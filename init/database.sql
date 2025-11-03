CREATE TABLE "Caregiver" (
  "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "carepartner_id" integer,
  "name" text,
  "role" text
);

CREATE TABLE "Carepartner" (
  "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "name" text
);

CREATE TABLE "Module" (
  "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "module_type" text,
  "caregiver_id" integer,
  "carepartner_id" integer,
  "metadata_id" integer
);

CREATE TABLE "Configuration" (
  "id" integer GENERATED ALWAYS AS IDENTITY UNIQUE,
  "carepartner_id" integer,
  "module_id" integer,
  "occurence_type" text,
  "occurence_definition" text[],
  "position" integer,
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
  "filename" text,
  "extension" text
);

CREATE TABLE "VideoMedia" (
  "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "uri" text,
  "filename" text,
  "extension" text
);

CREATE TABLE "AudioMedia" (
  "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "uri" text,
  "filename" text,
  "extension" text
);

CREATE TABLE "WeatherWidget" (
  "id" integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  "zipcode" integer
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
  "time" timestamp,
  "text" text
);