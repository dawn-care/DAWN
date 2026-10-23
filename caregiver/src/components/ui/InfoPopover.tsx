/**
 * InfoPopover.tsx
 *
 * User: Caregiver
 *
 * An info icon button that opens a small dismissible bubble explaining
 * a page or section in one or two plain sentences. Built on Radix Popover:
 * Escape, outside click, and the Close button all dismiss it.
 */

import * as Popover from "@radix-ui/react-popover";
import { Info } from "lucide-react";
import IconButton from "./IconButton";
import { FOCUS_RING } from "./styles";

interface InfoPopoverProps {
  /** Accessible name for the icon button, e.g. "About this page" */
  label: string;
  /** The explanation, one or two plain sentences */
  children: string;
}

export default function InfoPopover({ label, children }: InfoPopoverProps) {
  return (
    <Popover.Root>
      <Popover.Trigger asChild>
        <IconButton
          label={label}
          icon={<Info size={22} aria-hidden="true" />}
          className="text-ink-muted data-[state=open]:text-link"
        />
      </Popover.Trigger>
      <Popover.Portal>
        <Popover.Content
          align="start"
          sideOffset={8}
          className="z-50 w-[300px] max-w-[calc(100vw-32px)] flex flex-col gap-2 p-3.5 bg-surface border-[1.5px] border-border rounded-xl shadow-popover text-base leading-snug text-ink"
        >
          <p className="m-0">{children}</p>
          <Popover.Close
            className={`self-end h-10 px-2.5 bg-transparent text-[15px] font-bold text-link cursor-pointer ${FOCUS_RING}`}
          >
            Close
          </Popover.Close>
        </Popover.Content>
      </Popover.Portal>
    </Popover.Root>
  );
}
