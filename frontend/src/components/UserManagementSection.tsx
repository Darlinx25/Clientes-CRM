import { useState, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Divider,
  TextField,
  Button,
  Stack,
  Alert,
  IconButton,
  List,
  ListItem,
  ListItemText,
  ListItemIcon,
  Chip,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogContentText,
  DialogActions,
  Switch,
  Tooltip,
} from '@mui/material';
import PeopleIcon from '@mui/icons-material/People';
import LockResetIcon from '@mui/icons-material/LockReset';
import DeleteIcon from '@mui/icons-material/Delete';
import AdminPanelSettingsIcon from '@mui/icons-material/AdminPanelSettings';
import PersonIcon from '@mui/icons-material/Person';
import {
  getUsers,
  getCurrentUser,
  updateUser,
  deleteUser,
} from '../api/admin';
import type { User } from '../types';

export default function UserManagementSection() {
  const { t } = useTranslation();
  const [users, setUsers] = useState<User[]>([]);
  const [currentUserId, setCurrentUserId] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [passwordError, setPasswordError] = useState('');
  const [saving, setSaving] = useState(false);

  const [userToDelete, setUserToDelete] = useState<User | null>(null);
  const [deleting, setDeleting] = useState(false);

  const load = useCallback(async () => {
    try {
      setLoading(true);
      const [current, data] = await Promise.all([getCurrentUser(), getUsers(1, 100)]);
      setCurrentUserId(current.id);
      setUsers(data.users);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('users.loadError'));
    } finally {
      setLoading(false);
    }
  }, [t]);

  useEffect(() => {
    load();
  }, [load]);

  const openPasswordDialog = (user: User) => {
    setEditingUser(user);
    setNewPassword('');
    setConfirmPassword('');
    setPasswordError('');
  };

  const handleSavePassword = async () => {
    if (!editingUser) return;
    if (newPassword !== confirmPassword) {
      setPasswordError(t('users.passwordMismatch'));
      return;
    }
    setSaving(true);
    setPasswordError('');
    setError('');
    try {
      await updateUser(editingUser.id, { password: newPassword });
      setSuccess(t('users.passwordChanged'));
      setEditingUser(null);
    } catch (err) {
      setPasswordError(err instanceof Error ? err.message : t('users.updateError'));
    } finally {
      setSaving(false);
    }
  };

  const toggleAdmin = async (user: User) => {
    setError('');
    setSuccess('');
    try {
      const updated = await updateUser(user.id, { is_admin: !user.is_admin });
      setUsers(prev => prev.map(u => (u.id === updated.id ? updated : u)));
      setSuccess(t('users.adminChanged'));
    } catch (err) {
      setError(err instanceof Error ? err.message : t('users.updateError'));
    }
  };

  const handleDelete = async () => {
    if (!userToDelete) return;
    setDeleting(true);
    setError('');
    try {
      await deleteUser(userToDelete.id);
      setUsers(prev => prev.filter(u => u.id !== userToDelete.id));
      setSuccess(t('users.deleted'));
      setUserToDelete(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('users.deleteError'));
      setUserToDelete(null);
    } finally {
      setDeleting(false);
    }
  };

  return (
    <>
      <Card sx={{ mb: 2 }}>
        <CardContent sx={{ py: 1.5, '&:last-child': { pb: 1.5 } }}>
          <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 1 }}>
            <Box sx={{ display: 'flex', alignItems: 'center' }}>
              <PeopleIcon sx={{ mr: 1, color: 'text.secondary', fontSize: 20 }} />
              <Typography variant="subtitle1" sx={{ fontWeight: 500 }}>
                {t('users.title')}
              </Typography>
            </Box>
          </Box>
          <Divider sx={{ mb: 1.5 }} />

          <Stack spacing={1.5}>
            <Typography variant="body2" color="text.secondary">
              {t('users.description')}
            </Typography>

            {error && <Alert severity="error" sx={{ py: 0 }} onClose={() => setError('')}>{error}</Alert>}
            {success && <Alert severity="success" sx={{ py: 0 }} onClose={() => setSuccess('')}>{success}</Alert>}

            {loading ? (
              <Typography variant="body2" color="text.secondary">
                {t('users.loading')}
              </Typography>
            ) : users.length === 0 ? (
              <Typography variant="body2" color="text.secondary" sx={{ fontStyle: 'italic' }}>
                {t('users.noUsers')}
              </Typography>
            ) : (
              <List dense sx={{ py: 0 }}>
                {users.map(user => {
                  const isSelf = currentUserId !== null && user.id === currentUserId;
                  return (
                    <ListItem
                      key={user.id}
                      sx={{ px: 0 }}
                      secondaryAction={
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                          <Tooltip title={t('users.changePassword')}>
                            <IconButton size="small" onClick={() => openPasswordDialog(user)}>
                              <LockResetIcon fontSize="small" />
                            </IconButton>
                          </Tooltip>
                          <Tooltip title={user.is_admin ? t('users.roles.admin') : t('users.roles.user')}>
                            <Switch
                              size="small"
                              checked={user.is_admin}
                              onChange={() => toggleAdmin(user)}
                              disabled={isSelf}
                            />
                          </Tooltip>
                          <Tooltip title={t('users.delete')}>
                            <span>
                              <IconButton
                                size="small"
                                color="error"
                                disabled={isSelf}
                                onClick={() => setUserToDelete(user)}
                              >
                                <DeleteIcon fontSize="small" />
                              </IconButton>
                            </span>
                          </Tooltip>
                        </Box>
                      }
                    >
                      <ListItemIcon sx={{ minWidth: 36 }}>
                        {user.is_admin ? (
                          <AdminPanelSettingsIcon fontSize="small" color="primary" />
                        ) : (
                          <PersonIcon fontSize="small" />
                        )}
                      </ListItemIcon>
                      <ListItemText
                        primaryTypographyProps={{ component: 'div' }}
                        secondaryTypographyProps={{ component: 'div' }}
                        primary={
                          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                            <Typography variant="body2" sx={{ fontWeight: 500 }}>
                              {user.username}
                            </Typography>
                            {isSelf && (
                              <Chip
                                size="small"
                                label={t('users.self')}
                                sx={{ height: 18, fontSize: '0.7rem' }}
                              />
                            )}
                            <Chip
                              size="small"
                              label={user.is_admin ? t('users.roles.admin') : t('users.roles.user')}
                              color={user.is_admin ? 'primary' : 'default'}
                              sx={{ height: 18, fontSize: '0.7rem' }}
                            />
                          </Box>
                        }
                        secondary={
                          <Typography variant="caption" color="text.secondary">
                            {user.email}
                          </Typography>
                        }
                      />
                    </ListItem>
                  );
                })}
              </List>
            )}
          </Stack>
        </CardContent>
      </Card>

      {/* Change password dialog */}
      <Dialog open={!!editingUser} onClose={() => setEditingUser(null)} maxWidth="xs" fullWidth>
        <DialogTitle>
          {t('users.editDialog.title')} — {editingUser?.username ?? ''}
        </DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <TextField
              label={t('users.editDialog.password')}
              type="password"
              value={newPassword}
              onChange={e => setNewPassword(e.target.value)}
              size="small"
              fullWidth
            />
            <TextField
              label={t('users.confirmPassword')}
              type="password"
              value={confirmPassword}
              onChange={e => setConfirmPassword(e.target.value)}
              size="small"
              fullWidth
              error={confirmPassword !== '' && newPassword !== confirmPassword}
            />
            {passwordError && <Alert severity="error" sx={{ py: 0 }}>{passwordError}</Alert>}
            <Typography variant="caption" color="text.secondary">
              {t('users.passwordDescription')}
            </Typography>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setEditingUser(null)}>
            {t('users.cancel')}
          </Button>
          <Button
            onClick={handleSavePassword}
            variant="contained"
            disabled={saving || newPassword === ''}
          >
            {saving ? t('users.saving') : t('users.save')}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Delete confirmation dialog */}
      <Dialog open={!!userToDelete} onClose={() => setUserToDelete(null)}>
        <DialogTitle>{t('users.deleteDialog.title')}</DialogTitle>
        <DialogContent>
          <DialogContentText>
            {t('users.deleteDialog.message', { username: userToDelete?.username ?? '' })}
          </DialogContentText>
          <DialogContentText sx={{ mt: 1, color: 'warning.main' }}>
            {t('users.deleteDialog.warning')}
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setUserToDelete(null)}>
            {t('users.cancel')}
          </Button>
          <Button onClick={handleDelete} color="error" disabled={deleting} autoFocus>
            {t('users.delete')}
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
}