import React, { useEffect } from "react";
import { router, type Href } from "expo-router";
import PairScreen from "../src/pairing/PairScreen";
import { usePairing } from "../src/pairing/usePairing";

/** Pairs an unpaired tablet (`index.tsx` sends it here), then returns home. */
export default function PairRoute() {
  const { view, newCode } = usePairing();

  useEffect(() => {
    if (view.kind === "paired") router.replace("/" as Href);
  }, [view.kind]);

  return <PairScreen view={view} onNewCode={newCode} />;
}
