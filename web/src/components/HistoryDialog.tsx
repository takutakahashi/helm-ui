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
} from '@mui/material';
import { useReleaseHistory } from '../hooks/useReleases';
import type { Release } from '../types';

interface HistoryDialogProps {
  open: boolean;
  onClose: () => void;
  release: Release;
}

export default function HistoryDialog({ open, onClose, release }: HistoryDialogProps) {
  const { data: history, isLoading, error } = useReleaseHistory(
    release.namespace,
    release.name
  );

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
                  </TableRow>
                ))}
                {history?.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={6} align="center">
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
    </Dialog>
  );
}
