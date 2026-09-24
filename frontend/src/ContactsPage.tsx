import { useEffect, useState, useMemo, useRef, useCallback } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useContacts } from './hooks/useContacts';
import { getCurrentUser } from './api/admin';
import { resolveEnabledFields, ContactFieldKey } from './contactFields';
import AddContactDialog from './components/AddContactDialog';
import ExcelImportDialog from './components/ExcelImportDialog';
import {
  Box,
  Card,
  Avatar,
  Typography,
  Chip,
  Stack,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  Pagination,
  Button,
  FormControlLabel,
  Switch,
  TextField,
  InputAdornment,
} from '@mui/material';
import PersonAddIcon from '@mui/icons-material/PersonAdd';
import GridOnIcon from '@mui/icons-material/GridOn';
import SearchIcon from '@mui/icons-material/Search';
import ClearIcon from '@mui/icons-material/Clear';
import { ContactListSkeleton } from './components/LoadingSkeletons';

export default function ContactsPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const searchQuery = searchParams.get('search') || '';
  const page = parseInt(searchParams.get('page') || '1', 10);
  const [sortOption, setSortOption] = useState(() => {
    return localStorage.getItem('contacts-sort-option') || 'id-desc';
  });
  const [addDialogOpen, setAddDialogOpen] = useState(false);
  const [excelImportDialogOpen, setExcelImportDialogOpen] = useState(false);
  const [customFieldNames, setCustomFieldNames] = useState<string[]>([]);
  const [enabledFields, setEnabledFields] = useState<Set<ContactFieldKey>>(() => resolveEnabledFields(null));
  const [showArchived, setShowArchived] = useState(false);
  const [searchInput, setSearchInput] = useState(searchQuery);
  const pageSize = 10;

  // Keep the local input in sync when the URL query changes externally
  useEffect(() => {
    setSearchInput(searchQuery);
  }, [searchQuery]);

  // Debounced: push the search term into the URL (?search=...) so useContacts refetches
  useEffect(() => {
    if (searchInput === searchQuery) return;
    const timer = setTimeout(() => {
      setSearchParams(prev => {
        const params = new URLSearchParams(prev);
        if (searchInput.trim()) {
          params.set('search', searchInput.trim());
        } else {
          params.delete('search');
        }
        params.delete('page');
        return params;
      });
    }, 300);
    return () => clearTimeout(timer);
  }, [searchInput, searchQuery, setSearchParams]);

  // Parse sort option into field and order
  const [sortField, sortOrder] = sortOption.split('-');

  // Persist sort option to localStorage
  useEffect(() => {
    localStorage.setItem('contacts-sort-option', sortOption);
  }, [sortOption]);

  // Helper to update page in URL
  const setPage = useCallback((newPage: number) => {
    setSearchParams(prev => {
      const params = new URLSearchParams(prev);
      if (newPage === 1) {
        params.delete('page');
      } else {
        params.set('page', String(newPage));
      }
      return params;
    });
  }, [setSearchParams]);

  // Memoize params to prevent infinite re-renders
  const contactParams = useMemo(() => ({
    page,
    limit: pageSize,
    search: searchQuery,
    sort: sortField,
    order: sortOrder,
    includeArchived: showArchived,
  }), [page, searchQuery, sortField, sortOrder, showArchived]);

  // Use custom hook for fetching contacts
  const { contacts, total: totalContacts, loading, refetch } = useContacts(contactParams);

  // Fetch custom field names + enabled fields for the list
  useEffect(() => {
    const fetchData = async () => {
      try {
        const user = await getCurrentUser();
        setCustomFieldNames(user.custom_field_names ?? []);
        setEnabledFields(resolveEnabledFields(user.enabled_contact_fields));
      } catch (err) {
        console.error('Error fetching custom field names:', err);
      }
    };
    fetchData();
  }, []);

  // Reset to page 1 when search changes (but not on initial mount)
  const prevFiltersRef = useRef({ searchQuery });
  useEffect(() => {
    const prev = prevFiltersRef.current;
    if (prev.searchQuery !== searchQuery) {
      setPage(1);
      prevFiltersRef.current = { searchQuery };
    }
  }, [searchQuery, setPage]);

  // With backend pagination, contacts are already filtered
  const filteredContacts = contacts;

  const handleContactAdded = (contactId: number) => {
    navigate(`/contacts/${contactId}`);
  };

  const handleImportComplete = async () => {
    await refetch();
  };
  
  return (
    <Box sx={{ maxWidth: 1200, mx: 'auto', mt: 2, p: 2 }}>
      <Typography variant="h5" gutterBottom sx={{ mb: 2 }}>
        {t('contacts.title')}
      </Typography>
      <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5} mb={2} alignItems="center">
        <FormControl sx={{ minWidth: 150 }} size="small">
          <InputLabel id="sort-select-label">{t('contacts.sortBy')}</InputLabel>
          <Select
            labelId="sort-select-label"
            value={sortOption}
            label={t('contacts.sortBy')}
            onChange={e => setSortOption(e.target.value)}
          >
            <MenuItem value="id-desc">{t('contacts.sort.recentlyAdded')}</MenuItem>
            <MenuItem value="id-asc">{t('contacts.sort.oldestFirst')}</MenuItem>
            <MenuItem value="firstname-asc">{t('contacts.sort.nameAZ')}</MenuItem>
            <MenuItem value="firstname-desc">{t('contacts.sort.nameZA')}</MenuItem>
            <MenuItem value="random-asc">{t('contacts.sort.random')}</MenuItem>
          </Select>
        </FormControl>
        <FormControlLabel
          control={
            <Switch
              checked={showArchived}
              onChange={(e) => setShowArchived(e.target.checked)}
              size="small"
            />
          }
          label={t('contacts.showArchived')}
          sx={{ ml: 0.5, whiteSpace: 'nowrap' }}
        />
        <Button
          variant="outlined"
          startIcon={<GridOnIcon />}
          onClick={() => setExcelImportDialogOpen(true)}
          sx={{ whiteSpace: 'nowrap' }}
        >
          {t('contacts.importExcel.button', 'Importar')}
        </Button>
        <Button
          variant="outlined"
          startIcon={<PersonAddIcon />}
          onClick={() => setAddDialogOpen(true)}
          sx={{ whiteSpace: 'nowrap' }}
        >
          {t('contacts.add.button')}
        </Button>
      </Stack>
      <Box sx={{ mb: 2, p: 1.5, bgcolor: 'action.hover', borderRadius: 1, display: 'flex', alignItems: 'center', gap: 1, flexWrap: 'wrap' }}>
        <TextField
          size="small"
          fullWidth
          placeholder={t('contacts.search')}
          value={searchInput}
          onChange={(e) => setSearchInput(e.target.value)}
          slotProps={{
            input: {
              startAdornment: (
                <InputAdornment position="start">
                  <SearchIcon />
                </InputAdornment>
              ),
              endAdornment: searchInput ? (
                <InputAdornment position="end">
                  <ClearIcon
                    fontSize="small"
                    sx={{ cursor: 'pointer', color: 'text.secondary' }}
                    onClick={() => setSearchInput('')}
                  />
                </InputAdornment>
              ) : null,
            },
          }}
        />
      </Box>
      {loading ? (
        <ContactListSkeleton count={10} />
      ) : (
        <>
          <Stack spacing={2}>
            {filteredContacts.map(contact => (
              <Card
                key={contact.ID}
                component={Link}
                to={`/contacts/${contact.ID}`}
                sx={{
                  display: 'flex',
                  alignItems: 'center',
                  p: 1.5,
                  cursor: 'pointer',
                  textDecoration: 'none',
                  color: 'inherit',
                  bgcolor: contact.archived ? 'action.disabledBackground' : undefined,
                  '&:hover': {
                    bgcolor: 'action.hover'
                  }
                }}
              >
                <Avatar src={contact.photo_thumbnail || undefined} sx={{ width: 48, height: 48, mr: 1.5, bgcolor: 'primary.main' }}>
                  {contact.firstname.charAt(0)}
                </Avatar>
                <Box sx={{ flex: 1 }}>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <Typography variant="body1" sx={{ fontWeight: 500 }}>
                      {[contact.firstname, contact.nickname && `"${contact.nickname}"`, contact.lastname].filter(Boolean).join(' ')}
                    </Typography>
                    {contact.archived && (
                      <Chip
                        label={t('contacts.archived')}
                        size="small"
                        color="default"
                        sx={{ height: 20, fontSize: '0.7rem' }}
                      />
                    )}
                  </Box>
                  {contact.rut && (
                    <Typography variant="body2" color="text.secondary" sx={{ fontSize: '0.8rem' }}>
                      {t('client.rut', 'RUT')}: {contact.rut}
                    </Typography>
                  )}
                </Box>
              </Card>
            ))}
          </Stack>
          {totalContacts > 0 && (
            <Box sx={{ display: 'flex', justifyContent: 'center', mt: 2 }}>
              <Pagination
                count={Math.max(1, Math.ceil(totalContacts / pageSize))}
                page={page}
                onChange={(_, value) => setPage(value)}
                color="primary"
                size="large"
              />
            </Box>
          )}
        </>
      )}
      <AddContactDialog
        open={addDialogOpen}
        onClose={() => setAddDialogOpen(false)}
        onContactAdded={handleContactAdded}
        customFieldNames={customFieldNames}
        enabledFields={enabledFields}
      />
      <ExcelImportDialog
        open={excelImportDialogOpen}
        onClose={() => setExcelImportDialogOpen(false)}
        onImportComplete={handleImportComplete}
      />
    </Box>
  );
}
