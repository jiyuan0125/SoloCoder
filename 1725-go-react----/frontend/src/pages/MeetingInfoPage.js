import React, { useState, useEffect } from 'react';
import {
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Chip,
  Box,
  Typography,
  IconButton,
  MenuItem,
} from '@mui/material';
import { Edit as EditIcon, Delete as DeleteIcon, Add as AddIcon } from '@mui/icons-material';
import { meetingAPI } from '../services/api';
import { MEETING_STATUS, MEETING_STATUS_LABELS } from '../types';

const MeetingInfoPage = () => {
  const [meetings, setMeetings] = useState([]);
  const [openDialog, setOpenDialog] = useState(false);
  const [editingMeeting, setEditingMeeting] = useState(null);
  const [formData, setFormData] = useState({
    name: '',
    abbreviation: '',
    start_date: '',
    end_date: '',
    location: '',
    topics: '',
    submission_deadline: '',
    review_deadline: '',
    notification_deadline: '',
    status: MEETING_STATUS.PREPARING,
  });

  useEffect(() => {
    loadMeetings();
  }, []);

  const loadMeetings = async () => {
    try {
      const response = await meetingAPI.getAll();
      setMeetings(response.data);
    } catch (error) {
      console.error('Failed to load meetings:', error);
    }
  };

  const handleOpenCreate = () => {
    setEditingMeeting(null);
    setFormData({
      name: '',
      abbreviation: '',
      start_date: '',
      end_date: '',
      location: '',
      topics: '',
      submission_deadline: '',
      review_deadline: '',
      notification_deadline: '',
      status: MEETING_STATUS.PREPARING,
    });
    setOpenDialog(true);
  };

  const handleOpenEdit = (meeting) => {
    setEditingMeeting(meeting);
    setFormData({
      name: meeting.name || '',
      abbreviation: meeting.abbreviation || '',
      start_date: meeting.start_date ? meeting.start_date.split('T')[0] : '',
      end_date: meeting.end_date ? meeting.end_date.split('T')[0] : '',
      location: meeting.location || '',
      topics: meeting.topics || '',
      submission_deadline: meeting.submission_deadline ? meeting.submission_deadline.split('T')[0] : '',
      review_deadline: meeting.review_deadline ? meeting.review_deadline.split('T')[0] : '',
      notification_deadline: meeting.notification_deadline ? meeting.notification_deadline.split('T')[0] : '',
      status: meeting.status || MEETING_STATUS.PREPARING,
    });
    setOpenDialog(true);
  };

  const handleClose = () => {
    setOpenDialog(false);
    setEditingMeeting(null);
  };

  const handleSubmit = async () => {
    try {
      const data = {
        ...formData,
        start_date: formData.start_date ? new Date(formData.start_date).toISOString() : null,
        end_date: formData.end_date ? new Date(formData.end_date).toISOString() : null,
        submission_deadline: formData.submission_deadline ? new Date(formData.submission_deadline).toISOString() : null,
        review_deadline: formData.review_deadline ? new Date(formData.review_deadline).toISOString() : null,
        notification_deadline: formData.notification_deadline ? new Date(formData.notification_deadline).toISOString() : null,
      };

      if (editingMeeting) {
        await meetingAPI.update(editingMeeting.id, data);
      } else {
        await meetingAPI.create(data);
      }
      handleClose();
      loadMeetings();
    } catch (error) {
      console.error('Failed to save meeting:', error);
      alert(error.response?.data?.error || '保存失败');
    }
  };

  const handleDelete = async (id) => {
    if (window.confirm('确定要删除这个会议吗？')) {
      try {
        await meetingAPI.delete(id);
        loadMeetings();
      } catch (error) {
        console.error('Failed to delete meeting:', error);
      }
    }
  };

  const formatDate = (dateStr) => {
    if (!dateStr) return '-';
    return new Date(dateStr).toLocaleDateString('zh-CN');
  };

  const getStatusColor = (status) => {
    switch (status) {
      case MEETING_STATUS.ACCEPTING_SUBMISSIONS: return 'success';
      case MEETING_STATUS.REVIEWING: return 'primary';
      case MEETING_STATUS.NOTIFYING: return 'warning';
      case MEETING_STATUS.ENDED: return 'default';
      default: return 'info';
    }
  };

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Typography variant="h5">会议信息管理</Typography>
        <Button variant="contained" startIcon={<AddIcon />} onClick={handleOpenCreate}>
          新建会议
        </Button>
      </Box>

      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>名称</TableCell>
              <TableCell>缩写</TableCell>
              <TableCell>举办日期</TableCell>
              <TableCell>地点</TableCell>
              <TableCell>征稿截止</TableCell>
              <TableCell>状态</TableCell>
              <TableCell>操作</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {meetings.map((meeting) => (
              <TableRow key={meeting.id}>
                <TableCell>{meeting.name}</TableCell>
                <TableCell>
                  <Chip label={meeting.abbreviation} variant="outlined" />
                </TableCell>
                <TableCell>
                  {formatDate(meeting.start_date)} - {formatDate(meeting.end_date)}
                </TableCell>
                <TableCell>{meeting.location}</TableCell>
                <TableCell>{formatDate(meeting.submission_deadline)}</TableCell>
                <TableCell>
                  <Chip
                    label={MEETING_STATUS_LABELS[meeting.status] || meeting.status}
                    color={getStatusColor(meeting.status)}
                  />
                </TableCell>
                <TableCell>
                  <IconButton onClick={() => handleOpenEdit(meeting)} color="primary">
                    <EditIcon />
                  </IconButton>
                  <IconButton onClick={() => handleDelete(meeting.id)} color="error">
                    <DeleteIcon />
                  </IconButton>
                </TableCell>
              </TableRow>
            ))}
            {meetings.length === 0 && (
              <TableRow>
                <TableCell colSpan={7} align="center">
                  暂无会议数据
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </TableContainer>

      <Dialog open={openDialog} onClose={handleClose} maxWidth="md" fullWidth>
        <DialogTitle>
          {editingMeeting ? '编辑会议' : '新建会议'}
        </DialogTitle>
        <DialogContent>
          <Box sx={{ pt: 2, display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 2 }}>
            <TextField
              label="会议名称"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              fullWidth
              required
            />
            <TextField
              label="会议缩写"
              value={formData.abbreviation}
              onChange={(e) => setFormData({ ...formData, abbreviation: e.target.value })}
              fullWidth
              required
              helperText="如 ICSE2026"
            />
            <TextField
              label="开始日期"
              type="date"
              value={formData.start_date}
              onChange={(e) => setFormData({ ...formData, start_date: e.target.value })}
              fullWidth
              InputLabelProps={{ shrink: true }}
            />
            <TextField
              label="结束日期"
              type="date"
              value={formData.end_date}
              onChange={(e) => setFormData({ ...formData, end_date: e.target.value })}
              fullWidth
              InputLabelProps={{ shrink: true }}
            />
            <TextField
              label="举办地点"
              value={formData.location}
              onChange={(e) => setFormData({ ...formData, location: e.target.value })}
              fullWidth
            />
            <TextField
              label="状态"
              select
              value={formData.status}
              onChange={(e) => setFormData({ ...formData, status: e.target.value })}
              fullWidth
            >
              {Object.entries(MEETING_STATUS_LABELS).map(([value, label]) => (
                <MenuItem key={value} value={value}>
                  {label}
                </MenuItem>
              ))}
            </TextField>
            <TextField
              label="征稿主题"
              value={formData.topics}
              onChange={(e) => setFormData({ ...formData, topics: e.target.value })}
              fullWidth
              multiline
              rows={2}
              helperText="软件工程、AI、数据库等"
            />
            <TextField
              label="投稿截止日期"
              type="date"
              value={formData.submission_deadline}
              onChange={(e) => setFormData({ ...formData, submission_deadline: e.target.value })}
              fullWidth
              InputLabelProps={{ shrink: true }}
            />
            <TextField
              label="审稿截止日期"
              type="date"
              value={formData.review_deadline}
              onChange={(e) => setFormData({ ...formData, review_deadline: e.target.value })}
              fullWidth
              InputLabelProps={{ shrink: true }}
            />
            <TextField
              label="录用通知日期"
              type="date"
              value={formData.notification_deadline}
              onChange={(e) => setFormData({ ...formData, notification_deadline: e.target.value })}
              fullWidth
              InputLabelProps={{ shrink: true }}
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose}>取消</Button>
          <Button onClick={handleSubmit} variant="contained">
            {editingMeeting ? '保存' : '创建'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default MeetingInfoPage;
