import React, { useState, useEffect } from 'react';
import {
  Paper,
  Grid,
  Typography,
  Box,
  TextField,
  Button,
  Chip,
  MenuItem,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  FormControl,
  InputLabel,
  Select,
  Snackbar,
  Alert,
  List,
  ListItem,
  ListItemText,
  IconButton,
} from '@mui/material';
import {
  Delete as DeleteIcon,
  Add as AddIcon,
  Warning as WarningIcon,
} from '@mui/icons-material';
import { meetingAPI, scheduleAPI } from '../services/api';
import { PAPER_STATUS } from '../types';

const SchedulePage = () => {
  const [meetings, setMeetings] = useState([]);
  const [selectedMeeting, setSelectedMeeting] = useState('');
  const [schedules, setSchedules] = useState([]);
  const [acceptedPapers, setAcceptedPapers] = useState([]);
  const [sessions, setSessions] = useState([]);
  const [scheduleDialogOpen, setScheduleDialogOpen] = useState(false);
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });

  const [scheduleForm, setScheduleForm] = useState({
    paper_id: '',
    day: 1,
    session: '',
    start_time: '',
    end_time: '',
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

  const loadScheduleData = async (meetingId) => {
    if (!meetingId) return;
    try {
      const [papersResp, scheduleResp, sessionsResp] = await Promise.all([
        meetingAPI.getPapers(meetingId),
        meetingAPI.getSchedule(meetingId),
        meetingAPI.getSessions(meetingId),
      ]);
      setSchedules(scheduleResp.data);
      setSessions(sessionsResp.data);
      const accepted = papersResp.data.filter(p => p.status === PAPER_STATUS.ACCEPTED);
      const scheduledPaperIds = scheduleResp.data.map(s => s.paper_id);
      setAcceptedPapers(accepted.filter(p => !scheduledPaperIds.includes(p.id)));
    } catch (error) {
      console.error('Failed to load schedule data:', error);
    }
  };

  const handleMeetingChange = (e) => {
    const meetingId = e.target.value;
    setSelectedMeeting(meetingId);
    loadScheduleData(meetingId);
  };

  const showSnackbar = (message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
  };

  const handleCloseSnackbar = () => {
    setSnackbar({ ...snackbar, open: false });
  };

  const handleOpenSchedule = () => {
    if (!selectedMeeting) {
      showSnackbar('请先选择会议', 'warning');
      return;
    }
    if (acceptedPapers.length === 0) {
      showSnackbar('没有待安排的录用论文', 'warning');
      return;
    }
    setScheduleForm({
      paper_id: acceptedPapers[0]?.id.toString() || '',
      day: 1,
      session: '',
      start_time: '',
      end_time: '',
    });
    setScheduleDialogOpen(true);
  };

  const handleSchedulePaper = async () => {
    if (!scheduleForm.paper_id || !scheduleForm.day || !scheduleForm.start_time || !scheduleForm.end_time) {
      showSnackbar('请填写完整信息', 'warning');
      return;
    }

    try {
      await scheduleAPI.create({
        meeting_id: parseInt(selectedMeeting),
        paper_id: parseInt(scheduleForm.paper_id),
        day: parseInt(scheduleForm.day),
        session: scheduleForm.session,
        start_time: scheduleForm.start_time,
        end_time: scheduleForm.end_time,
      });
      setScheduleDialogOpen(false);
      showSnackbar('日程安排成功', 'success');
      loadScheduleData(selectedMeeting);
    } catch (error) {
      console.error('Failed to schedule paper:', error);
      showSnackbar(error.response?.data?.error || '安排失败', 'error');
    }
  };

  const handleDeleteSchedule = async (scheduleId) => {
    if (!window.confirm('确定要删除这个日程安排吗？')) {
      return;
    }
    try {
      await scheduleAPI.delete(scheduleId);
      showSnackbar('日程已删除', 'success');
      loadScheduleData(selectedMeeting);
    } catch (error) {
      console.error('Failed to delete schedule:', error);
      showSnackbar('删除失败', 'error');
    }
  };

  const getSchedulesByDayAndSession = (day, timeOfDay) => {
    return schedules.filter(s => {
      const session = sessions.find(se => se.id === s.session_id);
      return s.day === day && (!session || session.time_of_day === timeOfDay);
    });
  };

  const formatPaperTitle = (title, index) => `论文 ${index + 1}`;

  const days = [1, 2, 3];
  const timeSlots = [
    { key: 'morning', label: '上午' },
    { key: 'afternoon', label: '下午' },
  ];

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Typography variant="h5">日程管理</Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={handleOpenSchedule}
          disabled={!selectedMeeting}
        >
          安排论文
        </Button>
      </Box>

      <Paper sx={{ p: 2, mb: 2 }}>
        <TextField
          label="选择会议"
          select
          value={selectedMeeting}
          onChange={handleMeetingChange}
          fullWidth
          helperText="请选择要管理日程的会议"
        >
          <MenuItem value="">-- 请选择会议 --</MenuItem>
          {meetings.map((m) => (
            <MenuItem key={m.id} value={m.id.toString()}>
              {m.name} ({m.abbreviation})
            </MenuItem>
          ))}
        </TextField>
      </Paper>

      {selectedMeeting && (
        <Paper sx={{ p: 2 }}>
          <Grid container spacing={2}>
            <Grid item xs={1}>
              <Box />
            </Grid>
            {days.map(day => (
              <Grid item xs key={day}>
                <Typography align="center" variant="h6">
                  第 {day} 天
                </Typography>
              </Grid>
            ))}

            {timeSlots.map(slot => (
              <React.Fragment key={slot.key}>
                <Grid item xs={1}>
                  <Typography variant="subtitle2" sx={{ pt: 1 }}>{slot.label}</Typography>
                </Grid>
                {days.map(day => {
                  const daySchedules = getSchedulesByDayAndSession(day, slot.key);
                  return (
                    <Grid item xs key={`${day}-${slot.key}`}>
                      <Paper sx={{ p: 1, minHeight: 100, bgcolor: '#fafafa' }}>
                        {daySchedules.length === 0 ? (
                          <Typography variant="caption" color="text.secondary">
                            暂无安排
                          </Typography>
                        ) : (
                          <List dense>
                            {daySchedules.map(schedule => (
                              <ListItem
                                key={schedule.id}
                                secondaryAction={
                                  <IconButton
                                    edge="end"
                                    size="small"
                                    onClick={() => handleDeleteSchedule(schedule.id)}
                                  >
                                    <DeleteIcon fontSize="small" />
                                  </IconButton>
                                }
                                sx={{
                                  bgcolor: schedule.has_conflict ? 'warning.light' : 'transparent',
                                  borderRadius: 1,
                                  mb: 0.5,
                                }}
                              >
                                <ListItemText
                                  primary={
                                    <Box sx={{ display: 'flex', alignItems: 'center' }}>
                                      {formatPaperTitle(
                                        schedule.paper?.title || '',
                                        schedules.findIndex(s => s.id === schedule.id)
                                      )}
                                      {schedule.has_conflict && (
                                        <WarningIcon
                                          fontSize="small"
                                          color="warning"
                                          sx={{ ml: 1 }}
                                        />
                                      )}
                                    </Box>
                                  }
                                  secondary={
                                    <Typography variant="caption">
                                      {new Date(schedule.start_time).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })}
                                      {' - '}
                                      {new Date(schedule.end_time).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })}
                                    </Typography>
                                  }
                                />
                              </ListItem>
                            ))}
                          </List>
                        )}
                      </Paper>
                    </Grid>
                  );
                })}
              </React.Fragment>
            ))}
          </Grid>

          {schedules.some(s => s.has_conflict) && (
            <Box sx={{ mt: 3, p: 2, bgcolor: 'warning.light', borderRadius: 1 }}>
              <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                <WarningIcon color="warning" sx={{ mr: 1 }} />
                <Typography variant="subtitle1">冲突警告</Typography>
              </Box>
              <Typography variant="body2" color="text.secondary">
                黄色背景标记表示同一作者的不同论文被安排在同一时间段。建议调整安排以避免冲突。
              </Typography>
            </Box>
          )}

          {acceptedPapers.length > 0 && (
            <Box sx={{ mt: 3 }}>
              <Typography variant="h6" sx={{ mb: 2 }}>待安排的录用论文</Typography>
              <Grid container spacing={2}>
                {acceptedPapers.map((paper, index) => (
                  <Grid item xs={12} sm={6} md={4} key={paper.id}>
                    <Paper sx={{ p: 2 }}>
                      <Typography variant="subtitle2">
                        {formatPaperTitle(paper.title, index)}
                      </Typography>
                      <Typography variant="caption" color="text.secondary">
                        {paper.paper_number}
                      </Typography>
                    </Paper>
                  </Grid>
                ))}
              </Grid>
            </Box>
          )}
        </Paper>
      )}

      <Dialog open={scheduleDialogOpen} onClose={() => setScheduleDialogOpen(false)}>
        <DialogTitle>安排论文日程</DialogTitle>
        <DialogContent>
          <Box sx={{ pt: 2, minWidth: 400 }}>
            <FormControl fullWidth sx={{ mb: 2 }}>
              <InputLabel>选择论文</InputLabel>
              <Select
                value={scheduleForm.paper_id}
                label="选择论文"
                onChange={(e) => setScheduleForm({ ...scheduleForm, paper_id: e.target.value })}
              >
                {acceptedPapers.map((paper, index) => (
                  <MenuItem key={paper.id} value={paper.id.toString()}>
                    {formatPaperTitle(paper.title, index)} ({paper.paper_number})
                  </MenuItem>
                ))}
              </Select>
            </FormControl>

            <FormControl fullWidth sx={{ mb: 2 }}>
              <InputLabel>日期</InputLabel>
              <Select
                value={scheduleForm.day}
                label="日期"
                onChange={(e) => setScheduleForm({ ...scheduleForm, day: e.target.value })}
              >
                {days.map(d => (
                  <MenuItem key={d} value={d}>第 {d} 天</MenuItem>
                ))}
              </Select>
            </FormControl>

            <FormControl fullWidth sx={{ mb: 2 }}>
              <InputLabel>Session</InputLabel>
              <Select
                value={scheduleForm.session}
                label="Session"
                onChange={(e) => setScheduleForm({ ...scheduleForm, session: e.target.value })}
              >
                <MenuItem value="morning">上午</MenuItem>
                <MenuItem value="afternoon">下午</MenuItem>
              </Select>
            </FormControl>

            <TextField
              label="开始时间"
              type="datetime-local"
              value={scheduleForm.start_time}
              onChange={(e) => setScheduleForm({ ...scheduleForm, start_time: e.target.value })}
              fullWidth
              sx={{ mb: 2 }}
              InputLabelProps={{ shrink: true }}
              helperText="每个报告20分钟（15分钟演讲+5分钟问答）"
            />

            <TextField
              label="结束时间"
              type="datetime-local"
              value={scheduleForm.end_time}
              onChange={(e) => setScheduleForm({ ...scheduleForm, end_time: e.target.value })}
              fullWidth
              InputLabelProps={{ shrink: true }}
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setScheduleDialogOpen(false)}>取消</Button>
          <Button onClick={handleSchedulePaper} variant="contained">
            安排
          </Button>
        </DialogActions>
      </Dialog>

      <Snackbar open={snackbar.open} autoHideDuration={6000} onClose={handleCloseSnackbar}>
        <Alert onClose={handleCloseSnackbar} severity={snackbar.severity} sx={{ width: '100%' }}>
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Box>
  );
};

export default SchedulePage;
