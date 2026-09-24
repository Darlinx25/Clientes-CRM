import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Box, Typography, Paper, Link } from '@mui/material';
import { Note } from '../api/notes';
import { useDateFormat } from '../DateFormatProvider';

interface NoteEditedInfoProps {
  note: Note;
  showEditorName?: boolean;
}

// Discreet "edited" indicator: shows the edit date and lets the user view the
// note's original content. Renders nothing for untouched notes. When
// showEditorName is set and the note was edited by someone other than its
// author, the editor's name is shown too.
export default function NoteEditedInfo({ note, showEditorName }: NoteEditedInfoProps) {
  const { t } = useTranslation();
  const { formatDate } = useDateFormat();
  const [showOriginal, setShowOriginal] = useState(false);

  const original =
    note.original_content !== undefined &&
    note.original_content !== null &&
    note.original_content !== '' &&
    note.original_content !== note.content;

  if (!original || !note.UpdatedAt) return null;

  const editorName =
    showEditorName &&
    note.edited_by_name &&
    note.author_name &&
    note.edited_by_name !== note.author_name
      ? note.edited_by_name
      : undefined;

  const editedDate = formatDate(note.UpdatedAt);

  return (
    <Box sx={{ mt: 1 }}>
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
        <Typography variant="caption" color="text.secondary">
          {editorName
            ? t('noteEdited.editedBy', 'Editada {{date}} por {{name}}', {
                date: editedDate,
                name: editorName,
              })
            : `${t('noteEdited.edited', 'Editada')} ${editedDate}`}
        </Typography>
        <Link
          component="button"
          variant="caption"
          underline="hover"
          onClick={() => setShowOriginal((s) => !s)}
          sx={{ cursor: 'pointer' }}
        >
          {showOriginal
            ? t('noteEdited.hideOriginal', 'Ocultar original')
            : t('noteEdited.viewOriginal', 'Ver original')}
        </Link>
      </Box>
      {showOriginal && (
        <Paper
          variant="outlined"
          sx={{ mt: 0.5, p: 1, bgcolor: 'action.hover' }}
        >
          {note.original_title && note.original_title !== note.title && (
            <Typography variant="subtitle2" sx={{ mb: 0.5 }}>
              {note.original_title}
            </Typography>
          )}
          <Typography variant="caption" color="text.secondary" sx={{ whiteSpace: 'pre-wrap' }}>
            {note.original_content}
          </Typography>
        </Paper>
      )}
    </Box>
  );
}