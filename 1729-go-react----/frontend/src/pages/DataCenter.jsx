import React, { useState, useEffect } from 'react';
import { format } from 'date-fns';
import { dataAPI, projectsAPI, usersAPI } from '../services/api';

const formatLabels = {
  csv: 'CSV',
  json: 'JSON',
  excel: 'Excel',
  image: '图片',
  other: '其他',
};

export default function DataCenter() {
  const [dataList, setDataList] = useState([]);
  const [projects, setProjects] = useState({});
  const [users, setUsers] = useState({});
  const [search, setSearch] = useState('');
  const [filterProject, setFilterProject] = useState('');

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      const [dataRes, projRes, userRes] = await Promise.all([
        dataAPI.getAll(),
        projectsAPI.getAll(),
        usersAPI.getAll(),
      ]);
      setDataList(dataRes.data);
      const projMap = {};
      projRes.data.forEach((p) => (projMap[p.id] = p));
      setProjects(projMap);
      setUsers(userRes.data);
    } catch (err) {
      alert('加载数据失败');
    }
  };

  const filtered = dataList.filter((d) => {
    if (filterProject && d.project_id !== filterProject) return false;
    if (search) {
      const s = search.toLowerCase();
      if (!d.name.toLowerCase().includes(s) && !d.description.toLowerCase().includes(s)) {
        return false;
      }
    }
    return true;
  });

  const handleDownload = async (dataId) => {
    try {
      const res = await dataAPI.download(dataId);
      alert(`下载记录已创建：${res.data.log_id}`);
      loadData();
    } catch (err) {
      alert(err.response?.data?.error || '下载失败');
    }
  };

  return (
    <div>
      <div className="page-header">
        <h1>数据共享中心</h1>
      </div>

      <div className="search-bar">
        <input
          type="text"
          placeholder="搜索数据名称或描述..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <select value={filterProject} onChange={(e) => setFilterProject(e.target.value)}>
          <option value="">所有项目</option>
          {Object.values(projects).map((p) => (
            <option key={p.id} value={p.id}>
              {p.name}
            </option>
          ))}
        </select>
      </div>

      {filtered.length === 0 ? (
        <p style={{ textAlign: 'center', padding: '3rem', color: '#718096' }}>
          暂无数据
        </p>
      ) : (
        <table className="data-table">
          <thead>
            <tr>
              <th>名称</th>
              <th>关联项目</th>
              <th>格式</th>
              <th>文件大小</th>
              <th>上传人</th>
              <th>上传时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {filtered.map((data) => (
              <tr key={data.id}>
                <td>
                  <div>
                    <strong>{data.name}</strong>
                    <div style={{ fontSize: '0.8rem', color: '#718096' }}>
                      {data.description}
                    </div>
                  </div>
                </td>
                <td>{projects[data.project_id]?.name || data.project_id}</td>
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
  );
}
