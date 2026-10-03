export type MediaKind = "video" | "audio";

/** What every player takes, whatever it plays. */
export type MediaPlayerProps = {
  uri: string;
  /** Called once when playback reaches the end. */
  onFinished: () => void;
  /** Called once if the media can't be played. */
  onFailed: (reason: string) => void;
};
