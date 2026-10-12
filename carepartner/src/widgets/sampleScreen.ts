import { resolveApiUri } from "../api/resolveApiUri";
import type { WidgetPlacement } from "./types";

/**
 * A sample always-on screen, arranged like the visual design, for reviewing
 * widgets before the server's screen endpoint exists.
 */
export const sampleScreen: WidgetPlacement[] = [
  { id: "orientation", renderer: "orientation", col: 0, row: 0, width: 2, height: 2, data: { showDate: true } },
  {
    id: "photo",
    renderer: "photo",
    col: 2,
    row: 0,
    width: 2,
    height: 3,
    // The sunset photo every local backend is seeded with.
    data: { uri: resolveApiUri("/photo/file?id=1"), name: "Sunset" },
  },
  { id: "weather", renderer: "tile", col: 0, row: 2, width: 1, height: 1, data: { icon: "partly-cloudy", word: "Chilly" } },
  { id: "sports", renderer: "tile", col: 1, row: 2, width: 1, height: 1, data: { icon: "baseball", word: "Won", label: "Yankees" } },
];
