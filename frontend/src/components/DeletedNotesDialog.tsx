import { useTranslation } from 'react-i18next';
import {
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Box,
  Typography,
  Chip,
} from '@mui/material';
import AppDialog from './AppDialog';
import { Note } from '../api/notes';
import { useDateFormat } from '../DateFormatProvider';
import NoteEditedInfo from './NoteEditedInfo';

interface DeletedNotesDialogProps {
  open: boolean;
  notes: Note[];
  onClose: () => void;
}

// Discreet read-only list of a contact's deleted notes, so information can be
// audited without cluttering the main timeline.
export default function DeletedNotesDialog({
  open,
  notes,
  onClose,
}: DeletedNotesDialogProps) {
  const { t } = useTranslation();
  const { formatDate } = useDateFormat();

  return (
    <AppDialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>{t('contactDetail.deletedNotes', 'Notas eliminadas')}</DialogTitle>
      <DialogContent dividers>
        {notes.length === 0 ? (
          <Typography variant="body2" color="text.secondary">
            {t('contactDetail.noDeletedNotes', 'No hay notas eliminadas')}
          </Typography>
        ) : (
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.5 }}>
            {notes.map((note) => (
              <Box
                key={note.ID}
                sx={{
                  p: 1.5,
                  border: '1px solid',
                  borderColor: 'divider',
                  borderRadius: 1,
                  opacity: 0.85,
                }}
              >
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 0.5 }}>
                  <Typography variant="subtitle2" sx={{ fontWeight: 600, flex: 1, mr: 1 }}>
                    {note.title || t('contactDetail.note')}
                  </Typography>
                  <Chip
                    size="small"
                    variant="outlined"
                    color="error"
                    sx={{ height: 20, fontSize: '0.7rem' }}
                    label={`${t('contactDetail.deletedAt', 'Eliminada')} ${note.deleted_at ? formatDate(note.deleted_at) : ''}`}
                  />
                </Box>
                <Typography variant="body2" color="text.secondary" sx={{ whiteSpace: 'pre-wrap' }}>
                  {note.content}
                </Typography>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mt: 0.5 }}>
                  <Typography variant="caption" color="text.disabled">
                    {formatDate(note.date)} {note.author_name ? `· ${note.author_name}` : ''}
                  </Typography>
                </Box>
                <NoteEditedInfo note={note} />
              </Box>
            ))}
          </Box>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>{t('common.close', 'Cerrar')}</Button>
      </DialogActions>
    </AppDialog>
  );
}