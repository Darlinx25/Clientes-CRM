import { useCallback, useEffect, useState } from 'react';
import {
  Alert,
  Box,
  Button,
  CircularProgress,
  Tab,
  Tabs,
  Typography,
} from '@mui/material';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { ApiError } from './api/client';
import {
  endAdminPanelSession,
  getAdminPanelBackups,
  getAdminPanelOverview,
  getAdminPanelTables,
  isAdminSessionRequiredError,
} from './api/adminPanel';
import type {
  AdminPanelBackup,
  AdminPanelOverview,
  AdminPanelTableSummary,
} from './types/adminPanel';
import AdminBackupsTab from './components/admin/AdminBackupsTab';
import AdminDataTab from './components/admin/AdminDataTab';
import AdminOverviewTab from './components/admin/AdminOverviewTab';
import AdminReauthDialog from './components/admin/AdminReauthDialog';

export default function AdminPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [needsAuth, setNeedsAuth] = useState(true);
  const [forbidden, setForbidden] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [tab, setTab] = useState(0);
  const [overview, setOverview] = useState<AdminPanelOverview | null>(null);
  const [tables, setTables] = useState<AdminPanelTableSummary[]>([]);
  const [backups, setBackups] = useState<AdminPanelBackup[]>([]);
  const [backupDir, setBackupDir] = useState('');

  const handleApiError = useCallback((err: unknown) => {
    if (err instanceof ApiError && err.status === 401) {
      if (isAdminSessionRequiredError(err)) {
        setNeedsAuth(true);
      } else {
        localStorage.removeItem('user_info');
        window.location.href = '/login';
      }
      return true;
    }
    if (err instanceof ApiError && err.status === 403) {
      setForbidden(true);
      return true;
    }
    return false;
  }, []);

  const loadData = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const [overviewData, tableData, backupData] = await Promise.all([
        getAdminPanelOverview(),
        getAdminPanelTables(),
        getAdminPanelBackups(),
      ]);
      setOverview(overviewData);
      setTables(tableData);
      setBackups(backupData.backups);
      setBackupDir(backupData.backup_dir);
      setForbidden(false);
      setNeedsAuth(false);
    } catch (err) {
      if (!handleApiError(err)) {
        setError(err instanceof ApiError ? err.getDisplayMessage() : t('admin.errors.load'));
      }
    } finally {
      setLoading(false);
    }
  }, [handleApiError, t]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  const handleAuthenticated = () => {
    setNeedsAuth(false);
    void loadData();
  };

  const handleSessionExpired = useCallback(() => {
    setNeedsAuth(true);
  }, []);

  const handleCancel = async () => {
    try {
      await endAdminPanelSession();
    } catch {
      // Best effort: leaving the panel must not depend on the server response.
    }
    navigate('/contacts');
  };

  if (forbidden) {
    return (
      <Box>
        <Alert
          severity="error"
          action={
            <Button color="inherit" size="small" onClick={() => navigate('/contacts')}>
              {t('admin.reauth.cancel')}
            </Button>
          }
        >
          {t('admin.errors.forbidden')}
        </Alert>
      </Box>
    );
  }

  return (
    <Box>
      <Typography variant="h5" sx={{ mb: 0.5 }}>
        {t('admin.title')}
      </Typography>
      <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
        {t('admin.subtitle')}
      </Typography>

      {error && (
        <Alert
          severity="error"
          sx={{ mb: 2 }}
          action={
            <Button color="inherit" size="small" onClick={() => void loadData()}>
              {t('admin.retry')}
            </Button>
          }
        >
          {error}
        </Alert>
      )}

      {loading && !overview ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', py: 8 }}>
          <CircularProgress />
        </Box>
      ) : overview ? (
        <>
          <Tabs value={tab} onChange={(_, value) => setTab(value)} sx={{ mb: 2 }}>
            <Tab label={t('admin.tabs.overview')} />
            <Tab label={t('admin.tabs.data')} />
            <Tab label={t('admin.tabs.backups')} />
          </Tabs>
          {tab === 0 && <AdminOverviewTab overview={overview} />}
          {tab === 1 && (
            <AdminDataTab tables={tables} onSessionExpired={handleSessionExpired} />
          )}
          {tab === 2 && (
            <AdminBackupsTab
              backups={backups}
              backupDir={backupDir}
              onSessionExpired={handleSessionExpired}
            />
          )}
        </>
      ) : null}

      <AdminReauthDialog
        open={needsAuth}
        onAuthenticated={handleAuthenticated}
        onCancel={() => void handleCancel()}
      />
    </Box>
  );
}
