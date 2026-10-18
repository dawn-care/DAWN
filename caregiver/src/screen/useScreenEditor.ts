/**
 * useScreenEditor.ts
 *
 * User: Caregiver
 *
 * Edit mode for the home page. Edit screen copies the saved widgets into a
 * draft that every change works on; Cancel throws the draft away, so nothing
 * reaches the tablet until the draft is saved. A new saved layout (another
 * carepartner, a reload, or a save) ends editing.
 *
 * FUTURE IMPLEMENTATION (#333): placing, changing and removing widgets, and
 * saving the draft, are added in later steps.
 */

import { useCallback, useState } from "react";
import type { EditorWidget, LoadedLayout } from "./layout";

export function useScreenEditor(saved: LoadedLayout | null) {
  // null while not editing
  const [draft, setDraft] = useState<EditorWidget[] | null>(null);

  // Drop the draft when the layout it was copied from is replaced. Adjusting
  // state during render like this avoids drawing one frame with the old draft.
  const [draftOf, setDraftOf] = useState(saved);
  if (saved !== draftOf) {
    setDraftOf(saved);
    setDraft(null);
  }

  const start = useCallback(() => {
    if (saved) setDraft(saved.widgets.map((widget) => ({ ...widget })));
  }, [saved]);

  const cancel = useCallback(() => setDraft(null), []);

  return { editing: draft !== null, draft, start, cancel };
}
