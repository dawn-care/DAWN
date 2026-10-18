import { usePlayReports } from "./usePlayReports";

/** Reports scheduled plays to the server while mounted. Draws nothing. */
export default function PlayReports() {
  usePlayReports();
  return null;
}
