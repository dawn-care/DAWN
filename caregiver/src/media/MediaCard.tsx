/**
 * MediaCard.tsx
 *
 * User: Caregiver
 *
 * One item in the media grid: a thumbnail (the photo itself, or an icon for
 * video and audio) with its type and length badge, an optional "New" badge,
 * its name, and where it is used. The whole card is a toggle button that
 * selects the item for the detail panel.
 */

import { useState } from "react";
import { Image, Music, Video } from "lucide-react";
import type { Media } from "./api";
import { fileUrl } from "./api";
import type { MediaKind } from "./constants";
import { badgeLabel, mediaStatus } from "./labels";
import Badge from "../components/ui/Badge";
import { FOCUS_RING } from "../components/ui/styles";

const KIND_ICONS: Record<MediaKind, typeof Video> = {
  photo: Image,
  video: Video,
  audio: Music,
};

interface MediaCardProps {
  media: Media;
  selected: boolean;
  /** Just saved by the caregiver; shows the "New" badge */
  isNew?: boolean;
  onSelect: () => void;
}

export default function MediaCard({ media, selected, isNew = false, onSelect }: MediaCardProps) {
  // Browsers can't draw every photo type (HEIC outside Safari); show the icon instead
  const [photoBroken, setPhotoBroken] = useState(false);
  const status = mediaStatus(media);
  const Icon = KIND_ICONS[media.kind];
  const showPhoto = media.kind === "photo" && !photoBroken;

  return (
    <button
      type="button"
      aria-pressed={selected}
      onClick={onSelect}
      className={`flex flex-col gap-1.5 px-1.5 pt-1.5 pb-2.5 bg-white rounded-xl border-[3px] text-left text-dawn-ink cursor-pointer ${
        selected ? "border-dawn-accent" : "border-dawn-line"
      } ${FOCUS_RING}`}
    >
      <span className="relative h-[92px] flex items-center justify-center overflow-hidden rounded-lg bg-dawn-avatar text-dawn-avatar-ink">
        {showPhoto ? (
          <img
            src={fileUrl(media)}
            alt=""
            loading="lazy"
            onError={() => setPhotoBroken(true)}
            className="w-full h-full object-cover"
          />
        ) : (
          <Icon size={32} aria-hidden="true" />
        )}
        <span className="absolute left-1.5 top-1.5">
          <Badge>{badgeLabel(media)}</Badge>
        </span>
        {isNew && (
          <span className="absolute right-1.5 top-1.5">
            <Badge tone="accent">New</Badge>
          </span>
        )}
      </span>
      <span className="px-1 text-[15px] font-bold leading-tight break-words">{media.display_name}</span>
      <span
        className={`px-1 text-sm leading-tight ${status.active ? "text-dawn-success" : "text-dawn-muted"}`}
      >
        {status.text}
      </span>
    </button>
  );
}
