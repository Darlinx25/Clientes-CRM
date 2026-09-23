import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Box,
  Typography,
  Alert,
  LinearProgress,
  Table,
  TableBody,
  TableCell,
  TableRow,
  Chip,
} from '@mui/material';
import AppDialog from './AppDialog';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import WarningIcon from '@mui/icons-material/Warning';
import ErrorIcon from '@mui/icons-material/Error';
import { uploadExcelForImport, ExcelImportResult } from '../api/excelImport';
import { getErrorMessage } from '../utils/errorHandler';
import { useSnackbar } from '../context/SnackbarContext';

const REQUIRED_COLUMNS = [
  'Cliente',
  'Tipo de Empresa',
  'Persona de Contacto',
  'Email',
  'Celular',
  'RUT',
  'Número de Empresa',
];

interface ExcelImportDialogProps {
  open: boolean;
  onClose: () => void;
  onImportComplete: () => void;
}

export default function ExcelImportDialog({
  open,
  onClose,
  onImportComplete,
}: ExcelImportDialogProps) {
  const { t } = useTranslation();
  const { showSuccess, showError: showSnackbarError } = useSnackbar();

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [dragOver, setDragOver] = useState(false);
  const [result, setResult] = useState<ExcelImportResult | null>(null);

  const resetDialog = () => {
    setLoading(false);
    setError(null);
    setDragOver(false);
    setResult(null);
  };

  const handleClose = () => {
    resetDialog();
    onClose();
  };

  const handleFile = async (file: File) => {
    const fileName = file.name.toLowerCase();
    const isXlsx = fileName.endsWith('.xlsx') || fileName.endsWith('.xlsm');

    if (!isXlsx) {
      setError(t('contacts.importExcel.errors.invalidFile', 'Seleccione un archivo Excel (.xlsx)'));
      return;
    }

    if (file.size > 20 * 1024 * 1024) {
      setError(t('contacts.importExcel.errors.fileTooLarge', 'El archivo es muy grande. Tamaño máximo: 20 MB'));
      return;
    }

    setLoading(true);
    setError(null);
    setResult(null);

    try {
      const res = await uploadExcelForImport(file);
      setResult(res);
      if (res.contacts_created > 0 || res.contacts_updated > 0) {
        showSuccess(
          t('contacts.importExcel.result.success', 'Importación completada: {{created}} clientes nuevos, {{updated}} actualizados', {
            created: res.contacts_created,
            updated: res.contacts_updated,
          })
        );
        onImportComplete();
      }
    } catch (err) {
      const message = getErrorMessage(err);
      setError(message);
      showSnackbarError(message);
    } finally {
      setLoading(false);
    }
  };

  const handleFileInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      handleFile(file);
    }
    event.target.value = '';
  };

  const handleDragOver = (event: React.DragEvent) => {
    event.preventDefault();
    setDragOver(true);
  };

  const handleDragLeave = () => {
    setDragOver(false);
  };

  const handleDrop = (event: React.DragEvent) => {
    event.preventDefault();
    setDragOver(false);
    const file = event.dataTransfer.files?.[0];
    if (file) {
      handleFile(file);
    }
  };

  const renderUpload = () => (
    <Box sx={{ py: 3 }}>
      <Alert severity="info" sx={{ mb: 2 }}>
        {t(
          'contacts.importExcel.instructions',
          'El archivo debe tener una fila de encabezados con las siguientes columnas (el orden no importa):'
        )}
      </Alert>
      <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5, mb: 2 }}>
        {REQUIRED_COLUMNS.map((col) => (
          <Chip key={col} label={col} size="small" variant="outlined" />
        ))}
      </Box>
      <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
        {t(
          'contacts.importExcel.dedupInfo',
          'Los clientes se agrupan por RUT: si el mismo RUT aparece en varias filas con distinto tipo o número de empresa, se guarda un solo cliente con todas sus empresas.'
        )}
      </Typography>
      <Box
        sx={{
          border: '2px dashed',
          borderColor: dragOver ? 'primary.main' : 'grey.400',
          borderRadius: 2,
          p: 5,
          textAlign: 'center',
          cursor: 'pointer',
          bgcolor: dragOver ? 'action.hover' : 'background.paper',
          transition: 'all 0.2s',
        }}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        onClick={() => document.getElementById('excel-import-file-input')?.click()}
      >
        <input
          id="excel-import-file-input"
          type="file"
          accept=".xlsx,.xlsm"
          style={{ display: 'none' }}
          onChange={handleFileInputChange}
        />
        <CloudUploadIcon sx={{ fontSize: 48, color: 'grey.500', mb: 2 }} />
        <Typography variant="h6" gutterBottom>
          {t('contacts.importExcel.upload.dragDrop', 'Arrastre un archivo Excel aquí, o haga clic para seleccionar')}
        </Typography>
        <Typography variant="body2" color="text.secondary">
          {t('contacts.importExcel.upload.supportedFormats', 'Formatos soportados: .xlsx (Excel)')}
        </Typography>
      </Box>
    </Box>
  );

  const renderResult = () => {
    if (!result) return null;
    return (
      <Box sx={{ py: 2 }}>
        <Alert severity="success" sx={{ mb: 2 }}>
          {t('contacts.importExcel.result.title', 'Importación completada')}
        </Alert>
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
          <Typography>
            <CheckCircleIcon color="success" sx={{ verticalAlign: 'middle', mr: 1 }} />
            {t('contacts.importExcel.result.created', '{{count}} clientes creados', { count: result.contacts_created })}
          </Typography>
          <Typography>
            <WarningIcon color="warning" sx={{ verticalAlign: 'middle', mr: 1 }} />
            {t('contacts.importExcel.result.updated', '{{count}} clientes actualizados', { count: result.contacts_updated })}
          </Typography>
          <Typography>
            <CheckCircleIcon color="info" sx={{ verticalAlign: 'middle', mr: 1 }} />
            {t('contacts.importExcel.result.companies', '{{count}} empresas agregadas', { count: result.companies_added })}
          </Typography>
          <Typography color="text.secondary">
            {t('contacts.importExcel.result.rows', '{{count}} filas procesadas, {{skipped}} omitidas', {
              count: result.rows_processed,
              skipped: result.rows_skipped,
            })}
          </Typography>
        </Box>
        {result.errors && result.errors.length > 0 && (
          <Box sx={{ mt: 2 }}>
            <Typography variant="subtitle2" color="error" gutterBottom>
              {t('contacts.importExcel.result.errors', 'Errores')}:
            </Typography>
            <Table size="small">
              <TableBody>
                {result.errors.map((err, i) => (
                  <TableRow key={i}>
                    <TableCell>
                      <Typography variant="body2" color="error">
                        <ErrorIcon sx={{ verticalAlign: 'middle', mr: 0.5, fontSize: 16 }} />
                        {err}
                      </Typography>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </Box>
        )}
      </Box>
    );
  };

  return (
    <AppDialog open={open} onClose={handleClose} maxWidth="md" fullWidth>
      <DialogTitle>{t('contacts.importExcel.title', 'Cargar clientes desde Excel')}</DialogTitle>
      <DialogContent dividers>
        {loading && <LinearProgress sx={{ mb: 2 }} />}
        {error && (
          <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError(null)}>
            {error}
          </Alert>
        )}
        {result ? renderResult() : renderUpload()}
      </DialogContent>
      <DialogActions>
        {result ? (
          <Button variant="contained" onClick={handleClose}>
            {t('contacts.importExcel.result.done', 'Hecho')}
          </Button>
        ) : (
          <>
            <Button onClick={handleClose} disabled={loading}>
              {t('common.cancel', 'Cancelar')}
            </Button>
            <Button
              variant="contained"
              onClick={() => document.getElementById('excel-import-file-input')?.click()}
              disabled={loading}
            >
              {t('contacts.importExcel.upload.selectFile', 'Seleccionar archivo')}
            </Button>
          </>
        )}
      </DialogActions>
    </AppDialog>
  );
}