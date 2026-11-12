import { useEffect, useState } from 'react';
import {
  Alert,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  TextField,
} from '@mui/material';
import { useTranslation } from 'react-i18next';
import { ApiError } from '../../api/client';
import { startAdminPanelSession } from '../../api/adminPanel';

interface AdminReauthDialogProps {
  open: boolean;
  onAuthenticated: () => void;
  onCancel: () => void;
}

export default function AdminReauthDialog({
  open,
  onAuthenticated,
  onCancel,
}: AdminReauthDialogProps) {
  const { t } = useTranslation();
  const [password, setPassword] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (open) {
      setPassword('');
      setError('');
      setSubmitting(false);
    }
  }, [open]);

  const handleSubmit = async () => {
    if (!password || submitting) {
      return;
    }
    setSubmitting(true);
    setError('');
    try {
      await startAdminPanelSession(password);
      onAuthenticated();
    } catch (err) {
      if (err instanceof ApiError && err.code === 'INVALID_CREDENTIALS') {
        setError(t('admin.reauth.invalid'));
      } else if (err instanceof ApiError && err.status === 401) {
        localStorage.removeItem('user_info');
        window.location.href = '/login';
      } else {
        setError(err instanceof ApiError ? err.getDisplayMessage() : t('admin.errors.load'));
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onClose={submitting ? undefined : onCancel} maxWidth="xs" fullWidth>
      <DialogTitle>{t('admin.reauth.title')}</DialogTitle>
      <DialogContent>
        <DialogContentText sx={{ mb: 2 }}>{t('admin.reauth.description')}</DialogContentText>
        {error && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {error}
          </Alert>
        )}
        <TextField
          autoFocus
          fullWidth
          type="password"
          label={t('admin.reauth.password')}
          value={password}
          onChange={(event) => setPassword(event.target.value)}
          onKeyDown={(event) => {
            if (event.key === 'Enter') {
              void handleSubmit();
            }
          }}
          disabled={submitting}
        />
      </DialogContent>
      <DialogActions>
        <Button onClick={onCancel} disabled={submitting}>
          {t('admin.reauth.cancel')}
        </Button>
        <Button
          variant="contained"
          onClick={() => void handleSubmit()}
          disabled={!password || submitting}
        >
          {submitting ? t('admin.reauth.unlocking') : t('admin.reauth.unlock')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
