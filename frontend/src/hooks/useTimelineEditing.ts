import { useState } from 'react';
import { updateNote, deleteNote, Note } from '../api/notes';
import { handleError, ErrorNotifier } from '../utils/errorHandler';

export function useTimelineEditing(
  onRefresh: () => Promise<void>,
  notifier?: ErrorNotifier
) {
  const [editingNote, setEditingNote] = useState<Note | null>(null);
  const [editNoteValues, setEditNoteValues] = useState<{
    noteTitle?: string;
    noteContent?: string;
    noteDate?: string;
  }>({});

  const handleStartEditNote = (note: Note) => {
    setEditingNote(note);
    setEditNoteValues({
      noteTitle: note.title || '',
      noteContent: note.content || '',
      noteDate: note.date ? new Date(note.date).toISOString().split('T')[0] : ''
    });
  };

  const handleCancelEditNote = () => {
    setEditingNote(null);
    setEditNoteValues({});
  };

  const handleUpdateNote = async (noteId: number) => {
    if (!editNoteValues.noteContent?.trim()) return;

    try {
      await updateNote(noteId, {
        title: editNoteValues.noteTitle || '',
        content: editNoteValues.noteContent,
        date: editNoteValues.noteDate ? new Date(editNoteValues.noteDate).toISOString() : new Date().toISOString(),
      });
      await onRefresh();
      handleCancelEditNote();
    } catch (err) {
      handleError(err, { operation: 'updating note' }, notifier);
    }
  };

  const handleDeleteNote = async (noteId: number) => {
    try {
      await deleteNote(noteId);
      await onRefresh();
      handleCancelEditNote();
    } catch (err) {
      handleError(err, { operation: 'deleting note' }, notifier);
    }
  };

  return {
    editingNote,
    editNoteValues,
    setEditNoteValues,
    handleStartEditNote,
    handleCancelEditNote,
    handleUpdateNote,
    handleDeleteNote
  };
}
