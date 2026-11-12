import { useState } from 'react';
import {
  Alert,
  Box,
  Card,
  CardContent,
  Chip,
  Divider,
  IconButton,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Tooltip,
  Typography,
} from '@mui/material';
import DownloadIcon from '@mui/icons-material/Download';
import { useTranslation } from 'react-i18next';
import { ApiError } from '../../api/client';
import {
  downloadAdminPanelBackup,
  isAdminSessionRequiredError,
} from '../../api/adminPanel';
import type { AdminPanelBackup } from '../../types/adminPanel';

interface AdminBackupsTabProps {
  backups: AdminPanelBackup[];
  backupDir: string;
  onSessionExpired: () => void;
}

function formatBytes(bytes: number): string {
  if (!bytes) {
    return '0 B';
  }
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const index = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
  return `${(bytes / Math.pow(1024, index)).toFixed(index === 0 ? 0 : 1)} ${units[index]}`;
}

function formatDate(value: string, fallback: string): string {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? fallback : date.toLocaleString();
}

export default function AdminBackupsTab({
  backups,
  backupDir,
  onSessionExpired,
}: AdminBackupsTabProps) {
  const { t } = useTranslation();
  const [downloading, setDownloading] = useState('');
  const [error, setError] = useState('');

  const handleDownload = async (filename: string) => {
    setDownloading(filename);
    setError('');
    try {
      const blob = await downloadAdminPanelBackup(filename);
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = filename;
      document.body.appendChild(link);
      link.click();
      link.remove();
      URL.revokeObjectURL(url);
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        if (isAdminSessionRequiredError(err)) {
          onSessionExpired();
        } else {
          localStorage.removeItem('user_info');
          window.location.href = '/login';
        }
      } else {
        setError(err instanceof ApiError ? err.getDisplayMessage() : t('admin.errors.download'));
      }
    } finally {
      setDownloading('');
    }
  };

  return (
    <Card>
      <CardContent>
        <Typography variant="h6">{t('admin.backups.title')}</Typography>
        <Divider sx={{ my: 1.5 }} />
        <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
          {t('admin.backups.description')}
        </Typography>
        <Typography variant="caption" color="text.secondary" sx={{ wordBreak: 'break-all' }}>
          {`${t('admin.backups.folder')}: ${backupDir}`}
        </Typography>

        {error && (
          <Alert severity="error" sx={{ mt: 2 }}>
            {error}
          </Alert>
        )}

        <Box sx={{ mt: 2 }}>
          {backups.length === 0 ? (
            <Alert severity="info">{t('admin.backups.empty')}</Alert>
          ) : (
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>{t('admin.backups.filename')}</TableCell>
                  <TableCell>{t('admin.backups.kind')}</TableCell>
                  <TableCell>{t('admin.backups.size')}</TableCell>
                  <TableCell>{t('admin.backups.created')}</TableCell>
                  <TableCell align="right">{t('admin.backups.download')}</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {backups.map((backup) => (
                  <TableRow key={backup.filename}>
                    <TableCell sx={{ fontFamily: 'monospace', wordBreak: 'break-all' }}>
                      {backup.filename}
                    </TableCell>
                    <TableCell>
                      <Chip
                        size="small"
                        label={
                          backup.kind === 'pre_upgrade'
                            ? t('admin.backups.kinds.preUpgrade')
                            : t('admin.backups.kinds.backup')
                        }
                        color={backup.kind === 'pre_upgrade' ? 'warning' : 'default'}
                      />
                    </TableCell>
                    <TableCell>{formatBytes(backup.size_bytes)}</TableCell>
                    <TableCell>{formatDate(backup.created_at, t('admin.notAvailable'))}</TableCell>
                    <TableCell align="right">
                      <Tooltip title={t('admin.backups.download')}>
                        <span>
                          <IconButton
                            size="small"
                            onClick={() => void handleDownload(backup.filename)}
                            disabled={downloading === backup.filename}
                          >
                            <DownloadIcon fontSize="small" />
                          </IconButton>
                        </span>
                      </Tooltip>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </Box>
      </CardContent>
    </Card>
  );
}
