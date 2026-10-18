/**
 * What every player takes, whatever it plays.
 *
 * Every player keeps the same contract, so `MediaPlayer` can play any kind the same way:
 * - It starts on its own as soon as it is on screen.
 * - It calls `onStarted` once when the media is actually playing (a photo:
 *   once it has loaded), and never after it has ended.
 * - It calls exactly one of `onFinished` or `onFailed`, once.
 * - Once taken off screen, it stops and reports nothing more.
 */
export type PlayerProps = {
  uri: string;
  /** Called once when the media is actually playing, so reports say "started" only then. */
  onStarted?: () => void;
  /** Called once when playback reaches the end. */
  onFinished: () => void;
  /** Called once if the media can't be played. */
  onFailed: (reason: string) => void;
};
