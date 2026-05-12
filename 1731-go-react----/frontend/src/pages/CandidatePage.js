import React, { useState, useEffect } from 'react';
import { batchApi, occupationApi } from '../api';

const LEVEL_OPTIONS = ['初级', '中级', '高级', '技师', '高级技师'];

function CandidatePage() {
  const [activeTab, setActiveTab] = useState('register');
  const [batches, setBatches] = useState([]);
  const [occupations, setOccupations] = useState([]);
  const [loading, setLoading] = useState(true);
  const [selectedBatch, setSelectedBatch] = useState(null);
  const [registerForm, setRegisterForm] = useState({
    name: '',
    idCard: '',
    phone: '',
    appliedLevel: '初级',
  });
  const [scoreForm, setScoreForm] = useState({
    candidateId: '',
    scores: [],
  });
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      const [batchesRes, occsRes] = await Promise.all([
        batchApi.getAll(),
        occupationApi.getAll(),
      ]);
      setBatches(batchesRes.data);
      setOccupations(occsRes.data);
    } catch (err) {
      console.error('加载失败:', err);
    } finally {
      setLoading(false);
    }
  };

  const selectBatch = (batch) => {
    setSelectedBatch(batch);
    setError('');
    setSuccess('');
    setScoreForm({ candidateId: '', scores: [] });
  };

  const handleRegister = async (e) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    if (!selectedBatch) {
      setError('请先选择考试批次');
      return;
    }

    try {
      await batchApi.registerCandidate(selectedBatch.id, registerForm);
      setSuccess('报名成功！');
      setRegisterForm({ name: '', idCard: '', phone: '', appliedLevel: '初级' });
      loadData();
      
      const batchRes = await batchApi.getById(selectedBatch.id);
      setSelectedBatch(batchRes.data);
    } catch (err) {
      setError(err.response?.data?.error || '报名失败');
    }
  };

  const handleSelectCandidate = (candidate) => {
    const occ = occupations.find(o => o.id === selectedBatch.occupationId);
    const levelConfig = occ?.levels.find(l => l.level === selectedBatch.level);
    const subjects = levelConfig?.subjects || ['理论知识', '实操技能'];
    
    const initialScores = subjects.map(subject => ({
      subject,
      score: candidate.scores?.[subject] || 0,
    }));
    
    setScoreForm({
      candidateId: candidate.id,
      scores: initialScores,
    });
  };

  const handleScoreChange = (index, value) => {
    const newScores = [...scoreForm.scores];
    newScores[index].score = parseFloat(value) || 0;
    setScoreForm({ ...scoreForm, scores: newScores });
  };

  const handleEnterScores = async (e) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    if (!selectedBatch || !scoreForm.candidateId) {
      setError('请选择批次和考生');
      return;
    }

    try {
      await batchApi.enterScores(selectedBatch.id, scoreForm);
      setSuccess('成绩录入成功！');
      loadData();
      
      const batchRes = await batchApi.getById(selectedBatch.id);
      setSelectedBatch(batchRes.data);
    } catch (err) {
      setError(err.response?.data?.error || '成绩录入失败');
    }
  };

  const registerableBatches = batches.filter(b => b.status !== '已结束');
  const completedBatches = batches.filter(b => b.status === '已结束');

  if (loading) {
    return <div className="page-container">加载中...</div>;
  }

  return (
    <div className="page-container">
      <h2 className="page-title" style={{ marginBottom: 24 }}>考生管理</h2>

      <div className="tabs">
        <button
          className={`tab ${activeTab === 'register' ? 'active' : ''}`}
          onClick={() => setActiveTab('register')}
        >
          考生报名
        </button>
        <button
          className={`tab ${activeTab === 'scores' ? 'active' : ''}`}
          onClick={() => setActiveTab('scores')}
        >
          成绩录入
        </button>
      </div>

      {error && <div className="alert alert-error">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}

      {activeTab === 'register' ? (
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 24 }}>
          <div>
            <h3 style={{ marginBottom: 16 }}>选择考试批次</h3>
            {registerableBatches.length === 0 ? (
              <p style={{ color: '#666' }}>暂无可报名的批次</p>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                {registerableBatches.map((batch) => (
                  <div
                    key={batch.id}
                    onClick={() => selectBatch(batch)}
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
                      {batch.occupationName} · {batch.level} · {new Date(batch.examDate).toLocaleDateString()}
                    </div>
                    <div style={{ fontSize: 12, color: '#999', marginTop: 4 }}>
                      已报名: {batch.candidates.length} 人 · 状态: {batch.status}
                    </div>
                  </div>
                ))}
              </div>
            )}

            {selectedBatch && (
              <div style={{ marginTop: 24 }}>
                <h4 style={{ marginBottom: 12 }}>已报名考生</h4>
                {selectedBatch.candidates.length === 0 ? (
                  <p style={{ color: '#666' }}>暂无考生报名</p>
                ) : (
                  <table className="table">
                    <thead>
                      <tr>
                        <th>姓名</th>
                        <th>身份证号</th>
                        <th>电话</th>
                        <th>等级</th>
                      </tr>
                    </thead>
                    <tbody>
                      {selectedBatch.candidates.map((c) => (
                        <tr key={c.id}>
                          <td>{c.name}</td>
                          <td>{c.idCard}</td>
                          <td>{c.phone}</td>
                          <td>{c.appliedLevel}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                )}
              </div>
            )}
          </div>

          <div>
            <h3 style={{ marginBottom: 16 }}>报名信息</h3>
            <form onSubmit={handleRegister}>
              <div className="form-group">
                <label className="form-label">姓名 *</label>
                <input
                  className="form-input"
                  value={registerForm.name}
                  onChange={(e) => setRegisterForm({ ...registerForm, name: e.target.value })}
                  placeholder="请输入姓名"
                />
              </div>

              <div className="form-group">
                <label className="form-label">身份证号 * (18位)</label>
                <input
                  className="form-input"
                  value={registerForm.idCard}
                  onChange={(e) => setRegisterForm({ ...registerForm, idCard: e.target.value })}
                  placeholder="请输入18位身份证号"
                  maxLength={18}
                />
              </div>

              <div className="form-group">
                <label className="form-label">联系电话 *</label>
                <input
                  className="form-input"
                  value={registerForm.phone}
                  onChange={(e) => setRegisterForm({ ...registerForm, phone: e.target.value })}
                  placeholder="请输入联系电话"
                />
              </div>

              <div className="form-group">
                <label className="form-label">报考等级 *</label>
                <select
                  className="form-select"
                  value={registerForm.appliedLevel}
                  onChange={(e) => setRegisterForm({ ...registerForm, appliedLevel: e.target.value })}
                >
                  {LEVEL_OPTIONS.map(level => (
                    <option key={level} value={level}>{level}</option>
                  ))}
                </select>
              </div>

              <button
                type="submit"
                className="btn btn-primary"
                disabled={!selectedBatch}
              >
                提交报名
              </button>
            </form>
          </div>
        </div>
      ) : (
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 24 }}>
          <div>
            <h3 style={{ marginBottom: 16 }}>选择已结束的批次</h3>
            {completedBatches.length === 0 ? (
              <p style={{ color: '#666' }}>暂无可录入成绩的批次（需要先将批次状态流转为"已结束"）</p>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                {completedBatches.map((batch) => (
                  <div
                    key={batch.id}
                    onClick={() => selectBatch(batch)}
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
                    <div style={{ fontSize: 12, color: '#999', marginTop: 4 }}>
                      考生数: {batch.candidates.length}
                    </div>
                  </div>
                ))}
              </div>
            )}

            {selectedBatch && (
              <div style={{ marginTop: 24 }}>
                <h4 style={{ marginBottom: 12 }}>选择考生</h4>
                {selectedBatch.candidates.length === 0 ? (
                  <p style={{ color: '#666' }}>该批次暂无考生</p>
                ) : (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                    {selectedBatch.candidates.map((c) => (
                      <div
                        key={c.id}
                        onClick={() => handleSelectCandidate(c)}
                        style={{
                          padding: 10,
                          border: `2px solid ${scoreForm.candidateId === c.id ? '#388e3c' : '#eee'}`,
                          borderRadius: 4,
                          cursor: 'pointer',
                          backgroundColor: scoreForm.candidateId === c.id ? '#e8f5e9' : 'white',
                        }}
                      >
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                          <div>
                            <span style={{ fontWeight: 500 }}>{c.name}</span>
                            <span style={{ color: '#999', marginLeft: 8, fontSize: 12 }}>{c.idCard}</span>
                          </div>
                          {c.hasTakenExam && (
                            <span className={`status-badge ${c.isPassed ? 'status-valid' : 'status-cancelled'}`}>
                              {c.isPassed ? '通过' : '未通过'}
                            </span>
                          )}
                        </div>
                        {c.scores && Object.keys(c.scores).length > 0 && (
                          <div style={{ fontSize: 12, color: '#666', marginTop: 4 }}>
                            成绩: {Object.entries(c.scores).map(([k, v]) => `${k}:${v}`).join(', ')}
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}
          </div>

          <div>
            <h3 style={{ marginBottom: 16 }}>录入成绩</h3>
            {!scoreForm.candidateId ? (
              <p style={{ color: '#666' }}>请先选择批次和考生</p>
            ) : (
              <form onSubmit={handleEnterScores}>
                {scoreForm.scores.map((item, index) => (
                  <div key={item.subject} className="form-group">
                    <label className="form-label">{item.subject} (0-100分)</label>
                    <input
                      type="number"
                      min="0"
                      max="100"
                      step="0.5"
                      className="form-input"
                      value={item.score}
                      onChange={(e) => handleScoreChange(index, e.target.value)}
                      placeholder="请输入成绩"
                    />
                  </div>
                ))}
                <div style={{ fontSize: 12, color: '#666', marginBottom: 16 }}>
                  * 所有科目均达到60分才算通过
                </div>
                <button type="submit" className="btn btn-primary">
                  保存成绩
                </button>
              </form>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

export default CandidatePage;
