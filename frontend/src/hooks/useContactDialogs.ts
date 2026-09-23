import { useState } from 'react';
import { createNote } from '../api/notes';
import { handleError, ErrorNotifier, getErrorMessage } from '../utils/errorHandler';

export function useContactDialogs(
  contactId: string | undefined,
  onRefresh: () => Promise<void>,
  notifier?: ErrorNotifier
) {
  const [noteDialogOpen, setNoteDialogOpen] = useState(false);

  const handleSaveNote = async (title: string, content: string, date: string, _companyId?: number) => {
    if (!contactId) return;

    try {
      await createNote(contactId, {
        title,
        content,
        date: new Date(date).toISOString(),
        contact_id: parseInt(contactId)
      });
      await onRefresh();
    } catch (err) {
      handleError(err, { operation: 'saving note' }, notifier);
      throw new Error(getErrorMessage(err));
    }
  };

  return {
    noteDialogOpen,
    setNoteDialogOpen,
    handleSaveNote
  };
}
