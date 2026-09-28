import { useTranslation } from 'react-i18next';
import { Paper, Box, TextField, Button, SxProps, Theme } from '@mui/material';
import { useDateFormat } from '../DateFormatProvider';

interface NotesFilterBarProps {
  search: string;
  onSearchChange: (value: string) => void;
  fromDate: string;
  toDate: string;
  onFromDateChange: (value: string) => void;
  onToDateChange: (value: string) => void;
  compact?: boolean;
  sx?: SxProps<Theme>;
}

export default function NotesFilterBar({
  search,
  onSearchChange,
  fromDate,
  toDate,
  onFromDateChange,
  onToDateChange,
  compact,
  sx,
}: NotesFilterBarProps) {
  const { t } = useTranslation();
  const { getDatePlaceholder } = useDateFormat();

  return (
    <Paper sx={[{ p: 1.5, mb: 2 }, ...(sx ? (Array.isArray(sx) ? sx : [sx]) : [])]}>
      <Box display="flex" gap={1.5} flexWrap="wrap" alignItems="center">
        <TextField
          size="small"
          label={t('timelinePage.search')}
          value={search}
          onChange={(e) => onSearchChange(e.target.value)}
          variant="outlined"
          sx={
            compact
              ? { flex: '1 1 140px', minWidth: 120 }
              : { flex: '1 1 200px', minWidth: 0 }
          }
        />
        <TextField
          size="small"
          label={t('timelinePage.fromDate')}
          type="date"
          value={fromDate}
          onChange={(e) => onFromDateChange(e.target.value)}
          variant="outlined"
          slotProps={{ inputLabel: { shrink: true }, input: { placeholder: getDatePlaceholder() } }}
          sx={
            compact
              ? { flex: '0 1 130px', minWidth: 0, maxWidth: 160 }
              : { flex: '1 1 140px', minWidth: 0, maxWidth: 220 }
          }
        />
        <TextField
          size="small"
          label={t('timelinePage.toDate')}
          type="date"
          value={toDate}
          onChange={(e) => onToDateChange(e.target.value)}
          variant="outlined"
          slotProps={{ inputLabel: { shrink: true }, input: { placeholder: getDatePlaceholder() } }}
          sx={
            compact
              ? { flex: '0 1 130px', minWidth: 0, maxWidth: 160 }
              : { flex: '1 1 140px', minWidth: 0, maxWidth: 220 }
          }
        />
        {(fromDate || toDate) && (
          <Button
            size="small"
            onClick={() => { onFromDateChange(''); onToDateChange(''); }}
            sx={{ alignSelf: 'center', color: 'text.secondary', textTransform: 'none', whiteSpace: 'nowrap' }}
          >
            {t('timelinePage.clearDates')}
          </Button>
        )}
      </Box>
    </Paper>
  );
}