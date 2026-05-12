import { useState, useEffect } from 'react';
import { Routes, Route, useNavigate, useParams } from 'react-router-dom';
import { api } from './api/api';
import type { Paper, User, PaperStatus, PublicationInfo, ReviewDecision, Author } from './types';
import { statusDisplayMap, statusColorMap, decisionDisplayMap, decisionColorMap } from './types';

function formatDate(dateStr: string): string {
  const d = new Date(dateStr);
  return d.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });
}

function roleDisplayName(role: string): string {
  const map: Record<string, string> = {
    author: '作者',
    editor: '编辑',
    reviewer: '审稿人',
    admin: '管理员',
  };
  return map[role] || role;
}

function authorsToString(authors: Author[]): string {
  return authors.map((a, i) => {
    let name = a.name;
    const badges: string[] = [];
    if (a.is_first) badges.push('一作');
    if (a.is_corresponding) badges.push('通讯');
    if (badges.length > 0) name += ` (${badges.join(', ')})`;
    if (i === 0) return name;
    return ', ' + name;
  }).join('');
}

function Header({ users, currentUserId, onUserChange }: {
  users: User[];
  currentUserId: string;
  onUserChange: (id: string) => void;
}) {
  const currentUser = users.find(u => u.id === currentUserId);
  return (
    <header className="app-header">
      <h1>📄 论文全生命周期管理平台</h1>
      <div className="user-selector">
        <span>当前身份：</span>
        <select value={currentUserId} onChange={e => onUserChange(e.target.value)}>
          {users.map(u => (
            <option key={u.id} value={u.id}>{u.name}</option>
          ))}
        </select>
        {currentUser && (
          <span className="badge" style={{ backgroundColor: '#f97316' }}>
            {roleDisplayName(currentUser.role)}
          </span>
        )}
      </div>
    </header>
  );
}

function PaperCard({ paper, onClick }: { paper: Paper; onClick: () => void }) {
  return (
    <div className="paper-card" onClick={onClick}>
      <div className="paper-card-header">
        <div className="paper-title">{paper.title}</div>
        <span
          className="badge"
          style={{ backgroundColor: statusColorMap[paper.status], whiteSpace: 'nowrap' }}
        >
          {statusDisplayMap[paper.status]}
        </span>
      </div>
      <div className="paper-meta">
        <span>👥 {authorsToString(paper.authors)}</span>
        <span>📅 提交时间：{formatDate(paper.created_at)}</span>
        {paper.target_journal && <span>📚 {paper.target_journal}</span>}
        {paper.subject_category && <span>🏷️ {paper.subject_category}</span>}
      </div>
      {paper.keywords.length > 0 && (
        <div className="paper-keywords">
          {paper.keywords.map((kw, i) => (
            <span key={i} className="keyword-tag">#{kw}</span>
          ))}
        </div>
      )}
    </div>
  );
}

function PaperListPage({ userId, onViewPaper, users }: {
  userId: string;
  onViewPaper: (id: string) => void;
  users: User[];
}) {
  const [papers, setPapers] = useState<Paper[]>([]);
  const [loading, setLoading] = useState(true);
  const [statusFilter, setStatusFilter] = useState('');
  const [keywordFilter, setKeywordFilter] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);

  const loadPapers = async () => {
    setLoading(true);
    try {
      const params: { status?: string; keyword?: string } = {};
      if (statusFilter) params.status = statusFilter;
      if (keywordFilter) params.keyword = keywordFilter;
      const data = await api.getPapers(userId, params);
      setPapers(data);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadPapers();
  }, [userId, statusFilter, keywordFilter]);

  return (
    <div>
      <div className="filter-bar">
        <select value={statusFilter} onChange={e => setStatusFilter(e.target.value)}>
          <option value="">全部状态</option>
          <option value="draft">草稿</option>
          <option value="pending_review">待审稿</option>
          <option value="in_review">审稿中</option>
          <option value="revision">修改中</option>
          <option value="accepted">已录用</option>
          <option value="published">已发表</option>
          <option value="rejected">已拒稿</option>
        </select>
        <input
          type="text"
          placeholder="🔍 搜索标题或关键词..."
          value={keywordFilter}
          onChange={e => setKeywordFilter(e.target.value)}
        />
        <button className="btn btn-primary" onClick={() => setShowCreateModal(true)}>
          + 新建论文
        </button>
      </div>

      {loading ? (
        <div className="loading">加载中...</div>
      ) : papers.length === 0 ? (
        <div className="empty-state">
          <h3>暂无论文</h3>
          <p>点击上方按钮创建第一篇论文</p>
        </div>
      ) : (
        <div className="paper-list">
          {papers.map(p => (
            <PaperCard key={p.id} paper={p} onClick={() => onViewPaper(p.id)} />
          ))}
        </div>
      )}

      {showCreateModal && (
        <CreatePaperModal
          userId={userId}
          users={users}
          onClose={() => setShowCreateModal(false)}
          onCreated={() => {
            setShowCreateModal(false);
            loadPapers();
          }}
        />
      )}
    </div>
  );
}

