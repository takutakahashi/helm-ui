import { useState } from 'react';
import {
  AppBar,
  Toolbar,
  Typography,
  Container,
  Tabs,
  Tab,
  Box,
  CssBaseline,
  ThemeProvider,
  createTheme,
} from '@mui/material';
import ReleaseList from './components/ReleaseList';
import RepositoryList from './components/RepositoryList';

const theme = createTheme({
  palette: {
    mode: 'light',
    primary: {
      main: '#1976d2',
    },
  },
});

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel({ children, value, index }: TabPanelProps) {
  return (
    <div role="tabpanel" hidden={value !== index}>
      {value === index && <Box py={3}>{children}</Box>}
    </div>
  );
}

function App() {
  const [tabIndex, setTabIndex] = useState(0);

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <AppBar position="static">
        <Toolbar>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            Helm Version Manager
          </Typography>
        </Toolbar>
      </AppBar>
      <Container maxWidth="lg" sx={{ mt: 2 }}>
        <Tabs
          value={tabIndex}
          onChange={(_, newValue) => setTabIndex(newValue)}
          sx={{ borderBottom: 1, borderColor: 'divider' }}
        >
          <Tab label="Releases" />
          <Tab label="Repositories" />
        </Tabs>
        <TabPanel value={tabIndex} index={0}>
          <ReleaseList />
        </TabPanel>
        <TabPanel value={tabIndex} index={1}>
          <RepositoryList />
        </TabPanel>
      </Container>
    </ThemeProvider>
  );
}

export default App;
