import React, { useState, useEffect } from 'react';
import { api } from '../utils/api';

function TreeNode({ node, onAddChild, onUpdate, onDelete }) {
  const [expanded, setExpanded] = useState(true);
  const [editing, setEditing] = useState(false);
  const [editName, setEditName] = useState(node.name);
  const [showAdd, setShowAdd] = useState(false);
  const [newName, setNewName] = useState('');

  const hasChildren = node.children && node.children.length > 0;

  const handleUpdate = async () => {
    try {
      await api.updateKnowledgePoint(node.id, { name: editName });
      setEditing(false);
      onUpdate();
    } catch (error) {
      alert(error.message);
    }
  };

  const handleAddChild = async () => {
    try {
      await api.createKnowledgePoint({
        name: newName,
        subject: node.subject,
        parent_id: node.id,
      });
      setShowAdd(false);
      setNewName('');
      onUpdate();
    } catch (error) {
      alert(error.message);
    }
  };

  const handleDelete = async () => {
    if (!confirm('确定要删除此知识点吗？')) return;
    try {
      await api.deleteKnowledgePoint(node.id);
      onUpdate();
    } catch (error) {
      alert(error.message);
    }
  };

  return (
    <div className="tree-node">
      <div className="tree-node-content">
        <button
          className="tree-toggle"
          onClick={() => hasChildren && setExpanded(!expanded)}
        >
          {hasChildren ? (expanded ? '▼' : '▶') : '•'}
        </button>
        {editing ? (
          <>
            <input
              type="text"
              value={editName}
              onChange={(e) => setEditName(e.target.value)}
              style={{ padding: '0.25rem' }}
            />
            <button className="btn btn-success" onClick={handleUpdate}>保存</button>
            <button className="btn btn-secondary" onClick={() => setEditing(false)}>取消</button>
          </>
        ) : (
          <>
            <span className="tree-name">{node.name}</span>
            <span style={{ color: '#666', fontSize: '0.8rem' }}>
              (L{node.level})
            </span>
            <div className="tree-actions">
              {node.level < 4 && (
                <button className="btn btn-primary" onClick={() => setShowAdd(true)}>
                  +子节点
                </button>
              )}
              <button className="btn btn-secondary" onClick={() => setEditing(true)}>
                编辑
              </button>
              <button className="btn btn-danger" onClick={handleDelete}>
                删除
              </button>
            </div>
          </>
        )}
      </div>

      {showAdd && (
        <div style={{ marginLeft: '2rem', marginTop: '0.5rem' }}>
          <input
            type="text"
            placeholder="新节点名称"
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            style={{ padding: '0.25rem', marginRight: '0.5rem' }}
          />
          <button className="btn btn-success" onClick={handleAddChild}>添加</button>
          <button className="btn btn-secondary" onClick={() => setShowAdd(false)}>取消</button>
        </div>
      )}

      {expanded && hasChildren && (
        <div>
          {node.children.map((child) => (
            <TreeNode
              key={child.id}
              node={child}
              onAddChild={onAddChild}
              onUpdate={onUpdate}
              onDelete={onDelete}
            />
          ))}
        </div>
      )}
    </div>
  );
}

function KnowledgeManagement() {
  const [knowledgePoints, setKnowledgePoints] = useState([]);
  const [loading, setLoading] = useState(true);
  const [subject, setSubject] = useState('');
  const [newSubject, setNewSubject] = useState('');
  const [newName, setNewName] = useState('');
  const [error, setError] = useState('');

  const loadKnowledgePoints = async () => {
    try {
      setLoading(true);
      const data = await api.getKnowledgePoints(subject);
      setKnowledgePoints(data.data || []);
      setError('');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadKnowledgePoints();
  }, [subject]);

  const handleCreateRoot = async () => {
    if (!newSubject || !newName) {
      alert('请填写学科和名称');
      return;
    }
    try {
      await api.createKnowledgePoint({
        name: newName,
        subject: newSubject,
        parent_id: null,
      });
      setNewName('');
      loadKnowledgePoints();
    } catch (err) {
      alert(err.message);
    }
  };

  return (
    <div>
      <div className="card">
        <h2>知识点管理</h2>

        <div style={{ marginBottom: '1.5rem' }}>
          <label style={{ marginRight: '0.5rem' }}>筛选学科：</label>
          <select
            value={subject}
            onChange={(e) => setSubject(e.target.value)}
            style={{ padding: '0.5rem', minWidth: '200px' }}
          >
            <option value="">全部学科</option>
            {[...new Set(knowledgePoints.map(kp => kp.subject))].map(s => (
              <option key={s} value={s}>{s}</option>
            ))}
          </select>
        </div>

        <div className="card" style={{ background: '#f8f9fa' }}>
          <h3>添加根知识点</h3>
          <div style={{ display: 'flex', gap: '1rem', alignItems: 'flex-end' }}>
            <div className="form-group" style={{ flex: 1 }}>
              <label>学科</label>
              <input
                type="text"
                value={newSubject}
                onChange={(e) => setNewSubject(e.target.value)}
                placeholder="例如：数学"
              />
            </div>
            <div className="form-group" style={{ flex: 2 }}>
              <label>知识点名称</label>
              <input
                type="text"
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                placeholder="例如：初等数学"
              />
            </div>
            <button className="btn btn-primary" onClick={handleCreateRoot}>
              添加根节点
            </button>
          </div>
        </div>

        {loading ? (
          <div className="empty-state">
            <p>加载中...</p>
          </div>
        ) : error ? (
          <div className="empty-state" style={{ color: '#dc3545' }}>
            <p>{error}</p>
          </div>
        ) : knowledgePoints.length === 0 ? (
          <div className="empty-state">
            <p>暂无知识点，请添加根知识点开始</p>
          </div>
        ) : (
          <div>
            {knowledgePoints.map((root) => (
              <TreeNode
                key={root.id}
                node={root}
                onUpdate={loadKnowledgePoints}
                onDelete={loadKnowledgePoints}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

export default KnowledgeManagement;
