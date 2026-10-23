/**
 * PhotoPicker.tsx
 *
 * User: Caregiver
 *
 * A row of selectable photo thumbnails ("Which photo"). Used by the photo
 * widget pop-up and the audio background form. Built on Radix RadioGroup,
 * so exactly one photo is chosen and arrow keys move between them. A long
 * list scrolls sideways instead of squeezing the thumbnails.
 */

import * as RadixRadioGroup from "@radix-ui/react-radio-group";
import { Check } from "lucide-react";
import { FOCUS_RING } from "./styles";

export interface PickerPhoto {
  id: string;
  /** Accessible name, e.g. the photo's title */
  label: string;
  /** Thumbnail image URL */
  src: string;
}

interface PhotoPickerProps {
  photos: PickerPhoto[];
  /** Id of the chosen photo, or "" when none is chosen */
  value: string;
  onChange: (id: string) => void;
  /** Names the group for screen readers */
  label: string;
  /** Shown instead of the row when there are no photos */
  emptyText?: string;
}

export default function PhotoPicker({
  photos,
  value,
  onChange,
  label,
  emptyText = "No photos to choose from.",
}: PhotoPickerProps) {
  if (photos.length === 0) {
    return <p className="m-0 text-base text-ink-muted">{emptyText}</p>;
  }

  return (
    // p-1 leaves room for the focus ring, which the scroll container would otherwise clip
    <RadixRadioGroup.Root
      aria-label={label}
      value={value}
      onValueChange={onChange}
      className="flex gap-2 p-1 -m-1 overflow-x-auto"
    >
      {photos.map((photo) => (
        <div key={photo.id} className="flex-1 min-w-[72px] flex flex-col gap-1">
          <RadixRadioGroup.Item
            value={photo.id}
            aria-label={photo.label}
            className={`relative w-full h-[54px] p-0 overflow-hidden rounded-lg border-2 border-border bg-border bg-cover bg-center cursor-pointer data-[state=checked]:border-[3px] data-[state=checked]:border-accent ${FOCUS_RING}`}
            // JSON.stringify quotes the URL so spaces and parentheses can't break the CSS
            style={{ backgroundImage: `url(${JSON.stringify(photo.src)})` }}
          >
            <RadixRadioGroup.Indicator className="absolute right-[3px] bottom-[3px] w-5 h-5 rounded-full bg-accent text-on-accent flex items-center justify-center">
              <Check size={12} strokeWidth={3.4} aria-hidden="true" />
            </RadixRadioGroup.Indicator>
          </RadixRadioGroup.Item>
          {/* The name is already the button's accessible name, so it is only drawn here */}
          <span aria-hidden="true" title={photo.label} className="text-sm leading-tight text-center truncate">
            {photo.label}
          </span>
        </div>
      ))}
    </RadixRadioGroup.Root>
  );
}
