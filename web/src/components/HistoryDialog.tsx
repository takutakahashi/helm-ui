import { useState } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  CircularProgress,
  Alert,
  Box,
  Chip,
  IconButton,
  Tooltip,
  Typography,
} from '@mui/material';
import { Restore as RestoreIcon } from '@mui/icons-material';
import { useReleaseHistory } from '../hooks/useReleases';
import { rollbackRelease } from '../api/client';
import type { Release } from '../types';

interface HistoryDialogProps {
  open: boolean;
  onClose: () => void;
  release: Release;
  onRollbackSuccess?: () => void;
}

export default function HistoryDialog({ open, onClose, release, onRollbackSuccess }: HistoryDialogProps) {
  const { data: history, isLoading, error } = useReleaseHistory(
    release.namespace,
    release.name
  );
  const [rollbackConfirm, setRollbackConfirm] = useState<{
    open: boolean;
    revision: number;
    chart: string;
    appVersion: string;
  } | null>(null);
  const [rollbackLoading, setRollbackLoading] = useState(false);
  const [rollbackError, setRollbackError] = useState<string | null>(null);

  const handleRollbackClick = (revision: number, chart: string, appVersion: string) => {
    setRollbackConfirm({ open: true, revision, chart, appVersion });
  };

  const handleRollbackConfirm = async () => {
    if (!rollbackConfirm) return;

    setRollbackLoading(true);
    setRollbackError(null);

    try {
      await rollbackRelease(release.namespace, release.name, rollbackConfirm.revision);
      setRollbackConfirm(null);
      onClose();
      onRollbackSuccess?.();
    } catch (err) {
      setRollbackError((err as Error).message);
    } finally {
      setRollbackLoading(false);
    }
  };

  const handleRollbackCancel = () => {
    setRollbackConfirm(null);
    setRollbackError(null);
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
      <DialogTitle>
        Release History - {release.namespace}/{release.name}
      </DialogTitle>
      <DialogContent>
        {error && (
          <Alert severity="error" sx={{ mb: 2 }}>
            Failed to load history: {(error as Error).message}
          </Alert>
        )}

        {isLoading ? (
          <Box display="flex" justifyContent="center" p={4}>
            <CircularProgress />
          </Box>
        ) : (
          <TableContainer component={Paper} variant="outlined">
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Revision</TableCell>
                  <TableCell>Updated</TableCell>
                  <TableCell>Status</TableCell>
                  <TableCell>Chart Version</TableCell>
                  <TableCell>App Version</TableCell>
                  <TableCell>Description</TableCell>
                  <TableCell align="center">Actions</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {history?.map((h) => (
                  <TableRow
                    key={h.revision}
                    sx={{
                      backgroundColor:
                        h.revision === release.revision ? 'action.selected' : undefined,
                    }}
                  >
                    <TableCell>{h.revision}</TableCell>
                    <TableCell>{new Date(h.updated).toLocaleString()}</TableCell>
                    <TableCell>
                      <Chip label={h.status} size="small" />
                    </TableCell>
                    <TableCell>{h.chart}</TableCell>
                    <TableCell>{h.appVersion}</TableCell>
                    <TableCell>{h.description}</TableCell>
                    <TableCell align="center">
                      {h.revision !== release.revision && (
                        <Tooltip title="Rollback to this revision">
                          <IconButton
                            size="small"
                            color="primary"
                            onClick={() => handleRollbackClick(h.revision, h.chart, h.appVersion)}
                          >
                            <RestoreIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                      )}
                      {h.revision === release.revision && (
                        <Chip label="Current" size="small" color="primary" />
                      )}
                    </TableCell>
                  </TableRow>
                ))}
                {history?.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={7} align="center">
                      No history found
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </TableContainer>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Close</Button>
      </DialogActions>

      {/* Rollback Confirmation Dialog */}
      <Dialog
        open={rollbackConfirm?.open || false}
        onClose={handleRollbackCancel}
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle>Confirm Rollback</DialogTitle>
        <DialogContent>
          {rollbackError && (
            <Alert severity="error" sx={{ mb: 2 }}>
              Failed to rollback: {rollbackError}
            </Alert>
          )}
          <Typography variant="body1" gutterBottom>
            Are you sure you want to rollback <strong>{release.namespace}/{release.name}</strong> to revision{' '}
            <strong>{rollbackConfirm?.revision}</strong>?
          </Typography>
          <Box sx={{ mt: 2 }}>
            <Typography variant="body2" color="text.secondary">
              Chart Version: <strong>{rollbackConfirm?.chart}</strong>
            </Typography>
            <Typography variant="body2" color="text.secondary">
              App Version: <strong>{rollbackConfirm?.appVersion}</strong>
            </Typography>
          </Box>
          <Alert severity="warning" sx={{ mt: 2 }}>
            This will create a new revision with the configuration from revision {rollbackConfirm?.revision}.
          </Alert>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleRollbackCancel} disabled={rollbackLoading}>
            Cancel
          </Button>
          <Button
            onClick={handleRollbackConfirm}
            variant="contained"
            color="primary"
            disabled={rollbackLoading}
          >
            {rollbackLoading ? 'Rolling back...' : 'Rollback'}
          </Button>
        </DialogActions>
      </Dialog>
    </Dialog>
  );
}
