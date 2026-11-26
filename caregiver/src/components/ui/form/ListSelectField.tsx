/**
 * ListSelectField.tsx
 *
 * User: Caregiver
 *
 * A labeled dropdown for a long list of choices, like the schedule form's
 * start times. It looks like SelectField, but the list is drawn by the app
 * (Radix Select) instead of the browser: it opens under the field, is never
 * taller than the room left on screen, and opens scrolled to the current
 * choice in its middle, the way Google Calendar's time list does. A native <select> with
 * dozens of options can run off the bottom of the screen (#481).
 *
 * The list scrolls with the mouse wheel or its scrollbar. Radix's scroll
 * buttons (arrows that scroll while hovered) are left out on purpose, and the
 * scrollbar Radix hides is shown again.
 *
 * Radix gives it the keyboard behavior of a real select: arrow keys, Home and
 * End, typing to jump to a choice, Escape to close, and the listbox roles for
 * screen readers.
 */

import * as RadixSelect from "@radix-ui/react-select";
import { Check, ChevronDown } from "lucide-react";
import { forwardRef, useId, useRef } from "react";
import { FIELD_CLASSES } from "../styles";
import type { SelectOption } from "./SelectField";

interface ListSelectFieldProps {
  label: string;
  value: string;
  onValueChange: (value: string) => void;
  options: SelectOption[];
}

/** The ref is the field's button, so a form can focus it */
const ListSelectField = forwardRef<HTMLButtonElement, ListSelectFieldProps>(function ListSelectField(
  { label, value, onValueChange, options },
  ref
) {
  const id = useId();
  const viewport = useRef<HTMLDivElement>(null);

  // Radix scrolls the chosen choice just into view, at the list's edge; put it in the middle instead.
  // Two frames, so this runs after the list is drawn and Radix has done its own scroll
  const centerChosen = (open: boolean) => {
    if (!open) return;
    requestAnimationFrame(() =>
      requestAnimationFrame(() => {
        const list = viewport.current;
        const chosen = list?.querySelector<HTMLElement>('[data-state="checked"]');
        if (list && chosen) list.scrollTop = chosen.offsetTop - (list.clientHeight - chosen.offsetHeight) / 2;
      })
    );
  };

  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={id} className="text-base font-bold text-ink">
        {label}
      </label>
      <RadixSelect.Root value={value} onValueChange={onValueChange} onOpenChange={centerChosen}>
        <RadixSelect.Trigger
          ref={ref}
          id={id}
          className={`${FIELD_CLASSES} flex items-center justify-between gap-2 font-normal text-left cursor-pointer`}
        >
          <RadixSelect.Value />
          <RadixSelect.Icon className="text-ink-muted">
            <ChevronDown size={18} aria-hidden="true" />
          </RadixSelect.Icon>
        </RadixSelect.Trigger>

        <RadixSelect.Portal>
          <RadixSelect.Content
            position="popper"
            side="bottom"
            sideOffset={4}
            collisionPadding={8}
            // React passes key presses up through the portal; Escape here only closes the list, not a form around it
            onKeyDown={(event) => {
              if (event.key === "Escape") event.stopPropagation();
            }}
            className="z-50 flex flex-col w-[var(--radix-select-trigger-width)] max-h-[min(20rem,var(--radix-select-content-available-height))] overflow-hidden rounded-[10px] border-2 border-ink-muted bg-surface text-ink shadow-popover"
          >
            {/* Radix hides the scrollbar to make room for its scroll buttons; this shows it again */}
            <RadixSelect.Viewport ref={viewport} className="p-1 ![scrollbar-width:thin]">
              {options.map((option) => (
                <RadixSelect.Item
                  key={option.value}
                  value={option.value}
                  className="relative flex items-center h-10 pl-8 pr-3 rounded-md text-base cursor-pointer select-none outline-none data-[highlighted]:bg-accent/15 data-[state=checked]:font-bold"
                >
                  <RadixSelect.ItemIndicator className="absolute left-2 inline-flex">
                    <Check size={16} strokeWidth={2.6} aria-hidden="true" />
                  </RadixSelect.ItemIndicator>
                  <RadixSelect.ItemText>{option.label}</RadixSelect.ItemText>
                </RadixSelect.Item>
              ))}
            </RadixSelect.Viewport>
          </RadixSelect.Content>
        </RadixSelect.Portal>
      </RadixSelect.Root>
    </div>
  );
});

export default ListSelectField;
