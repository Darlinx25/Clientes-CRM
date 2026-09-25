import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import {
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  Typography,
  Stack,
  Box,
  Divider,
  Chip,
  Checkbox,
  FormControlLabel,
  IconButton
} from '@mui/material';
import BusinessIcon from '@mui/icons-material/Business';
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import AppDialog from './AppDialog';
import MultiValueField from './MultiValueField';
import { createContact, ContactValue } from '../api/contacts';
import { getCompanyTypes, CompanyType } from '../api/companyTypes';
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

interface CompanyDraft {
  id: number;
  company_number: string;
  typeIds: number[];
  customType?: string;
}

const emptyForm = {
  firstname: '',
  anniversary: '',
  rut: '',
  documento: '',
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

  const [companies, setCompanies] = useState<CompanyDraft[]>([]);
  const [typeOptions, setTypeOptions] = useState<CompanyType[]>([]);
  const [companyDialogOpen, setCompanyDialogOpen] = useState(false);
  const [editingCompany, setEditingCompany] = useState<CompanyDraft | null>(null);
  const [draftNumber, setDraftNumber] = useState('');
  const [draftTypeIds, setDraftTypeIds] = useState<number[]>([]);
  const [draftOtherType, setDraftOtherType] = useState(false);
  const [draftCustomType, setDraftCustomType] = useState('');
  const [creatingType, setCreatingType] = useState(false);

  useEffect(() => {
    if (!open) return;
    let cancelled = false;
    getCompanyTypes()
      .then(types => {
        if (!cancelled) setTypeOptions(types);
      })
      .catch(() => {
        // Not critical for the form itself
      });
    return () => {
      cancelled = true;
    };
  }, [open]);

  const handleChange = (field: string) => (event: React.ChangeEvent<HTMLInputElement>) => {
    if (field === 'anniversary') {
      setFormData({ ...formData, anniversary: autoFormatBirthdayInput(event.target.value, formData.anniversary) });
    } else {
      setFormData({ ...formData, [field]: event.target.value });
    }
  };

  const openAddCompany = () => {
    setEditingCompany(null);
    setDraftNumber('');
    setDraftTypeIds([]);
    setDraftOtherType(false);
    setDraftCustomType('');
    setCompanyDialogOpen(true);
  };

  const openEditCompany = (company: CompanyDraft) => {
    setEditingCompany(company);
    setDraftNumber(company.company_number);
    setDraftTypeIds(company.typeIds);
    setDraftOtherType(!!company.customType);
    setDraftCustomType(company.customType || '');
    setCompanyDialogOpen(true);
  };

  const closeCompanyDialog = () => {
    setCompanyDialogOpen(false);
    setEditingCompany(null);
    setDraftOtherType(false);
    setDraftCustomType('');
  };

  const toggleDraftType = (id: number) => {
    setDraftTypeIds(prev => (prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id]));
  };

  const handleCompanyDialogSave = async () => {
    if (!draftNumber.trim()) {
      showError(t('companies.numberRequired', 'Ingrese el número de empresa'));
      return;
    }
    if (draftOtherType && !draftCustomType.trim()) {
      showError(t('companies.otherTypeRequired', 'Ingrese el tipo de empresa o desmarque "Otro"'));
      return;
    }
    setCreatingType(true);
    try {
      const draft: CompanyDraft = {
        id: editingCompany ? editingCompany.id : Date.now(),
        company_number: draftNumber.trim(),
        typeIds: draftTypeIds,
        customType: draftOtherType ? draftCustomType.trim() : '',
      };
      setCompanies(prev =>
        editingCompany ? prev.map(c => (c.id === editingCompany.id ? draft : c)) : [...prev, draft]
      );
      closeCompanyDialog();
    } finally {
      setCreatingType(false);
    }
  };

  const handleDeleteCompany = (id: number) => {
    setCompanies(prev => prev.filter(c => c.id !== id));
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
        documento: formData.documento.trim(),
        contact_person: formData.contact_person.trim(),
        contact_information: formData.contact_information.trim(),
        emails: cleanEmails,
        phones: cleanPhones,
        // Derived primary scalars keep search/list and the backend in sync
        email: cleanEmails[0]?.value || '',
        phone: cleanPhones[0]?.value || '',
        // Inline companies attached to the new client
        companies:
          companies.length > 0
            ? companies.map(c => ({
                ID: 0,
                contact_id: 0,
                company_number: c.company_number,
                custom_type: c.customType,
                types: c.typeIds.map(id => ({ ID: id } as CompanyType))
              }))
            : undefined
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
    setCompanies([]);
    setCompanyDialogOpen(false);
    setEditingCompany(null);
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
            {isOn('documento') && (
              <TextField label={t('client.documento', 'Documento')} fullWidth value={formData.documento} onChange={handleChange('documento')} />
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
            />
          )}

          {isOn('emails') && (
            <MultiValueField label={t('contacts.email')} value={emails} onChange={setEmails} valueType="email" defaultType="home" hideType />
          )}
          {isOn('phones') && (
            <MultiValueField label={t('contacts.phone')} value={phones} onChange={setPhones} valueType="tel" defaultType="cell" hideType />
          )}

          <Box>
            <Divider sx={{ mb: 1.5 }} />
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1.5 }}>
              <Typography variant="subtitle1" sx={{ display: 'flex', alignItems: 'center', fontWeight: 600 }}>
                <BusinessIcon sx={{ mr: 0.5, color: 'text.secondary', fontSize: '1.2rem' }} />
                {t('companies.title', 'Empresas')}
              </Typography>
              <Button startIcon={<AddIcon />} size="small" variant="outlined" onClick={openAddCompany}>
                {t('companies.add', 'Agregar')}
              </Button>
            </Box>

            {companies.length === 0 ? (
              <Typography variant="body2" color="text.secondary">
                {t('companies.none', 'Este cliente no tiene empresas registradas.')}
              </Typography>
            ) : (
              <Stack spacing={1}>
                {companies.map((company) => (
                  <Box
                    key={company.id}
                    sx={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 1,
                      p: 1,
                      border: 1,
                      borderColor: 'divider',
                      borderRadius: 1
                    }}
                  >
                    <Box sx={{ flex: 1, minWidth: 0 }}>
                      <Typography variant="body1" sx={{ fontWeight: 500 }} noWrap>
                        {t('companies.numberLabel', 'Nº Empresa')}: {company.company_number || '—'}
                      </Typography>
                      <Stack direction="row" spacing={0.5} flexWrap="wrap" gap={0.5} sx={{ mt: 0.5 }}>
                        {company.typeIds.map(id => {
                          const ct = typeOptions.find(x => x.ID === id);
                          return ct ? (
                            <Chip key={ct.ID} label={ct.name} size="small" variant="outlined" sx={{ height: 20, fontSize: '0.72rem' }} />
                          ) : null;
                        })}
                        {company.customType && (
                          <Chip label={company.customType} size="small" color="primary" sx={{ height: 20, fontSize: '0.72rem' }} />
                        )}
                      </Stack>
                    </Box>
                    <IconButton size="small" onClick={() => openEditCompany(company)}>
                      <EditIcon fontSize="small" />
                    </IconButton>
                    <IconButton size="small" color="error" onClick={() => handleDeleteCompany(company.id)}>
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  </Box>
                ))}
              </Stack>
            )}
          </Box>

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

      <AppDialog open={companyDialogOpen} onClose={closeCompanyDialog} maxWidth="sm" fullWidth>
        <DialogTitle>
          {editingCompany ? t('companies.edit', 'Editar empresa') : t('companies.add', 'Agregar empresa')}
        </DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <TextField
              label={t('companies.companyNumber', 'Número de Empresa')}
              fullWidth
              value={draftNumber}
              onChange={(e) => setDraftNumber(e.target.value)}
              required
            />
            <Box>
              <Typography variant="subtitle2" sx={{ mb: 0.5 }}>
                {t('companies.types', 'Tipos de Empresa')}
              </Typography>
              <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 0, maxHeight: 280, overflowY: 'auto' }}>
                {typeOptions.map(ct => (
                  <FormControlLabel
                    key={ct.ID}
                    control={
                      <Checkbox
                        size="small"
                        checked={draftTypeIds.includes(ct.ID)}
                        onChange={() => toggleDraftType(ct.ID)}
                      />
                    }
                    label={<Typography variant="body2">{ct.name}</Typography>}
                  />
                ))}
                <FormControlLabel
                  control={
                    <Checkbox
                      size="small"
                      checked={draftOtherType}
                      onChange={(e) => setDraftOtherType(e.target.checked)}
                    />
                  }
                  label={<Typography variant="body2">{t('companies.otherType', 'Otro')}</Typography>}
                />
              </Box>
              {draftOtherType && (
                <TextField
                  label={t('companies.otherTypeLabel', 'Escriba el tipo de empresa')}
                  fullWidth
                  size="small"
                  value={draftCustomType}
                  onChange={e => setDraftCustomType(e.target.value)}
                  autoFocus
                />
              )}
            </Box>
            <Divider />
            <Typography variant="caption" color="text.secondary">
              {t('companies.subNote', 'Un RUT puede tener varios números de empresa, y cada número varios tipos de empresa.')}
            </Typography>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={closeCompanyDialog}>
            {t('common.cancel', 'Cancelar')}
          </Button>
          <Button onClick={handleCompanyDialogSave} variant="contained" disabled={creatingType}>
            {creatingType ? t('common.saving', 'Guardando...') : t('common.save', 'Guardar')}
          </Button>
        </DialogActions>
      </AppDialog>
    </AppDialog>
  );
}