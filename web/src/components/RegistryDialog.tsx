import { useState } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  CircularProgress,
  Alert,
  Typography,
  Box,
} from '@mui/material';
import { useRegistry, useSetRegistry } from '../hooks/useReleases';
import type { Release } from '../types';

interface RegistryDialogProps {
  open: boolean;
  onClose: () => void;
  release: Release;
}

export default function RegistryDialog({ open, onClose, release }: RegistryDialogProps) {
  const [registry, setRegistry] = useState<string | null>(null);
  const { data: currentMapping, isLoading: loadingRegistry, error: registryError } = useRegistry(
    release.namespace,
    release.name
  );
  const { mutate: saveRegistry, isPending: isSaving, error: saveError } = useSetRegistry();

  const displayValue = registry ?? currentMapping?.registry ?? '';

  const handleSave = () => {
    if (!displayValue.trim()) return;

    saveRegistry(
      {
        namespace: release.namespace,
        name: release.name,
        request: { registry: displayValue.trim() },
      },
      {
        onSuccess: () => {
          setRegistry(null);
          onClose();
        },
      }
    );
  };

  const handleClose = () => {
    setRegistry(null);
    onClose();
  };

  const isNotFound = registryError && (registryError as { response?: { status?: number } }).response?.status === 404;

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle>Set OCI Registry</DialogTitle>
      <DialogContent>
        <Box mb={2}>
          <Typography variant="body2" color="textSecondary">
            Release: {release.namespace}/{release.name}
          </Typography>
          <Typography variant="body2" color="textSecondary">
            Chart: {release.chart}
          </Typography>
        </Box>

        {registryError && !isNotFound && (
          <Alert severity="error" sx={{ mb: 2 }}>
            Failed to load registry: {(registryError as Error).message}
          </Alert>
        )}

        {isNotFound && (
          <Alert severity="info" sx={{ mb: 2 }}>
            No registry configured yet. Please set one below.
          </Alert>
        )}

        {saveError && (
          <Alert severity="error" sx={{ mb: 2 }}>
            Failed to save registry: {(saveError as Error).message}
          </Alert>
        )}

        {loadingRegistry ? (
          <Box display="flex" justifyContent="center" p={2}>
            <CircularProgress size={24} />
          </Box>
        ) : (
          <TextField
            fullWidth
            label="OCI Registry URL"
            placeholder="ghcr.io/myorg/charts"
            value={displayValue}
            onChange={(e) => setRegistry(e.target.value)}
            sx={{ mt: 2 }}
            helperText="Enter the OCI registry URL where the chart is stored (e.g., ghcr.io/myorg/charts)"
          />
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose} disabled={isSaving}>
          Cancel
        </Button>
        <Button
          onClick={handleSave}
          variant="contained"
          disabled={!displayValue.trim() || isSaving}
        >
          {isSaving ? <CircularProgress size={24} /> : 'Save'}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
