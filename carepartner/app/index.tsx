import React from "react";
import { Redirect } from "expo-router";
import NightScreen from "../src/display/NightScreen";
import MediaPlayer from "../src/playback/MediaPlayer";
import { usePlayback } from "../src/playback/PlaybackProvider";
import AlwaysOnScreen from "../src/screen/AlwaysOnScreen";
import { TouchArea } from "../src/telemetry/TouchTelemetry";

/** Shows whatever the playback machine says the screen should show. */
export default function Index() {
  const { state, screen, started, ended } = usePlayback();

  switch (screen) {
    case "starting":
      return null;
    case "unpaired":
      return <Redirect href="/pair" />;
    case "playing": {
      const { playId, item } = state.playing!;
      return (
        <TouchArea screen="playing" target={{ type: "media", mediaKind: item.kind }} style={{ flex: 1 }}>
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
      return <AlwaysOnScreen />;
  }
}
