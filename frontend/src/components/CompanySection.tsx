import { useState, useCallback, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Box,
  Typography,
  IconButton,
  Button,
  Stack,
  Chip,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Checkbox,
  FormControlLabel,
  Divider,
} from '@mui/material';
import BusinessIcon from '@mui/icons-material/Business';
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import AppDialog from './AppDialog';
import {
  Company,
  CompanyInput,
  createCompany,
  updateCompany,
  deleteCompany,
} from '../api/contacts';
import { getCompanyTypes, CompanyType } from '../api/companyTypes';
import { useSnackbar } from '../context/SnackbarContext';
import { handleError } from '../utils/errorHandler';

interface CompanySectionProps {
  contactId: number;
  companies: Company[];
  onChange: (companies: Company[]) => void;
}

interface CompanyDialogState {
  open: boolean;
  company: Company | null;
  companyNumber: string;
  typeIds: number[];
  otherType: boolean;
  customType: string;
}

export default function CompanySection({ contactId, companies, onChange }: CompanySectionProps) {
  const { t } = useTranslation();
  const { showSuccess, showError } = useSnackbar();
  const [dialog, setDialog] = useState<CompanyDialogState>({
    open: false,
    company: null,
    companyNumber: '',
    typeIds: [],
    otherType: false,
    customType: '',
  });
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState<number | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Company | null>(null);
  const [typeOptions, setTypeOptions] = useState<CompanyType[]>([]);

  useEffect(() => {
    let cancelled = false;
    getCompanyTypes()
      .then(types => {
        if (!cancelled) {
          setTypeOptions(types);
        }
      })
      .catch(() => {
        // Not critical for the form itself
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const openAdd = () => {
    setDialog({ open: true, company: null, companyNumber: '', typeIds: [], otherType: false, customType: '' });
  };

  const openEdit = (company: Company) => {
    setDialog({
      open: true,
      company,
      companyNumber: company.company_number,
      typeIds: (company.types || []).map(ct => ct.ID),
      otherType: !!company.custom_type,
      customType: company.custom_type || '',
    });
  };

  const closeDialog = () => {
    setDialog((d) => ({ ...d, open: false }));
  };

  const toggleType = (id: number) => {
    setDialog((d) => ({
      ...d,
      typeIds: d.typeIds.includes(id) ? d.typeIds.filter(x => x !== id) : [...d.typeIds, id],
    }));
  };

  const handleSave = async () => {
    if (!dialog.companyNumber.trim()) {
      showError(t('companies.numberRequired', 'Ingrese el número de empresa'));
      return;
    }

    if (dialog.otherType && !dialog.customType.trim()) {
      showError(t('companies.otherTypeRequired', 'Ingrese el tipo de empresa o desmarque "Otro"'));
      return;
    }

    const payload: CompanyInput = {
      company_number: dialog.companyNumber.trim(),
      type_ids: dialog.typeIds,
      custom_type: dialog.otherType ? dialog.customType.trim() : '',
    };

    setSaving(true);
    try {
      if (dialog.company) {
        const updated = await updateCompany(contactId, dialog.company.ID, payload);
        onChange(companies.map((c) => (c.ID === updated.ID ? updated : c)));
        showSuccess(t('companies.updated', 'Empresa actualizada'));
      } else {
        const created = await createCompany(contactId, payload);
        onChange([...companies, created]);
        showSuccess(t('companies.added', 'Empresa agregada'));
      }

      closeDialog();
    } catch (err) {
      handleError(err, { operation: 'saving company' }, { showError });
    } finally {
      setSaving(false);
    }
  };

  const handleDeleteRequest = (company: Company) => {
    setDeleteTarget(company);
  };

  const handleDeleteConfirm = useCallback(async () => {
    if (!deleteTarget) return;
    setDeleting(deleteTarget.ID);
    try {
      await deleteCompany(contactId, deleteTarget.ID);
      onChange(companies.filter((c) => c.ID !== deleteTarget.ID));
      showSuccess(t('companies.deleted', 'Empresa eliminada'));
      setDeleteTarget(null);
    } catch (err) {
      handleError(err, { operation: 'deleting company' }, { showError });
    } finally {
      setDeleting(null);
    }
  }, [contactId, companies, onChange, deleteTarget, showError, showSuccess, t]);

  return (
    <>
      <Divider sx={{ mb: 1.5 }} />
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1.5 }}>
        <Typography variant="subtitle1" sx={{ display: 'flex', alignItems: 'center', fontWeight: 600 }}>
          <BusinessIcon sx={{ mr: 0.5, color: 'text.secondary', fontSize: '1.2rem' }} />
          {t('companies.title', 'Empresas')}
        </Typography>
        <Button startIcon={<AddIcon />} size="small" variant="outlined" onClick={openAdd}>
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
              key={company.ID}
              sx={{
                display: 'flex',
                alignItems: 'center',
                gap: 1,
                p: 1,
                border: 1,
                borderColor: 'divider',
                borderRadius: 1,
                '&:hover .action-icon': {
                  opacity: 1
                }
              }}
            >
              <Box sx={{ flex: 1, minWidth: 0 }}>
                <Typography variant="body1" sx={{ fontWeight: 500 }} noWrap>
                  {t('companies.numberLabel', 'Nº Empresa')}: {company.company_number || '—'}
                </Typography>
                <Stack direction="row" spacing={0.5} flexWrap="wrap" gap={0.5} sx={{ mt: 0.5 }}>
                  {(company.types || []).map(ct => (
                    <Chip key={ct.ID} label={ct.name} size="small" variant="outlined" sx={{ height: 20, fontSize: '0.72rem' }} />
                  ))}
                  {company.custom_type && (
                    <Chip label={company.custom_type} size="small" color="primary" sx={{ height: 20, fontSize: '0.72rem' }} />
                  )}
                </Stack>
              </Box>
              <IconButton
                className="action-icon"
                size="small"
                onClick={() => openEdit(company)}
                sx={{ opacity: 0, transition: 'opacity 0.2s' }}
              >
                <EditIcon fontSize="small" />
              </IconButton>
              <IconButton
                className="action-icon"
                size="small"
                color="error"
                disabled={deleting === company.ID}
                onClick={() => handleDeleteRequest(company)}
                sx={{ opacity: 0, transition: 'opacity 0.2s' }}
              >
                <DeleteIcon fontSize="small" />
              </IconButton>
            </Box>
          ))}
        </Stack>
      )}

      <AppDialog open={dialog.open} onClose={closeDialog} maxWidth="sm" fullWidth>
        <DialogTitle>
          {dialog.company ? t('companies.edit', 'Editar empresa') : t('companies.add', 'Agregar empresa')}
        </DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <TextField
              label={t('companies.companyNumber', 'Número de Empresa')}
              fullWidth
              value={dialog.companyNumber}
              onChange={(e) => setDialog({ ...dialog, companyNumber: e.target.value })}
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
                        checked={dialog.typeIds.includes(ct.ID)}
                        onChange={() => toggleType(ct.ID)}
                      />
                    }
                    label={<Typography variant="body2">{ct.name}</Typography>}
                  />
                ))}
                <FormControlLabel
                  control={
                    <Checkbox
                      size="small"
                      checked={dialog.otherType}
                      onChange={(e) => setDialog({ ...dialog, otherType: e.target.checked })}
                    />
                  }
                  label={<Typography variant="body2">{t('companies.otherType', 'Otro')}</Typography>}
                />
              </Box>
              {dialog.otherType && (
                <TextField
                  label={t('companies.otherTypeLabel', 'Escriba el tipo de empresa')}
                  fullWidth
                  size="small"
                  value={dialog.customType}
                  onChange={e => setDialog({ ...dialog, customType: e.target.value })}
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
          <Button onClick={closeDialog} disabled={saving}>
            {t('common.cancel', 'Cancelar')}
          </Button>
          <Button onClick={handleSave} variant="contained" disabled={saving}>
            {saving ? t('common.saving', 'Guardando...') : t('common.save', 'Guardar')}
          </Button>
        </DialogActions>
      </AppDialog>

      <AppDialog open={!!deleteTarget} onClose={() => setDeleteTarget(null)} maxWidth="xs">
        <DialogTitle>{t('companies.deleteTitle', 'Eliminar empresa')}</DialogTitle>
        <DialogContent>
          <Typography variant="body2">
            {t('companies.deleteConfirm', '¿Eliminar esta empresa?')}
            {deleteTarget ? ` (Nº ${deleteTarget.company_number || '—'})` : ''}
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteTarget(null)} disabled={deleting !== null}>
            {t('common.cancel', 'Cancelar')}
          </Button>
          <Button onClick={handleDeleteConfirm} color="error" variant="contained" disabled={deleting !== null}>
            {deleting !== null ? t('common.deleting', 'Eliminando...') : t('companies.delete', 'Eliminar')}
          </Button>
        </DialogActions>
      </AppDialog>
    </>
  );
}