import { useState } from 'react';
import { Card, CardContent, Stack, Box, Tabs, Tab, Typography } from '@mui/material';
import EmailIcon from '@mui/icons-material/Email';
import PhoneIcon from '@mui/icons-material/Phone';
import EventIcon from '@mui/icons-material/Event';
import BadgeIcon from '@mui/icons-material/Badge';
import NotesIcon from '@mui/icons-material/Notes';
import ContactsIcon from '@mui/icons-material/Contacts';
import { useTranslation } from 'react-i18next';
import EditableField from './EditableField';
import EditableArrayField from './EditableArrayField';
import MultiValueField from './MultiValueField';
import { Contact, ContactValue, Company } from '../api/contacts';
import { ContactFieldKey, resolveEnabledFields } from '../contactFields';
import { useDateFormat } from '../DateFormatProvider';
import CompanySection from './CompanySection';

interface ContactInformationProps {
  contact: Partial<Contact>;
  editingField: string | null;
  editValue: string;
  validationError: string;
  onEditStart: (field: string, value: string) => void;
  onEditCancel: () => void;
  onEditSave: (field: string) => void;
  onEditValueChange: (value: string) => void;
  onUpdateContact: (partial: Partial<Contact>) => Promise<void>;
  enabledFields?: Set<ContactFieldKey>;
  onCompaniesChange: (companies: Company[]) => void;
}

const iconSx = { mr: 1, color: 'text.secondary', fontSize: '1.2rem' };
const cloneValues = <T extends object>(v: T[]): T[] => v.map((x) => ({ ...x }));

export default function ContactInformation({
  contact,
  editingField,
  editValue,
  validationError,
  onEditStart,
  onEditCancel,
  onEditSave,
  onEditValueChange,
  onUpdateContact,
  enabledFields,
  onCompaniesChange,
}: ContactInformationProps) {
  const { t } = useTranslation();
  const { formatBirthday, getBirthdayPlaceholder } = useDateFormat();
  const [activeTab, setActiveTab] = useState(0);
  const enabled = enabledFields ?? resolveEnabledFields(null);
  const isOn = (key: ContactFieldKey) => enabled.has(key);

  const renderValueList = (rows: ContactValue[] | undefined) => {
    if (!rows || rows.length === 0) return <Typography variant="body2" color="text.disabled">—</Typography>;
    return (
      <Stack>
        {rows.map((r, i) => (
          <Typography key={i} variant="body2">
            {r.value}
          </Typography>
        ))}
      </Stack>
    );
  };

  return (
    <Card sx={{ flex: 1 }}>
      <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
        <Tabs value={activeTab} onChange={(_, newValue) => setActiveTab(newValue)} aria-label="contact information tabs">
          <Tab label={t('contactDetail.generalInfo')} />
        </Tabs>
      </Box>

      {/* General Information Tab */}
      {activeTab === 0 && (
        <CardContent sx={{ py: 2 }}>
          <Stack spacing={2}>
            {isOn('rut') && (
              <EditableField
                icon={<BadgeIcon sx={iconSx} />}
                label={t('client.rut', 'RUT')}
                field="rut"
                value={contact.rut || ''}
                isEditing={editingField === 'rut'}
                editValue={editValue}
                validationError={validationError}
                onEditStart={onEditStart}
                onEditCancel={onEditCancel}
                onEditSave={onEditSave}
                onEditValueChange={onEditValueChange}
              />
            )}

            {isOn('contact_person') && (
              <EditableField
                icon={<ContactsIcon sx={iconSx} />}
                label={t('client.contactPerson', 'Persona de Contacto')}
                field="contact_person"
                value={contact.contact_person || ''}
                isEditing={editingField === 'contact_person'}
                editValue={editValue}
                validationError={validationError}
                onEditStart={onEditStart}
                onEditCancel={onEditCancel}
                onEditSave={onEditSave}
                onEditValueChange={onEditValueChange}
              />
            )}

            {isOn('anniversary') && (
              <EditableField
                icon={<EventIcon sx={iconSx} />}
                label={t('client.startDate')}
                field="anniversary"
                value={contact.anniversary || ''}
                formattedDisplayValue={contact.anniversary ? formatBirthday(contact.anniversary) : undefined}
                placeholder={getBirthdayPlaceholder()}
                isEditing={editingField === 'anniversary'}
                editValue={editValue}
                validationError={validationError}
                onEditStart={onEditStart}
                onEditCancel={onEditCancel}
                onEditSave={onEditSave}
                onEditValueChange={onEditValueChange}
              />
            )}

            {isOn('emails') && (
              <EditableArrayField<ContactValue[]>
                icon={<EmailIcon sx={iconSx} />}
                label={t('contactDetail.email')}
                value={contact.emails || []}
                cloneValue={cloneValues}
                renderDisplay={renderValueList}
                renderEditor={(draft, setDraft) => (
                  <MultiValueField label={t('contacts.email')} value={draft} onChange={setDraft} valueType="email" defaultType="home" hideType />
                )}
                onSave={(draft) => {
                  const clean = draft.filter((e) => e.value.trim());
                  return onUpdateContact({ emails: clean, email: clean[0]?.value || '' });
                }}
              />
            )}

            {isOn('phones') && (
              <EditableArrayField<ContactValue[]>
                icon={<PhoneIcon sx={iconSx} />}
                label={t('contactDetail.phone')}
                value={contact.phones || []}
                cloneValue={cloneValues}
                renderDisplay={renderValueList}
                renderEditor={(draft, setDraft) => (
                  <MultiValueField label={t('contacts.phone')} value={draft} onChange={setDraft} valueType="tel" defaultType="cell" hideType />
                )}
                onSave={(draft) => {
                  const clean = draft.filter((p) => p.value.trim());
                  return onUpdateContact({ phones: clean, phone: clean[0]?.value || '' });
                }}
              />
            )}

            {isOn('contact_information') && (
              <EditableField
                icon={<NotesIcon sx={{ ...iconSx, mt: 0.5 }} />}
                label={t('client.comment')}
                field="contact_information"
                value={contact.contact_information || ''}
                multiline
                isEditing={editingField === 'contact_information'}
                editValue={editValue}
                validationError={validationError}
                onEditStart={onEditStart}
                onEditCancel={onEditCancel}
                onEditSave={onEditSave}
                onEditValueChange={onEditValueChange}
              />
            )}

            <CompanySection
              contactId={contact.ID || 0}
              companies={contact.companies || []}
              onChange={onCompaniesChange}
            />
          </Stack>
        </CardContent>
      )}
    </Card>
  );
}