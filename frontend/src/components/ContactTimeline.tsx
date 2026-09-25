import { ChangeEvent } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Typography,
  Paper,
  IconButton,
  Box,
  TextField,
  InputAdornment,
  Pagination,
  Stack,
  Button,
} from '@mui/material';
import SearchIcon from '@mui/icons-material/Search';
import {
  Timeline,
  TimelineItem,
  TimelineSeparator,
  TimelineConnector,
  TimelineContent,
  TimelineDot
} from '@mui/lab';
import NoteIcon from '@mui/icons-material/Note';
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import { Note } from '../api/notes';
import { useDateFormat } from '../DateFormatProvider';
import NoteEditedInfo from './NoteEditedInfo';

interface ContactTimelineProps {
  notes: Note[];
  onEditNote: (note: Note) => void;
  search: string;
  onSearchChange: (value: string) => void;
  fromDate: string;
  toDate: string;
  onFromDateChange: (value: string) => void;
  onToDateChange: (value: string) => void;
  hasFilters: boolean;
  page: number;
  totalPages: number;
  onPageChange: (_: ChangeEvent<unknown>, value: number) => void;
  onAddNote?: () => void;
  onViewDeleted?: () => void;
}

export default function ContactTimeline({
  notes,
  onEditNote,
  search,
  onSearchChange,
  fromDate,
  toDate,
  onFromDateChange,
  onToDateChange,
  hasFilters,
  page,
  totalPages,
  onPageChange,
  onAddNote,
  onViewDeleted,
}: ContactTimelineProps) {
  const { t } = useTranslation();
  const { formatDate, getDatePlaceholder } = useDateFormat();

  return (
    <>
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
        {onAddNote && (
          <IconButton
            color="primary"
            size="small"
            onClick={onAddNote}
            title={t('contactDetail.addNote')}
            className="contact-add-note-btn"
            sx={{
              width: 27,
              height: 27,
              flexShrink: 0,
              bgcolor: 'primary.main',
              color: 'white',
              '&:hover': { bgcolor: 'primary.dark' },
              '& .MuiSvgIcon-root': { fontSize: '1.2rem' },
            }}
          >
            <AddIcon />
          </IconButton>
        )}
        <Paper sx={{ p: 1, flex: 1, minWidth: 0 }}>
          <Stack direction="row" spacing={1} sx={{ flexWrap: 'wrap', alignItems: 'center' }}>
            <TextField
              size="small"
              placeholder={t('notes.search')}
              value={search}
              onChange={e => onSearchChange(e.target.value)}
              sx={{ flex: 1, minWidth: 160 }}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <SearchIcon fontSize="small" />
                  </InputAdornment>
                ),
              }}
            />
            <TextField
              size="small"
              label={t('notes.fromDate')}
              type="date"
              value={fromDate}
              onChange={e => onFromDateChange(e.target.value)}
              sx={{ width: 140 }}
              slotProps={{ inputLabel: { shrink: true }, input: { placeholder: getDatePlaceholder() } }}
            />
            <TextField
              size="small"
              label={t('notes.toDate')}
              type="date"
              value={toDate}
              onChange={e => onToDateChange(e.target.value)}
              sx={{ width: 140 }}
              slotProps={{ inputLabel: { shrink: true }, input: { placeholder: getDatePlaceholder() } }}
            />
          </Stack>
        </Paper>
      </Box>

      {notes.length === 0 ? (
        <Typography variant="body2" color="text.secondary">
          {hasFilters ? t('notes.noResults') : t('contactDetail.noActivity')}
        </Typography>
      ) : (
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
                <TimelineContent sx={{ minWidth: 0 }}>
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
                      <Typography variant="subtitle2" sx={{ fontWeight: 600, flex: 1, mr: 1, overflowWrap: 'anywhere' }}>
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
                    <Typography
                      variant="body2"
                      sx={{
                        mt: 1,
                        whiteSpace: 'pre-wrap',
                        overflowWrap: 'anywhere',
                        wordBreak: 'break-word',
                        maxHeight: 260,
                        overflowY: 'auto',
                      }}
                    >
                      {note.content}
                    </Typography>
                    <NoteEditedInfo note={note} showEditorName />
                  </Paper>
                </TimelineContent>
              </TimelineItem>
            );
          })}
        </Timeline>
      )}

      {notes.length > 0 && totalPages > 1 && (
        <Box display="flex" justifyContent="center" mt={2}>
          <Pagination
            color="primary"
            count={totalPages}
            page={page}
            onChange={onPageChange}
            showFirstButton
            showLastButton
          />
        </Box>
      )}

      <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'flex-end', mt: 1 }}>
        {onViewDeleted && (
          <Button
            size="small"
            onClick={onViewDeleted}
            sx={{
              minWidth: 0,
              p: '4px 6px',
              color: 'text.disabled',
              textTransform: 'none',
              fontSize: '0.75rem',
              whiteSpace: 'nowrap',
            }}
          >
            {t('contactDetail.viewDeletedNotes')}
          </Button>
        )}
      </Box>
    </>
  );
}