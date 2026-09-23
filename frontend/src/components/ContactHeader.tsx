import { Box, Card, CardContent, Avatar, Typography, Chip, IconButton, TextField, Button } from '@mui/material';
import EditIcon from '@mui/icons-material/Edit';
import SaveIcon from '@mui/icons-material/Save';
import CloseIcon from '@mui/icons-material/Close';
import DeleteIcon from '@mui/icons-material/Delete';
import CameraAltIcon from '@mui/icons-material/CameraAlt';
import ArchiveIcon from '@mui/icons-material/Archive';
import UnarchiveIcon from '@mui/icons-material/Unarchive';
import { useTranslation } from 'react-i18next';

export interface ProfileValues {
  firstname: string;
}

interface ContactHeaderProps {
  contact: {
    ID: number;
    firstname: string;
    archived?: boolean;
  };
  profilePic: string;
  editingProfile: boolean;
  profileValues: ProfileValues;
  onStartEditProfile: () => void;
  onCancelEditProfile: () => void;
  onSaveProfile: () => void;
  onDeleteContact: () => void;
  onProfileValueChange: (values: ProfileValues) => void;
  onUploadProfilePicture: () => void;
  onArchiveContact?: () => void;
  onUnarchiveContact?: () => void;
}

export default function ContactHeader({
  contact,
  profilePic,
  editingProfile,
  profileValues,
  onStartEditProfile,
  onCancelEditProfile,
  onSaveProfile,
  onDeleteContact,
  onProfileValueChange,
  onUploadProfilePicture,
  onArchiveContact,
  onUnarchiveContact
}: ContactHeaderProps) {
  const { t } = useTranslation();

  return (
    <Card sx={{ mb: 1.5 }}>
      <CardContent sx={{ py: 1.5, '&:last-child': { pb: 1.5 } }}>
        <Box sx={{ display: 'flex', alignItems: 'flex-start' }}>
          <Box
            sx={{
              position: 'relative',
              '&:hover .camera-badge': {
                opacity: 1
              }
            }}
          >
            <Avatar
              src={profilePic || undefined}
              sx={{
                width: 90,
                height: 90,
                cursor: 'pointer',
                bgcolor: 'primary.main',
                fontSize: '2rem',
                '&:hover': { opacity: 0.8 }
              }}
              onClick={onUploadProfilePicture}
            >
              {contact.firstname.charAt(0)}
            </Avatar>
            <IconButton
              className="camera-badge"
              size="small"
              onClick={onUploadProfilePicture}
              sx={{
                position: 'absolute',
                bottom: -4,
                right: -4,
                bgcolor: 'primary.main',
                color: 'white',
                width: 22,
                height: 22,
                opacity: 0,
                transition: 'opacity 0.2s',
                '&:hover': { bgcolor: 'primary.dark' }
              }}
            >
              <CameraAltIcon sx={{ fontSize: 14 }} />
            </IconButton>
          </Box>
          <Box sx={{ flex: 1, ml: 1.5 }}>
            {editingProfile ? (
              // Edit Mode
              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
                <TextField
                  label={t('client.denomination')}
                  value={profileValues.firstname}
                  onChange={(e) => onProfileValueChange({ firstname: e.target.value })}
                  size="small"
                  required
                  autoFocus
                />
                <Box sx={{ display: 'flex', gap: 1, justifyContent: 'space-between' }}>
                  <IconButton
                    size="small"
                    color="error"
                    onClick={onDeleteContact}
                    title={t('contactDetail.deleteContact')}
                  >
                    <DeleteIcon />
                  </IconButton>
                  <Box sx={{ display: 'flex', gap: 1 }}>
                    <IconButton size="small" color="primary" onClick={onSaveProfile}>
                      <SaveIcon />
                    </IconButton>
                    <IconButton size="small" onClick={onCancelEditProfile}>
                      <CloseIcon />
                    </IconButton>
                  </Box>
                </Box>
              </Box>
            ) : (
              // View Mode
              <>
                {contact.archived && (
                  <Chip
                    label={t('contactDetail.archivedBadge')}
                    color="warning"
                    size="small"
                    sx={{ mb: 1 }}
                  />
                )}
                <Box
                  sx={{
                    display: 'flex',
                    alignItems: 'flex-start',
                    justifyContent: 'space-between',
                    flexWrap: 'wrap',
                    gap: 0.5,
                    '&:hover .edit-icon': {
                      opacity: 1
                    }
                  }}
                >
                  <Box sx={{ display: 'flex', alignItems: 'center' }}>
                    <Typography variant="h5" sx={{ fontWeight: 500, lineHeight: 1.2 }}>
                      {contact.firstname}
                    </Typography>
                    <IconButton
                      className="edit-icon"
                      size="small"
                      onClick={onStartEditProfile}
                      sx={{
                        ml: 1,
                        opacity: 0,
                        transition: 'opacity 0.2s'
                      }}
                    >
                      <EditIcon fontSize="small" />
                    </IconButton>
                  </Box>
                  <Box sx={{ display: 'flex', gap: 1 }}>
                    {contact.archived ? (
                      onUnarchiveContact && (
                        <Button
                          variant="outlined"
                          size="small"
                          color="success"
                          startIcon={<UnarchiveIcon />}
                          onClick={onUnarchiveContact}
                        >
                          {t('contactDetail.unarchive')}
                        </Button>
                      )
                    ) : (
                      <>
                        {onArchiveContact && (
                          <Button
                            variant="outlined"
                            size="small"
                            color="warning"
                            startIcon={<ArchiveIcon />}
                            onClick={onArchiveContact}
                          >
                            {t('contactDetail.archive')}
                          </Button>
                        )}
                      </>
                    )}
                  </Box>
                </Box>
              </>
            )}
          </Box>
        </Box>
      </CardContent>
    </Card>
  );
}