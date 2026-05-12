import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { format } from 'date-fns';
import { projectsAPI, departmentsAPI, usersAPI } from '../services/api';
import ProjectForm from '../components/ProjectForm';

const statusLabels = {
  preparing: '筹备中',
  in_progress: '进行中',
  completed: '已结题',
};

const statusOptions = [
  { value: 'preparing', label: '筹备中' },
  { value: 'in_progress', label: '进行中' },
  { value: 'completed', label: '已结题' },
];

export default function ProjectKanban() {
  const [projects, setProjects] = useState([]);
  const [departments, setDepartments] = useState({});
  const [users, setUsers] = useState({});
  const [showForm, setShowForm] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      const [projRes, deptRes, userRes] = await Promise.all([
        projectsAPI.getAll(),
        departmentsAPI.getAll(),
        usersAPI.getAll(),
      ]);
      setProjects(projRes.data);
      setDepartments(deptRes.data);
      setUsers(userRes.data);
    } catch (err) {
      alert('加载数据失败');
    }
  };

  const handleCreateProject = async (data) => {
    try {
      await projectsAPI.create(data);
      setShowForm(false);
      loadData();
    } catch (err) {
      alert(err.response?.data?.error || '创建失败');
    }
  };

  const handleUpdateStatus = async (projectId, status) => {
    try {
      await projectsAPI.updateStatus(projectId, status);
      loadData();
    } catch (err) {
      alert(err.response?.data?.error || '更新状态失败');
    }
  };

  const handleDelete = async (projectId) => {
    if (!confirm('确定要删除该项目吗？')) return;
    try {
      await projectsAPI.delete(projectId);
      loadData();
    } catch (err) {
      alert('删除失败');
    }
  };

  return (
    <div>
      <div className="page-header">
        <h1>项目看板</h1>
        <button className="btn btn-primary" onClick={() => setShowForm(true)}>
          新建项目
        </button>
      </div>

      {projects.length === 0 ? (
        <p style={{ textAlign: 'center', padding: '3rem', color: '#718096' }}>
          暂无项目，点击上方按钮创建
        </p>
      ) : (
        <div className="project-kanban">
          {projects.map((project) => (
            <div
              key={project.id}
              className={`project-card status-${project.status}`}
            >
              <h3>{project.name}</h3>
              <span className="status-badge">
                {statusLabels[project.status]}
              </span>
              <div className="meta">
                <strong>牵头部门：</strong>
                {departments[project.lead_department]?.name || project.lead_department}
              </div>
              <div className="meta">
                <strong>负责人：</strong>
                {users[project.leader_id]?.name || project.leader_id}
              </div>
              <div className="meta">
                <strong>开始日期：</strong>
                {format(new Date(project.start_date), 'yyyy-MM-dd')}
              </div>
              <div className="meta">
                <strong>截止日期：</strong>
                {format(new Date(project.end_date), 'yyyy-MM-dd')}
              </div>

              <div className="actions">
                <button
                  className="btn btn-primary"
                  onClick={() => navigate(`/projects/${project.id}`)}
                >
                  详情
                </button>

                {project.status !== 'completed' && (
                  <select
                    style={{ padding: '0.5rem', borderRadius: '6px', border: '1px solid #e2e8f0' }}
                    value={project.status}
                    onChange={(e) => handleUpdateStatus(project.id, e.target.value)}
                  >
                    {statusOptions.map((opt) => (
                      <option key={opt.value} value={opt.value}>
                        {opt.label}
                      </option>
                    ))}
                  </select>
                )}

                <button
                  className="btn btn-danger"
                  onClick={() => handleDelete(project.id)}
                >
                  删除
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {showForm && (
        <ProjectForm
          departments={departments}
          users={users}
          onSubmit={handleCreateProject}
          onClose={() => setShowForm(false)}
        />
      )}
    </div>
  );
}
