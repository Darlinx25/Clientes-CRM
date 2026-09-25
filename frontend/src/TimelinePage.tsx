import { useState, useMemo, ChangeEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  Box,
  Typography,
  Paper,
  TextField,
  Button,
  Pagination,
  Link,
} from '@mui/material';
import {
  Timeline,
  TimelineItem,
  TimelineSeparator,
  TimelineConnector,
  TimelineContent,
  TimelineDot,
} from '@mui/lab';
import TimelineIcon from '@mui/icons-material/Timeline';
import AddIcon from '@mui/icons-material/Add';
import { ListSkeleton } from './components/LoadingSkeletons';
import { useNotes } from './hooks/useNotes';
import { useDebouncedValue } from './hooks/useDebounce';
import { createUnassignedNote, Note } from './api/notes';
import AddNoteDialog from './components/AddNoteDialog';
import NoteEditedInfo from './components/NoteEditedInfo';
import { handleError } from './utils/errorHandler';
import { useDateFormat } from './DateFormatProvider';

const NOTES_PER_PAGE = 15;

// ISO date (YYYY-MM-DD) a whole number of months back from today, local time.
function isoMonthsAgo(months: number): string {
  const d = new Date();
  d.setMonth(d.getMonth() - months);
  const mm = `${d.getMonth() + 1}`.padStart(2, '0');
  const dd = `${d.getDate()}`.padStart(2, '0');
  return `${d.getFullYear()}-${mm}-${dd}`;
}

