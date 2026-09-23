import { useState, useEffect } from 'react';
import {
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Button,
  Box,
  Autocomplete,
} from '@mui/material';
import AppDialog from './AppDialog';
import { useTranslation } from 'react-i18next';
import { getContacts, Contact } from '../api/contacts';
import { useDebouncedValue } from '../hooks/useDebounce';

interface AddNoteDialogProps {
  open: boolean;
  onClose: () => void;
  onSave: (title: string, content: string, date: string, contactId?: number) => Promise<void>;
  contactRequired?: boolean;
}

export default function AddNoteDialog({ open, onClose, onSave, contactRequired = false }: AddNoteDialogProps) {
  const { t } = useTranslation();
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [selected, setSelected] = useState<Contact | null>(null);
  const [searchInput, setSearchInput] = useState('');
  const debouncedSearch = useDebouncedValue(searchInput, 300);
  const [options, setOptions] = useState<Contact[]>([]);
  const [optionsLoading, setOptionsLoading] = useState(false);
  const [error, setError] = useState('');
  const [saving, setSaving] = useState(false);

  // Search clients by name (server-side) for the autocomplete picker.
  useEffect(() => {
    if (!open || !contactRequired) return;
    let active = true;
    setOptionsLoading(true);
    getContacts({ search: debouncedSearch.trim(), limit: 25, archived: false })
      .then((res) => {
        if (active) setOptions(res.contacts);
      })
      .catch(() => {
        if (active) setOptions([]);
      })
      .finally(() => {
        if (active) setOptionsLoading(false);
      });
    return () => {
      active = false;
    };
  }, [open, contactRequired, debouncedSearch]);

  const fullName = (c: Contact) => `${c.firstname} ${c.lastname || ''}`.trim() || c.nickname || String(c.ID);

  const handleSave = async () => {
    if (!content.trim()) {
      setError(t('noteDialog.required'));
      return;
    }

    if (contactRequired && !selected) {
      setError(t('noteDialog.clientRequired'));
      return;
    }

    setSaving(true);
    try {
      // Date is always the creation date, not user-selectable.
      await onSave(title.trim(), content, new Date().toISOString(), selected?.ID);
      handleClose();
    } catch (err) {
      setError(t('noteDialog.saveError'));
    } finally {
      setSaving(false);
    }
  };

  const handleClose = () => {
    setTitle('');
    setContent('');
    setSelected(null);
    setSearchInput('');
    setError('');
    onClose();
  };

  return (
    <AppDialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle>{t('noteDialog.title')}</DialogTitle>
      <DialogContent>
        <Box sx={{ pt: 1, display: 'flex', flexDirection: 'column', gap: 2 }}>
          {contactRequired && (
            <Autocomplete
              options={options}
              value={selected}
              onChange={(_e, v: Contact | null) => {
                setSelected(v);
                setError('');
              }}
              onInputChange={(_e, v) => setSearchInput(v)}
              getOptionLabel={(o) => fullName(o)}
              isOptionEqualToValue={(o, v) => o.ID === v.ID}
              loading={optionsLoading}
              renderInput={(params) => (
                <TextField
                  {...params}
                  label={t('noteDialog.client')}
                  placeholder={t('noteDialog.searchClient')}
                  required
                  error={contactRequired && !selected && error === t('noteDialog.clientRequired')}
                  helperText={error === t('noteDialog.clientRequired') ? error : undefined}
                />
              )}
            />
          )}
          <TextField
            label={t('noteDialog.noteTitle')}
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            fullWidth
            autoFocus
          />
          <TextField
            label={t('noteDialog.content')}
            placeholder={t('noteDialog.contentPlaceholder')}
            multiline
            rows={4}
            value={content}
            onChange={(e) => {
              setContent(e.target.value);
              setError('');
            }}
            error={error === t('noteDialog.required')}
            helperText={error === t('noteDialog.required') ? error : ''}
            fullWidth
            required
          />
          <TextField
            disabled
            label={t('noteDialog.date')}
            value={new Date().toISOString().split('T')[0]}
            fullWidth
          />
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose} disabled={saving}>
          {t('noteDialog.cancel')}
        </Button>
        <Button onClick={handleSave} variant="contained" disabled={saving}>
          {t('noteDialog.save')}
        </Button>
      </DialogActions>
    </AppDialog>
  );
}