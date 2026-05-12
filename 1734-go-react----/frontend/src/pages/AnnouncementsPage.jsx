import React, { useState, useEffect } from 'react';
import api from '../api';
import { useAuth } from '../context/AuthContext';

const AnnouncementsPage = () => {
  const { user, isAdmin, isTeacher, isParent, refreshUser } = useAuth();
  const [announcements, setAnnouncements] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [selectedAnnouncement, setSelectedAnnouncement] = useState(null);
  const [formData, setFormData] = useState({
    title: '',
    content: '',
    scope_type: 'all',
    scope_value: '',
    urgency: 'normal',
  });

  useEffect(() => {
    fetchAnnouncements();
    const interval = setInterval(fetchAnnouncements, 5000);
    return () => clearInterval(interval);
  }, []);

  const fetchAnnouncements = async () => {
    try {
      const endpoint = isAdmin || isTeacher ? '/announcements/all' : '/announcements';
      const response = await api.get(endpoint);
      setAnnouncements(response.data);
    } catch (error) {
      console.error('获取通知失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      await api.post('/announcements', formData);
      setShowCreateModal(false);
      setFormData({
        title: '',
        content: '',
        scope_type: 'all',
        scope_value: '',
        urgency: 'normal',
      });
      fetchAnnouncements();
    } catch (error) {
      alert(error.response?.data?.error || '发布失败');
    }
  };

  const handleView = async (item) => {
    setSelectedAnnouncement(item);
    if (isParent && !item.isRead) {
      try {
        await api.post(`/announcements/${item.announcement.id}/read`);
        fetchAnnouncements();
        refreshUser();
      } catch (error) {
        console.error('标记已读失败:', error);
      }
    }
  };

  const getUrgencyStyle = (urgency) => {
    switch (urgency) {
      case 'urgent':
        return { background: '#fee2e2', color: '#dc2626', label: '紧急' };
      case 'important':
        return { background: '#fef3c7', color: '#d97706', label: '重要' };
      default:
        return { background: '#e5e7eb', color: '#6b7280', label: '普通' };
    }
  };

  const getScopeLabel = (item) => {
    switch (item.announcement.scope_type) {
      case 'all':
        return '全校';
      case 'grade':
        return item.announcement.scope_value;
      case 'class':
        return item.announcement.scope_value;
      default:
        return '';
    }
  };

  if (loading) {
    return <div style={{ padding: '40px', textAlign: 'center' }}>加载中...</div>;
  }

  return (
    <div style={styles.container}>
      <div style={styles.header}>
        <h2 style={styles.title}>📢 通知公告</h2>
        {(isAdmin || isTeacher) && (
          <button style={styles.createButton} onClick={() => setShowCreateModal(true)}>
            + 发布通知
          </button>
        )}
      </div>

      <div style={styles.list}>
        {announcements.length === 0 ? (
          <div style={styles.empty}>暂无通知</div>
        ) : (
          announcements.map((item) => {
            const urgencyStyle = getUrgencyStyle(item.announcement.urgency);
            return (
              <div
                key={item.announcement.id}
                style={{
                  ...styles.card,
                  ...(item.announcement.urgency === 'urgent' ? styles.cardUrgent : {}),
                  ...(isParent && !item.isRead ? styles.cardUnread : {}),
                }}
                onClick={() => handleView(item)}
              >
                <div style={styles.cardHeader}>
                  <span style={{ ...styles.urgencyBadge, background: urgencyStyle.background, color: urgencyStyle.color }}>
                    {urgencyStyle.label}
                  </span>
                  <span style={styles.scopeBadge}>
                    {getScopeLabel(item)}
                  </span>
                  {isParent && !item.isRead && (
                    <span style={styles.unreadDot}></span>
                  )}
                </div>
                <h3 style={styles.cardTitle}>{item.announcement.title}</h3>
                <p style={styles.cardContent}>
                  {item.announcement.content.substring(0, 100)}
                  {item.announcement.content.length > 100 ? '...' : ''}
                </p>
                <div style={styles.cardFooter}>
                  <span style={styles.meta}>
                    发布人: {item.authorName}
                  </span>
                  <span style={styles.meta}>
                    {new Date(item.announcement.published_at).toLocaleString()}
                  </span>
                </div>
                {item.stats && (
                  <div style={styles.stats}>
                    <span style={styles.statItem}>已读: {item.stats.read_count}</span>
                    <span style={styles.statItem}>未读: {item.stats.unread_count}</span>
                    <span style={styles.statItem}>
                      共 {item.stats.total_parents} 位家长
                    </span>
                  </div>
                )}
              </div>
            );
          })
        )}
      </div>

      {showCreateModal && (
        <div style={styles.modalOverlay} onClick={() => setShowCreateModal(false)}>
          <div style={styles.modal} onClick={(e) => e.stopPropagation()}>
            <h3 style={styles.modalTitle}>发布新通知</h3>
            <form onSubmit={handleSubmit} style={styles.form}>
              <div style={styles.formGroup}>
                <label style={styles.label}>标题 *</label>
                <input
                  type="text"
                  style={styles.input}
                  value={formData.title}
                  onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                  required
                />
              </div>

              <div style={styles.formGroup}>
                <label style={styles.label}>正文 *</label>
                <textarea
                  style={styles.textarea}
                  value={formData.content}
                  onChange={(e) => setFormData({ ...formData, content: e.target.value })}
                  rows={6}
                  required
                />
              </div>

              <div style={styles.row}>
                <div style={styles.formGroup}>
                  <label style={styles.label}>发布范围 *</label>
                  <select
                    style={styles.select}
                    value={formData.scope_type}
                    onChange={(e) => setFormData({ ...formData, scope_type: e.target.value, scope_value: '' })}
                  >
                    <option value="all">全校</option>
                    <option value="grade">指定年级</option>
                    <option value="class">指定班级</option>
                  </select>
                </div>

                <div style={styles.formGroup}>
                  <label style={styles.label}>紧急程度</label>
                  <select
                    style={styles.select}
                    value={formData.urgency}
                    onChange={(e) => setFormData({ ...formData, urgency: e.target.value })}
                  >
                    <option value="normal">普通</option>
                    <option value="important">重要</option>
                    <option value="urgent">紧急</option>
                  </select>
                </div>
              </div>

              {formData.scope_type !== 'all' && (
                <div style={styles.formGroup}>
                  <label style={styles.label}>
                    {formData.scope_type === 'grade' ? '年级名称' : '班级名称'} *
                  </label>
                  <input
                    type="text"
                    style={styles.input}
                    value={formData.scope_value}
                    onChange={(e) => setFormData({ ...formData, scope_value: e.target.value })}
                    placeholder={formData.scope_type === 'grade' ? '如：一年级' : '如：一年级1班'}
                    required
                  />
                </div>
              )}

              <div style={styles.modalActions}>
                <button
                  type="button"
                  style={styles.cancelButton}
                  onClick={() => setShowCreateModal(false)}
                >
                  取消
                </button>
                <button type="submit" style={styles.submitButton}>
                  发布
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {selectedAnnouncement && (
        <div style={styles.modalOverlay} onClick={() => setSelectedAnnouncement(null)}>
          <div style={styles.modal} onClick={(e) => e.stopPropagation()}>
            <div style={styles.detailHeader}>
              <span style={{
                ...styles.urgencyBadge,
                ...getUrgencyStyle(selectedAnnouncement.announcement.urgency),
              }}>
                {getUrgencyStyle(selectedAnnouncement.announcement.urgency).label}
              </span>
              <span style={styles.scopeBadge}>{getScopeLabel(selectedAnnouncement)}</span>
            </div>
            <h2 style={styles.detailTitle}>{selectedAnnouncement.announcement.title}</h2>
            <div style={styles.detailMeta}>
              <span>发布人: {selectedAnnouncement.authorName}</span>
              <span>{new Date(selectedAnnouncement.announcement.published_at).toLocaleString()}</span>
            </div>
            <div style={styles.detailContent}>
              {selectedAnnouncement.announcement.content}
            </div>
            {selectedAnnouncement.stats && (
              <div style={styles.detailStats}>
                <span>已读: {selectedAnnouncement.stats.read_count}</span>
                <span>未读: {selectedAnnouncement.stats.unread_count}</span>
                <span>共 {selectedAnnouncement.stats.total_parents} 位家长</span>
              </div>
            )}
            <div style={styles.modalActions}>
              <button
                style={styles.cancelButton}
                onClick={() => setSelectedAnnouncement(null)}
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
    maxWidth: '900px',
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
  list: {
    display: 'flex',
    flexDirection: 'column',
    gap: '16px',
  },
  empty: {
    textAlign: 'center',
    padding: '60px',
    color: '#999',
  },
  card: {
    background: 'white',
    borderRadius: '12px',
    padding: '20px',
    cursor: 'pointer',
    boxShadow: '0 2px 8px rgba(0,0,0,0.06)',
    transition: 'box-shadow 0.2s',
  },
  cardUrgent: {
    borderLeft: '4px solid #dc2626',
  },
  cardUnread: {
    borderLeft: '4px solid #4f46e5',
  },
  cardHeader: {
    display: 'flex',
    gap: '10px',
    alignItems: 'center',
    marginBottom: '12px',
  },
  urgencyBadge: {
    padding: '4px 12px',
    borderRadius: '20px',
    fontSize: '12px',
    fontWeight: 600,
  },
  scopeBadge: {
    padding: '4px 12px',
    background: '#eef2ff',
    color: '#4f46e5',
    borderRadius: '20px',
    fontSize: '12px',
  },
  unreadDot: {
    width: '8px',
    height: '8px',
    background: '#4f46e5',
    borderRadius: '50%',
    marginLeft: 'auto',
  },
  cardTitle: {
    margin: '0 0 10px 0',
    fontSize: '18px',
    color: '#333',
  },
  cardContent: {
    margin: '0 0 12px 0',
    color: '#666',
    lineHeight: 1.6,
  },
  cardFooter: {
    display: 'flex',
    justifyContent: 'space-between',
    fontSize: '13px',
    color: '#999',
  },
  meta: {},
  stats: {
    marginTop: '12px',
    paddingTop: '12px',
    borderTop: '1px solid #eee',
    display: 'flex',
    gap: '20px',
    fontSize: '13px',
    color: '#666',
  },
  statItem: {},
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
  },
  modal: {
    background: 'white',
    borderRadius: '12px',
    padding: '30px',
    width: '100%',
    maxWidth: '550px',
    maxHeight: '80vh',
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
  textarea: {
    padding: '12px 16px',
    border: '1px solid #ddd',
    borderRadius: '8px',
    fontSize: '14px',
    resize: 'vertical',
  },
  select: {
    padding: '12px 16px',
    border: '1px solid #ddd',
    borderRadius: '8px',
    fontSize: '14px',
    background: 'white',
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
  detailHeader: {
    display: 'flex',
    gap: '10px',
    marginBottom: '16px',
  },
  detailTitle: {
    margin: '0 0 12px 0',
    color: '#333',
  },
  detailMeta: {
    display: 'flex',
    gap: '20px',
    fontSize: '13px',
    color: '#999',
    marginBottom: '20px',
    paddingBottom: '16px',
    borderBottom: '1px solid #eee',
  },
  detailContent: {
    lineHeight: 1.8,
    color: '#444',
    whiteSpace: 'pre-wrap',
  },
  detailStats: {
    marginTop: '20px',
    paddingTop: '16px',
    borderTop: '1px solid #eee',
    display: 'flex',
    gap: '20px',
    fontSize: '13px',
    color: '#666',
  },
};

export default AnnouncementsPage;
