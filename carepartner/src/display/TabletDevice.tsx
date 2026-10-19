import { useKeepAwake } from "expo-keep-awake";
import { useHiddenSystemBars } from "./useHiddenSystemBars";

/**
 * What only a real tablet does with its screen: keeps it on, and hides
 * Android's status and navigation bars. Draws nothing. Left out in preview mode,
 * where the app runs in the caregiver's browser inside an iframe.
 */
export default function TabletDevice() {
  useKeepAwake();
  useHiddenSystemBars();
  return null;
}
