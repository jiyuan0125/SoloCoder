import React, { useState, useEffect } from 'react';
import { statsApi, batchApi, certificateApi } from '../api';

function DashboardPage() {
  const [stats, setStats] = useState([]);
  const [batchCount, setBatchCount] = useState(0);
  const [certCount, setCertCount] = useState(0);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      const [statsRes, batchesRes, certsRes] = await Promise.all([
        statsApi.getPassRate(),
        batchApi.getAll(),
        certificateApi.getAll(),
      ]);
      setStats(statsRes.data);
      setBatchCount(batchesRes.data.length);
      setCertCount(certsRes.data.length);
    } catch (error) {
      console.error('加载数据失败:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return <div className="page-container">加载中...</div>;
  }

  return (
    <div>
      <div className="page-container" style={{ marginBottom: 24 }}>
        <h2 className="page-title" style={{ marginBottom: 24 }}>统计概览</h2>
        <div className="stats-grid">
          <div className="stat-card">
            <div className="stat-label">考试批次总数</div>
            <div className="stat-value">{batchCount}</div>
          </div>
          <div className="stat-card">
            <div className="stat-label">证书颁发总数</div>
            <div className="stat-value">{certCount}</div>
          </div>
        </div>
      </div>

      <div className="page-container">
        <h2 className="page-title" style={{ marginBottom: 24 }}>各职业各等级通过率统计</h2>
        {stats.length === 0 ? (
          <p style={{ color: '#666' }}>暂无统计数据</p>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>职业名称</th>
                <th>等级</th>
                <th>参考人数</th>
                <th>通过人数</th>
                <th>通过率</th>
              </tr>
            </thead>
            <tbody>
              {stats.map((stat, index) => (
                <tr key={index}>
                  <td>{stat.occupationName}</td>
                  <td>{stat.level}</td>
                  <td>{stat.totalTaken}</td>
                  <td>{stat.totalPassed}</td>
                  <td>{stat.totalTaken > 0 ? stat.passRate.toFixed(2) + '%' : '-'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}

export default DashboardPage;
