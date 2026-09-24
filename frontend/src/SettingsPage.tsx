import { FormEvent, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Box,
  Card,
  CardContent,
  Typography,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Divider,
  TextField,
  Button,
  Stack,
  Alert
} from '@mui/material';
import { SelectChangeEvent } from '@mui/material/Select';
import LockResetIcon from '@mui/icons-material/LockReset';
import DarkModeIcon from '@mui/icons-material/DarkMode';
import AccountCircleIcon from '@mui/icons-material/AccountCircle';
import { changePassword } from './api/auth';
import { getCurrentUser } from './api/admin';
import { isAdmin } from './auth';
import { ThemePreference, useThemePreference } from './AppThemeProvider';
import UserManagementSection from './components/UserManagementSection';

export default function SettingsPage() {
  const { t } = useTranslation();
  const { preference: themePreference, setPreference: setThemePreference } = useThemePreference();
  const [username, setUsername] = useState('');
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [passwordError, setPasswordError] = useState('');
  const [passwordSuccess, setPasswordSuccess] = useState('');
  const [changingPassword, setChangingPassword] = useState(false);

  const handleThemeChange = (event: SelectChangeEvent<ThemePreference>) => {
    setThemePreference(event.target.value as ThemePreference);
  };

  // Show which account is changing its password
  useEffect(() => {
    getCurrentUser()
      .then(user => setUsername(user.username))
      .catch(() => setUsername(''));
  }, []);

  const handlePasswordChange = async (event: FormEvent) => {
    event.preventDefault();
    setPasswordError('');
    setPasswordSuccess('');

    if (newPassword !== confirmPassword) {
      setPasswordError(t('settings.password.mismatch'));
      return;
    }

    setChangingPassword(true);

    try {
      const message = await changePassword(currentPassword, newPassword);
      setPasswordSuccess(message || t('settings.password.success'));
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : t('settings.password.error');
      setPasswordError(errorMessage);
    } finally {
      setChangingPassword(false);
    }
  };

  return (
    <Box sx={{ maxWidth: 1200, mx: 'auto', mt: 2, p: 2 }}>
      <Typography variant="h5" gutterBottom sx={{ mb: 1.5 }}>
        {t('settings.title')}
      </Typography>

      <Card sx={{ mb: 2 }}>
        <CardContent sx={{ py: 1.5, '&:last-child': { pb: 1.5 } }}>
          <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
            <DarkModeIcon sx={{ mr: 1, color: 'text.secondary', fontSize: 20 }} />
            <Typography variant="subtitle1" sx={{ fontWeight: 500 }}>
              {t('settings.theme.title')}
            </Typography>
          </Box>
          <Divider sx={{ mb: 1.5 }} />

          <FormControl fullWidth size="small">
            <InputLabel id="theme-select-label">
              {t('settings.theme.label')}
            </InputLabel>
            <Select
              labelId="theme-select-label"
              value={themePreference}
              label={t('settings.theme.label')}
              onChange={handleThemeChange}
            >
              <MenuItem value="system">{t('settings.theme.options.system')}</MenuItem>
              <MenuItem value="light">{t('settings.theme.options.light')}</MenuItem>
              <MenuItem value="dark">{t('settings.theme.options.dark')}</MenuItem>
            </Select>
          </FormControl>

          <Typography variant="caption" color="text.secondary" sx={{ mt: 1, display: 'block' }}>
            {t('settings.theme.description')}
          </Typography>
        </CardContent>
      </Card>

      <Card sx={{ mb: 2 }}>
        <CardContent sx={{ py: 1.5, '&:last-child': { pb: 1.5 } }}>
          <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
            <LockResetIcon sx={{ mr: 1, color: 'text.secondary', fontSize: 20 }} />
            <Typography variant="subtitle1" sx={{ fontWeight: 500 }}>
              {t('settings.password.title')}
            </Typography>
          </Box>
          <Divider sx={{ mb: 1.5 }} />

          <form onSubmit={handlePasswordChange}>
            <Stack spacing={1.5}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.75 }}>
                <AccountCircleIcon fontSize="small" sx={{ color: 'text.secondary' }} />
                <Typography variant="body2" color="text.secondary">
                  {t('settings.password.username')}: <strong>{username}</strong>
                </Typography>
              </Box>
              <Typography variant="body2" color="text.secondary">
                {t('settings.password.description')}
              </Typography>
              {passwordError && <Alert severity="error" sx={{ py: 0 }}>{passwordError}</Alert>}
              {passwordSuccess && <Alert severity="success" sx={{ py: 0 }}>{passwordSuccess}</Alert>}
              <TextField
                label={t('settings.password.current')}
                type="password"
                value={currentPassword}
                onChange={event => setCurrentPassword(event.target.value)}
                fullWidth
                required
                size="small"
              />
              <TextField
                label={t('settings.password.new')}
                type="password"
                value={newPassword}
                onChange={event => setNewPassword(event.target.value)}
                fullWidth
                required
                size="small"
              />
              <TextField
                label={t('settings.password.confirm')}
                type="password"
                value={confirmPassword}
                onChange={event => setConfirmPassword(event.target.value)}
                fullWidth
                required
                size="small"
              />
              <Button type="submit" variant="contained" size="small" disabled={changingPassword}>
                {changingPassword ? t('settings.password.changing') : t('settings.password.changeButton')}
              </Button>
            </Stack>
          </form>
        </CardContent>
      </Card>

      {isAdmin() && <UserManagementSection />}
    </Box>
  );
}