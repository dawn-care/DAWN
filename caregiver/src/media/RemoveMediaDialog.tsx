/**
 * RemoveMediaDialog.tsx
 *
 * User: Caregiver
 *
 * Asks the caregiver to confirm removing a photo, video or audio file. It
 * only asks: the page does the removing, and shows an Undo toast afterwards,
 * so the dialog needs no warning text.
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
      {null}
    </Dialog>
  );
}
