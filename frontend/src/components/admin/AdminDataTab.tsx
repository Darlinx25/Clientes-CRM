import { useCallback, useEffect, useState } from 'react';
import {
  Alert,
  Box,
  Card,
  CardContent,
  Checkbox,
  CircularProgress,
  FormControlLabel,
  MenuItem,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TablePagination,
  TableRow,
  TextField,
  Typography,
} from '@mui/material';
import { useTranslation } from 'react-i18next';
import { ApiError } from '../../api/client';
import {
  getAdminPanelTable,
  isAdminSessionRequiredError,
} from '../../api/adminPanel';
import type {
  AdminPanelTableRows,
  AdminPanelTableSummary,
} from '../../types/adminPanel';

interface AdminDataTabProps {
  tables: AdminPanelTableSummary[];
  onSessionExpired: () => void;
}

function renderCell(value: unknown, yes: string, no: string, empty: string) {
  if (value === null || value === undefined || value === '') {
    return empty;
  }
  if (typeof value === 'boolean') {
    return value ? yes : no;
  }
  if (typeof value === 'object') {
    return JSON.stringify(value);
  }
  return String(value);
}

export default function AdminDataTab({ tables, onSessionExpired }: AdminDataTabProps) {
  const { t } = useTranslation();
  const [selectedTable, setSelectedTable] = useState('');
  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [includeDeleted, setIncludeDeleted] = useState(false);
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(25);
  const [data, setData] = useState<AdminPanelTableRows | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!selectedTable && tables.length > 0) {
      setSelectedTable(tables[0].name);
    }
  }, [selectedTable, tables]);

  useEffect(() => {
    const id = window.setTimeout(() => setDebouncedSearch(search), 300);
    return () => window.clearTimeout(id);
  }, [search]);

  const loadRows = useCallback(async () => {
    if (!selectedTable) {
      return;
    }
    setLoading(true);
    setError('');
    try {
      const result = await getAdminPanelTable(selectedTable, {
        page: page + 1,
        limit: rowsPerPage,
        search: debouncedSearch,
        includeDeleted,
      });
      setData(result);
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        if (isAdminSessionRequiredError(err)) {
          onSessionExpired();
        } else {
          localStorage.removeItem('user_info');
          window.location.href = '/login';
        }
      } else {
        setError(err instanceof ApiError ? err.getDisplayMessage() : t('admin.errors.load'));
      }
    } finally {
      setLoading(false);
    }
  }, [debouncedSearch, includeDeleted, onSessionExpired, page, rowsPerPage, selectedTable, t]);

  useEffect(() => {
    void loadRows();
  }, [loadRows]);

  return (
    <Card>
      <CardContent>
        <Stack
          direction={{ xs: 'column', md: 'row' }}
          spacing={2}
          alignItems={{ xs: 'stretch', md: 'center' }}
        >
          <TextField
            select
            size="small"
            label={t('admin.data.table')}
            value={selectedTable}
            onChange={(event) => {
              setSelectedTable(event.target.value);
              setPage(0);
            }}
            sx={{ minWidth: 240 }}
          >
            {tables.map((table) => (
              <MenuItem key={table.name} value={table.name}>
                {`${table.label} (${table.row_count})`}
              </MenuItem>
            ))}
          </TextField>
          <TextField
            size="small"
            fullWidth
            label={t('admin.data.search')}
            value={search}
            onChange={(event) => {
              setSearch(event.target.value);
              setPage(0);
            }}
          />
          <FormControlLabel
            control={
              <Checkbox
                checked={includeDeleted}
                onChange={(event) => {
                  setIncludeDeleted(event.target.checked);
                  setPage(0);
                }}
              />
            }
            label={t('admin.data.includeDeleted')}
          />
        </Stack>

        {error && (
          <Alert severity="error" sx={{ mt: 2 }}>
            {error}
          </Alert>
        )}

        {loading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', py: 6 }}>
            <CircularProgress />
          </Box>
        ) : data ? (
          <>
            <TableContainer sx={{ mt: 2, maxHeight: 520 }}>
              <Table stickyHeader size="small">
                <TableHead>
                  <TableRow>
                    {data.columns.map((column) => (
                      <TableCell key={column}>{column}</TableCell>
                    ))}
                  </TableRow>
                </TableHead>
                <TableBody>
                  {data.rows.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={data.columns.length || 1}>
                        <Typography variant="body2" color="text.secondary">
                          {t('admin.data.empty')}
                        </Typography>
                      </TableCell>
                    </TableRow>
                  ) : (
                    data.rows.map((row, index) => (
                      <TableRow key={index}>
                        {data.columns.map((column) => (
                          <TableCell key={column}>
                            {renderCell(
                              row[column],
                              t('admin.data.yes'),
                              t('admin.data.no'),
                              t('admin.data.emptyValue')
                            )}
                          </TableCell>
                        ))}
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </TableContainer>
            <TablePagination
              component="div"
              count={data.total}
              page={page}
              onPageChange={(_, newPage) => setPage(newPage)}
              rowsPerPage={rowsPerPage}
              onRowsPerPageChange={(event) => {
                setRowsPerPage(parseInt(event.target.value, 10));
                setPage(0);
              }}
              rowsPerPageOptions={[25, 50, 100]}
              labelRowsPerPage={t('admin.data.rowsPerPage')}
            />
          </>
        ) : null}
      </CardContent>
    </Card>
  );
}
