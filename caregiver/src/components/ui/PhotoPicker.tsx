/**
 * PhotoPicker.tsx
 *
 * User: Caregiver
 *
 * A row of selectable photo thumbnails ("Which photo"). Used by the photo
 * widget pop-up and the audio background form. Built on Radix RadioGroup,
 * so exactly one photo is chosen and arrow keys move between them.
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
}

export default function PhotoPicker({ photos, value, onChange, label }: PhotoPickerProps) {
  return (
    <RadixRadioGroup.Root
      aria-label={label}
      value={value}
      onValueChange={onChange}
      className="flex gap-2"
    >
      {photos.map((photo) => (
        <RadixRadioGroup.Item
          key={photo.id}
          value={photo.id}
          aria-label={photo.label}
          className={`relative flex-1 h-[54px] p-0 overflow-hidden rounded-lg border-2 border-dawn-line-strong bg-dawn-avatar bg-cover bg-center cursor-pointer data-[state=checked]:border-[3px] data-[state=checked]:border-dawn-accent ${FOCUS_RING}`}
          style={{ backgroundImage: `url(${photo.src})` }}
        >
          <RadixRadioGroup.Indicator className="absolute right-[3px] bottom-[3px] w-5 h-5 rounded-full bg-dawn-accent text-white flex items-center justify-center">
            <Check size={12} strokeWidth={3.4} aria-hidden="true" />
          </RadixRadioGroup.Indicator>
        </RadixRadioGroup.Item>
      ))}
    </RadixRadioGroup.Root>
  );
}
