import { useState, useMemo, ChangeEvent, MouseEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  Box,
  Typography,
  Paper,
  Button,
  Pagination,
  Link,
  Avatar,
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
import { getInitials, avatarColorFor } from './utils/avatar';
import { useDateFormat } from './DateFormatProvider';
import NotesFilterBar from './components/NotesFilterBar';

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
  const { formatDate } = useDateFormat();
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

      <NotesFilterBar
        search={searchInput}
        onSearchChange={handleSearchChange}
        fromDate={fromDate}
        toDate={toDate}
        onFromDateChange={handleFromDateChange}
        onToDateChange={handleToDateChange}
      />

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
            const isContactClickable = contactName != null && note.contact != null;
            const openContact = (e: MouseEvent<HTMLElement>) => {
              e.preventDefault();
              navigate(`/contacts/${note.contact!.ID}`);
            };
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
                    <Box display="flex" alignItems="center" justifyContent="space-between">
                      <Box display="flex" alignItems="center" sx={{ flex: 1, minWidth: 0 }}>
                        {contactName && isContactClickable && (
                          <Avatar
                            component="button"
                            onClick={openContact}
                            title={contactName}
                            sx={{
                              width: 44,
                              height: 44,
                              flexShrink: 0,
                              bgcolor: avatarColorFor(contactName),
                              fontSize: '1rem',
                              fontWeight: 700,
                              color: '#fff',
                              cursor: 'pointer',
                              border: 'none',
                            }}
                          >
                            {getInitials(contactName)}
                          </Avatar>
                        )}
                        {contactName && isContactClickable ? (
                          <Link
                            component="button"
                            variant="h6"
                            underline="hover"
                            onClick={openContact}
                            sx={{
                              display: 'block',
                              ml: 1.5,
                              fontWeight: 700,
                              fontSize: '1.06rem',
                              lineHeight: 1.3,
                              textAlign: 'left',
                              textTransform: 'none',
                              cursor: 'pointer',
                              overflowWrap: 'anywhere',
                            }}
                          >
                            {contactName}
                          </Link>
                        ) : (
                          <Typography variant="h6" sx={{ fontWeight: 700, fontSize: '1.06rem', lineHeight: 1.3 }}>
                            {note.title || t('contactDetail.note')}
                          </Typography>
                        )}
                      </Box>
                      {formatDate(note.date) && (
                        <Typography
                          variant="caption"
                          color="text.secondary"
                          sx={{ whiteSpace: 'nowrap', ml: 1, flexShrink: 0 }}
                        >
                          {formatDate(note.date)}
                        </Typography>
                      )}
                    </Box>
                    {isContactClickable && note.title && (
                      <Typography
                        variant="subtitle2"
                        sx={{
                          color: 'text.primary',
                          fontWeight: 600,
                          mt: 1,
                          overflowWrap: 'anywhere',
                        }}
                      >
                        {note.title}
                      </Typography>
                    )}
                    {note.content && (
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
                    )}
                    <Box sx={{ mt: 0.75 }}>
                      <NoteEditedInfo note={note} />
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
