import React, { useState, useEffect } from 'react';
import { api } from '../utils/api';

const QUESTION_TYPES = [
  { value: 'single_choice', label: '单选题' },
  { value: 'multiple_choice', label: '多选题' },
  { value: 'true_false', label: '判断题' },
  { value: 'fill_blank', label: '填空题' },
  { value: 'essay', label: '简答题' },
];

const STATUS_BADGES = {
  draft: 'badge-draft',
  pending_review: 'badge-pending',
  approved: 'badge-approved',
  published: 'badge-published',
  offline: 'badge-offline',
  rejected: 'badge-rejected',
};

const STATUS_LABELS = {
  draft: '草稿',
  pending_review: '待审核',
  approved: '已审核',
  published: '已上架',
  offline: '已下架',
  rejected: '已退回',
};

function QuestionManagement() {
  const [questions, setQuestions] = useState([]);
  const [knowledgePoints, setKnowledgePoints] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [filterStatus, setFilterStatus] = useState('');
  const [filterType, setFilterType] = useState('');
  const [filterDifficulty, setFilterDifficulty] = useState('');

  const [formData, setFormData] = useState({
    question_number: '',
    content: '',
    type: 'single_choice',
    knowledge_point_id: '',
    difficulty: 3,
    correct_answer: '',
    options: '',
    explanation: '',
    score: 1,
  });

  const loadData = async () => {
    try {
      setLoading(true);
      const [kpData, qData] = await Promise.all([
        api.getKnowledgePoints(),
        api.getQuestions({
          status: filterStatus,
          type: filterType,
          difficulty: filterDifficulty,
        }),
      ]);
      setKnowledgePoints(kpData.data || []);
      setQuestions(qData.data || []);
    } catch (error) {
      alert(error.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [filterStatus, filterType, filterDifficulty]);

  const flattenKP = (nodes, result = []) => {
    nodes.forEach((node) => {
      if (!node.children || node.children.length === 0) {
        result.push(node);
      }
      if (node.children) {
        flattenKP(node.children, result);
      }
    });
    return result;
  };

  const leafKP = flattenKP(knowledgePoints);

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      await api.createQuestion({
        ...formData,
        knowledge_point_id: parseInt(formData.knowledge_point_id),
        difficulty: parseInt(formData.difficulty),
        score: parseFloat(formData.score),
      });
      setShowModal(false);
      setFormData({
        question_number: '',
        content: '',
        type: 'single_choice',
        knowledge_point_id: '',
        difficulty: 3,
        correct_answer: '',
        options: '',
        explanation: '',
        score: 1,
      });
      loadData();
    } catch (error) {
      alert(error.message);
    }
  };

  const handleStatusAction = async (id, action) => {
    try {
      switch (action) {
        case 'submit':
          await api.submitQuestionForReview(id);
          break;
        case 'approve':
          await api.approveQuestion(id);
          break;
        case 'publish':
          await api.publishQuestion(id);
          break;
        case 'offline':
          await api.offlineQuestion(id);
          break;
        case 'reject':
          await api.rejectQuestion(id);
          break;
      }
      loadData();
    } catch (error) {
      alert(error.message);
    }
  };

  const getActions = (question) => {
    const actions = [];
    switch (question.status) {
      case 'draft':
      case 'rejected':
        actions.push({ action: 'submit', label: '提交审核', class: 'btn-primary' });
        break;
      case 'pending_review':
        actions.push({ action: 'approve', label: '通过', class: 'btn-success' });
        actions.push({ action: 'reject', label: '退回', class: 'btn-danger' });
        break;
      case 'approved':
        actions.push({ action: 'publish', label: '上架', class: 'btn-success' });
        break;
      case 'published':
        actions.push({ action: 'offline', label: '下架', class: 'btn-warning' });
        break;
    }
    return actions;
  };

  return (
    <div>
      <div className="card">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
          <h2>题库管理</h2>
          <button className="btn btn-primary" onClick={() => setShowModal(true)}>
            + 添加题目
          </button>
        </div>

        <div style={{ display: 'flex', gap: '1rem', marginBottom: '1.5rem', flexWrap: 'wrap' }}>
          <select
            value={filterStatus}
            onChange={(e) => setFilterStatus(e.target.value)}
            style={{ padding: '0.5rem' }}
          >
            <option value="">全部状态</option>
            {Object.entries(STATUS_LABELS).map(([value, label]) => (
              <option key={value} value={value}>{label}</option>
            ))}
          </select>
          <select
            value={filterType}
            onChange={(e) => setFilterType(e.target.value)}
            style={{ padding: '0.5rem' }}
          >
            <option value="">全部类型</option>
            {QUESTION_TYPES.map((t) => (
              <option key={t.value} value={t.value}>{t.label}</option>
            ))}
          </select>
          <select
            value={filterDifficulty}
            onChange={(e) => setFilterDifficulty(e.target.value)}
            style={{ padding: '0.5rem' }}
          >
            <option value="">全部难度</option>
            {[1, 2, 3, 4, 5].map((d) => (
              <option key={d} value={d}>难度 {d}</option>
            ))}
          </select>
        </div>

        {loading ? (
          <div className="empty-state">
            <p>加载中...</p>
          </div>
        ) : questions.length === 0 ? (
          <div className="empty-state">
            <p>暂无题目</p>
            <button className="btn btn-primary" onClick={() => setShowModal(true)}>
              添加第一道题目
            </button>
          </div>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>编号</th>
                <th>内容</th>
                <th>类型</th>
                <th>知识点</th>
                <th>难度</th>
                <th>分值</th>
                <th>状态</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {questions.map((q) => (
                <tr key={q.id}>
                  <td style={{ fontWeight: '600' }}>{q.question_number}</td>
                  <td style={{ maxWidth: '300px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {q.content}
                  </td>
                  <td>{QUESTION_TYPES.find(t => t.value === q.type)?.label || q.type}</td>
                  <td>{q.knowledge_point?.name || '-'}</td>
                  <td>L{q.difficulty}</td>
                  <td>{q.score}</td>
                  <td>
                    <span className={`badge ${STATUS_BADGES[q.status]}`}>
                      {STATUS_LABELS[q.status]}
                    </span>
                  </td>
                  <td>
                    {getActions(q).map(({ action, label, class: cls }) => (
                      <button
                        key={action}
                        className={`btn ${cls}`}
                        style={{ fontSize: '0.8rem', padding: '0.25rem 0.5rem' }}
                        onClick={() => handleStatusAction(q.id, action)}
                      >
                        {label}
                      </button>
                    ))}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {showModal && (
        <div className="modal-backdrop" onClick={() => setShowModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h3>添加题目</h3>
            <form onSubmit={handleSubmit}>
              <div className="form-group">
                <label>题目编号 *</label>
                <input
                  type="text"
                  value={formData.question_number}
                  onChange={(e) => setFormData({ ...formData, question_number: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label>题目类型</label>
                <select
                  value={formData.type}
                  onChange={(e) => setFormData({ ...formData, type: e.target.value })}
                >
                  {QUESTION_TYPES.map((t) => (
                    <option key={t.value} value={t.value}>{t.label}</option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label>关联知识点（叶子节点） *</label>
                <select
                  value={formData.knowledge_point_id}
                  onChange={(e) => setFormData({ ...formData, knowledge_point_id: e.target.value })}
                  required
                >
                  <option value="">请选择</option>
                  {leafKP.map((kp) => (
                    <option key={kp.id} value={kp.id}>
                      {kp.subject} - {kp.name}
                    </option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label>难度等级 (1-5)</label>
                <select
                  value={formData.difficulty}
                  onChange={(e) => setFormData({ ...formData, difficulty: e.target.value })}
                >
                  {[1, 2, 3, 4, 5].map((d) => (
                    <option key={d} value={d}>L{d}</option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label>题目内容 *</label>
                <textarea
                  value={formData.content}
                  onChange={(e) => setFormData({ ...formData, content: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label>选项（JSON格式，如 ["A", "B", "C", "D"]）</label>
                <textarea
                  value={formData.options}
                  onChange={(e) => setFormData({ ...formData, options: e.target.value })}
                  placeholder='["A. 选项1", "B. 选项2"]'
                />
              </div>
              <div className="form-group">
                <label>正确答案 *</label>
                <input
                  type="text"
                  value={formData.correct_answer}
                  onChange={(e) => setFormData({ ...formData, correct_answer: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label>解析</label>
                <textarea
                  value={formData.explanation}
                  onChange={(e) => setFormData({ ...formData, explanation: e.target.value })}
                />
              </div>
              <div className="form-group">
                <label>分值</label>
                <input
                  type="number"
                  min="0"
                  step="0.5"
                  value={formData.score}
                  onChange={(e) => setFormData({ ...formData, score: e.target.value })}
                />
              </div>
              <div className="modal-actions">
                <button type="button" className="btn btn-secondary" onClick={() => setShowModal(false)}>
                  取消
                </button>
                <button type="submit" className="btn btn-primary">
                  保存
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

export default QuestionManagement;
