import { useState } from 'react';
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  IconButton,
  Typography,
  Box,
  CircularProgress,
  Alert,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
} from '@mui/material';
import DeleteIcon from '@mui/icons-material/Delete';
import RefreshIcon from '@mui/icons-material/Refresh';
import AddIcon from '@mui/icons-material/Add';
import {
  useRepositories,
  useAddRepository,
  useRemoveRepository,
  useUpdateRepository,
} from '../hooks/useRepositories';

export default function RepositoryList() {
  const { data: repositories, isLoading, error } = useRepositories();
  const { mutate: addRepo, isPending: isAdding } = useAddRepository();
  const { mutate: removeRepo, isPending: isRemoving } = useRemoveRepository();
  const { mutate: updateRepo, isPending: isUpdating } = useUpdateRepository();

  const [dialogOpen, setDialogOpen] = useState(false);
  const [newRepoName, setNewRepoName] = useState('');
  const [newRepoUrl, setNewRepoUrl] = useState('');
  const [addError, setAddError] = useState<string | null>(null);

  const handleAdd = () => {
    if (!newRepoName || !newRepoUrl) return;

    addRepo(
      { name: newRepoName, url: newRepoUrl },
      {
        onSuccess: () => {
          setDialogOpen(false);
          setNewRepoName('');
          setNewRepoUrl('');
          setAddError(null);
        },
        onError: (err) => {
          setAddError((err as Error).message);
        },
      }
    );
  };

  const handleDialogClose = () => {
    setDialogOpen(false);
    setNewRepoName('');
    setNewRepoUrl('');
    setAddError(null);
  };

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" p={4}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return <Alert severity="error">Failed to load repositories: {(error as Error).message}</Alert>;
  }

  return (
    <>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={2}>
        <Typography variant="h5">Helm Repositories</Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={() => setDialogOpen(true)}
        >
          Add Repository
        </Button>
      </Box>

      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Name</TableCell>
              <TableCell>URL</TableCell>
              <TableCell>Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {repositories?.map((repo) => (
              <TableRow key={repo.name}>
                <TableCell>{repo.name}</TableCell>
                <TableCell>{repo.url}</TableCell>
                <TableCell>
                  <IconButton
                    onClick={() => updateRepo(repo.name)}
                    disabled={isUpdating}
                    title="Update Repository"
                    size="small"
                  >
                    <RefreshIcon />
                  </IconButton>
                  <IconButton
                    onClick={() => removeRepo(repo.name)}
                    disabled={isRemoving}
                    title="Remove Repository"
                    size="small"
                    color="error"
                  >
                    <DeleteIcon />
                  </IconButton>
                </TableCell>
              </TableRow>
            ))}
            {repositories?.length === 0 && (
              <TableRow>
                <TableCell colSpan={3} align="center">
                  No repositories configured
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </TableContainer>

      <Dialog open={dialogOpen} onClose={handleDialogClose} maxWidth="sm" fullWidth>
        <DialogTitle>Add Repository</DialogTitle>
        <DialogContent>
          {addError && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {addError}
            </Alert>
          )}
          <TextField
            autoFocus
            margin="dense"
            label="Name"
            fullWidth
            value={newRepoName}
            onChange={(e) => setNewRepoName(e.target.value)}
          />
          <TextField
            margin="dense"
            label="URL"
            fullWidth
            value={newRepoUrl}
            onChange={(e) => setNewRepoUrl(e.target.value)}
            placeholder="https://charts.example.com"
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={handleDialogClose} disabled={isAdding}>
            Cancel
          </Button>
          <Button
            onClick={handleAdd}
            variant="contained"
            disabled={!newRepoName || !newRepoUrl || isAdding}
          >
            {isAdding ? <CircularProgress size={24} /> : 'Add'}
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
}
