import React, { useState, useEffect } from 'react';
import api from '../api';
import { useAuth } from '../context/AuthContext';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  BarChart,
  Bar,
} from 'recharts';

const ScoresPage = () => {
  const { user, isAdmin, isTeacher, isParent } = useAuth();
  const [selectedChild, setSelectedChild] = useState(null);
  const [selectedSemester, setSelectedSemester] = useState('');
  const [scores, setScores] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showRecordModal, setShowRecordModal] = useState(false);
  const [showHistoryModal, setShowHistoryModal] = useState(false);
  const [historySubject, setHistorySubject] = useState(null);
  const [historyData, setHistoryData] = useState([]);
  const [selectedSubject, setSelectedSubject] = useState(null);

  const [recordForm, setRecordForm] = useState({
    class: '',
    subject: '',
    semester: '',
    exam_date: '',
    scores: [],
  });

  const [classStudents, setClassStudents] = useState([]);
  const [classStats, setClassStats] = useState(null);

  useEffect(() => {
    if (isParent && user?.children?.length > 0) {
      setSelectedChild(user.children[0]);
    }
  }, [user]);

  useEffect(() => {
    if (isParent && selectedChild) {
      fetchStudentScores();
    }
  }, [selectedChild, selectedSemester]);

  const fetchStudentScores = async () => {
    if (!selectedChild) return;
    setLoading(true);
    try {
      const params = selectedSemester ? { semester: selectedSemester } : {};
      const response = await api.get(`/scores/student/${selectedChild.id}`, { params });
      setScores(response.data.scores);
    } catch (error) {
      console.error('获取成绩失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchClassStudents = async () => {
    if (!recordForm.class) return;
    try {
      const response = await api.get(`/scores/class/${recordForm.class}/students`);
      const students = response.data.students.map((s) => ({
        ...s,
        score: null,
      }));
      setClassStudents(students);
      setRecordForm({ ...recordForm, scores: [] });
    } catch (error) {
      console.error('获取班级学生失败:', error);
    }
  };

  const fetchClassStats = async () => {
    if (!recordForm.class || !recordForm.subject || !recordForm.semester) return;
    try {
      const response = await api.get('/scores/class-stats', {
        params: {
          class: recordForm.class,
          subject: recordForm.subject,
          semester: recordForm.semester,
        },
      });
      setClassStats(response.data);
    } catch (error) {
      console.error('获取班级统计失败:', error);
    }
  };

  useEffect(() => {
    fetchClassStats();
  }, [recordForm.class, recordForm.subject, recordForm.semester]);

  const handleViewHistory = async (subject) => {
    setHistorySubject(subject);
    try {
      const response = await api.get(
        `/scores/student/${selectedChild.id}/subject/${subject}`
      );
      const history = response.data.history.map((h) => ({
        ...h,
        examDate: new Date(h.exam_date).toLocaleDateString(),
        score: h.score,
      }));
      setHistoryData(history);
      setShowHistoryModal(true);
    } catch (error) {
      console.error('获取历史成绩失败:', error);
    }
  };

  const handleStudentScoreChange = (studentId, value) => {
    const numValue = value === '' ? null : parseFloat(value);
    const newStudents = classStudents.map((s) =>
      s.id === studentId ? { ...s, score: numValue } : s
    );
    setClassStudents(newStudents);
  };

  const handleRecordSubmit = async (e) => {
    e.preventDefault();
    try {
      const scoresToSubmit = classStudents.map((s) => ({
        student_id: s.id,
        score: s.score,
      }));

      await api.post('/scores/batch', {
        class: recordForm.class,
        subject: recordForm.subject,
        semester: recordForm.semester,
        exam_date: recordForm.exam_date,
        scores: scoresToSubmit,
      });

      alert('成绩录入成功！');
      fetchClassStats();
    } catch (error) {
      alert(error.response?.data?.error || '录入失败');
    }
  };

  const getScoreColor = (score) => {
    if (score === null) return '#999';
    if (score >= 90) return '#16a34a';
    if (score >= 80) return '#2563eb';
    if (score >= 60) return '#d97706';
    return '#dc2626';
  };

  const getScoreLabel = (score) => {
    if (score === null) return '缺考';
    if (score >= 90) return '优秀';
    if (score >= 80) return '良好';
    if (score >= 70) return '中等';
    if (score >= 60) return '及格';
    return '不及格';
  };

  if (loading && isParent && selectedChild) {
    return <div style={{ padding: '40px', textAlign: 'center' }}>加载中...</div>;
  }

  return (
    <div style={styles.container}>
      <div style={styles.header}>
        <h2 style={styles.title}>📊 成绩管理</h2>
        {(isAdmin || isTeacher) && (
          <button
            style={styles.createButton}
            onClick={() => setShowRecordModal(true)}
          >
            + 录入成绩
          </button>
        )}
      </div>

      {isParent && (
        <>
          <div style={styles.filterBar}>
            {user?.children?.length > 1 && (
              <div style={styles.filterItem}>
                <label>选择孩子：</label>
                <select
                  style={styles.select}
                  value={selectedChild?.id || ''}
                  onChange={(e) => {
                    const child = user.children.find((c) => c.id === e.target.value);
                    setSelectedChild(child);
                  }}
                >
                  {user.children.map((child) => (
                    <option key={child.id} value={child.id}>
                      {child.name} ({child.class})
                    </option>
                  ))}
                </select>
              </div>
            )}
            <div style={styles.filterItem}>
              <label>学期：</label>
              <select
                style={styles.select}
                value={selectedSemester}
                onChange={(e) => setSelectedSemester(e.target.value)}
              >
                <option value="">全部学期</option>
                <option value="2024-2025上学期">2024-2025上学期</option>
                <option value="2024-2025下学期">2024-2025下学期</option>
              </select>
            </div>
          </div>

          {selectedChild && scores.length === 0 ? (
            <div style={styles.empty}>暂无成绩记录</div>
          ) : (
            <div style={styles.scoreGrid}>
              {scores.map((score) => (
                <div
                  key={score.id}
                  style={styles.scoreCard}
                  onClick={() => handleViewHistory(score.subject)}
                >
                  <div style={styles.scoreHeader}>
                    <span style={styles.subjectName}>{score.subject}</span>
                    <span style={styles.semester}>{score.semester}</span>
                  </div>
                  <div style={styles.scoreMain}>
                    <span
                      style={{ ...styles.scoreValue, color: getScoreColor(score.score) }}
                    >
                      {score.score !== null ? score.score : '-'}
                    </span>
                    <span style={styles.scoreLabel}>
                      {score.score !== null ? getScoreLabel(score.score) : '缺考'}
                    </span>
                  </div>
                  <div style={styles.scoreFooter}>
                    <span>点击查看趋势</span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </>
      )}

      {(isAdmin || isTeacher) && !showRecordModal && (
        <div style={styles.teacherTips}>
          <p>点击右上角"录入成绩"按钮批量录入学生成绩</p>
        </div>
      )}

      {showRecordModal && (
        <div
          style={styles.modalOverlay}
          onClick={() => setShowRecordModal(false)}
        >
          <div style={styles.largeModal} onClick={(e) => e.stopPropagation()}>
            <h3 style={styles.modalTitle}>批量录入成绩</h3>

            <form onSubmit={handleRecordSubmit} style={styles.form}>
              <div style={styles.row}>
                <div style={styles.formGroup}>
                  <label style={styles.label}>班级 *</label>
                  <input
                    type="text"
                    style={styles.input}
                    value={recordForm.class}
                    onChange={(e) =>
                      setRecordForm({ ...recordForm, class: e.target.value })
                    }
                    onBlur={fetchClassStudents}
                    placeholder="如：一年级1班"
                    required
                  />
                </div>
                <div style={styles.formGroup}>
                  <label style={styles.label}>学科 *</label>
                  <select
                    style={styles.select}
                    value={recordForm.subject}
                    onChange={(e) =>
                      setRecordForm({ ...recordForm, subject: e.target.value })
                    }
                    required
                  >
                    <option value="">请选择</option>
                    <option value="语文">语文</option>
                    <option value="数学">数学</option>
                    <option value="英语">英语</option>
                    <option value="物理">物理</option>
                    <option value="化学">化学</option>
                  </select>
                </div>
              </div>

              <div style={styles.row}>
                <div style={styles.formGroup}>
                  <label style={styles.label}>学期 *</label>
                  <select
                    style={styles.select}
                    value={recordForm.semester}
                    onChange={(e) =>
                      setRecordForm({ ...recordForm, semester: e.target.value })
                    }
                    required
                  >
                    <option value="">请选择</option>
                    <option value="2024-2025上学期">2024-2025上学期</option>
                    <option value="2024-2025下学期">2024-2025下学期</option>
                  </select>
                </div>
                <div style={styles.formGroup}>
                  <label style={styles.label}>考试日期 *</label>
                  <input
                    type="date"
                    style={styles.input}
                    value={recordForm.exam_date}
                    onChange={(e) =>
                      setRecordForm({ ...recordForm, exam_date: e.target.value })
                    }
                    required
                  />
                </div>
              </div>

              {classStudents.length > 0 && (
                <div style={styles.tableContainer}>
                  <table style={styles.table}>
                    <thead>
                      <tr>
                        <th style={styles.th}>学号</th>
                        <th style={styles.th}>姓名</th>
                        <th style={styles.th}>分数 (0-150)</th>
                      </tr>
                    </thead>
                    <tbody>
                      {classStudents.map((student) => (
                        <tr key={student.id}>
                          <td style={styles.td}>{student.student_no}</td>
                          <td style={styles.td}>{student.name}</td>
                          <td style={styles.td}>
                            <input
                              type="number"
                              style={styles.scoreInput}
                              value={student.score ?? ''}
                              onChange={(e) =>
                                handleStudentScoreChange(student.id, e.target.value)
                              }
                              placeholder="空为缺考"
                              min="0"
                              max="150"
                            />
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}

              {classStats && classStats.valid_scores > 0 && (
                <div style={styles.statsSection}>
                  <h4 style={styles.statsTitle}>本次录入班级统计</h4>
                  <div style={styles.statsGrid}>
                    <div style={styles.statCard}>
                      <div style={styles.statLabel}>平均分</div>
                      <div style={styles.statValue}>
                        {classStats.average?.toFixed(1) || '-'}
                      </div>
                    </div>
                    <div style={styles.statCard}>
                      <div style={styles.statLabel}>最高分</div>
                      <div style={styles.statValue}>{classStats.max_score || '-'}</div>
                    </div>
                    <div style={styles.statCard}>
                      <div style={styles.statLabel}>最低分</div>
                      <div style={styles.statValue}>{classStats.min_score || '-'}</div>
                    </div>
                    <div style={styles.statCard}>
                      <div style={styles.statLabel}>有效人数</div>
                      <div style={styles.statValue}>
                        {classStats.valid_scores}/{classStats.total_students}
                      </div>
                    </div>
                  </div>

                  {classStats.segments && (
                    <div style={styles.chartContainer}>
                      <ResponsiveContainer width="100%" height={200}>
                        <BarChart data={classStats.segments}>
                          <CartesianGrid strokeDasharray="3 3" />
                          <XAxis dataKey="label" />
                          <YAxis />
                          <Tooltip />
                          <Bar dataKey="count" fill="#4f46e5" />
                        </BarChart>
                      </ResponsiveContainer>
                    </div>
                  )}
                </div>
              )}

              <div style={styles.modalActions}>
                <button
                  type="button"
                  style={styles.cancelButton}
                  onClick={() => setShowRecordModal(false)}
                >
                  关闭
                </button>
                {classStudents.length > 0 && (
                  <button type="submit" style={styles.submitButton}>
                    录入成绩
                  </button>
                )}
              </div>
            </form>
          </div>
        </div>
      )}

      {showHistoryModal && historyData.length > 0 && (
        <div
          style={styles.modalOverlay}
          onClick={() => setShowHistoryModal(false)}
        >
          <div style={styles.largeModal} onClick={(e) => e.stopPropagation()}>
            <h3 style={styles.modalTitle}>
              {historySubject} 成绩趋势
              {selectedChild && ` - ${selectedChild.name}`}
            </h3>

            <div style={styles.chartContainer}>
              <ResponsiveContainer width="100%" height={300}>
                <LineChart data={historyData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="examDate" />
                  <YAxis domain={[0, 150]} />
                  <Tooltip />
                  <Legend />
                  <Line
                    type="monotone"
                    dataKey="score"
                    name="分数"
                    stroke="#4f46e5"
                    strokeWidth={2}
                    dot={{ r: 4 }}
                  />
                </LineChart>
              </ResponsiveContainer>
            </div>

            <table style={styles.table}>
              <thead>
                <tr>
                  <th style={styles.th}>考试日期</th>
                  <th style={styles.th}>学期</th>
                  <th style={styles.th}>分数</th>
                  <th style={styles.th}>等级</th>
                </tr>
              </thead>
              <tbody>
                {historyData.map((h) => (
                  <tr key={h.id}>
                    <td style={styles.td}>{h.examDate}</td>
                    <td style={styles.td}>{h.semester}</td>
                    <td style={styles.td}>
                      <span style={{ color: getScoreColor(h.score) }}>
                        {h.score !== null ? h.score : '缺考'}
                      </span>
                    </td>
                    <td style={styles.td}>{getScoreLabel(h.score)}</td>
                  </tr>
                ))}
              </tbody>
            </table>

            <div style={styles.modalActions}>
              <button
                style={styles.cancelButton}
                onClick={() => setShowHistoryModal(false)}
              >
                关闭
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

const styles = {
  container: {
    maxWidth: '1200px',
    margin: '0 auto',
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '24px',
  },
  title: {
    margin: 0,
    color: '#333',
  },
  createButton: {
    padding: '10px 20px',
    background: '#4f46e5',
    color: 'white',
    border: 'none',
    borderRadius: '8px',
    cursor: 'pointer',
    fontSize: '14px',
    fontWeight: 500,
  },
  filterBar: {
    display: 'flex',
    gap: '20px',
    marginBottom: '24px',
    background: 'white',
    padding: '16px 20px',
    borderRadius: '8px',
  },
  filterItem: {
    display: 'flex',
    alignItems: 'center',
    gap: '10px',
  },
  select: {
    padding: '8px 12px',
    border: '1px solid #ddd',
    borderRadius: '6px',
    fontSize: '14px',
    background: 'white',
  },
  empty: {
    textAlign: 'center',
    padding: '60px',
    color: '#999',
    background: 'white',
    borderRadius: '12px',
  },
  scoreGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fill, minmax(220px, 1fr))',
    gap: '16px',
  },
  scoreCard: {
    background: 'white',
    borderRadius: '12px',
    padding: '20px',
    cursor: 'pointer',
    boxShadow: '0 2px 8px rgba(0,0,0,0.06)',
    transition: 'transform 0.2s',
  },
  scoreHeader: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '16px',
  },
  subjectName: {
    fontSize: '16px',
    fontWeight: 600,
    color: '#333',
  },
  semester: {
    fontSize: '12px',
    color: '#999',
  },
  scoreMain: {
    textAlign: 'center',
    marginBottom: '12px',
  },
  scoreValue: {
    fontSize: '48px',
    fontWeight: 'bold',
  },
  scoreLabel: {
    display: 'block',
    fontSize: '14px',
    color: '#666',
    marginTop: '4px',
  },
  scoreFooter: {
    textAlign: 'center',
    fontSize: '12px',
    color: '#999',
    borderTop: '1px solid #eee',
    paddingTop: '12px',
  },
  teacherTips: {
    background: 'white',
    padding: '40px',
    borderRadius: '12px',
    textAlign: 'center',
    color: '#666',
  },
  modalOverlay: {
    position: 'fixed',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    background: 'rgba(0,0,0,0.5)',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    zIndex: 1000,
    padding: '20px',
  },
  largeModal: {
    background: 'white',
    borderRadius: '12px',
    padding: '30px',
    width: '100%',
    maxWidth: '900px',
    maxHeight: '90vh',
    overflowY: 'auto',
  },
  modalTitle: {
    margin: '0 0 24px 0',
    color: '#333',
  },
  form: {
    display: 'flex',
    flexDirection: 'column',
    gap: '16px',
  },
  formGroup: {
    display: 'flex',
    flexDirection: 'column',
    gap: '8px',
    flex: 1,
  },
  row: {
    display: 'grid',
    gridTemplateColumns: '1fr 1fr',
    gap: '16px',
  },
  label: {
    color: '#555',
    fontSize: '14px',
    fontWeight: 500,
  },
  input: {
    padding: '12px 16px',
    border: '1px solid #ddd',
    borderRadius: '8px',
    fontSize: '14px',
  },
  tableContainer: {
    maxHeight: '300px',
    overflowY: 'auto',
  },
  table: {
    width: '100%',
    borderCollapse: 'collapse',
  },
  th: {
    background: '#f8f9fa',
    padding: '12px',
    textAlign: 'left',
    borderBottom: '2px solid #eee',
    fontSize: '14px',
    color: '#666',
  },
  td: {
    padding: '10px 12px',
    borderBottom: '1px solid #eee',
    fontSize: '14px',
  },
  scoreInput: {
    width: '100px',
    padding: '8px',
    border: '1px solid #ddd',
    borderRadius: '6px',
    textAlign: 'center',
  },
  statsSection: {
    background: '#f8f9fa',
    padding: '20px',
    borderRadius: '12px',
    marginTop: '16px',
  },
  statsTitle: {
    margin: '0 0 16px 0',
    fontSize: '16px',
    color: '#333',
  },
  statsGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(4, 1fr)',
    gap: '12px',
    marginBottom: '20px',
  },
  statCard: {
    background: 'white',
    padding: '16px',
    borderRadius: '8px',
    textAlign: 'center',
  },
  statLabel: {
    fontSize: '12px',
    color: '#999',
    marginBottom: '4px',
  },
  statValue: {
    fontSize: '24px',
    fontWeight: 'bold',
    color: '#4f46e5',
  },
  chartContainer: {
    background: 'white',
    padding: '20px',
    borderRadius: '8px',
  },
  modalActions: {
    display: 'flex',
    gap: '12px',
    justifyContent: 'flex-end',
    marginTop: '20px',
  },
  cancelButton: {
    padding: '10px 24px',
    background: '#f5f5f5',
    border: 'none',
    borderRadius: '8px',
    cursor: 'pointer',
    fontSize: '14px',
    color: '#666',
  },
  submitButton: {
    padding: '10px 24px',
    background: '#4f46e5',
    color: 'white',
    border: 'none',
    borderRadius: '8px',
    cursor: 'pointer',
    fontSize: '14px',
  },
};

export default ScoresPage;
