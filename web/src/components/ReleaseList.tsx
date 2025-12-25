import { useState } from 'react';
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Chip,
  IconButton,
  Typography,
  Box,
  CircularProgress,
  Alert,
} from '@mui/material';
import UpgradeIcon from '@mui/icons-material/Upgrade';
import HistoryIcon from '@mui/icons-material/History';
import { useReleases } from '../hooks/useReleases';
import type { Release } from '../types';
import VersionDialog from './VersionDialog';
import HistoryDialog from './HistoryDialog';

const getStatusColor = (status: string) => {
  switch (status.toLowerCase()) {
    case 'deployed':
      return 'success';
    case 'failed':
      return 'error';
    case 'pending':
    case 'pending-install':
    case 'pending-upgrade':
      return 'warning';
    default:
      return 'default';
  }
};

export default function ReleaseList() {
  const { data: releases, isLoading, error } = useReleases();
  const [selectedRelease, setSelectedRelease] = useState<Release | null>(null);
  const [versionDialogOpen, setVersionDialogOpen] = useState(false);
  const [historyDialogOpen, setHistoryDialogOpen] = useState(false);

  const handleOpenVersionDialog = (release: Release) => {
    setSelectedRelease(release);
    setVersionDialogOpen(true);
  };

  const handleOpenHistoryDialog = (release: Release) => {
    setSelectedRelease(release);
    setHistoryDialogOpen(true);
  };

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" p={4}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return <Alert severity="error">Failed to load releases: {(error as Error).message}</Alert>;
  }

  return (
    <>
      <Typography variant="h5" gutterBottom>
        Helm Releases
      </Typography>
      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Name</TableCell>
              <TableCell>Namespace</TableCell>
              <TableCell>Chart</TableCell>
              <TableCell>Version</TableCell>
              <TableCell>App Version</TableCell>
              <TableCell>Status</TableCell>
              <TableCell>Updated</TableCell>
              <TableCell>Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {releases?.map((release) => (
              <TableRow key={`${release.namespace}/${release.name}`}>
                <TableCell>{release.name}</TableCell>
                <TableCell>{release.namespace}</TableCell>
                <TableCell>{release.chart}</TableCell>
                <TableCell>{release.chartVersion}</TableCell>
                <TableCell>{release.appVersion}</TableCell>
                <TableCell>
                  <Chip
                    label={release.status}
                    color={getStatusColor(release.status)}
                    size="small"
                  />
                </TableCell>
                <TableCell>
                  {new Date(release.updated).toLocaleString()}
                </TableCell>
                <TableCell>
                  <IconButton
                    onClick={() => handleOpenVersionDialog(release)}
                    title="Change Version"
                    size="small"
                  >
                    <UpgradeIcon />
                  </IconButton>
                  <IconButton
                    onClick={() => handleOpenHistoryDialog(release)}
                    title="View History"
                    size="small"
                  >
                    <HistoryIcon />
                  </IconButton>
                </TableCell>
              </TableRow>
            ))}
            {releases?.length === 0 && (
              <TableRow>
                <TableCell colSpan={8} align="center">
                  No releases found
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </TableContainer>

      {selectedRelease && (
        <>
          <VersionDialog
            open={versionDialogOpen}
            onClose={() => setVersionDialogOpen(false)}
            release={selectedRelease}
          />
          <HistoryDialog
            open={historyDialogOpen}
            onClose={() => setHistoryDialogOpen(false)}
            release={selectedRelease}
          />
        </>
      )}
    </>
  );
}
