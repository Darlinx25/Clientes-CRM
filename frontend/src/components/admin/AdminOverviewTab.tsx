import {
  Alert,
  Box,
  Card,
  CardContent,
  Chip,
  Divider,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Typography,
} from '@mui/material';
import { useTranslation } from 'react-i18next';
import type { AdminPanelOverview } from '../../types/adminPanel';

interface AdminOverviewTabProps {
  overview: AdminPanelOverview;
}

function formatBytes(bytes: number): string {
  if (!bytes) {
    return '0 B';
  }
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const index = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
  return `${(bytes / Math.pow(1024, index)).toFixed(index === 0 ? 0 : 1)} ${units[index]}`;
}

function formatDate(value: string | null, fallback: string): string {
  if (!value) {
    return fallback;
  }
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? fallback : date.toLocaleString();
}

function InfoItem({ label, value }: { label: string; value: string }) {
  return (
    <Box>
      <Typography variant="caption" color="text.secondary">
        {label}
      </Typography>
      <Typography variant="body2" sx={{ wordBreak: 'break-all' }}>
        {value}
      </Typography>
    </Box>
  );
}

export default function AdminOverviewTab({ overview }: AdminOverviewTabProps) {
  const { t } = useTranslation();
  const notAvailable = t('admin.notAvailable');

  const schemaStatus = overview.schema.dirty
    ? t('admin.overview.dirty')
    : overview.schema.up_to_date
      ? t('admin.overview.upToDate')
      : t('admin.overview.pending');

  const counts = [
    { label: t('admin.overview.users'), value: overview.counts.users },
    { label: t('admin.overview.contacts'), value: overview.counts.contacts },
    { label: t('admin.overview.archivedContacts'), value: overview.counts.archived_contacts },
    { label: t('admin.overview.deletedContacts'), value: overview.counts.deleted_contacts },
    { label: t('admin.overview.notes'), value: overview.counts.notes },
    { label: t('admin.overview.deletedNotes'), value: overview.counts.deleted_notes },
    { label: t('admin.overview.activities'), value: overview.counts.activities },
    { label: t('admin.overview.reminders'), value: overview.counts.reminders },
    { label: t('admin.overview.reminderCompletions'), value: overview.counts.reminder_completions },
    { label: t('admin.overview.companies'), value: overview.counts.companies },
    { label: t('admin.overview.relationships'), value: overview.counts.relationships },
  ];

  return (
    <Stack spacing={2}>
      <Card>
        <CardContent>
          <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 1 }}>
            <Typography variant="h6">{t('admin.overview.system')}</Typography>
            <Chip size="small" color="primary" label={`v${overview.app_version}`} />
          </Box>
          <Divider sx={{ my: 1.5 }} />
          <Box
            sx={{
              display: 'grid',
              gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr' },
              gap: 2,
            }}
          >
            <InfoItem
              label={t('admin.overview.schema')}
              value={`${overview.schema.current} / ${overview.schema.expected}`}
            />
            <InfoItem label={t('admin.overview.schemaStatus')} value={schemaStatus} />
            <InfoItem
              label={t('admin.overview.database')}
              value={formatBytes(overview.db_size_bytes)}
            />
            <InfoItem label={t('admin.overview.wal')} value={formatBytes(overview.wal_size_bytes)} />
            <InfoItem label={t('admin.overview.backupDir')} value={overview.backup_dir || notAvailable} />
            <InfoItem
              label={t('admin.overview.latestBackup')}
              value={formatDate(
                overview.backups.latest_created_at,
                overview.backups.latest_filename || t('admin.overview.noBackups')
              )}
            />
          </Box>
          {overview.schema.dirty && (
            <Alert severity="warning" sx={{ mt: 2 }}>
              {t('admin.overview.dirtyWarning')}
            </Alert>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardContent>
          <Typography variant="h6">{t('admin.overview.counts')}</Typography>
          <Divider sx={{ my: 1.5 }} />
          <Box
            sx={{
              display: 'grid',
              gridTemplateColumns: { xs: '1fr 1fr', md: '1fr 1fr 1fr' },
              gap: 1.5,
            }}
          >
            {counts.map((item) => (
              <InfoItem key={item.label} label={item.label} value={String(item.value)} />
            ))}
          </Box>
        </CardContent>
      </Card>

      <Card>
        <CardContent>
          <Typography variant="h6">{t('admin.overview.jobs')}</Typography>
          <Divider sx={{ my: 1.5 }} />
          {overview.jobs.length === 0 ? (
            <Typography variant="body2" color="text.secondary">
              {t('admin.overview.noJobs')}
            </Typography>
          ) : (
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>{t('admin.overview.jobs')}</TableCell>
                  <TableCell>{t('admin.overview.lastRun')}</TableCell>
                  <TableCell>{t('admin.overview.locked')}</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {overview.jobs.map((job) => (
                  <TableRow key={job.job_name}>
                    <TableCell sx={{ fontFamily: 'monospace' }}>{job.job_name}</TableCell>
                    <TableCell>{formatDate(job.last_run_at, notAvailable)}</TableCell>
                    <TableCell>
                      {job.locked_at
                        ? `${formatDate(job.locked_at, notAvailable)}${job.locked_by ? ` · ${job.locked_by}` : ''}`
                        : t('admin.overview.notLocked')}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </Stack>
  );
}
