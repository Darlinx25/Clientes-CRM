import { useState, useEffect, Suspense, useMemo } from 'react';
import ContactsPage from './ContactsPage';
import ContactDetailPage from './ContactDetailPage';
import TimelinePage from './TimelinePage';
import SettingsPage from './SettingsPage';
import LoginPage from './LoginPage';
import RegisterPage from './RegisterPage';
import { getToken, logoutAndRedirect, fetchAndCacheUserInfo } from './auth';
import { BrowserRouter as Router, Routes, Route, Link, useLocation, Navigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  Box,
  Button,
  CircularProgress,
  Typography,
  IconButton,
  Tooltip,
} from '@mui/material';
import ContactsIcon from '@mui/icons-material/Contacts';
import TimelineIcon from '@mui/icons-material/Timeline';
import SettingsIcon from '@mui/icons-material/Settings';
import LogoutIcon from '@mui/icons-material/Logout';
import './App.css';

// Scroll to top on route change
function ScrollToTop() {
  const { pathname } = useLocation();

  useEffect(() => {
    window.scrollTo(0, 0);
  }, [pathname]);

  return null;
}

// Inner component that can use useLocation (must be inside Router)
function AppContent({ token, setToken }: { token: string | null; setToken: (token: string | null) => void }) {
  const { t } = useTranslation();
  const location = useLocation();

  const handleLogout = async () => {
    await logoutAndRedirect();
  };

  // eslint-disable-next-line react-hooks/exhaustive-deps -- token changes trigger admin status recalculation
  const mainNavItems = useMemo(() => [
    { text: t('nav.contacts'), icon: <ContactsIcon />, path: '/contacts' },
    { text: t('nav.timeline'), icon: <TimelineIcon />, path: '/timeline' },
  ], [t]);

  // Check if current path matches the nav item (handle exact match for "/" and prefix match for others)
  const isActiveRoute = (path: string) => {
    if (path === '/') {
      return location.pathname === '/';
    }
    return location.pathname.startsWith(path);
  };

  if (!token) {
    return (
      <Box sx={{ p: 2, width: '100%' }}>
        <Routes>
          <Route path="/register" element={<RegisterPage />} />
          <Route path="*" element={<LoginPage setToken={setToken} />} />
        </Routes>
      </Box>
    );
  }

  return (
    <>
      {/* Rigid lateral sidebar */}
      <Box
        component="nav"
        sx={{
          width: { xs: 200, sm: 220 },
          flexShrink: 0,
          position: 'sticky',
          top: 0,
          height: '100vh',
          bgcolor: 'background.paper',
          borderRight: 1,
          borderColor: 'divider',
          display: 'flex',
          flexDirection: 'column',
          p: 1.5,
        }}
      >
        <Typography
          variant="h6"
          component={Link}
          to="/contacts"
          sx={{
            textDecoration: 'none',
            color: 'inherit',
            px: 1,
            mb: 2,
            '&:hover': { opacity: 0.8 }
          }}
        >
          {t('app.title')}
        </Typography>

        <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'stretch', gap: 0.5 }}>
          {mainNavItems.map((item) => (
            <Button
              key={item.path}
              component={Link}
              to={item.path}
              startIcon={item.icon}
              color={isActiveRoute(item.path) ? 'primary' : 'inherit'}
              variant={isActiveRoute(item.path) ? 'contained' : 'text'}
              sx={{ justifyContent: 'flex-start', textAlign: 'left' }}
            >
              {item.text}
            </Button>
          ))}
        </Box>

        <Box sx={{ flexGrow: 1 }} />

        <Box
          sx={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'flex-start',
            gap: 0.5,
          }}
        >
          <Tooltip title={t('nav.settings')}>
            <IconButton
              component={Link}
              to="/settings"
              color={isActiveRoute('/settings') ? 'primary' : 'default'}
              size="large"
            >
              <SettingsIcon />
            </IconButton>
          </Tooltip>
          <Tooltip title={t('app.logout')}>
            <IconButton onClick={handleLogout} size="large">
              <LogoutIcon />
            </IconButton>
          </Tooltip>
        </Box>
      </Box>

      <Box component="main" sx={{ flexGrow: 1, minWidth: 0, p: 2 }}>
        <Routes>
          <Route path="/contacts" element={<Suspense fallback={<Box display="flex" justifyContent="center" mt={4}><CircularProgress /></Box>}><ContactsPage /></Suspense>} />
          <Route path="/contacts/:id" element={<Suspense fallback={<Box display="flex" justifyContent="center" mt={4}><CircularProgress /></Box>}><ContactDetailPage /></Suspense>} />
          <Route path="/timeline" element={<Suspense fallback={<Box display="flex" justifyContent="center" mt={4}><CircularProgress /></Box>}><TimelinePage /></Suspense>} />
          <Route path="/settings" element={<Suspense fallback={<Box display="flex" justifyContent="center" mt={4}><CircularProgress /></Box>}><SettingsPage /></Suspense>} />
          <Route path="/" element={<Navigate to="/contacts" replace />} />
          <Route path="/login" element={<Navigate to="/contacts" replace />} />
          <Route path="/register" element={<Navigate to="/contacts" replace />} />
          <Route path="*" element={<Navigate to="/contacts" replace />} />
        </Routes>
      </Box>
    </>
  );
}

function App() {
  const [token, setToken] = useState(getToken());

  useEffect(() => {
    // Restore session after OIDC redirect: the server sets the auth cookie but
    // localStorage is empty, so we fetch user info once to populate it.
    if (!getToken()) {
      fetchAndCacheUserInfo().then(info => {
        if (info) setToken(getToken());
      });
    }
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    // Listen for storage changes (e.g., logout in another tab)
    const onStorage = (e: StorageEvent) => {
      if (e.key === 'user_info') {
        setToken(getToken());
      }
    };
    window.addEventListener('storage', onStorage);
    return () => window.removeEventListener('storage', onStorage);
  }, []);

  return (
    <Router>
      <ScrollToTop />
      <Box sx={{ display: 'flex', minHeight: '100vh' }}>
        <AppContent token={token} setToken={setToken} />
      </Box>
    </Router>
  );
}

export default App;