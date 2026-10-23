/**
 * Dialog.tsx
 *
 * User: Caregiver
 *
 * Shared pop-up shell for the add/edit/schedule forms: a dimmed backdrop,
 * a titled panel with a close button, a body, and an optional footer for
 * action buttons. Built on Radix Dialog, which traps focus, closes on
 * Escape, and returns focus to the opener.
 */

import * as RadixDialog from "@radix-ui/react-dialog";
import { X } from "lucide-react";
import type { ReactNode } from "react";
import IconButton from "./IconButton";

interface DialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  children: ReactNode;
  /** Action buttons, shown below the body */
  footer?: ReactNode;
}

export default function Dialog({ open, onOpenChange, title, children, footer }: DialogProps) {
  return (
    <RadixDialog.Root open={open} onOpenChange={onOpenChange}>
      <RadixDialog.Portal>
        <RadixDialog.Overlay className="fixed inset-0 z-40 bg-nav/60" />
        <RadixDialog.Content
          aria-describedby={undefined}
          className="fixed z-50 left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 w-[420px] max-w-[calc(100vw-32px)] max-h-[calc(100vh-32px)] overflow-y-auto flex flex-col gap-3 p-4 bg-surface border-[1.5px] border-border rounded-[14px] shadow-dialog text-ink"
        >
          <div className="flex items-start justify-between gap-2">
            <RadixDialog.Title className="m-0 min-w-0 text-lg font-extrabold break-words">{title}</RadixDialog.Title>
            <RadixDialog.Close asChild>
              <IconButton label="Close" icon={<X size={18} strokeWidth={2.4} aria-hidden="true" />} />
            </RadixDialog.Close>
          </div>
          {children}
          {footer && <div className="flex justify-end gap-2">{footer}</div>}
        </RadixDialog.Content>
      </RadixDialog.Portal>
    </RadixDialog.Root>
  );
}
