import React, { useState, useEffect } from 'react';
import {
  Paper,
  Grid,
  List,
  ListItem,
  ListItemButton,
  ListItemText,
  Typography,
  Box,
  TextField,
  Button,
  Chip,
  Divider,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Rating,
  Snackbar,
  Alert,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
} from '@mui/material';
import { meetingAPI, reviewAPI, userAPI } from '../services/api';
import { REVIEW_RECOMMENDATIONS, REVIEW_RECOMMENDATION_LABELS } from '../types';

const ReviewPage = () => {
  const [meetings, setMeetings] = useState([]);
  const [selectedMeeting, setSelectedMeeting] = useState('');
  const [papers, setPapers] = useState([]);
  const [selectedPaper, setSelectedPaper] = useState(null);
  const [reviewers, setReviewers] = useState([]);
  const [assignDialogOpen, setAssignDialogOpen] = useState(false);
  const [selectedReviewer, setSelectedReviewer] = useState('');
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' });

  const [reviewForm, setReviewForm] = useState({
    originality: 0,
    technical_quality: 0,
    relevance: 0,
    clarity: 0,
    recommendation: '',
    comments: '',
    confidential_comments: '',
  });

  useEffect(() => {
    loadMeetings();
    loadReviewers();
  }, []);

  const loadMeetings = async () => {
    try {
      const response = await meetingAPI.getAll();
      setMeetings(response.data);
    } catch (error) {
      console.error('Failed to load meetings:', error);
    }
  };

  const loadReviewers = async () => {
    try {
      const response = await userAPI.getReviewers();
      setReviewers(response.data);
    } catch (error) {
      console.error('Failed to load reviewers:', error);
    }
  };

  const loadPapers = async (meetingId) => {
    if (!meetingId) return;
    try {
      const response = await meetingAPI.getPapers(meetingId);
      setPapers(response.data);
      setSelectedPaper(null);
    } catch (error) {
      console.error('Failed to load papers:', error);
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

  const handleSelectPaper = (paper) => {
    setSelectedPaper(paper);
    setReviewForm({
      originality: 0,
      technical_quality: 0,
      relevance: 0,
      clarity: 0,
      recommendation: '',
      comments: '',
      confidential_comments: '',
    });
  };

  const handleOpenAssign = () => {
    if (!selectedPaper) {
      showSnackbar('请先选择论文', 'warning');
      return;
    }
    setSelectedReviewer('');
    setAssignDialogOpen(true);
  };

  const handleAssignReviewer = async () => {
    if (!selectedReviewer) {
      showSnackbar('请选择审稿人', 'warning');
      return;
    }
    try {
      await reviewAPI.assign({
        paper_id: selectedPaper.id,
        reviewer_id: parseInt(selectedReviewer),
        meeting_id: parseInt(selectedMeeting),
      });
      setAssignDialogOpen(false);
      showSnackbar('审稿人分配成功', 'success');
      loadPapers(selectedMeeting);
    } catch (error) {
      console.error('Failed to assign reviewer:', error);
      showSnackbar(error.response?.data?.error || '分配失败', 'error');
    }
  };

  const handleSubmitReview = async () => {
    if (!selectedPaper) return;

    const activeReview = selectedPaper.reviews?.find(r => r.submitted_at === null);
    if (!activeReview) {
      showSnackbar('该论文没有待提交的评审', 'warning');
      return;
    }

    if (reviewForm.originality < 1 || reviewForm.technical_quality < 1 ||
        reviewForm.relevance < 1 || reviewForm.clarity < 1) {
      showSnackbar('请为所有评分项打分（1-5分）', 'warning');
      return;
    }
    if (!reviewForm.recommendation) {
      showSnackbar('请选择总体推荐', 'warning');
      return;
    }
    if (reviewForm.comments.length < 100) {
      showSnackbar('评审意见至少需要100字', 'warning');
      return;
    }

    try {
      await reviewAPI.submit(activeReview.id, reviewForm);
      showSnackbar('评审提交成功', 'success');
      loadPapers(selectedMeeting);
      setSelectedPaper(null);
    } catch (error) {
      console.error('Failed to submit review:', error);
      showSnackbar(error.response?.data?.error || '提交失败', 'error');
    }
  };

  const formatPaperTitle = (title, index) => `论文 ${index + 1}`;

  const formatReviewerName = (name, index) => {
    const labels = ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H'];
    return `审稿人${labels[index % labels.length]}`;
  };

  return (
    <Box>
      <Typography variant="h5" sx={{ mb: 2 }}>审稿工作台</Typography>

      <Paper sx={{ p: 2, mb: 2 }}>
        <TextField
          label="选择会议"
          select
          value={selectedMeeting}
          onChange={handleMeetingChange}
          fullWidth
          helperText="请选择要管理审稿的会议"
        >
          <MenuItem value="">-- 请选择会议 --</MenuItem>
          {meetings.map((m) => (
            <MenuItem key={m.id} value={m.id.toString()}>
              {m.name} ({m.abbreviation})
            </MenuItem>
          ))}
        </TextField>
      </Paper>

      <Grid container spacing={2}>
        <Grid item xs={4}>
          <Paper sx={{ height: 'calc(100vh - 300px)', overflow: 'auto' }}>
            <Box sx={{ p: 2, borderBottom: 1, borderColor: 'divider' }}>
              <Typography variant="h6">待审论文列表</Typography>
              <Button
                variant="outlined"
                size="small"
                onClick={handleOpenAssign}
                disabled={!selectedPaper}
                sx={{ mt: 1 }}
              >
                分配审稿人
              </Button>
            </Box>
            <List>
              {papers.map((paper, index) => (
                <React.Fragment key={paper.id}>
                  <ListItem disablePadding>
                    <ListItemButton
                      selected={selectedPaper?.id === paper.id}
                      onClick={() => handleSelectPaper(paper)}
                    >
                      <ListItemText
                        primary={formatPaperTitle(paper.title, index)}
                        secondary={
                          <Box>
                            <Chip
                              label={paper.paper_number}
                              size="small"
                              variant="outlined"
                              sx={{ mr: 1 }}
                            />
                            <Chip
                              label={`${paper.reviews?.filter(r => r.submitted_at).length}/${paper.reviews?.length || 0} 已审`}
                              size="small"
                              color={paper.reviews?.every(r => r.submitted_at) ? 'success' : 'default'}
                            />
                          </Box>
                        }
                      />
                    </ListItemButton>
                  </ListItem>
                  <Divider />
                </React.Fragment>
              ))}
              {papers.length === 0 && selectedMeeting && (
                <ListItem>
                  <ListItemText primary="暂无论文" />
                </ListItem>
              )}
            </List>
          </Paper>
        </Grid>

        <Grid item xs={8}>
          <Paper sx={{ p: 3, height: 'calc(100vh - 300px)', overflow: 'auto' }}>
            {selectedPaper ? (
              <Box>
                <Typography variant="h6" sx={{ mb: 2 }}>
                  {formatPaperTitle(selectedPaper.title, papers.indexOf(selectedPaper))}
                </Typography>

                <Typography variant="subtitle2" color="text.secondary" sx={{ mb: 1 }}>
                  论文编号: {selectedPaper.paper_number}
                </Typography>
                <Typography variant="subtitle2" color="text.secondary" sx={{ mb: 2 }}>
                  主题领域: {selectedPaper.topic_area || '-'}
                </Typography>

                <Box sx={{ mb: 3 }}>
                  <Typography variant="subtitle1" sx={{ mb: 1 }}>摘要</Typography>
                  <Paper sx={{ p: 2, bgcolor: '#fafafa' }}>
                    <Typography variant="body2">{selectedPaper.abstract}</Typography>
                  </Paper>
                </Box>

                <Box sx={{ mb: 3 }}>
                  <Typography variant="subtitle1" sx={{ mb: 1 }}>关键词</Typography>
                  <Typography variant="body2">{selectedPaper.keywords || '-'}</Typography>
                </Box>

                <Divider sx={{ my: 3 }} />

                <Typography variant="h6" sx={{ mb: 2 }}>评审表单</Typography>

                <Box sx={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 3, mb: 3 }}>
                  <Box>
                    <Typography component="legend">原创性</Typography>
                    <Rating
                      value={reviewForm.originality}
                      onChange={(e, newValue) => setReviewForm({ ...reviewForm, originality: newValue })}
                      max={5}
                    />
                    <Typography variant="caption" color="text.secondary">1-5分</Typography>
                  </Box>
                  <Box>
                    <Typography component="legend">技术质量</Typography>
                    <Rating
                      value={reviewForm.technical_quality}
                      onChange={(e, newValue) => setReviewForm({ ...reviewForm, technical_quality: newValue })}
                      max={5}
                    />
                    <Typography variant="caption" color="text.secondary">1-5分</Typography>
                  </Box>
                  <Box>
                    <Typography component="legend">相关性</Typography>
                    <Rating
                      value={reviewForm.relevance}
                      onChange={(e, newValue) => setReviewForm({ ...reviewForm, relevance: newValue })}
                      max={5}
                    />
                    <Typography variant="caption" color="text.secondary">1-5分</Typography>
                  </Box>
                  <Box>
                    <Typography component="legend">清晰度</Typography>
                    <Rating
                      value={reviewForm.clarity}
                      onChange={(e, newValue) => setReviewForm({ ...reviewForm, clarity: newValue })}
                      max={5}
                    />
                    <Typography variant="caption" color="text.secondary">1-5分</Typography>
                  </Box>
                </Box>

                <FormControl fullWidth sx={{ mb: 3 }}>
                  <InputLabel>总体推荐</InputLabel>
                  <Select
                    value={reviewForm.recommendation}
                    label="总体推荐"
                    onChange={(e) => setReviewForm({ ...reviewForm, recommendation: e.target.value })}
                  >
                    {Object.entries(REVIEW_RECOMMENDATION_LABELS).map(([value, label]) => (
                      <MenuItem key={value} value={value}>{label}</MenuItem>
                    ))}
                  </Select>
                </FormControl>

                <TextField
                  label="评审意见"
                  value={reviewForm.comments}
                  onChange={(e) => setReviewForm({ ...reviewForm, comments: e.target.value })}
                  fullWidth
                  multiline
                  rows={4}
                  helperText={`${reviewForm.comments.length} 字（至少100字）`}
                  sx={{ mb: 3 }}
                />

                <TextField
                  label="机密意见（仅程序主席可见）"
                  value={reviewForm.confidential_comments}
                  onChange={(e) => setReviewForm({ ...reviewForm, confidential_comments: e.target.value })}
                  fullWidth
                  multiline
                  rows={2}
                  sx={{ mb: 3 }}
                />

                <Button
                  variant="contained"
                  onClick={handleSubmitReview}
                  disabled={!selectedPaper.reviews?.some(r => r.submitted_at === null)}
                >
                  提交评审
                </Button>

                {selectedPaper.reviews && selectedPaper.reviews.length > 0 && (
                  <Box sx={{ mt: 4 }}>
                    <Typography variant="h6" sx={{ mb: 2 }}>已分配审稿人</Typography>
                    {selectedPaper.reviews.map((review, index) => (
                      <Paper key={review.id} sx={{ p: 2, mb: 2 }}>
                        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                          <Typography>
                            {formatReviewerName(review.reviewer?.name || '', index)}
                          </Typography>
                          <Chip
                            label={review.submitted_at ? '已提交' : '待评审'}
                            color={review.submitted_at ? 'success' : 'warning'}
                            size="small"
                          />
                        </Box>
                        {review.submitted_at && review.recommendation && (
                          <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                            推荐: {REVIEW_RECOMMENDATION_LABELS[review.recommendation]}
                          </Typography>
                        )}
                      </Paper>
                    ))}
                  </Box>
                )}
              </Box>
            ) : (
              <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}>
                <Typography color="text.secondary">请从左侧选择一篇论文</Typography>
              </Box>
            )}
          </Paper>
        </Grid>
      </Grid>

      <Dialog open={assignDialogOpen} onClose={() => setAssignDialogOpen(false)}>
        <DialogTitle>分配审稿人</DialogTitle>
        <DialogContent>
          <TextField
            label="选择审稿人"
            select
            value={selectedReviewer}
            onChange={(e) => setSelectedReviewer(e.target.value)}
            fullWidth
            sx={{ mt: 2, minWidth: 300 }}
          >
            <MenuItem value="">-- 请选择审稿人 --</MenuItem>
            {reviewers.map((r) => (
              <MenuItem key={r.id} value={r.id.toString()}>
                {r.name} ({r.email})
              </MenuItem>
            ))}
          </TextField>
          <Typography variant="caption" color="text.secondary" sx={{ mt: 2, display: 'block' }}>
            每位审稿人同一轮最多审8篇论文
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setAssignDialogOpen(false)}>取消</Button>
          <Button onClick={handleAssignReviewer} variant="contained">
            分配
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

export default ReviewPage;