function CreatePaperModal({ userId, users, onClose, onCreated }: {
  userId: string;
  users: User[];
  onClose: () => void;
  onCreated: () => void;
}) {
  const [title, setTitle] = useState('');
  const [abstract, setAbstract] = useState('');
  const [keywords, setKeywords] = useState('');
  const [subjectCategory, setSubjectCategory] = useState('');
  const [targetJournal, setTargetJournal] = useState('');
  const [authors, setAuthors] = useState<Author[]>([
    { id: '', name: '', affiliation: '', is_first: true, is_corresponding: true },
  ]);
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const addAuthor = () => {
    setAuthors([...authors, {
      id: '', name: '', affiliation: '', is_first: false, is_corresponding: false
    }]);
  };

  const removeAuthor = (idx: number) => {
    if (authors.length > 1) {
      setAuthors(authors.filter((_, i) => i !== idx));
    }
  };

  const updateAuthor = (idx: number, field: keyof Author, value: any) => {
    const updated = [...authors];
    (updated[idx] as any)[field] = value;
    if (field === 'is_first' && value === true) {
      updated.forEach((a, i) => { if (i !== idx) a.is_first = false; });
    }
    if (field === 'is_corresponding' && value === true) {
      updated.forEach((a, i) => { if (i !== idx) a.is_corresponding = false; });
    }
    setAuthors(updated);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) { setError('请输入论文标题'); return; }
    const validAuthors = authors.filter(a => a.name.trim());
    if (validAuthors.length === 0) { setError('请至少添加一位作者'); return; }
    setSubmitting(true);
    setError('');
    try {
      await api.createPaper(userId, {
        title: title.trim(),
        abstract: abstract.trim(),
        keywords: keywords,
        authors: validAuthors.map((a, i) => ({ ...a, id: a.id || `a${i}` })),
        subject_category: subjectCategory.trim(),
        target_journal: targetJournal.trim(),
      });
      onCreated();
    } catch (err: any) {
      const msg = err.response?.data?.message || err.message || '创建失败';
      setError(msg);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h3>新建论文</h3>
          <button className="modal-close" onClick={onClose}>&times;</button>
        </div>
        <form onSubmit={handleSubmit}>
          <div className="modal-body">
            <div className="form-group">
              <label>论文标题 *</label>
              <input value={title} onChange={e => setTitle(e.target.value)} placeholder="请输入论文标题" />
            </div>
            <div className="form-group">
              <label>摘要</label>
              <textarea value={abstract} onChange={e => setAbstract(e.target.value)} placeholder="请输入论文摘要" />
            </div>
            <div className="form-group">
              <label>关键词（逗号分隔）</label>
              <input value={keywords} onChange={e => setKeywords(e.target.value)} placeholder="例如：深度学习, 神经网络, AI" />
            </div>
            <div className="form-group">
              <label>学科分类</label>
              <input value={subjectCategory} onChange={e => setSubjectCategory(e.target.value)} placeholder="例如：计算机科学" />
            </div>
            <div className="form-group">
              <label>投稿期刊</label>
              <input value={targetJournal} onChange={e => setTargetJournal(e.target.value)} placeholder="例如：Nature" />
            </div>
            <div className="form-group">
              <label>作者列表 *</label>
              <div className="authors-editor">
                {authors.map((a, i) => (
                  <div key={i} className="author-row">
                    <input placeholder="姓名" value={a.name} onChange={e => updateAuthor(i, 'name', e.target.value)} />
                    <input placeholder="单位" value={a.affiliation} onChange={e => updateAuthor(i, 'affiliation', e.target.value)} />
                    <div className="checkbox-group">
                      <label className="checkbox-label">
                        <input type="checkbox" checked={a.is_first} onChange={e => updateAuthor(i, 'is_first', e.target.checked)} /> 第一作者
                      </label>
                      <label className="checkbox-label">
                        <input type="checkbox" checked={a.is_corresponding} onChange={e => updateAuthor(i, 'is_corresponding', e.target.checked)} /> 通讯作者
                      </label>
                    </div>
                    {authors.length > 1 && (
                      <button type="button" className="btn btn-danger btn-sm" onClick={() => removeAuthor(i)}>删除</button>
                    )}
                  </div>
                ))}
                <button type="button" className="btn btn-outline btn-sm" onClick={addAuthor}>+ 添加作者</button>
              </div>
            </div>
            {error && <div className="error-message">{error}</div>}
          </div>
          <div className="modal-footer">
            <button type="button" className="btn btn-outline" onClick={onClose} disabled={submitting}>取消</button>
            <button type="submit" className="btn btn-primary" disabled={submitting}>
              {submitting ? '提交中...' : '创建论文'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

function PaperDetailPage({ userId, paperId, onBack, users, onReload }: {
  userId: string;
  paperId: string;
  onBack: () => void;
  users: User[];
  onReload: () => void;
}) {
  const [paper, setPaper] = useState<Paper | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showReviewerModal, setShowReviewerModal] = useState(false);
  const [showReviewModal, setShowReviewModal] = useState(false);
  const [showPublishModal, setShowPublishModal] = useState(false);
  const [message, setMessage] = useState('');

  const currentUser = users.find(u => u.id === userId);

  const loadPaper = async () => {
    setLoading(true);
    try {
      const data = await api.getPaper(userId, paperId);
      setPaper(data);
    } catch (err: any) {
      setError(err.response?.data?.message || '加载失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadPaper();
  }, [userId, paperId]);

  const showMessage = (msg: string) => {
    setMessage(msg);
    setTimeout(() => setMessage(''), 3000);
  };

  const handleSubmitReview = async () => {
    try {
      await api.submitForReview(userId, paperId);
      showMessage('已提交审稿');
      loadPaper();
    } catch (err: any) {
      showMessage('操作失败：' + (err.response?.data?.message || err.message));
    }
  };

  const handleApprove = async () => {
    try {
      await api.approveAction(userId, paperId);
      showMessage('已通过');
      loadPaper();
    } catch (err: any) {
      showMessage('操作失败：' + (err.response?.data?.message || err.message));
    }
  };

  const handleReject = async () => {
    try {
      await api.rejectAction(userId, paperId);
      showMessage('已拒稿');
      loadPaper();
    } catch (err: any) {
      showMessage('操作失败：' + (err.response?.data?.message || err.message));
    }
  };

  const handleCancel = async () => {
    try {
      await api.cancelAction(userId, paperId);
      showMessage('已撤回');
      loadPaper();
    } catch (err: any) {
      showMessage('操作失败：' + (err.response?.data?.message || err.message));
    }
  };

  if (loading) return <div className="loading">加载中...</div>;
  if (error) return <div className="error-message" style={{ padding: 40, textAlign: 'center' }}>{error}</div>;
  if (!paper) return null;

  const canSubmit = ['draft', 'revision'].includes(paper.status);
  const canAssign = paper.status === 'pending_review' && (currentUser?.role === 'editor' || currentUser?.role === 'admin');
  const canReview = paper.status === 'in_review' && (currentUser?.role === 'reviewer' || currentUser?.role === 'admin');
  const canPublish = paper.status === 'accepted' && currentUser?.role === 'admin';
  const canManualReview = paper.status === 'in_review' && (currentUser?.role === 'editor' || currentUser?.role === 'admin');
  const canCancel = ['pending_review', 'in_review'].includes(paper.status);

  const sortedHistory = [...paper.version_history].sort((a, b) =>
    new Date(b.changed_at).getTime() - new Date(a.changed_at).getTime()
  );

  return (
    <div>
      <div className="back-link" onClick={onBack}>← 返回列表</div>

      {message && (
        <div style={{ padding: 12, marginBottom: 16, borderRadius: 8, backgroundColor: '#dbeafe', color: '#1d4ed8' }}>
          {message}
        </div>
      )}

      <div className="paper-detail-page">
        <div className="paper-detail-header">
          <h2>{paper.title}</h2>
          <div style={{ display: 'flex', alignItems: 'center', gap: 12, flexWrap: 'wrap' }}>
            <span
              className="badge"
              style={{ backgroundColor: statusColorMap[paper.status], fontSize: 14, padding: '6px 14px' }}
            >
              {statusDisplayMap[paper.status]}
            </span>
            <span style={{ color: '#6b7280', fontSize: 13 }}>创建时间：{formatDate(paper.created_at)}</span>
          </div>

          <div className="paper-actions">
            {canSubmit && <button className="btn btn-primary" onClick={handleSubmitReview}>📤 提交审稿</button>}
            {canAssign && <button className="btn btn-warning" onClick={() => setShowReviewerModal(true)}>👥 分配审稿人</button>}
            {canReview && <button className="btn btn-success" onClick={() => setShowReviewModal(true)}>📝 提交评审</button>}
            {canManualReview && (
              <>
                <button className="btn btn-success btn-sm" onClick={handleApprove}>✓ 人工通过</button>
                <button className="btn btn-danger btn-sm" onClick={handleReject}>✗ 人工拒稿</button>
              </>
            )}
            {canCancel && <button className="btn btn-secondary btn-sm" onClick={handleCancel}>↩ 撤回</button>}
            {canPublish && <button className="btn btn-success" onClick={() => setShowPublishModal(true)}>📖 发表论文</button>}
          </div>
        </div>

        <div className="paper-detail-section">
          <h3>📋 基本信息</h3>
          <div className="info-grid">
            <div className="info-item"><span className="info-label">标题</span><span className="info-value">{paper.title}</span></div>
            <div className="info-item"><span className="info-label">学科分类</span><span className="info-value">{paper.subject_category || '-'}</span></div>
            <div className="info-item"><span className="info-label">投稿期刊</span><span className="info-value">{paper.target_journal || '-'}</span></div>
            <div className="info-item"><span className="info-label">更新时间</span><span className="info-value">{formatDate(paper.updated_at)}</span></div>
          </div>
          {paper.abstract && (
            <div className="divider" />
          )}
          {paper.abstract && (
            <div>
              <span className="info-label" style={{ display: 'block', marginBottom: 6 }}>摘要</span>
              <p style={{ color: '#374151', lineHeight: 1.8 }}>{paper.abstract}</p>
            </div>
          )}
          <div className="divider" />
          <div>
            <span className="info-label" style={{ display: 'block', marginBottom: 8 }}>作者列表</span>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 12 }}>
              {paper.authors.map((a, i) => (
                <div key={i} style={{ padding: '8px 12px', backgroundColor: '#f3f4f6', borderRadius: 8 }}>
                  <strong>{a.name}</strong>
                  {a.is_first && <span className="author-badge-first" style={{ marginLeft: 6 }}>一作</span>}
                  {a.is_corresponding && <span className="author-badge-corresponding" style={{ marginLeft: 6 }}>通讯</span>}
                  {a.affiliation && <div style={{ fontSize: 12, color: '#6b7280', marginTop: 2 }}>{a.affiliation}</div>}
                </div>
              ))}
            </div>
          </div>
          {paper.keywords.length > 0 && (
            <>
              <div className="divider" />
              <div className="paper-keywords">
                {paper.keywords.map((kw, i) => (
                  <span key={i} className="keyword-tag">#{kw}</span>
                ))}
              </div>
            </>
          )}
        </div>

        {paper.publication_info && (
          <div className="paper-detail-section">
            <h3>📖 发表记录</h3>
            <div className="publication-info">
              <h4>✓ 已发表</h4>
              <div className="info-grid">
                <div className="info-item"><span className="info-label">期刊</span><span className="info-value">{paper.publication_info.journal_name}</span></div>
                <div className="info-item"><span className="info-label">卷号</span><span className="info-value">{paper.publication_info.volume}</span></div>
                <div className="info-item"><span className="info-label">期号</span><span className="info-value">{paper.publication_info.issue}</span></div>
                <div className="info-item"><span className="info-label">页码</span><span className="info-value">{paper.publication_info.page_start} - {paper.publication_info.page_end}</span></div>
                <div className="info-item"><span className="info-label">DOI</span><span className="info-value">{paper.publication_info.doi}</span></div>
              </div>
            </div>
          </div>
        )}

        {paper.assigned_reviewers.length > 0 && (
          <div className="paper-detail-section">
            <h3>👥 分配的审稿人</h3>
            <div className="section-desc">共 {paper.assigned_reviewers.length} 位审稿人，已提交 {paper.assigned_reviewers.filter(r => r.submitted).length} 位</div>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 12 }}>
              {paper.assigned_reviewers.map((r, i) => (
                <div key={i} style={{
                  padding: '10px 14px',
                  backgroundColor: r.submitted ? '#ecfdf5' : '#fef3c7',
                  borderRadius: 8,
                  border: `1px solid ${r.submitted ? '#a7f3d0' : '#fde68a'}`,
                }}>
                  <strong>{r.name}</strong>
                  <div style={{ fontSize: 12, color: '#6b7280', marginTop: 2 }}>{r.affiliation}</div>
                  <span className="badge" style={{
                    backgroundColor: r.submitted ? '#10b981' : '#f59e0b',
                    marginTop: 6,
                    display: 'inline-block',
                  }}>
                    {r.submitted ? '已提交' : '等待中'}
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}

        {paper.reviews.length > 0 && (
          <div className="paper-detail-section">
            <h3>📝 审稿意见</h3>
            <div className="section-desc">共 {paper.reviews.length} 条评审意见</div>
            <div className="review-list">
              {paper.reviews.map(r => (
                <div key={r.id} className="review-card" style={{ borderColor: decisionColorMap[r.decision] }}>
                  <div className="review-header">
                    <div>
                      <span className="reviewer-name">{r.reviewer_name}</span>
                      <span style={{ fontSize: 12, color: '#6b7280', marginLeft: 10 }}>{formatDate(r.submitted_at)}</span>
                    </div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                      <span
                        className="badge"
                        style={{ backgroundColor: decisionColorMap[r.decision] }}
                      >
                        {decisionDisplayMap[r.decision]}
                      </span>
                      <span className="review-score" style={{ color: decisionColorMap[r.decision] }}>{r.score}</span>
                    </div>
                  </div>
                  <div className="review-comments">{r.comments || '（无详细意见）'}</div>
                </div>
              ))}
            </div>
          </div>
        )}

        <div className="paper-detail-section">
          <h3>📜 版本历史</h3>
          <div className="timeline">
            {sortedHistory.map((h, i) => (
              <div key={h.id} className="timeline-item">
                <div
                  className="timeline-dot"
                  style={{
                    backgroundColor: statusColorMap[h.status],
                    color: statusColorMap[h.status],
                  }}
                />
                <div className="timeline-content">
                  <h4>
                    <span className="badge" style={{ backgroundColor: statusColorMap[h.status], marginRight: 8 }}>
                      {statusDisplayMap[h.status]}
                    </span>
                    {h.description}
                  </h4>
                  <p>操作人：{h.operator_name}</p>
                  <div className="time">{formatDate(h.changed_at)}</div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {showReviewerModal && (
        <AssignReviewersModal
          paper={paper}
          users={users}
          userId={userId}
          onClose={() => setShowReviewerModal(false)}
          onAssigned={() => { setShowReviewerModal(false); loadPaper(); }}
        />
      )}

      {showReviewModal && (
        <SubmitReviewModal
          paper={paper}
          users={users}
          userId={userId}
          onClose={() => setShowReviewModal(false)}
          onSubmitted={() => { setShowReviewModal(false); loadPaper(); }}
        />
      )}

      {showPublishModal && (
        <PublishModal
          paperId={paperId}
          userId={userId}
          onClose={() => setShowPublishModal(false)}
          onPublished={() => { setShowPublishModal(false); loadPaper(); }}
        />
      )}
    </div>
  );
}

function AssignReviewersModal({ paper, users, userId, onClose, onAssigned }: {
  paper: Paper;
  users: User[];
  userId: string;
  onClose: () => void;
  onAssigned: () => void;
}) {
  const reviewers = users.filter(u => u.role === 'reviewer');
  const [selected, setSelected] = useState<string[]>([]);
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const toggle = (id: string) => {
    setSelected(selected.includes(id) ? selected.filter(s => s !== id) : [...selected, id]);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (selected.length === 0) { setError('请至少选择一位审稿人'); return; }
    setSubmitting(true);
    setError('');
    try {
      const reviewerList = selected.map(id => {
        const u = users.find(usr => usr.id === id)!;
        return { id: u.id, name: u.name, affiliation: u.name + ' - 专家', assigned_at: '', submitted: false };
      });
      await api.assignReviewers(userId, paper.id, reviewerList);
      onAssigned();
    } catch (err: any) {
      setError(err.response?.data?.message || err.message || '分配失败');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={e => e.stopPropagation()}>
        <form onSubmit={handleSubmit}>
          <div className="modal-header">
            <h3>分配审稿人</h3>
            <button type="button" className="modal-close" onClick={onClose}>&times;</button>
          </div>
          <div className="modal-body">
            <div className="section-desc">选择要分配的审稿专家（审稿人不能评审自己参与的论文）</div>
            <div className="reviewer-selector">
              {reviewers.map(r => {
                const isAuthor = paper.authors.some(a => a.id === r.id || a.name === r.name);
                const isSelected = selected.includes(r.id);
                return (
                  <div
                    key={r.id}
                    className={`reviewer-item ${isSelected ? 'selected' : ''}`}
                    style={{ opacity: isAuthor ? 0.5 : 1 }}
                  >
                    <input
                      type="checkbox"
                      checked={isSelected}
                      disabled={isAuthor}
                      onChange={() => !isAuthor && toggle(r.id)}
                    />
                    <div>
                      <strong>{r.name}</strong>
                      <div style={{ fontSize: 12, color: '#6b7280' }}>{roleDisplayName(r.role)}</div>
                    </div>
                    {isAuthor && <span style={{ fontSize: 12, color: '#ef4444' }}>（论文作者，不可选）</span>}
                  </div>
                );
              })}
            </div>
            {error && <div className="error-message" style={{ marginTop: 12 }}>{error}</div>}
          </div>
          <div className="modal-footer">
            <button type="button" className="btn btn-outline" onClick={onClose} disabled={submitting}>取消</button>
            <button type="submit" className="btn btn-primary" disabled={submitting}>
              {submitting ? '提交中...' : '确认分配'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

function SubmitReviewModal({ paper, users, userId, onClose, onSubmitted }: {
  paper: Paper;
  users: User[];
  userId: string;
  onClose: () => void;
  onSubmitted: () => void;
}) {
  const [reviewerId, setReviewerId] = useState('');
  const [decision, setDecision] = useState<ReviewDecision>('accept');
  const [score, setScore] = useState(8);
  const [comments, setComments] = useState('');
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const availableReviewers = paper.assigned_reviewers.filter(r => !r.submitted);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!reviewerId) { setError('请选择审稿人'); return; }
    if (score < 1 || score > 10) { setError('分数必须在1-10之间'); return; }
    setSubmitting(true);
    setError('');
    try {
      await api.submitReview(userId, paper.id, {
        reviewer_id: reviewerId,
        decision,
        score,
        comments,
      });
      onSubmitted();
    } catch (err: any) {
      setError(err.response?.data?.message || err.message || '提交失败');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={e => e.stopPropagation()}>
        <form onSubmit={handleSubmit}>
          <div className="modal-header">
            <h3>提交评审意见</h3>
            <button type="button" className="modal-close" onClick={onClose}>&times;</button>
          </div>
          <div className="modal-body">
            <div className="form-group">
              <label>审稿人</label>
              <select value={reviewerId} onChange={e => setReviewerId(e.target.value)}>
                <option value="">请选择</option>
                {availableReviewers.map(r => (
                  <option key={r.id} value={r.id}>{r.name}</option>
                ))}
              </select>
              {availableReviewers.length === 0 && <div className="section-desc">所有审稿人均已提交评审</div>}
            </div>
            <div className="form-group">
              <label>评审结论</label>
              <select value={decision} onChange={e => setDecision(e.target.value as ReviewDecision)}>
                <option value="accept">通过</option>
                <option value="revision">修改后重审</option>
                <option value="reject">拒稿</option>
              </select>
            </div>
            <div className="form-group">
              <label>评分（1-10）</label>
              <input type="number" min={1} max={10} value={score} onChange={e => setScore(parseInt(e.target.value) || 0)} />
              <input type="range" min={1} max={10} value={score} onChange={e => setScore(parseInt(e.target.value))} style={{ width: '100%', marginTop: 8 }} />
            </div>
            <div className="form-group">
              <label>评审意见</label>
              <textarea value={comments} onChange={e => setComments(e.target.value)} placeholder="请输入详细评审意见..." />
            </div>
            {error && <div className="error-message">{error}</div>}
          </div>
          <div className="modal-footer">
            <button type="button" className="btn btn-outline" onClick={onClose} disabled={submitting}>取消</button>
            <button type="submit" className="btn btn-primary" disabled={submitting}>
              {submitting ? '提交中...' : '确认提交'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

function PublishModal({ paperId, userId, onClose, onPublished }: {
  paperId: string;
  userId: string;
  onClose: () => void;
  onPublished: () => void;
}) {
  const [form, setForm] = useState<PublicationInfo>({
    journal_name: '', volume: '', issue: '', page_start: '', page_end: '', doi: '',
  });
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setError('');
    try {
      await api.publishPaper(userId, paperId, form);
      onPublished();
    } catch (err: any) {
      setError(err.response?.data?.message || err.message || '发表失败');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={e => e.stopPropagation()}>
        <form onSubmit={handleSubmit}>
          <div className="modal-header">
            <h3>发表论文（仅管理员）</h3>
            <button type="button" className="modal-close" onClick={onClose}>&times;</button>
          </div>
          <div className="modal-body">
            <div className="form-group"><label>期刊名</label><input value={form.journal_name} onChange={e => setForm({ ...form, journal_name: e.target.value })} /></div>
            <div style={{ display: 'flex', gap: 12 }}>
              <div className="form-group" style={{ flex: 1 }}><label>卷号</label><input value={form.volume} onChange={e => setForm({ ...form, volume: e.target.value })} /></div>
              <div className="form-group" style={{ flex: 1 }}><label>期号</label><input value={form.issue} onChange={e => setForm({ ...form, issue: e.target.value })} /></div>
            </div>
            <div style={{ display: 'flex', gap: 12 }}>
              <div className="form-group" style={{ flex: 1 }}><label>起始页码</label><input value={form.page_start} onChange={e => setForm({ ...form, page_start: e.target.value })} /></div>
              <div className="form-group" style={{ flex: 1 }}><label>结束页码</label><input value={form.page_end} onChange={e => setForm({ ...form, page_end: e.target.value })} /></div>
            </div>
            <div className="form-group"><label>DOI</label><input value={form.doi} onChange={e => setForm({ ...form, doi: e.target.value })} /></div>
            {error && <div className="error-message">{error}</div>}
          </div>
          <div className="modal-footer">
            <button type="button" className="btn btn-outline" onClick={onClose} disabled={submitting}>取消</button>
            <button type="submit" className="btn btn-success" disabled={submitting}>
              {submitting ? '提交中...' : '确认发表'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

function App() {
  const [users, setUsers] = useState<User[]>([]);
  const [currentUserId, setCurrentUserId] = useState('author1');
  const [viewPaperId, setViewPaperId] = useState<string | null>(null);

  useEffect(() => {
    api.getUsers().then(setUsers).catch(console.error);
  }, []);

  return (
    <div className="app-container">
      <Header users={users} currentUserId={currentUserId} onUserChange={setCurrentUserId} />

      {viewPaperId ? (
        <PaperDetailPage
          userId={currentUserId}
          paperId={viewPaperId}
          onBack={() => setViewPaperId(null)}
          users={users}
          onReload={() => {}}
        />
      ) : (
        <PaperListPage
          userId={currentUserId}
          onViewPaper={setViewPaperId}
          users={users}
        />
      )}
    </div>
  );
}

export default App;
