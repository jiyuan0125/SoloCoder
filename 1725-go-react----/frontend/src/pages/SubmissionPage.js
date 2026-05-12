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
  FormControlLabel,
  Checkbox,
  CircularProgress,
  Snackbar,
  Alert,
} from '@mui/material';
import {
  Add as AddIcon,
  Visibility as VisibilityIcon,
  Send as SendIcon,
  Delete as DeleteIcon,
  Download as DownloadIcon,
} from '@mui/icons-material';
import { meetingAPI, paperAPI } from '../services/api';
import { PAPER_STATUS, PAPER_STATUS_LABELS } from '../types';

const SubmissionPage = () => {
  const [meetings, setMeetings] = useState([]);
  const [selectedMeeting, setSelectedMeeting] = useState('');
  const [papers, setPapers] = useState([]);
  const [loading, setLoading] = useState(false);
  const [openDialog, setOpenDialog] = useState(false);
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });

  const [formData, setFormData] = useState({
    title: '',
    abstract: '',
    keywords: '',
    topic_area: '',
    content: '',
    authors: [{ name: '', email: '', affiliation: '', is_first_author: false, is_corresponding: false }],
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

  const loadPapers = async (meetingId) => {
    if (!meetingId) return;
    setLoading(true);
    try {
      const response = await meetingAPI.getPapers(meetingId);
      setPapers(response.data);
    } catch (error) {
      console.error('Failed to load papers:', error);
      showSnackbar('加载论文失败', 'error');
    } finally {
      setLoading(false);
    }
  };

  const handleMeetingChange = (e) => {
    const meetingId = e.target.value;
    setSelectedMeeting(meetingId);
    loadPapers(meetingId);
  };

  const showSnackbar = (message, severity = 'success') => {
    setSnackbar({ open: true, message, severity });
  };

  const handleCloseSnackbar = () => {
    setSnackbar({ ...snackbar, open: false });
  };

  const handleOpenCreate = () => {
    if (!selectedMeeting) {
      showSnackbar('请先选择会议', 'warning');
      return;
    }
    setFormData({
      title: '',
      abstract: '',
      keywords: '',
      topic_area: '',
      content: '',
      authors: [{ name: '', email: '', affiliation: '', is_first_author: false, is_corresponding: false }],
    });
    setOpenDialog(true);
  };

  const handleClose = () => {
    setOpenDialog(false);
  };

  const handleAddAuthor = () => {
    setFormData({
      ...formData,
      authors: [...formData.authors, { name: '', email: '', affiliation: '', is_first_author: false, is_corresponding: false }],
    });
  };

  const handleRemoveAuthor = (index) => {
    if (formData.authors.length <= 1) return;
    const newAuthors = formData.authors.filter((_, i) => i !== index);
    setFormData({ ...formData, authors: newAuthors });
  };

  const handleAuthorChange = (index, field, value) => {
    const newAuthors = [...formData.authors];
    newAuthors[index][field] = value;
    setFormData({ ...formData, authors: newAuthors });
  };

  const handleSubmit = async () => {
    if (!formData.title.trim()) {
      showSnackbar('请输入论文标题', 'error');
      return;
    }
    if (formData.abstract.length < 200) {
      showSnackbar('摘要至少需要200字', 'error');
      return;
    }
    if (formData.abstract.length > 500) {
      showSnackbar('摘要最多500字', 'error');
      return;
    }

    const validAuthors = formData.authors.filter(a => a.name && a.email);
    if (validAuthors.length === 0) {
      showSnackbar('至少需要一位作者', 'error');
      return;
    }

    try {
      const paperData = {
        meeting_id: parseInt(selectedMeeting),
        ...formData,
        authors: validAuthors,
      };
      await paperAPI.create(paperData);
      handleClose();
      loadPapers(selectedMeeting);
      showSnackbar('论文创建成功', 'success');
    } catch (error) {
      console.error('Failed to create paper:', error);
      showSnackbar(error.response?.data?.error || '创建失败', 'error');
    }
  };

  const handleSubmitPaper = async (paperId) => {
    if (!window.confirm('确定要提交这篇论文吗？提交后将进入格式审查流程。')) {
      return;
    }
    try {
      await paperAPI.submit(paperId);
      loadPapers(selectedMeeting);
      showSnackbar('论文提交成功', 'success');
    } catch (error) {
      console.error('Failed to submit paper:', error);
      showSnackbar(error.response?.data?.error || '提交失败', 'error');
    }
  };

  const handleExport = async () => {
    if (!selectedMeeting) {
      showSnackbar('请先选择会议', 'warning');
      return;
    }
    try {
      const response = await meetingAPI.exportPapers(selectedMeeting);
      const blob = new Blob([JSON.stringify(response.data, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `papers_${selectedMeeting}.json`;
      a.click();
      URL.revokeObjectURL(url);
      showSnackbar('导出成功', 'success');
    } catch (error) {
      console.error('Failed to export:', error);
      showSnackbar('导出失败', 'error');
    }
  };

  const getStatusColor = (status) => {
    switch (status) {
      case PAPER_STATUS.ACCEPTED: return 'success';
      case PAPER_STATUS.REJECTED: return 'error';
      case PAPER_STATUS.UNDER_REVIEW: return 'primary';
      case PAPER_STATUS.FORMAT_CHECKING: return 'info';
      case PAPER_STATUS.REVISION: return 'warning';
      default: return 'default';
    }
  };

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Typography variant="h5">投稿管理</Typography>
        <Box>
          <Button
            variant="outlined"
            startIcon={<DownloadIcon />}
            onClick={handleExport}
            sx={{ mr: 1 }}
          >
            导出论文
          </Button>
          <Button variant="contained" startIcon={<AddIcon />} onClick={handleOpenCreate}>
            新建投稿
          </Button>
        </Box>
      </Box>

      <Paper sx={{ p: 2, mb: 2 }}>
        <TextField
          label="选择会议"
          select
          value={selectedMeeting}
          onChange={handleMeetingChange}
          fullWidth
          helperText="请先选择要管理的会议"
        >
          <MenuItem value="">-- 请选择会议 --</MenuItem>
          {meetings.map((m) => (
            <MenuItem key={m.id} value={m.id.toString()}>
              {m.name} ({m.abbreviation})
            </MenuItem>
          ))}
        </TextField>
      </Paper>

      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CircularProgress />
        </Box>
      ) : (
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>论文编号</TableCell>
                <TableCell>标题</TableCell>
                <TableCell>主题领域</TableCell>
                <TableCell>状态</TableCell>
                <TableCell>提交时间</TableCell>
                <TableCell>操作</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {papers.map((paper) => (
                <TableRow key={paper.id}>
                  <TableCell>
                    <Chip label={paper.paper_number} variant="outlined" />
                  </TableCell>
                  <TableCell>{paper.title}</TableCell>
                  <TableCell>{paper.topic_area || '-'}</TableCell>
                  <TableCell>
                    <Chip
                      label={PAPER_STATUS_LABELS[paper.status] || paper.status}
                      color={getStatusColor(paper.status)}
                    />
                  </TableCell>
                  <TableCell>
                    {paper.submission_time ? new Date(paper.submission_time).toLocaleString('zh-CN') : '-'}
                  </TableCell>
                  <TableCell>
                    {paper.status === PAPER_STATUS.DRAFT && (
                      <IconButton
                        onClick={() => handleSubmitPaper(paper.id)}
                        color="primary"
                        title="提交论文"
                      >
                        <SendIcon />
                      </IconButton>
                    )}
                  </TableCell>
                </TableRow>
              ))}
              {papers.length === 0 && selectedMeeting && (
                <TableRow>
                  <TableCell colSpan={6} align="center">
                    暂无论文数据
                  </TableCell>
                </TableRow>
              )}
              {!selectedMeeting && (
                <TableRow>
                  <TableCell colSpan={6} align="center">
                    请选择会议查看论文
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      <Dialog open={openDialog} onClose={handleClose} maxWidth="lg" fullWidth>
        <DialogTitle>新建投稿</DialogTitle>
        <DialogContent>
          <Box sx={{ pt: 2 }}>
            <TextField
              label="论文标题"
              value={formData.title}
              onChange={(e) => setFormData({ ...formData, title: e.target.value })}
              fullWidth
              required
              sx={{ mb: 2 }}
            />
            <TextField
              label="摘要"
              value={formData.abstract}
              onChange={(e) => setFormData({ ...formData, abstract: e.target.value })}
              fullWidth
              multiline
              rows={4}
              required
              helperText={`${formData.abstract.length}/500 字（建议200-500字）`}
              sx={{ mb: 2 }}
            />
            <TextField
              label="关键词"
              value={formData.keywords}
              onChange={(e) => setFormData({ ...formData, keywords: e.target.value })}
              fullWidth
              helperText="用逗号分隔，建议3-5个关键词"
              sx={{ mb: 2 }}
            />
            <TextField
              label="主题领域"
              value={formData.topic_area}
              onChange={(e) => setFormData({ ...formData, topic_area: e.target.value })}
              fullWidth
              sx={{ mb: 2 }}
            />
            <TextField
              label="论文正文"
              value={formData.content}
              onChange={(e) => setFormData({ ...formData, content: e.target.value })}
              fullWidth
              multiline
              rows={6}
              sx={{ mb: 3 }}
            />

            <Typography variant="h6" sx={{ mb: 2 }}>作者信息</Typography>
            {formData.authors.map((author, index) => (
              <Paper key={index} sx={{ p: 2, mb: 2 }}>
                <Box sx={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr auto auto', gap: 2, alignItems: 'center' }}>
                  <TextField
                    label="姓名"
                    value={author.name}
                    onChange={(e) => handleAuthorChange(index, 'name', e.target.value)}
                    required
                  />
                  <TextField
                    label="邮箱"
                    type="email"
                    value={author.email}
                    onChange={(e) => handleAuthorChange(index, 'email', e.target.value)}
                    required
                  />
                  <TextField
                    label="单位"
                    value={author.affiliation}
                    onChange={(e) => handleAuthorChange(index, 'affiliation', e.target.value)}
                  />
                  <FormControlLabel
                    control={
                      <Checkbox
                        checked={author.is_first_author}
                        onChange={(e) => handleAuthorChange(index, 'is_first_author', e.target.checked)}
                      />
                    }
                    label="第一作者"
                  />
                  <FormControlLabel
                    control={
                      <Checkbox
                        checked={author.is_corresponding}
                        onChange={(e) => handleAuthorChange(index, 'is_corresponding', e.target.checked)}
                      />
                    }
                    label="通讯作者"
                  />
                  {formData.authors.length > 1 && (
                    <IconButton onClick={() => handleRemoveAuthor(index)} color="error">
                      <DeleteIcon />
                    </IconButton>
                  )}
                </Box>
              </Paper>
            ))}
            <Button onClick={handleAddAuthor} startIcon={<AddIcon />}>
              添加作者
            </Button>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose}>取消</Button>
          <Button onClick={handleSubmit} variant="contained">
            创建
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

export default SubmissionPage;
