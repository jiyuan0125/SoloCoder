import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { format } from 'date-fns';
import {
  projectsAPI,
  tasksAPI,
  dataAPI,
  usersAPI,
  departmentsAPI,
} from '../services/api';
import TaskForm from '../components/TaskForm';
import DataForm from '../components/DataForm';

const statusLabels = {
  pending: '待开始',
  in_progress: '进行中',
  completed: '已完成',
  cancelled: '已取消',
};

const priorityLabels = {
  high: '高',
  medium: '中',
  low: '低',
};

const formatLabels = {
  csv: 'CSV',
  json: 'JSON',
  excel: 'Excel',
  image: '图片',
  other: '其他',
};

export default function ProjectDetail() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [project, setProject] = useState(null);
  const [tasks, setTasks] = useState([]);
  const [dataList, setDataList] = useState([]);
  const [users, setUsers] = useState({});
  const [departments, setDepartments] = useState({});
  const [activeTab, setActiveTab] = useState('tasks');
  const [showTaskForm, setShowTaskForm] = useState(false);
  const [showDataForm, setShowDataForm] = useState(false);

  useEffect(() => {
    loadData();
  }, [id]);

  const loadData = async () => {
    try {
      const [projRes, tasksRes, dataRes, userRes, deptRes] = await Promise.all([
        projectsAPI.get(id),
        tasksAPI.getByProject(id),
        dataAPI.getAll({ project_id: id }),
        usersAPI.getAll(),
        departmentsAPI.getAll(),
      ]);
      setProject(projRes.data);
      setTasks(tasksRes.data);
      setDataList(dataRes.data);
      setUsers(userRes.data);
      setDepartments(deptRes.data);
    } catch (err) {
      alert('加载数据失败');
    }
  };

  const handleCreateTask = async (data) => {
    try {
      await tasksAPI.create(data);
      setShowTaskForm(false);
      loadData();
    } catch (err) {
      alert(err.response?.data?.error || '创建失败');
    }
  };

  const handleTaskStatusChange = async (taskId, status) => {
    try {
      await tasksAPI.updateStatus(taskId, status);
      loadData();
    } catch (err) {
      alert(err.response?.data?.error || '更新状态失败');
    }
  };

  const handleUploadData = async (data) => {
    try {
      await dataAPI.upload(data);
      setShowDataForm(false);
      loadData();
    } catch (err) {
      alert(err.response?.data?.error || '上传失败');
    }
  };

  const handleDownload = async (dataId) => {
    try {
      const res = await dataAPI.download(dataId);
      alert(`下载记录已创建：${res.data.log_id}`);
      loadData();
    } catch (err) {
      alert(err.response?.data?.error || '下载失败');
    }
  };

  if (!project) {
    return <div>加载中...</div>;
  }

  const statusMap = {
    preparing: '筹备中',
    in_progress: '进行中',
    completed: '已结题',
  };

  return (
    <div>
      <button className="btn btn-secondary" onClick={() => navigate('/')} style={{ marginBottom: '1rem' }}>
        ← 返回看板
      </button>

      <div className="project-detail">
        <h1>{project.name}</h1>
        <span className="status-badge">{statusMap[project.status]}</span>
        <div className="detail-meta" style={{ marginTop: '1rem' }}>
          <div>
            <strong>牵头部门：</strong>
            {departments[project.lead_department]?.name || project.lead_department}
          </div>
          <div>
            <strong>负责人：</strong>
            {users[project.leader_id]?.name || project.leader_id}
          </div>
          <div>
            <strong>开始：</strong>
            {format(new Date(project.start_date), 'yyyy-MM-dd')}
          </div>
          <div>
            <strong>截止：</strong>
            {format(new Date(project.end_date), 'yyyy-MM-dd')}
          </div>
          <div>
            <strong>参与部门：</strong>
            {project.participating_depts
              ?.map((d) => departments[d]?.name || d)
              .join('、') || '无'}
          </div>
        </div>
      </div>

      <div className="tabs">
        <div
          className={`tab ${activeTab === 'tasks' ? 'active' : ''}`}
          onClick={() => setActiveTab('tasks')}
        >
          子任务 ({tasks.length})
        </div>
        <div
          className={`tab ${activeTab === 'data' ? 'active' : ''}`}
          onClick={() => setActiveTab('data')}
        >
          数据共享 ({dataList.length})
        </div>
      </div>

      {activeTab === 'tasks' && (
        <div className="section">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
            <h2 style={{ margin: 0 }}>子任务列表</h2>
            {project.status === 'in_progress' && (
              <button className="btn btn-primary" onClick={() => setShowTaskForm(true)}>
                新建任务
              </button>
            )}
          </div>

          {tasks.length === 0 ? (
            <p style={{ color: '#718096', padding: '1rem' }}>暂无子任务</p>
          ) : (
            <table className="data-table">
              <thead>
                <tr>
                  <th>任务名称</th>
                  <th>负责人</th>
                  <th>优先级</th>
                  <th>截止日期</th>
                  <th>状态</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {tasks.map((task) => (
                  <tr key={task.id}>
                    <td>{task.title}</td>
                    <td>{users[task.assignee_id]?.name || task.assignee_id}</td>
                    <td className={`priority-${task.priority}`}>
                      {priorityLabels[task.priority]}
                    </td>
                    <td>{format(new Date(task.due_date), 'yyyy-MM-dd')}</td>
                    <td>
                      <span className={`badge badge-${task.status}`}>
                        {statusLabels[task.status]}
                      </span>
                    </td>
                    <td>
                      {task.status === 'pending' && (
                        <button
                          className="btn btn-primary"
                          style={{ padding: '0.25rem 0.5rem', fontSize: '0.8rem' }}
                          onClick={() => handleTaskStatusChange(task.id, 'in_progress')}
                        >
                          开始
                        </button>
                      )}
                      {task.status === 'in_progress' && (
                        <>
                          <button
                            className="btn btn-success"
                            style={{ padding: '0.25rem 0.5rem', fontSize: '0.8rem', marginRight: '0.5rem' }}
                            onClick={() => handleTaskStatusChange(task.id, 'completed')}
                          >
                            完成
                          </button>
                          <button
                            className="btn btn-danger"
                            style={{ padding: '0.25rem 0.5rem', fontSize: '0.8rem' }}
                            onClick={() => handleTaskStatusChange(task.id, 'cancelled')}
                          >
                            取消
                          </button>
                        </>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {activeTab === 'data' && (
        <div className="section">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
            <h2 style={{ margin: 0 }}>共享数据</h2>
            <button className="btn btn-primary" onClick={() => setShowDataForm(true)}>
              上传数据
            </button>
          </div>

          {dataList.length === 0 ? (
            <p style={{ color: '#718096', padding: '1rem' }}>暂无共享数据</p>
          ) : (
            <table className="data-table">
              <thead>
                <tr>
                  <th>名称</th>
                  <th>格式</th>
                  <th>文件大小</th>
                  <th>上传人</th>
                  <th>上传时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {dataList.map((data) => (
                  <tr key={data.id}>
                    <td>
                      <div>
                        <strong>{data.name}</strong>
                        <div style={{ fontSize: '0.8rem', color: '#718096' }}>
                          {data.description}
                        </div>
                      </div>
                    </td>
                    <td>{formatLabels[data.format]}</td>
                    <td>{(data.file_size / 1024).toFixed(1)} KB</td>
                    <td>{users[data.uploader_id]?.name || data.uploader_id}</td>
                    <td>{format(new Date(data.created_at), 'yyyy-MM-dd HH:mm')}</td>
                    <td>
                      <button
                        className="btn btn-primary"
                        style={{ padding: '0.25rem 0.5rem', fontSize: '0.8rem' }}
                        onClick={() => handleDownload(data.id)}
                      >
                        下载
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {showTaskForm && (
        <TaskForm
          projectId={id}
          project={project}
          users={users}
          departments={departments}
          onSubmit={handleCreateTask}
          onClose={() => setShowTaskForm(false)}
        />
      )}

      {showDataForm && (
        <DataForm
          projectId={id}
          onSubmit={handleUploadData}
          onClose={() => setShowDataForm(false)}
        />
      )}
    </div>
  );
}
