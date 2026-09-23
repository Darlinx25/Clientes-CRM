import { useTranslation } from 'react-i18next';
import {
  Typography,
  Paper,
  IconButton,
  Box,
} from '@mui/material';
import {
  Timeline,
  TimelineItem,
  TimelineSeparator,
  TimelineConnector,
  TimelineContent,
  TimelineDot
} from '@mui/lab';
import NoteIcon from '@mui/icons-material/Note';
import EditIcon from '@mui/icons-material/Edit';
import { Note } from '../api/notes';
import { useDateFormat } from '../DateFormatProvider';

interface ContactTimelineProps {
  notes: Note[];
  onEditNote: (note: Note) => void;
}

export default function ContactTimeline({ notes, onEditNote }: ContactTimelineProps) {
  const { t } = useTranslation();
  const { formatDate } = useDateFormat();

  if (notes.length === 0) {
    return (
      <Typography variant="body2" color="text.secondary">
        {t('contactDetail.noActivity')}
      </Typography>
    );
  }

  return (
    <Timeline
      position="right"
      sx={{
        px: 0,
        '& .MuiTimelineItem-root::before': {
          content: 'none',
        },
      }}
    >
      {notes.map((note, index) => {
        const itemDate = new Date(note.date);
        const isValidDate = !isNaN(itemDate.getTime());

        return (
          <TimelineItem key={note.ID}>
            <TimelineSeparator>
              <TimelineDot color="primary">
                <NoteIcon fontSize="small" />
              </TimelineDot>
              {index < notes.length - 1 && <TimelineConnector />}
            </TimelineSeparator>
            <TimelineContent>
              <Paper
                elevation={1}
                sx={{
                  p: 1.5,
                  '&:hover .action-icon': {
                    opacity: 1
                  }
                }}
              >
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <Typography variant="subtitle2" sx={{ fontWeight: 600, flex: 1, mr: 1 }}>
                    {note.title || t('contactDetail.note')}
                  </Typography>
                  <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end' }}>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                      <Typography variant="caption" color="text.secondary" sx={{ whiteSpace: 'nowrap' }}>
                        {isValidDate ? formatDate(note.date) : (note.date || 'N/A')}
                      </Typography>
                      <IconButton
                        className="action-icon"
                        size="small"
                        onClick={() => onEditNote(note)}
                        sx={{ p: 0.5, opacity: 0, transition: 'opacity 0.2s' }}
                      >
                        <EditIcon fontSize="small" />
                      </IconButton>
                    </Box>
                    {note.author_name && (
                      <Typography variant="caption" color="text.disabled" sx={{ whiteSpace: 'nowrap' }}>
                        {note.author_name}
                      </Typography>
                    )}
                  </Box>
                </Box>
                <Typography variant="body2" sx={{ mt: 1, whiteSpace: 'pre-wrap' }}>
                  {note.content}
                </Typography>
              </Paper>
            </TimelineContent>
          </TimelineItem>
        );
      })}
    </Timeline>
  );
}
