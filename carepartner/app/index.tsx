import React from "react";
import { Redirect } from "expo-router";
import NightScreen from "../src/display/NightScreen";
import MediaPlayer from "../src/playback/MediaPlayer";
import { usePlayback } from "../src/playback/PlaybackProvider";
import { PREVIEW } from "../src/preview/previewMode";
import PreviewUnavailable from "../src/preview/PreviewUnavailable";
import AlwaysOnScreen from "../src/screen/AlwaysOnScreen";

/** Shows whatever the playback machine says the screen should show. */
export default function Index() {
  const { state, screen, ended } = usePlayback();

  switch (screen) {
    case "starting":
      return null;
    case "unpaired":
      // A preview without a carepartner has nothing to show, and must never pair
      return PREVIEW ? <PreviewUnavailable /> : <Redirect href="/pair" />;
    case "playing": {
      const { playId, item } = state.playing!;
      return (
        <MediaPlayer
          key={playId}
          item={item}
          onFinished={() => ended(playId, "finished")}
          onFailed={(reason) => {
            console.warn("Playback failed:", reason);
            ended(playId, "failed");
          }}
        />
      );
    }
    case "night":
      return <NightScreen />;
    case "alwaysOn":
      return <AlwaysOnScreen />;
  }
}
