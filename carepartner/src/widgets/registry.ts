import type { ComponentType } from "react";
import Headline from "./renderers/Headline";
import Message from "./renderers/Message";
import Orientation from "./renderers/Orientation";
import Photo from "./renderers/Photo";
import Reminder from "./renderers/Reminder";
import Tile from "./renderers/Tile";
import type { RendererName, WidgetPlacement } from "./types";

type Entry<R extends RendererName> = {
  component: ComponentType<{ data: Extract<WidgetPlacement, { renderer: R }>["data"] }>;
  /** Whether the widget sits in a tile frame or directly on the screen background. */
  framed: boolean;
};

/** Every look the tablet can draw. Adding a renderer is one file plus one line here. */
export const registry: { [R in RendererName]: Entry<R> } = {
  orientation: { component: Orientation, framed: false },
  photo: { component: Photo, framed: false },
  tile: { component: Tile, framed: true },
  message: { component: Message, framed: true },
  reminder: { component: Reminder, framed: true },
  headline: { component: Headline, framed: true },
};

/** The registry entry for a renderer name, or undefined for one this app version doesn't know. */
export function lookupRenderer(name: string): Entry<RendererName> | undefined {
  return Object.prototype.hasOwnProperty.call(registry, name)
    ? (registry[name as RendererName] as Entry<RendererName>)
    : undefined;
}
