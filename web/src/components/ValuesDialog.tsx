import { useState, useMemo } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  CircularProgress,
  Alert,
  Typography,
  Box,
  TextField,
} from '@mui/material';
import { useValues, useUpdateValues } from '../hooks/useReleases';
import type { Release } from '../types';

interface ValuesDialogProps {
  open: boolean;
  onClose: () => void;
  release: Release;
}

function formatYaml(obj: Record<string, unknown>, indent = 0): string {
  const lines: string[] = [];
  const prefix = '  '.repeat(indent);

  for (const [key, value] of Object.entries(obj)) {
    if (value === null || value === undefined) {
      lines.push(`${prefix}${key}: null`);
    } else if (typeof value === 'object' && !Array.isArray(value)) {
      lines.push(`${prefix}${key}:`);
      lines.push(formatYaml(value as Record<string, unknown>, indent + 1));
    } else if (Array.isArray(value)) {
      lines.push(`${prefix}${key}:`);
      value.forEach((item) => {
        if (typeof item === 'object' && item !== null) {
          const itemLines = formatYaml(item as Record<string, unknown>, indent + 2).split('\n');
          if (itemLines.length > 0) {
            lines.push(`${prefix}  - ${itemLines[0].trim()}`);
            itemLines.slice(1).forEach((l) => lines.push(`${prefix}    ${l.trim()}`));
          }
        } else {
          lines.push(`${prefix}  - ${JSON.stringify(item)}`);
        }
      });
    } else if (typeof value === 'string') {
      if (value.includes('\n')) {
        lines.push(`${prefix}${key}: |`);
        value.split('\n').forEach((l) => lines.push(`${prefix}  ${l}`));
      } else {
        lines.push(`${prefix}${key}: ${value}`);
      }
    } else {
      lines.push(`${prefix}${key}: ${JSON.stringify(value)}`);
    }
  }

  return lines.join('\n');
}

function parseYaml(yamlStr: string): Record<string, unknown> {
  // Simple YAML parser for basic key-value pairs
  // For production, consider using a proper YAML library
  const result: Record<string, unknown> = {};
  const lines = yamlStr.split('\n');
  const stack: { obj: Record<string, unknown>; indent: number }[] = [{ obj: result, indent: -1 }];

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    if (line.trim() === '' || line.trim().startsWith('#')) continue;

    const indent = line.search(/\S/);
    const content = line.trim();

    // Pop stack until we find the right parent
    while (stack.length > 1 && stack[stack.length - 1].indent >= indent) {
      stack.pop();
    }

    const current = stack[stack.length - 1].obj;

    if (content.includes(':')) {
      const colonIndex = content.indexOf(':');
      const key = content.substring(0, colonIndex).trim();
      const valueStr = content.substring(colonIndex + 1).trim();

      if (valueStr === '' || valueStr === '|') {
        // Nested object or multiline string
        const newObj: Record<string, unknown> = {};
        current[key] = newObj;
        stack.push({ obj: newObj, indent });
      } else if (valueStr === 'null') {
        current[key] = null;
      } else if (valueStr === 'true') {
        current[key] = true;
      } else if (valueStr === 'false') {
        current[key] = false;
      } else if (!isNaN(Number(valueStr))) {
        current[key] = Number(valueStr);
      } else if (valueStr.startsWith('"') && valueStr.endsWith('"')) {
        current[key] = valueStr.slice(1, -1);
      } else if (valueStr.startsWith("'") && valueStr.endsWith("'")) {
        current[key] = valueStr.slice(1, -1);
      } else {
        current[key] = valueStr;
      }
    }
  }

  return result;
}

export default function ValuesDialog({ open, onClose, release }: ValuesDialogProps) {
  const [editedValues, setEditedValues] = useState<string | null>(null);
  const [parseError, setParseError] = useState<string | null>(null);

  const { data: currentValues, isLoading: loadingValues, error: valuesError } = useValues(
    release.namespace,
    release.name
  );
  const { mutate: saveValues, isPending: isSaving, error: saveError } = useUpdateValues();

  const initialValues = useMemo(() => {
    if (currentValues) {
      return formatYaml(currentValues);
    }
    return '';
  }, [currentValues]);

  const displayValue = editedValues ?? initialValues;

  const handleSave = () => {
    try {
      const parsed = parseYaml(displayValue);
      setParseError(null);

      saveValues(
        {
          namespace: release.namespace,
          name: release.name,
          request: { values: parsed },
        },
        {
          onSuccess: () => {
            setEditedValues(null);
            onClose();
          },
        }
      );
    } catch (e) {
      setParseError((e as Error).message);
    }
  };

  const handleClose = () => {
    setEditedValues(null);
    setParseError(null);
    onClose();
  };

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="md" fullWidth>
      <DialogTitle>Edit Values</DialogTitle>
      <DialogContent>
        <Box mb={2}>
          <Typography variant="body2" color="textSecondary">
            Release: {release.namespace}/{release.name}
          </Typography>
          <Typography variant="body2" color="textSecondary">
            Chart: {release.chart} ({release.chartVersion})
          </Typography>
        </Box>

        {valuesError && (
          <Alert severity="error" sx={{ mb: 2 }}>
            Failed to load values: {(valuesError as Error).message}
          </Alert>
        )}

        {saveError && (
          <Alert severity="error" sx={{ mb: 2 }}>
            Failed to save values: {(saveError as Error).message}
          </Alert>
        )}

        {parseError && (
          <Alert severity="error" sx={{ mb: 2 }}>
            YAML parse error: {parseError}
          </Alert>
        )}

        {!release.hasRegistry && (
          <Alert severity="warning" sx={{ mb: 2 }}>
            Registry is not configured for this release. Please set a registry first before updating values.
          </Alert>
        )}

        {loadingValues ? (
          <Box display="flex" justifyContent="center" p={2}>
            <CircularProgress size={24} />
          </Box>
        ) : (
          <TextField
            fullWidth
            multiline
            rows={20}
            value={displayValue}
            onChange={(e) => setEditedValues(e.target.value)}
            sx={{
              mt: 2,
              '& .MuiInputBase-input': {
                fontFamily: 'monospace',
                fontSize: '0.875rem',
              },
            }}
            placeholder="# YAML format values"
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
          disabled={!displayValue.trim() || isSaving || !release.hasRegistry}
        >
          {isSaving ? <CircularProgress size={24} /> : 'Save & Deploy'}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
