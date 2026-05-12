import React, { useState, useEffect } from 'react';
import { certificateApi, batchApi } from '../api';

function CertificatePage() {
  const [activeTab, setActiveTab] = useState('list');
  const [certificates, setCertificates] = useState([]);
  const [expiringSoon, setExpiringSoon] = useState([]);
  const [batches, setBatches] = useState([]);
  const [loading, setLoading] = useState(true);
  const [selectedBatch, setSelectedBatch] = useState(null);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      const [certsRes, expiringRes, batchesRes] = await Promise.all([
        certificateApi.getAll(),
        certificateApi.getExpiringSoon(),
        batchApi.getAll(),
      ]);
      setCertificates(certsRes.data);
      setExpiringSoon(expiringRes.data);
      setBatches(batchesRes.data);
    } catch (err) {
      console.error('加载失败:', err);
    } finally {
      setLoading(false);
    }
  };

  const getStatusClass = (status) => {
    switch (status) {
      case '有效': return 'status-valid';
      case '注销': return 'status-cancelled';
      case '过期': return 'status-expired';
      default: return '';
    }
  };

  const handleIssueCertificate = async (candidate) => {
    if (!confirm(`确定要为考生 "${candidate.name}" 颁发证书吗？`)) return;
    
    try {
      await certificateApi.issue({
        candidateId: candidate.id,
        batchId: selectedBatch.id,
      });
      setSuccess('证书颁发成功！');
      setError('');
      loadData();
    } catch (err) {
      setError(err.response?.data?.error || '颁发证书失败');
      setSuccess('');
    }
  };

  const handleUpdateStatus = async (certId, newStatus) => {
    if (!confirm(`确定要将证书状态更新为"${newStatus}"吗？`)) return;
    
    try {
      await certificateApi.updateStatus(certId, newStatus);
      setSuccess('状态更新成功！');
      setError('');
      loadData();
    } catch (err) {
      setError(err.response?.data?.error || '更新失败');
      setSuccess('');
    }
  };

  const completedBatchesWithPassed = batches.filter(
    b => b.status === '已结束' && b.candidates.some(c => c.isPassed)
  );

  if (loading) {
    return <div className="page-container">加载中...</div>;
  }

  return (
    <div className="page-container">
      <h2 className="page-title" style={{ marginBottom: 24 }}>证书管理</h2>

      <div className="tabs">
        <button
          className={`tab ${activeTab === 'list' ? 'active' : ''}`}
          onClick={() => setActiveTab('list')}
        >
          证书列表
        </button>
        <button
          className={`tab ${activeTab === 'issue' ? 'active' : ''}`}
          onClick={() => setActiveTab('issue')}
        >
          颁发证书
        </button>
        <button
          className={`tab ${activeTab === 'expiring' ? 'active' : ''}`}
          onClick={() => setActiveTab('expiring')}
        >
          即将过期 ({expiringSoon.length})
        </button>
      </div>

      {error && <div className="alert alert-error">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}

      {activeTab === 'list' && (
        <>
          {certificates.length === 0 ? (
            <p style={{ color: '#666', textAlign: 'center', padding: 40 }}>
              暂无证书数据
            </p>
          ) : (
            <table className="table">
              <thead>
                <tr>
                  <th>证书编号</th>
                  <th>持证人</th>
                  <th>身份证号</th>
                  <th>职业</th>
                  <th>等级</th>
                  <th>发证日期</th>
                  <th>有效期至</th>
                  <th>状态</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {certificates.map((cert) => (
                  <tr key={cert.id}>
                    <td style={{ fontFamily: 'monospace', fontWeight: 500 }}>
                      {cert.certificateNo}
                    </td>
                    <td>{cert.name}</td>
                    <td>{cert.idCard}</td>
                    <td>{cert.occupationName}</td>
                    <td>{cert.level}</td>
                    <td>{new Date(cert.issueDate).toLocaleDateString()}</td>
                    <td>{new Date(cert.expiryDate).toLocaleDateString()}</td>
                    <td>
                      <span className={`status-badge ${getStatusClass(cert.status)}`}>
                        {cert.status}
                      </span>
                    </td>
                    <td>
                      {cert.status === '有效' && (
                        <div className="action-buttons">
                          <button
                            className="btn btn-danger btn-sm"
                            onClick={() => handleUpdateStatus(cert.id, '注销')}
                          >
                            注销
                          </button>
                        </div>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </>
      )}

      {activeTab === 'issue' && (
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 24 }}>
          <div>
            <h3 style={{ marginBottom: 16 }}>选择已结束且有通过考生的批次</h3>
            {completedBatchesWithPassed.length === 0 ? (
              <p style={{ color: '#666' }}>
                暂无可颁发证书的批次。需要：<br/>
                1. 考试批次状态为"已结束"<br/>
                2. 该批次有成绩录入且通过考试的考生
              </p>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                {completedBatchesWithPassed.map((batch) => (
                  <div
                    key={batch.id}
                    onClick={() => setSelectedBatch(batch)}
                    style={{
                      padding: 12,
                      border: `2px solid ${selectedBatch?.id === batch.id ? '#1976d2' : '#ddd'}`,
                      borderRadius: 4,
                      cursor: 'pointer',
                      backgroundColor: selectedBatch?.id === batch.id ? '#e3f2fd' : 'white',
                    }}
                  >
                    <div style={{ fontWeight: 600 }}>{batch.name}</div>
                    <div style={{ fontSize: 13, color: '#666' }}>
                      {batch.occupationName} · {batch.level}
                    </div>
                    <div style={{ fontSize: 12, color: '#388e3c', marginTop: 4 }}>
                      通过人数: {batch.candidates.filter(c => c.isPassed).length} 人
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          <div>
            <h3 style={{ marginBottom: 16 }}>通过的考生</h3>
            {!selectedBatch ? (
              <p style={{ color: '#666' }}>请先选择批次</p>
            ) : selectedBatch.candidates.filter(c => c.isPassed).length === 0 ? (
              <p style={{ color: '#666' }}>该批次暂无通过的考生</p>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                {selectedBatch.candidates
                  .filter(c => c.isPassed)
                  .map((candidate) => (
                    <div
                      key={candidate.id}
                      style={{
                        padding: 12,
                        border: '1px solid #ddd',
                        borderRadius: 4,
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center',
                      }}
                    >
                      <div>
                        <div style={{ fontWeight: 500 }}>{candidate.name}</div>
                        <div style={{ fontSize: 12, color: '#666' }}>{candidate.idCard}</div>
                      </div>
                      <button
                        className="btn btn-success btn-sm"
                        onClick={() => handleIssueCertificate(candidate)}
                      >
                        颁发证书
                      </button>
                    </div>
                  ))}
              </div>
            )}
          </div>
        </div>
      )}

      {activeTab === 'expiring' && (
        <>
          {expiringSoon.length === 0 ? (
            <p style={{ color: '#666', textAlign: 'center', padding: 40 }}>
              暂无即将过期的证书（到期前3个月内会显示在这里）
            </p>
          ) : (
            <table className="table">
              <thead>
                <tr>
                  <th>证书编号</th>
                  <th>持证人</th>
                  <th>职业</th>
                  <th>等级</th>
                  <th>有效期至</th>
                  <th>剩余天数</th>
                </tr>
              </thead>
              <tbody>
                {expiringSoon.map((cert) => {
                  const daysLeft = Math.ceil((new Date(cert.expiryDate) - new Date()) / (1000 * 60 * 60 * 24));
                  return (
                    <tr key={cert.id}>
                      <td style={{ fontFamily: 'monospace' }}>{cert.certificateNo}</td>
                      <td>{cert.name}</td>
                      <td>{cert.occupationName}</td>
                      <td>{cert.level}</td>
                      <td>{new Date(cert.expiryDate).toLocaleDateString()}</td>
                      <td>
                        <span className="status-badge status-expired">
                          {daysLeft} 天
                        </span>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          )}
        </>
      )}
    </div>
  );
}

export default CertificatePage;
