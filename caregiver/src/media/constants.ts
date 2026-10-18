/**
 * constants.ts
 *
 * User: Caregiver
 *
 * Fixed values for the media page: the file types the server accepts for
 * each kind, the 10 minute limit, and the wording for server error codes.
 * The server decides what it accepts by sniffing the file's bytes
 * (server/internal/controllers/media/media_controller.go); these lists
 * follow what it accepts (the docs/api/screen-and-media.md list leaves out
 * a few formats, e.g. AAC) so the picker and the messages match, but the
 * server stays the authority.
 */

export type MediaKind = "photo" | "video" | "audio";

// Longest video or audio the server accepts (MaxDurationSeconds)
export const MAX_DURATION_SECONDS = 600;

// Largest file the server accepts (MAX_MEDIA_UPLOAD_MB), 1 GB
export const MAX_FILE_BYTES = 1024 * 1024 * 1024;

export const KIND_LABELS: Record<MediaKind, string> = {
  photo: "Photo",
  video: "Video",
  audio: "Audio",
};

export interface AcceptedTypes {
  /** MIME types the browser reports for these files */
  mimeTypes: string[];
  /**
   * File extensions, lower case with the dot. Used when the browser gives no
   * MIME type or a generic one (HEIC and MOV often have none)
   */
  extensions: string[];
}

export const ACCEPTED_TYPES: Record<MediaKind, AcceptedTypes> = {
  photo: {
    mimeTypes: ["image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp", "image/heic", "image/heif"],
    extensions: [".jpg", ".jpeg", ".png", ".gif", ".webp", ".heic", ".heif"],
  },
  video: {
    mimeTypes: ["video/mp4", "video/quicktime", "video/webm", "video/ogg"],
    extensions: [".mp4", ".mov", ".webm", ".ogv", ".ogg"],
  },
  audio: {
    mimeTypes: [
      "audio/mpeg",
      "audio/mp3",
      "audio/wav",
      "audio/wave",
      "audio/x-wav",
      "audio/mp4",
      "audio/x-m4a",
      "audio/aac",
      "audio/ogg",
      "audio/webm",
    ],
    extensions: [".mp3", ".wav", ".m4a", ".aac", ".ogg", ".oga", ".webm"],
  },
};

// Value for <input type="file" accept>: every accepted MIME type and extension
export const FILE_INPUT_ACCEPT = Object.values(ACCEPTED_TYPES)
  .flatMap((types) => [...types.mimeTypes, ...types.extensions])
  .join(",");

// Shown when the server gives a code we have no wording for
export const GENERIC_ERROR = "Something went wrong. Please try again.";

// Caregiver-facing wording for the server's error codes (docs/api/screen-and-media.md), plus NETWORK_ERROR
export const ERROR_MESSAGES: Record<string, string> = {
  INVALID_REQUEST: GENERIC_ERROR,
  INTERNAL_ERROR: GENERIC_ERROR,
  // Not a server code: the request never got an answer
  NETWORK_ERROR: "Can't reach the server. Check your connection and try again.",
  NOT_FOUND: "This was deleted. Refresh to see the latest.",
  INVALID_KIND: "Choose photo, video or audio.",
  INVALID_FILE_TYPE: "That file type isn't supported.",
  INVALID_NAME: "Enter a name up to 200 characters.",
  FILE_TOO_LARGE: "That file is too large. Try a shorter clip.",
  MEDIA_TOO_LONG: "Videos and audio can be at most 10 minutes long.",
  UNREADABLE_MEDIA: "That file couldn't be read. Try exporting it again.",
};
