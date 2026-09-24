import { useEffect, useState, ChangeEvent } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  Contact,
  Company,
  getContact,
  updateContact,
  deleteContact,
  archiveContact,
  unarchiveContact
} from './api/contacts';
import { getCurrentUser } from './api/admin';
import { resolveEnabledFields, ContactFieldKey } from './contactFields';
import {
  getContactNotes,
  getDeletedContactNotes,
  Note
} from './api/notes';
import { useDebouncedValue } from './hooks/useDebounce';
import {
  Box,
  Card,
  CardContent,
  Button,
  Typography,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions
} from '@mui/material';
import { ContactDetailHeaderSkeleton, TimelineSkeleton } from './components/LoadingSkeletons';
import AddNoteDialog from './components/AddNoteDialog';
import EditTimelineItemDialog from './components/EditTimelineItemDialog';
import ContactHeader from './components/ContactHeader';
import ContactInformation from './components/ContactInformation';
import ContactTimeline from './components/ContactTimeline';
import DeletedNotesDialog from './components/DeletedNotesDialog';
import { useContactDialogs } from './hooks/useContactDialogs';
import { useTimelineEditing, } from './hooks/useTimelineEditing';
import { useSnackbar } from './context/SnackbarContext';
import { ApiError } from './api/client';
import { handleFetchError } from './utils/errorHandler';
import { useDateFormat } from './DateFormatProvider';

// Extended Contact type with optional relations for local state
interface ContactWithRelations extends Contact {
  notes?: Note[];
  companies?: Company[];
}

const CONTACT_FIELDS = [
  'ID', 'firstname', 'lastname', 'nickname', 'gender',
  'email', 'phone', 'birthday', 'address', 'how_we_met',
  'food_preference', 'work_information', 'contact_information',
  'photo', 'custom_fields', 'archived',
  'emails', 'phones', 'impps',
  'prefix', 'middle_name', 'suffix', 'organization', 'department',
  'job_title', 'role', 'anniversary', 'rut', 'contact_person', 'companies'
];

