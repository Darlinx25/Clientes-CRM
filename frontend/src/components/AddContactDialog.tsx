import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  Typography,
  Stack
} from '@mui/material';
import AppDialog from './AppDialog';
import MultiValueField from './MultiValueField';
import { createContact, ContactValue } from '../api/contacts';
import { useSnackbar } from '../context/SnackbarContext';
import { handleError, getErrorMessage } from '../utils/errorHandler';
import { useDateFormat } from '../DateFormatProvider';
import { ContactFieldKey, resolveEnabledFields } from '../contactFields';

interface AddContactDialogProps {
  open: boolean;
  onClose: () => void;
  onContactAdded: (contactId: number) => void;
  customFieldNames?: string[];
  enabledFields?: Set<ContactFieldKey>;
}

const emptyForm = {
  firstname: '',
  anniversary: '',
  rut: '',
  contact_person: '',
  contact_information: ''
};

export default function AddContactDialog({
  open,
  onClose,
  onContactAdded,
  enabledFields
}: AddContactDialogProps) {
  const { t } = useTranslation();
  const { showError, showSuccess } = useSnackbar();
  const { parseBirthdayInput, getBirthdayPlaceholder, autoFormatBirthdayInput } = useDateFormat();
  const enabled = enabledFields ?? resolveEnabledFields(null);
  const isOn = (key: ContactFieldKey) => enabled.has(key);

  const [formData, setFormData] = useState({ ...emptyForm });
  const [emails, setEmails] = useState<ContactValue[]>([]);
  const [phones, setPhones] = useState<ContactValue[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleChange = (field: string) => (event: React.ChangeEvent<HTMLInputElement>) => {
    if (field === 'anniversary') {
      setFormData({ ...formData, anniversary: autoFormatBirthdayInput(event.target.value, formData.anniversary) });
    } else {
      setFormData({ ...formData, [field]: event.target.value });
    }
  };

  const handleSubmit = async () => {
    if (!formData.firstname.trim()) {
      setError(t('contacts.add.requiredFields'));
      return;
    }

    let anniversaryISO = '';
    if (formData.anniversary.trim()) {
      const parsed = parseBirthdayInput(formData.anniversary);
      if (parsed === null) {
        setError(t('contactDetail.birthdayError'));
        return;
      }
      anniversaryISO = parsed;
    }

    setLoading(true);
    setError('');

    try {
      // Clean empty rows from multi-value fields
      const cleanEmails = emails.filter(e => e.value.trim());
      const cleanPhones = phones.filter(p => p.value.trim());

      const contactData = {
        firstname: formData.firstname,
        anniversary: anniversaryISO,
        rut: formData.rut.trim(),
        contact_person: formData.contact_person.trim(),
        contact_information: formData.contact_information.trim(),
        emails: cleanEmails,
        phones: cleanPhones,
        // Derived primary scalars keep search/list and the backend in sync
        email: cleanEmails[0]?.value || '',
        phone: cleanPhones[0]?.value || ''
      };

      const newContact = await createContact(contactData);

      onContactAdded(newContact.ID);
      showSuccess(t('contacts.add.success'));
      handleClose();
    } catch (err) {
      handleError(err, { operation: 'creating contact' }, { showError });
      const errorMessage = getErrorMessage(err);
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    setFormData({ ...emptyForm });
    setEmails([]);
    setPhones([]);
    setError('');
    onClose();
  };

  return (
    <AppDialog open={open} onClose={handleClose} maxWidth="md" fullWidth>
      <DialogTitle>{t('contacts.add.title')}</DialogTitle>
      <DialogContent>
        {error && (
          <Typography color="error" sx={{ mb: 2 }}>
            {error}
          </Typography>
        )}
        <Stack spacing={2} sx={{ mt: 1 }}>
          <TextField
            label={t('client.denomination')}
            fullWidth
            value={formData.firstname}
            onChange={handleChange('firstname')}
            required
          />

          <Stack direction="row" spacing={2}>
            {isOn('rut') && (
              <TextField label={t('client.rut', 'RUT')} fullWidth value={formData.rut} onChange={handleChange('rut')} />
            )}
            {isOn('contact_person') && (
              <TextField label={t('client.contactPerson', 'Persona de Contacto')} fullWidth value={formData.contact_person} onChange={handleChange('contact_person')} />
            )}
          </Stack>

          {isOn('anniversary') && (
            <TextField
              label={t('client.startDate')}
              fullWidth
              value={formData.anniversary}
              onChange={handleChange('anniversary')}
              placeholder={getBirthdayPlaceholder()}
              helperText={t('contacts.birthdayFormat')}
            />
          )}

          {isOn('emails') && (
            <MultiValueField label={t('contacts.email')} value={emails} onChange={setEmails} valueType="email" defaultType="home" />
          )}
          {isOn('phones') && (
            <MultiValueField label={t('contacts.phone')} value={phones} onChange={setPhones} valueType="tel" defaultType="cell" />
          )}

          {isOn('contact_information') && (
            <TextField
              label={t('client.comment')}
              fullWidth
              multiline
              rows={2}
              value={formData.contact_information}
              onChange={handleChange('contact_information')}
            />
          )}
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose} disabled={loading}>
          {t('common.cancel')}
        </Button>
        <Button onClick={handleSubmit} variant="contained" disabled={loading}>
          {loading ? t('common.saving') : t('contacts.add.create')}
        </Button>
      </DialogActions>
    </AppDialog>
  );
}