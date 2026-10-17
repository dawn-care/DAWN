import React from "react";
import { Redirect } from "expo-router";
import NightScreen from "../src/display/NightScreen";
import MediaPlayer from "../src/playback/MediaPlayer";
import { usePlayback } from "../src/playback/PlaybackProvider";
import AlwaysOnScreen from "../src/screen/AlwaysOnScreen";

/** Shows whatever the playback machine says the screen should show. */
export default function Index() {
  const { state, screen, ended } = usePlayback();

  switch (screen) {
    case "starting":
      return null;
    case "unpaired":
      return <Redirect href="/pair" />;
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
