import { useState } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  CircularProgress,
  Alert,
  Typography,
  Box,
  IconButton,
  Stack,
} from '@mui/material';
import RefreshIcon from '@mui/icons-material/Refresh';
import type { SelectChangeEvent } from '@mui/material';
import { useAvailableVersions, useUpgradeRelease } from '../hooks/useReleases';
import type { Release } from '../types';

interface VersionDialogProps {
  open: boolean;
  onClose: () => void;
  release: Release;
}

export default function VersionDialog({ open, onClose, release }: VersionDialogProps) {
  const [selectedVersion, setSelectedVersion] = useState('');
  const { data: versions, isLoading: loadingVersions, error: versionsError, refetch: refetchVersions, isFetching: isFetchingVersions } = useAvailableVersions(
    release.namespace,
    release.name
  );
  const { mutate: upgrade, isPending: isUpgrading, error: upgradeError } = useUpgradeRelease();

  const handleVersionChange = (event: SelectChangeEvent) => {
    setSelectedVersion(event.target.value);
  };

  const handleUpgrade = () => {
    if (!selectedVersion) return;

    upgrade(
      {
        namespace: release.namespace,
        name: release.name,
        request: { chartVersion: selectedVersion },
      },
      {
        onSuccess: () => {
          setSelectedVersion('');
          onClose();
        },
      }
    );
  };

  const handleClose = () => {
    setSelectedVersion('');
    onClose();
  };

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle>Change Version</DialogTitle>
      <DialogContent>
        <Box mb={2}>
          <Typography variant="body2" color="textSecondary">
            Release: {release.namespace}/{release.name}
          </Typography>
          <Typography variant="body2" color="textSecondary">
            Current Version: {release.chartVersion}
          </Typography>
        </Box>

        {versionsError && (
          <Alert severity="error" sx={{ mb: 2 }}>
            Failed to load versions: {(versionsError as Error).message}
          </Alert>
        )}

        {upgradeError && (
          <Alert severity="error" sx={{ mb: 2 }}>
            Failed to upgrade: {(upgradeError as Error).message}
          </Alert>
        )}

        {loadingVersions ? (
          <Box display="flex" justifyContent="center" p={2}>
            <CircularProgress size={24} />
          </Box>
        ) : (
          <Stack direction="row" spacing={1} alignItems="flex-end" sx={{ mt: 2 }}>
            <FormControl fullWidth>
              <InputLabel>Target Version</InputLabel>
              <Select
                value={selectedVersion}
                label="Target Version"
                onChange={handleVersionChange}
              >
                {versions?.map((v) => (
                  <MenuItem key={v.version} value={v.version}>
                    {v.version} (App: {v.appVersion})
                    {v.version === release.chartVersion && ' - Current'}
                  </MenuItem>
                ))}
                {versions?.length === 0 && (
                  <MenuItem disabled>No versions available</MenuItem>
                )}
              </Select>
            </FormControl>
            <IconButton
              onClick={() => refetchVersions()}
              disabled={isFetchingVersions}
              title="Refresh versions"
              size="small"
              sx={{ mb: 0.5 }}
            >
              {isFetchingVersions ? <CircularProgress size={20} /> : <RefreshIcon />}
            </IconButton>
          </Stack>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose} disabled={isUpgrading}>
          Cancel
        </Button>
        <Button
          onClick={handleUpgrade}
          variant="contained"
          disabled={!selectedVersion || isUpgrading || selectedVersion === release.chartVersion}
        >
          {isUpgrading ? <CircularProgress size={24} /> : 'Upgrade'}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
