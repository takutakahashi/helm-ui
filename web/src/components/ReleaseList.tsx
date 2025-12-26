import { useState, useMemo } from 'react';
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
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Stack,
} from '@mui/material';
import type { SelectChangeEvent } from '@mui/material';
import UpgradeIcon from '@mui/icons-material/Upgrade';
import HistoryIcon from '@mui/icons-material/History';
import SettingsIcon from '@mui/icons-material/Settings';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import { useReleases } from '../hooks/useReleases';
import type { Release, ReleaseFilter } from '../types';
import VersionDialog from './VersionDialog';
import HistoryDialog from './HistoryDialog';
import RegistryDialog from './RegistryDialog';

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
  const [namespaceFilter, setNamespaceFilter] = useState<string>('');
  const [hasRegistryFilter, setHasRegistryFilter] = useState<string>('');

  const filter: ReleaseFilter | undefined = useMemo(() => {
    const f: ReleaseFilter = {};
    if (namespaceFilter) {
      f.namespace = namespaceFilter;
    }
    if (hasRegistryFilter === 'true') {
      f.hasRegistry = true;
    } else if (hasRegistryFilter === 'false') {
      f.hasRegistry = false;
    }
    return Object.keys(f).length > 0 ? f : undefined;
  }, [namespaceFilter, hasRegistryFilter]);

  const { data: releases, isLoading, error } = useReleases(filter);

  // Get all releases without filters to extract unique namespaces
  const { data: allReleases } = useReleases();
  const namespaces = useMemo(() => {
    if (!allReleases) return [];
    const nsSet = new Set(allReleases.map((r) => r.namespace));
    return Array.from(nsSet).sort();
  }, [allReleases]);

  const [selectedRelease, setSelectedRelease] = useState<Release | null>(null);
  const [versionDialogOpen, setVersionDialogOpen] = useState(false);
  const [historyDialogOpen, setHistoryDialogOpen] = useState(false);
  const [registryDialogOpen, setRegistryDialogOpen] = useState(false);

  const handleOpenVersionDialog = (release: Release) => {
    setSelectedRelease(release);
    setVersionDialogOpen(true);
  };

  const handleOpenHistoryDialog = (release: Release) => {
    setSelectedRelease(release);
    setHistoryDialogOpen(true);
  };

  const handleOpenRegistryDialog = (release: Release) => {
    setSelectedRelease(release);
    setRegistryDialogOpen(true);
  };

  const handleNamespaceChange = (event: SelectChangeEvent) => {
    setNamespaceFilter(event.target.value);
  };

  const handleHasRegistryChange = (event: SelectChangeEvent) => {
    setHasRegistryFilter(event.target.value);
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

      <Stack direction="row" spacing={2} sx={{ mb: 2 }}>
        <FormControl size="small" sx={{ minWidth: 200 }}>
          <InputLabel id="namespace-filter-label">Namespace</InputLabel>
          <Select
            labelId="namespace-filter-label"
            value={namespaceFilter}
            label="Namespace"
            onChange={handleNamespaceChange}
          >
            <MenuItem value="">
              <em>All Namespaces</em>
            </MenuItem>
            {namespaces.map((ns) => (
              <MenuItem key={ns} value={ns}>
                {ns}
              </MenuItem>
            ))}
          </Select>
        </FormControl>

        <FormControl size="small" sx={{ minWidth: 200 }}>
          <InputLabel id="registry-filter-label">Registry Status</InputLabel>
          <Select
            labelId="registry-filter-label"
            value={hasRegistryFilter}
            label="Registry Status"
            onChange={handleHasRegistryChange}
          >
            <MenuItem value="">
              <em>All</em>
            </MenuItem>
            <MenuItem value="true">Registered</MenuItem>
            <MenuItem value="false">Not Registered</MenuItem>
          </Select>
        </FormControl>
      </Stack>

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
              <TableCell>Registry</TableCell>
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
                  {release.hasRegistry ? (
                    <CheckCircleIcon color="success" fontSize="small" />
                  ) : (
                    <Chip label="Not Set" size="small" variant="outlined" />
                  )}
                </TableCell>
                <TableCell>
                  {new Date(release.updated).toLocaleString()}
                </TableCell>
                <TableCell>
                  <IconButton
                    onClick={() => handleOpenRegistryDialog(release)}
                    title="Set Registry"
                    size="small"
                  >
                    <SettingsIcon />
                  </IconButton>
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
                <TableCell colSpan={9} align="center">
                  No releases found
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </TableContainer>

      {selectedRelease && (
        <>
          <RegistryDialog
            open={registryDialogOpen}
            onClose={() => setRegistryDialogOpen(false)}
            release={selectedRelease}
          />
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
