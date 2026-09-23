import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Box, Typography, Paper, Link } from '@mui/material';
import { Note } from '../api/notes';
import { useDateFormat } from '../DateFormatProvider';

interface NoteEditedInfoProps {
  note: Note;
}

// Discreet "edited" indicator: shows the edit date and lets the user view the
// note's original content. Renders nothing for untouched notes.
export default function NoteEditedInfo({ note }: NoteEditedInfoProps) {
  const { t } = useTranslation();
  const { formatDate } = useDateFormat();
  const [showOriginal, setShowOriginal] = useState(false);

  const original =
    note.original_content !== undefined &&
    note.original_content !== null &&
    note.original_content !== '' &&
    note.original_content !== note.content;

  if (!original || !note.UpdatedAt) return null;

  return (
    <Box sx={{ mt: 1 }}>
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
        <Typography variant="caption" color="text.secondary">
          {t('noteEdited.edited', 'Editada')} {formatDate(note.UpdatedAt)}
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