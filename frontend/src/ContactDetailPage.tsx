import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  Contact,
  Company,
  getContact,
  updateContact,
  getContactProfilePicture,
  deleteContact,
  uploadProfilePicture,
  archiveContact,
  unarchiveContact
} from './api/contacts';
import { getCurrentUser } from './api/admin';
import { resolveEnabledFields, ContactFieldKey } from './contactFields';
import { 
  getContactNotes, 
  Note 
} from './api/notes';
import {
  Box,
  Card,
  CardContent,
  Divider,
  Button,
  Typography
} from '@mui/material';
import { ContactDetailHeaderSkeleton, TimelineSkeleton } from './components/LoadingSkeletons';
import NoteIcon from '@mui/icons-material/Note';
import AddNoteDialog from './components/AddNoteDialog';
import EditTimelineItemDialog from './components/EditTimelineItemDialog';
import ContactHeader from './components/ContactHeader';
import ContactInformation from './components/ContactInformation';
import ContactTimeline from './components/ContactTimeline';
import ProfilePictureUploadDialog from './components/ProfilePictureUploadDialog';
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
  'circles', 'photo', 'custom_fields', 'archived',
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
  const [profilePic, setProfilePic] = useState<string>('');
  const [loading, setLoading] = useState(true);
  const [editingField, setEditingField] = useState<string | null>(null);
  const [editValue, setEditValue] = useState<string>('');
  const [validationError, setValidationError] = useState<string>('');
  const [notes, setNotes] = useState<Note[]>([]);
  
  // Profile editing state
  const [editingProfile, setEditingProfile] = useState(false);
  const [profileValues, setProfileValues] = useState({
    firstname: ''
  });

  // Profile picture upload state
  const [profilePictureDialogOpen, setProfilePictureDialogOpen] = useState(false);

  // Enabled extended contact fields (UI visibility)
  const [enabledFields, setEnabledFields] = useState<Set<ContactFieldKey>>(() => resolveEnabledFields(null));

  // Unified refresh function for notes
  const refreshNotesAndActivities = async () => {
    if (!id) return;

    try {
      const notesData = await getContactNotes(id);
      setNotes(notesData.notes || []);
    } catch (err) {
      handleFetchError(err, 'refreshing notes');
    }
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

    let currentBlobUrl: string | null = null;

    const fetchData = async () => {
      try {
        // First batch: parallel fetch of core data
        const [contactData, notesData, user] = await Promise.all([
          getContact(id, CONTACT_FIELDS),
          getContactNotes(id),
          getCurrentUser().catch(err => {
            console.error('Error fetching current user preferences:', err);
            return null;
          })
        ]);

        setContact(contactData);
        setNotes(notesData.notes || []);
        setEnabledFields(resolveEnabledFields(user?.enabled_contact_fields ?? null));

        // Only fetch profile picture if contact has one (avoid unnecessary 404)
        if (contactData.photo) {
          try {
            const blob = await getContactProfilePicture(id);
            if (blob) {
              currentBlobUrl = URL.createObjectURL(blob);
              setProfilePic(currentBlobUrl);
            } else {
              setProfilePic('');
            }
          } catch (err) {
            console.error('Error fetching profile picture:', err);
          }
        } else {
          setProfilePic('');
        }

        setLoading(false);
      } catch (err) {
        console.error('Error fetching data:', err);
        setLoading(false);
      }
    };

    fetchData();

    return () => {
      if (currentBlobUrl) {
        URL.revokeObjectURL(currentBlobUrl);
      }
    };
  }, [id]);

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

  const handleArchiveContact = async () => {
    if (!contact || !id) return;

    const confirmMessage = t('contactDetail.archiveConfirmation');
    if (!window.confirm(confirmMessage)) {
      return;
    }

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

  const handleUploadProfilePicture = async (croppedImageBlob: Blob) => {
    if (!id) return;

    await uploadProfilePicture(id, croppedImageBlob);

    // Refresh the profile picture
    const blob = await getContactProfilePicture(id);
    if (blob) {
      // Revoke old URL to prevent memory leaks
      if (profilePic) {
        URL.revokeObjectURL(profilePic);
      }
      setProfilePic(URL.createObjectURL(blob));
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
        profilePic={profilePic}
        editingProfile={editingProfile}
        profileValues={profileValues}
        onStartEditProfile={handleStartEditProfile}
        onCancelEditProfile={handleCancelEditProfile}
        onSaveProfile={handleSaveProfile}
        onDeleteContact={handleDeleteContact}
        onProfileValueChange={setProfileValues}
        onUploadProfilePicture={() => setProfilePictureDialogOpen(true)}
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
          <CardContent sx={{ py: 2 }}>
            <Box sx={{ display: 'flex', justifyContent: 'flex-end', alignItems: 'center', mb: 1.5, gap: 0.5 }}>
              <Button 
                startIcon={<NoteIcon />} 
                onClick={() => setNoteDialogOpen(true)}
                variant="outlined"
                size="small"
              >
                {t('contactDetail.addNote')}
              </Button>
            </Box>
            <Divider sx={{ mb: 2 }} />
            
            <ContactTimeline
              notes={[...notes].sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime())}
              onEditNote={handleStartEditNote}
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

      <ProfilePictureUploadDialog
        open={profilePictureDialogOpen}
        onClose={() => setProfilePictureDialogOpen(false)}
        onUpload={handleUploadProfilePicture}
      />
    </Box>
  );
}
