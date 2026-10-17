/**
 * RemoveMediaDialog.tsx
 *
 * User: Caregiver
 *
 * Asks the caregiver to confirm removing a photo, video or audio file, and
 * says what else goes with it. The warning is the same whether or not the
 * file is in use: the server removes whatever depends on it (widgets,
 * scheduled items). The dialog only asks: the page does the removing, and
 * shows an Undo toast afterwards.
 */

import Dialog from "../components/ui/Dialog";
import type { Media } from "./api";
import { PRIMARY_BUTTON, SECONDARY_BUTTON } from "./styles";

interface RemoveMediaDialogProps {
  media: Media;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  /** The caregiver confirmed; the dialog has asked to close by then */
  onConfirm: () => void;
}

export default function RemoveMediaDialog({ media, open, onOpenChange, onConfirm }: RemoveMediaDialogProps) {
  const confirm = () => {
    onOpenChange(false);
    onConfirm();
  };

  return (
    <Dialog
      open={open}
      onOpenChange={onOpenChange}
      title={`Remove "${media.display_name}"?`}
      footer={
        <>
          <button type="button" onClick={() => onOpenChange(false)} className={SECONDARY_BUTTON}>
            Cancel
          </button>
          <button type="button" onClick={confirm} className={PRIMARY_BUTTON}>
            Remove
          </button>
        </>
      }
    >
      <p className="m-0 text-base">This will be removed from the carepartner screen and all scheduled media.</p>
    </Dialog>
  );
}
