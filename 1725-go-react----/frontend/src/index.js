import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter, Routes, Route, Link, Navigate } from 'react-router-dom';
import { AppBar, Toolbar, Typography, Container, Tabs, Tab, Box } from '@mui/material';
import MeetingInfoPage from './pages/MeetingInfoPage';
import SubmissionPage from './pages/SubmissionPage';
import ReviewPage from './pages/ReviewPage';
import SchedulePage from './pages/SchedulePage';
import './styles.css';

function App() {
  const [value, setValue] = React.useState(0);

  const handleChange = (event, newValue) => {
    setValue(newValue);
  };

  return (
    <BrowserRouter>
      <Box sx={{ flexGrow: 1 }}>
        <AppBar position="static">
          <Toolbar>
            <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
              学术会议管理系统
            </Typography>
          </Toolbar>
        </AppBar>
        <Container maxWidth="lg" sx={{ mt: 4 }}>
          <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 2 }}>
            <Tabs value={value} onChange={handleChange} aria-label="navigation tabs">
              <Tab label="会议信息" component={Link} to="/meetings" />
              <Tab label="投稿管理" component={Link} to="/submissions" />
              <Tab label="审稿工作台" component={Link} to="/reviews" />
              <Tab label="日程管理" component={Link} to="/schedule" />
            </Tabs>
          </Box>
          <Routes>
            <Route path="/" element={<Navigate to="/meetings" replace />} />
            <Route path="/meetings" element={<MeetingInfoPage />} />
            <Route path="/submissions" element={<SubmissionPage />} />
            <Route path="/reviews" element={<ReviewPage />} />
            <Route path="/schedule" element={<SchedulePage />} />
          </Routes>
        </Container>
      </Box>
    </BrowserRouter>
  );
}

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(<App />);