const TimelinePage: React.FC = () => {
  const { t } = useTranslation();
  const { formatDate, getDatePlaceholder } = useDateFormat();
  const navigate = useNavigate();
  const [searchInput, setSearchInput] = useState('');
  const debouncedSearch = useDebouncedValue(searchInput, 400);
  const [page, setPage] = useState(1);
  const [fromDate, setFromDate] = useState(() => isoMonthsAgo(2));
  const [toDate, setToDate] = useState('');

  const notesParams = useMemo(
    () => ({
      page,
      limit: NOTES_PER_PAGE,
      search: debouncedSearch.trim() || undefined,
      fromDate: fromDate || undefined,
      toDate: toDate || undefined,
    }),
    [page, debouncedSearch, fromDate, toDate]
  );

  const {
    notes,
    total,
    page: serverPage,
    limit,
    loading,
    refetch,
  } = useNotes(undefined, notesParams);
  const [addDialogOpen, setAddDialogOpen] = useState(false);

  const handleSearchChange = (value: string) => {
    setSearchInput(value);
    setPage(1);
  };

  const handleFromDateChange = (value: string) => {
    setFromDate(value);
    setPage(1);
  };

  const handleToDateChange = (value: string) => {
    setToDate(value);
    setPage(1);
  };

  const handleClearDates = () => {
    setFromDate('');
    setToDate('');
    setPage(1);
  };

  const handlePageChange = (_: ChangeEvent<unknown>, value: number) => {
    setPage(value);
  };

  const currentPage = serverPage || page;
  const pageSize = limit || NOTES_PER_PAGE;
  const totalPages = Math.max(1, Math.ceil((total || 0) / pageSize));

  const handleAddNote = () => {
    setAddDialogOpen(true);
  };

  const handleNoteSave = async (title: string, content: string, date: string, contactId?: number) => {
    try {
      await createUnassignedNote({ title, content, date: new Date(date).toISOString(), contact_id: contactId });
      setAddDialogOpen(false);
      refetch();
    } catch (err) {
      handleError(err, { operation: 'creating note' });
      throw err;
    }
  };

  const formatContactName = (note: Note) => {
    const c = note.contact;
    if (!c) return null;
    return `${c.firstname} ${c.lastname || ''}`.trim();
  };

  const isInitialLoading = loading && notes.length === 0;
  const hasFilters = searchInput.trim().length > 0 || fromDate || toDate;

  return (
    <Box sx={{ maxWidth: 1200, mx: 'auto', mt: 2, p: 2 }}>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={2}>
        <Box display="flex" alignItems="center" gap={1}>
          <Typography variant="h5">{t('timelinePage.title')}</Typography>
        </Box>
        <Button variant="outlined" startIcon={<AddIcon />} onClick={handleAddNote}>
          {t('timelinePage.addNote')}
        </Button>
      </Box>

      <Paper sx={{ p: 1.5, mb: 2 }}>
        <Box display="flex" gap={2} flexWrap="wrap">
          <TextField
            size="small"
            label={t('timelinePage.search')}
            value={searchInput}
            onChange={(e) => handleSearchChange(e.target.value)}
            variant="outlined"
            sx={{ flex: 1, minWidth: 200 }}
          />
          <TextField
            size="small"
            label={t('timelinePage.fromDate')}
            type="date"
            value={fromDate}
            onChange={(e) => handleFromDateChange(e.target.value)}
            variant="outlined"
            slotProps={{ inputLabel: { shrink: true }, input: { placeholder: getDatePlaceholder() } }}
            sx={{ width: 160 }}
          />
          <TextField
            size="small"
            label={t('timelinePage.toDate')}
            type="date"
            value={toDate}
            onChange={(e) => handleToDateChange(e.target.value)}
            variant="outlined"
            slotProps={{ inputLabel: { shrink: true }, input: { placeholder: getDatePlaceholder() } }}
            sx={{ width: 160 }}
          />
          {(fromDate || toDate) && (
            <Button
              size="small"
              onClick={handleClearDates}
              sx={{ alignSelf: 'center', color: 'text.secondary', textTransform: 'none', whiteSpace: 'nowrap' }}
            >
              {t('timelinePage.clearDates')}
            </Button>
          )}
        </Box>
      </Paper>

      {isInitialLoading ? (
        <ListSkeleton count={8} />
      ) : notes.length === 0 ? (
        <Paper sx={{ p: 3, textAlign: 'center' }}>
          <Typography variant="body1" color="text.secondary">
            {hasFilters ? t('timelinePage.noResults') : t('timelinePage.noNotes')}
          </Typography>
        </Paper>
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
            const contactName = formatContactName(note);
            return (
              <TimelineItem key={note.ID}>
                <TimelineSeparator>
                  <TimelineDot color="primary">
                    <TimelineIcon fontSize="small" />
                  </TimelineDot>
                  {index < notes.length - 1 && <TimelineConnector />}
                </TimelineSeparator>
                <TimelineContent sx={{ flex: 0.8, minWidth: 0 }}>
                  <Paper elevation={2} sx={{ p: 2 }}>
                    <Box display="flex" justifyContent="space-between" alignItems="flex-start">
                      <Box sx={{ flex: 1, minWidth: 0 }}>
                        {note.title && (
                          <Typography variant="subtitle1" sx={{ fontWeight: 600, mb: 0.5, overflowWrap: 'anywhere' }}>
                            {note.title}
                          </Typography>
                        )}
                        <Typography
                          variant="body2"
                          sx={{
                            whiteSpace: 'pre-wrap',
                            overflowWrap: 'anywhere',
                            wordBreak: 'break-word',
                            maxHeight: 260,
                            overflowY: 'auto',
                          }}
                        >
                          {note.content}
                        </Typography>
                        <NoteEditedInfo note={note} />
                        {contactName && (
                          <Box sx={{ mt: 1 }}>
                            <Link
                              component="button"
                              variant="caption"
                              underline="hover"
                              onClick={(e) => {
                                e.preventDefault();
                                navigate(`/contacts/${note.contact!.ID}`);
                              }}
                              sx={{
                                cursor: 'pointer',
                                display: 'inline-flex',
                                alignItems: 'center',
                                px: 1,
                                py: 0.5,
                                borderRadius: '16px',
                                backgroundColor: 'rgba(76, 175, 80, 0.14)',
                                color: '#2e7d32',
                                fontWeight: 600,
                                textDecoration: 'none',
                                transition: 'background-color 0.2s, color 0.2s',
                                '&:hover': {
                                  backgroundColor: 'rgba(76, 175, 80, 0.26)',
                                  textDecoration: 'underline'
                                },
                                '&:focus-visible': {
                                  outline: '2px solid rgba(76, 175, 80, 0.5)',
                                  outlineOffset: 2,
                                  borderRadius: '16px'
                                }
                              }}
                            >
                              {contactName}
                            </Link>
                          </Box>
                        )}
                      </Box>
                      <Typography variant="caption" color="text.secondary" sx={{ whiteSpace: 'nowrap', ml: 1 }}>
                        {formatDate(note.date)}
                      </Typography>
                    </Box>
                  </Paper>
                </TimelineContent>
              </TimelineItem>
            );
          })}
        </Timeline>
      )}

      {totalPages > 1 && (
        <Box display="flex" justifyContent="center" mt={3}>
          <Pagination
            color="primary"
            count={totalPages}
            page={currentPage}
            onChange={handlePageChange}
            showFirstButton
            showLastButton
          />
        </Box>
      )}

      <AddNoteDialog
        open={addDialogOpen}
        onClose={() => setAddDialogOpen(false)}
        onSave={handleNoteSave}
        contactRequired
      />
    </Box>
  );
};

export default TimelinePage;
