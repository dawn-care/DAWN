import React from "react";
import { Redirect } from "expo-router";
import NightScreen from "../src/display/NightScreen";
import ChimeScreen from "../src/playback/ChimeScreen";
import MediaPlayer from "../src/playback/MediaPlayer";
import { usePlayback } from "../src/playback/PlaybackProvider";
import { PREVIEW } from "../src/preview/previewMode";
import PreviewUnavailable from "../src/preview/PreviewUnavailable";
import AlwaysOnScreen from "../src/screen/AlwaysOnScreen";
import { useScreen } from "../src/screen/ScreenProvider";
import { TouchArea } from "../src/telemetry/TouchTelemetry";

/** Shows whatever the playback machine says the screen should show. */
export default function Index() {
  const { state, screen, chimed, started, ended } = usePlayback();
  const loaded = useScreen();
  // On unless staff have switched it off for this resident (#387).
  const chime = loaded.screen?.chime ?? true;

  switch (screen) {
    case "starting":
      return null;
    case "unpaired":
      // A preview without a carepartner has nothing to show, and must never pair
      return PREVIEW ? <PreviewUnavailable /> : <Redirect href="/pair" />;
    case "playing": {
      const { playId, item, phase } = state.playing!;
      return (
        <TouchArea screen="playing" target={{ type: "media", mediaKind: item.kind }} style={{ flex: 1 }}>
          {phase === "chime" ? (
            <ChimeScreen key={playId} kind={item.kind} sound={chime} night={state.night} onDone={() => chimed(playId)} />
          ) : (
            <MediaPlayer
              key={playId}
              item={item}
              onStarted={() => started(playId)}
              onFinished={() => ended(playId, "finished")}
              onFailed={(reason) => {
                console.warn("Playback failed:", reason);
                ended(playId, "failed");
              }}
            />
          )}
        </TouchArea>
      );
    }
    case "night":
      return (
        <TouchArea screen="night" style={{ flex: 1 }}>
          <NightScreen />
        </TouchArea>
      );
    case "alwaysOn":
      // The tablet shows the day and time until a screen arrives; a preview would
      // pass that off as the carepartner's screen, so it says the load failed
      if (PREVIEW && !loaded.screen && loaded.failed) {
        return <PreviewUnavailable message="This screen couldn't be loaded." />;
      }
      return <AlwaysOnScreen />;
  }
}