export default function ContactDetailPage() {
  const { t } = useTranslation();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { showError } = useSnackbar();
  const { formatBirthdayForInput, parseBirthdayInput, autoFormatBirthdayInput } = useDateFormat();
  const [contact, setContact] = useState<ContactWithRelations | null>(null);
  const [loading, setLoading] = useState(true);
  const [editingField, setEditingField] = useState<string | null>(null);
  const [editValue, setEditValue] = useState<string>('');
  const [validationError, setValidationError] = useState<string>('');
  const [notes, setNotes] = useState<Note[]>([]);
  const NOTES_PER_PAGE = 5;
  const [notesPage, setNotesPage] = useState(1);
  const [notesTotal, setNotesTotal] = useState(0);
  const [notesLimit, setNotesLimit] = useState(NOTES_PER_PAGE);
  const [notesSearchInput, setNotesSearchInput] = useState('');
  const notesSearch = useDebouncedValue(notesSearchInput, 400);
  const [notesFromDate, setNotesFromDate] = useState('');
  const [notesToDate, setNotesToDate] = useState('');
  const [notesVersion, setNotesVersion] = useState(0);
  const [archiveDialogOpen, setArchiveDialogOpen] = useState(false);
  const [deletedNotes, setDeletedNotes] = useState<Note[]>([]);
  const [deletedNotesDialogOpen, setDeletedNotesDialogOpen] = useState(false);
  
  // Profile editing state
  const [editingProfile, setEditingProfile] = useState(false);
  const [profileValues, setProfileValues] = useState({
    firstname: ''
  });

  // Enabled extended contact fields (UI visibility)
  const [enabledFields, setEnabledFields] = useState<Set<ContactFieldKey>>(() => resolveEnabledFields(null));

  // Unified refresh function for notes: bumping the version re-runs the notes
  // effect, which refetches with the current page and filters.
  const refreshNotesAndActivities = async () => {
    setNotesVersion(v => v + 1);
  };

  // Custom hooks
  const {
    noteDialogOpen,
    setNoteDialogOpen,
    handleSaveNote
  } = useContactDialogs(id, refreshNotesAndActivities, { showError });

  const {
    editingNote,
    editNoteValues,
    setEditNoteValues,
    handleStartEditNote,
    handleCancelEditNote,
    handleUpdateNote,
    handleDeleteNote
  } = useTimelineEditing(refreshNotesAndActivities, { showError });

  // Fetch contact details and notes
  useEffect(() => {
    if (!id) return;

    const fetchData = async () => {
      try {
        // First batch: parallel fetch of core data (notes load in their own effect below)
        const [contactData, user] = await Promise.all([
          getContact(id, CONTACT_FIELDS),
          getCurrentUser().catch(err => {
            console.error('Error fetching current user preferences:', err);
            return null;
          })
        ]);

        setContact(contactData);
        setEnabledFields(resolveEnabledFields(user?.enabled_contact_fields ?? null));

        setLoading(false);
      } catch (err) {
        console.error('Error fetching data:', err);
        setLoading(false);
      }
    };

    fetchData();
  }, [id]);

  // Starting a different contact means back to the first page.
  useEffect(() => {
    setNotesPage(1);
  }, [id]);

  // Paginated notes: 5 per page, filtered by search text and date range.
  // The server applies the filters and returns the total for pagination.
  useEffect(() => {
    if (!id) return;
    let cancelled = false;

    (async () => {
      try {
        const data = await getContactNotes(id, {
          page: notesPage,
          limit: NOTES_PER_PAGE,
          search: notesSearch.trim() || undefined,
          fromDate: notesFromDate || undefined,
          toDate: notesToDate || undefined,
        });
        if (cancelled) return;
        const list = data.notes || [];
        setNotes(list);
        setNotesTotal(data.total ?? list.length);
        setNotesLimit(data.limit ?? NOTES_PER_PAGE);
        // If the current page came back empty (a filter narrowed the results or
        // the last item of this page was deleted), step back one page.
        if (notesPage > 1 && list.length === 0) {
          setNotesPage(p => Math.max(1, p - 1));
        }
      } catch (err) {
        if (!cancelled) {
          handleFetchError(err, 'refreshing notes');
        }
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [id, notesPage, notesSearch, notesFromDate, notesToDate, notesVersion]);

  const handleNotesSearchChange = (value: string) => {
    setNotesSearchInput(value);
    setNotesPage(1);
  };

  const handleNotesFromDateChange = (value: string) => {
    setNotesFromDate(value);
    setNotesPage(1);
  };

  const handleNotesToDateChange = (value: string) => {
    setNotesToDate(value);
    setNotesPage(1);
  };

  const handleNotesPageChange = (_: ChangeEvent<unknown>, value: number) => {
    setNotesPage(value);
  };

  const notesHasFilters = notesSearchInput.trim().length > 0 || !!notesFromDate || !!notesToDate;
  const notesTotalPages = Math.max(1, Math.ceil((notesTotal || 0) / (notesLimit || NOTES_PER_PAGE)));

  const validateBirthday = (value: string): boolean => {
    if (!value || value.trim() === '') return true;
    // Try to parse the birthday input - if it returns null, it's invalid
    const parsed = parseBirthdayInput(value);
    return parsed !== null;
  };

  const handleEditStart = (field: string, currentValue: string) => {
    setEditingField(field);
    // For date fields, convert from ISO to display format
    if ((field === 'birthday' || field === 'anniversary') && currentValue) {
      setEditValue(formatBirthdayForInput(currentValue));
    } else {
      setEditValue(currentValue || '');
    }
    setValidationError('');
  };

  const handleEditCancel = () => {
    setEditingField(null);
    setEditValue('');
    setValidationError('');
  };

  const handleEditSave = async (field: string) => {
    if (!contact) return;

    let valueToSave = editValue;

    if (field === 'birthday' || field === 'anniversary') {
      if (!validateBirthday(editValue)) {
        setValidationError(t('contactDetail.birthdayError'));
        return;
      }
      // Convert from display format to ISO format for storage
      const parsed = parseBirthdayInput(editValue);
      valueToSave = parsed || '';
    }

    // Handle custom fields
    if (field.startsWith('custom_field_')) {
      const customFieldName = field.replace('custom_field_', '');
      const updatedCustomFields = {
        ...(contact.custom_fields || {}),
        [customFieldName]: valueToSave
      };

      try {
        const updatedContact = await updateContact(id!, {
          ...toContactUpdatePayload(contact),
          custom_fields: updatedCustomFields
        });
        setContact(updatedContact);
        setEditingField(null);
        setEditValue('');
        setValidationError('');
      } catch (err) {
        console.error('Error updating contact custom field:', err);
        if (err instanceof ApiError) {
          const errorMessage = err.getDisplayMessage();
          setValidationError(errorMessage);
          showError(errorMessage);
        } else {
          showError(t('contactDetail.updateError'));
        }
      }
      return;
    }

    try {
      const updatedContact = await updateContact(id!, {
        ...toContactUpdatePayload(contact),
        [field]: valueToSave
      });
      setContact(updatedContact);
      setEditingField(null);
      setEditValue('');
      setValidationError('');
    } catch (err) {
      console.error('Error updating contact:', err);
      if (err instanceof ApiError) {
        const errorMessage = err.getDisplayMessage();
        setValidationError(errorMessage);
        showError(errorMessage);
      } else {
        showError(t('contactDetail.updateError'));
      }
    }
  };

  // Companies are managed by CompanySection through their own endpoints, so a
  // generic field edit must never rewrite the company list.
  const toContactUpdatePayload = (c: Contact): Partial<Contact> => {
    const { companies: _companies, ...rest } = c;
    return rest;
  };

  // Persist multi-valued / structured field updates (emails, phones, impps)
  const handleUpdateContactFields = async (partial: Partial<Contact>) => {
    if (!contact) return;
    try {
      const updatedContact = await updateContact(id!, { ...toContactUpdatePayload(contact), ...partial });
      setContact(updatedContact);
    } catch (err) {
      console.error('Error updating contact:', err);
      if (err instanceof ApiError) {
        showError(err.getDisplayMessage());
      } else {
        showError(t('contactDetail.updateError'));
      }
      throw err;
    }
  };

  const handleStartEditProfile = () => {
    if (!contact) return;
    setProfileValues({
      firstname: contact.firstname || ''
    });
    setEditingProfile(true);
  };

  const handleCancelEditProfile = () => {
    setEditingProfile(false);
    setProfileValues({ firstname: '' });
  };

  const handleSaveProfile = async () => {
    if (!contact || !profileValues.firstname.trim()) {
      alert(t('contactDetail.firstNameRequired'));
      return;
    }

    try {
      const updatedContact = await updateContact(id!, {
        ...toContactUpdatePayload(contact),
        firstname: profileValues.firstname.trim()
      });
      setContact(updatedContact);
      setEditingProfile(false);
    } catch (err) {
      console.error('Error updating profile:', err);
      if (err instanceof ApiError) {
        showError(err.getDisplayMessage());
      } else {
        showError(t('contactDetail.updateError'));
      }
    }
  };

  const handleDeleteContact = async () => {
    if (!contact || !id) return;

    const confirmMessage = t('contactDetail.confirmDeleteContact', {
      name: `${contact.firstname}`
    });

    if (!window.confirm(confirmMessage)) {
      return;
    }

    try {
      await deleteContact(id);
      navigate('/contacts');
    } catch (err) {
      console.error('Error deleting contact:', err);
      alert(t('contactDetail.deleteContactError'));
    }
  };

  const handleArchiveContact = () => {
    if (!contact || !id) return;
    setArchiveDialogOpen(true);
  };

  const handleConfirmArchive = async () => {
    if (!contact || !id) return;
    setArchiveDialogOpen(false);

    try {
      const updatedContact = await archiveContact(id);
      setContact({ ...contact, archived: updatedContact.archived });
    } catch (err) {
      console.error('Error archiving contact:', err);
      if (err instanceof ApiError) {
        showError(err.getDisplayMessage());
      } else {
        showError(t('contactDetail.updateError'));
      }
    }
  };

  const handleOpenDeletedNotes = async () => {
    if (!id) return;
    setDeletedNotesDialogOpen(true);
    try {
      const deletedData = await getDeletedContactNotes(id);
      setDeletedNotes(deletedData.notes || []);
    } catch (err) {
      handleFetchError(err, 'loading deleted notes');
    }
  };

  const handleUnarchiveContact = async () => {
    if (!contact || !id) return;

    try {
      const updatedContact = await unarchiveContact(id);
      setContact({ ...contact, archived: updatedContact.archived });
    } catch (err) {
      console.error('Error unarchiving contact:', err);
      if (err instanceof ApiError) {
        showError(err.getDisplayMessage());
      } else {
        showError(t('contactDetail.updateError'));
      }
    }
  };

  const handleCompaniesChange = (companies: Company[]) => {
    setContact((prev) => (prev ? { ...prev, companies } : prev));
  };

  if (loading) {
    return (
      <Box sx={{ maxWidth: 1200, mx: 'auto', mt: 1, px: 2, pb: 2 }}>
        <ContactDetailHeaderSkeleton />
        <Box sx={{ mt: 3 }}>
          <TimelineSkeleton count={5} />
        </Box>
      </Box>
    );
  }

  if (!contact) {
    return (
      <Box sx={{ maxWidth: 800, mx: 'auto', mt: 2, p: 2 }}>
        <Typography variant="h6">{t('contactDetail.notFound')}</Typography>
      </Box>
    );
  }

  return (
    <Box sx={{ maxWidth: 1200, mx: 'auto', mt: 1, px: 2, pb: 2 }}>

      {/* Contact Header Card */}
      <ContactHeader
        contact={contact}
        editingProfile={editingProfile}
        profileValues={profileValues}
        onStartEditProfile={handleStartEditProfile}
        onCancelEditProfile={handleCancelEditProfile}
        onSaveProfile={handleSaveProfile}
        onDeleteContact={handleDeleteContact}
        onProfileValueChange={setProfileValues}
        onArchiveContact={contact.archived ? undefined : handleArchiveContact}
        onUnarchiveContact={contact.archived ? handleUnarchiveContact : undefined}
      />

      {/* General Information and Timeline - Two Column Layout */}
      <Box sx={{ 
        display: 'flex', 
        flexDirection: { xs: 'column', md: 'row' }, 
        gap: 2 
      }}>
        {/* General Information */}
        <ContactInformation
          contact={contact}
          editingField={editingField}
          editValue={editValue}
          validationError={validationError}
          onEditStart={handleEditStart}
          onEditCancel={handleEditCancel}
          onEditSave={handleEditSave}
          onEditValueChange={(value) => {
            setEditValue(
              editingField === 'anniversary'
                ? autoFormatBirthdayInput(value, editValue)
                : value
            );
            setValidationError('');
          }}
          onUpdateContact={handleUpdateContactFields}
          enabledFields={enabledFields}
          onCompaniesChange={handleCompaniesChange}
        />

        {/* Timeline */}
        <Card sx={{ flex: 1 }}>
          <CardContent sx={{ py: 1.5 }}>
            <ContactTimeline
              notes={notes}
              onEditNote={handleStartEditNote}
              search={notesSearchInput}
              onSearchChange={handleNotesSearchChange}
              fromDate={notesFromDate}
              toDate={notesToDate}
              onFromDateChange={handleNotesFromDateChange}
              onToDateChange={handleNotesToDateChange}
              hasFilters={notesHasFilters}
              page={notesPage}
              totalPages={notesTotalPages}
              onPageChange={handleNotesPageChange}
              onAddNote={() => setNoteDialogOpen(true)}
              onViewDeleted={handleOpenDeletedNotes}
            />
          </CardContent>
        </Card>
      </Box>

      {/* Dialogs */}
      <AddNoteDialog
        open={noteDialogOpen}
        onClose={() => setNoteDialogOpen(false)}
        onSave={handleSaveNote}
      />

      {editingNote && (
        <EditTimelineItemDialog
          open={!!editingNote}
          onClose={handleCancelEditNote}
          onSave={() => handleUpdateNote(editingNote.ID)}
          onDelete={() => handleDeleteNote(editingNote.ID)}
          type="note"
          values={editNoteValues}
          onChange={setEditNoteValues}
          allContacts={[]}
        />
      )}

      <Dialog
        open={archiveDialogOpen}
        onClose={() => setArchiveDialogOpen(false)}
        maxWidth="xs"
        fullWidth
      >
        <DialogTitle>{t('contactDetail.archiveTitle')}</DialogTitle>
        <DialogContent>
          <Typography variant="body2">
            {t('contactDetail.archiveConfirmation')}
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setArchiveDialogOpen(false)}>
            {t('common.cancel')}
          </Button>
          <Button variant="contained" color="error" onClick={handleConfirmArchive}>
            {t('contactDetail.archive')}
          </Button>
        </DialogActions>
      </Dialog>

      <DeletedNotesDialog
        open={deletedNotesDialogOpen}
        notes={deletedNotes}
        onClose={() => setDeletedNotesDialogOpen(false)}
      />
    </Box>
  );
}
