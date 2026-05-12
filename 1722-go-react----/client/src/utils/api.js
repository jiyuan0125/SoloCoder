const API_BASE = 'http://localhost:8080/api';

async function request(path, options = {}) {
  const response = await fetch(API_BASE + path, {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    ...options,
    body: options.body ? JSON.stringify(options.body) : undefined,
  });

  const data = await response.json().catch(() => ({}));

  if (!response.ok) {
    throw new Error(data.error || `Request failed with status ${response.status}`);
  }

  return data;
}

export const api = {
  getKnowledgePoints: (subject) => request('/knowledge' + (subject ? `?subject=${encodeURIComponent(subject)}` : '')),
  createKnowledgePoint: (data) => request('/knowledge', { method: 'POST', body: data }),
  updateKnowledgePoint: (id, data) => request(`/knowledge/${id}`, { method: 'PUT', body: data }),
  deleteKnowledgePoint: (id) => request(`/knowledge/${id}`, { method: 'DELETE' }),

  getQuestions: (params = {}) => {
    const query = new URLSearchParams(params).toString();
    return request('/questions' + (query ? `?${query}` : ''));
  },
  createQuestion: (data) => request('/questions', { method: 'POST', body: data }),
  updateQuestion: (id, data) => request(`/questions/${id}`, { method: 'PUT', body: data }),
  deleteQuestion: (id) => request(`/questions/${id}`, { method: 'DELETE' }),
  submitQuestionForReview: (id) => request(`/questions/${id}/submit`, { method: 'POST' }),
  approveQuestion: (id) => request(`/questions/${id}/approve`, { method: 'POST' }),
  publishQuestion: (id) => request(`/questions/${id}/publish`, { method: 'POST' }),
  offlineQuestion: (id) => request(`/questions/${id}/offline`, { method: 'POST' }),
  rejectQuestion: (id) => request(`/questions/${id}/reject`, { method: 'POST' }),

  startExam: (studentId, billId) => request('/exams/start', {
    method: 'POST',
    body: { student_id: studentId, bill_id: billId },
  }),
  getCurrentQuestion: (examId) => request(`/exams/${examId}/current`),
  submitAnswer: (examId, answer, timeSpent) => request(`/exams/${examId}/submit`, {
    method: 'POST',
    body: { student_answer: answer, time_spent: timeSpent },
  }),
  markExamIncomplete: (examId) => request(`/exams/${examId}/incomplete`, { method: 'POST' }),
  getExamReport: (examId) => request(`/exams/${examId}/report`),

  getStudentReport: (studentId) => request(`/reports/student?student_id=${encodeURIComponent(studentId)}`),
  getExamHistory: (studentId) => request(`/reports/history?student_id=${encodeURIComponent(studentId)}`),

  getWrongAnswers: (studentId, knowledgePointId, sortBy = 'error_count') => {
    const params = new URLSearchParams({ student_id: studentId, sort_by: sortBy });
    if (knowledgePointId) params.set('knowledge_point_id', knowledgePointId);
    return request('/wrong-answers?' + params.toString());
  },
  addWrongAnswer: (data) => request('/wrong-answers', { method: 'POST', body: data }),
  updateWrongAnswer: (id, data) => request(`/wrong-answers/${id}`, { method: 'PUT', body: data }),
  deleteWrongAnswer: (id) => request(`/wrong-answers/${id}`, { method: 'DELETE' }),

  createBill: (data) => request('/bills', { method: 'POST', body: data }),
  getBills: (studentId) => request('/bills' + (studentId ? `?student_id=${encodeURIComponent(studentId)}` : '')),
  getBill: (id) => request(`/bills/${id}`),
  adjustBill: (id, newTotal) => request(`/bills/${id}/adjust`, { method: 'POST', body: { new_total: newTotal } }),
};
