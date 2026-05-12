import React, { useState, useEffect } from 'react';
import { checkups, residents } from '../api';
import { formatDate, getBMICategory, isHypertension } from '../utils';

const CheckupsPage = ({ selectedResident }) => {
  const [resident, setResident] = useState(selectedResident);
  const [checkupList, setCheckupList] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [showForm, setShowForm] = useState(false);
  const [newCheckup, setNewCheckup] = useState({
    date: new Date().toISOString().split('T')[0],
    height: '',
    weight: '',
    systolicBP: '',
    diastolicBP: '',
    heartRate: '',
    visionLeft: '',
    visionRight: '',
    isInitial: false,
  });

  const loadCheckups = async (residentID) => {
    if (!residentID) return;
    setLoading(true);
    setError('');
    try {
      const data = await checkups.get(residentID);
      setCheckupList(data);
    } catch (err) {
      setError('加载体检记录失败：' + err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (selectedResident) {
      setResident(selectedResident);
      loadCheckups(selectedResident.id);
    }
  }, [selectedResident]);

  const handleAddCheckup = async (e) => {
    e.preventDefault();
    setError('');
    try {
      const checkupData = {
        ...newCheckup,
        residentID: resident.id,
        height: parseFloat(newCheckup.height),
        weight: parseFloat(newCheckup.weight),
        systolicBP: parseInt(newCheckup.systolicBP),
        diastolicBP: parseInt(newCheckup.diastolicBP),
        heartRate: parseInt(newCheckup.heartRate),
        visionLeft: parseFloat(newCheckup.visionLeft),
        visionRight: parseFloat(newCheckup.visionRight),
        date: new Date(newCheckup.date).toISOString(),
      };

      const created = await checkups.create(checkupData);
      setCheckupList([created, ...checkupList]);
      setShowForm(false);
      setNewCheckup({
        date: new Date().toISOString().split('T')[0],
        height: '',
        weight: '',
        systolicBP: '',
        diastolicBP: '',
        heartRate: '',
        visionLeft: '',
        visionRight: '',
        isInitial: false,
      });
    } catch (err) {
      setError('添加体检记录失败：' + err.message);
    }
  };

  const searchResident = async (idCard) => {
    if (!idCard) return;
    try {
      const results = await residents.search('', idCard, '');
      if (results.length > 0) {
        setResident(results[0]);
        loadCheckups(results[0].id);
      } else {
        setError('未找到该居民');
      }
    } catch (err) {
      setError('搜索居民失败：' + err.message);
    }
  };

  return (
    <div className="page checkups-page">
      <h1>体检记录管理</h1>

      {!resident && (
        <div className="resident-selector">
          <h3>选择居民</h3>
          <div className="search-bar">
            <input
              type="text"
              placeholder="输入身份证号搜索居民"
              onKeyPress={(e) => e.key === 'Enter' && searchResident(e.target.value)}
            />
            <button onClick={(e) => searchResident(e.target.previousElementSibling.value)}>
              搜索
            </button>
          </div>
        </div>
      )}

      {resident && (
        <div className="resident-info">
          <h3>当前居民：{resident.name} ({resident.gender === 'male' ? '男' : '女'})</h3>
          <p>身份证号：{resident.idCard}</p>
          <button onClick={() => setResident(null)} className="btn-secondary">
            更换居民
          </button>
          <button onClick={() => setShowForm(!showForm)} className="btn-primary">
            {showForm ? '取消录入' : '录入体检'}
          </button>
        </div>
      )}

      {error && <div className="error-message">{error}</div>}

      {resident && showForm && (
        <div className="create-form">
          <h3>录入体检记录</h3>
          <form onSubmit={handleAddCheckup}>
            <div className="form-row">
              <label>
                体检日期：
                <input
                  type="date"
                  required
                  value={newCheckup.date}
                  onChange={(e) => setNewCheckup({ ...newCheckup, date: e.target.value })}
                />
              </label>
              <label>
                是否为初始体检：
                <input
                  type="checkbox"
                  checked={newCheckup.isInitial}
                  onChange={(e) => setNewCheckup({ ...newCheckup, isInitial: e.target.checked })}
                />
              </label>
            </div>
            <div className="form-row">
              <label>
                身高 (cm)：
                <input
                  type="number"
                  step="0.1"
                  required
                  value={newCheckup.height}
                  onChange={(e) => setNewCheckup({ ...newCheckup, height: e.target.value })}
                />
              </label>
              <label>
                体重 (kg)：
                <input
                  type="number"
                  step="0.1"
                  required
                  value={newCheckup.weight}
                  onChange={(e) => setNewCheckup({ ...newCheckup, weight: e.target.value })}
                />
              </label>
            </div>
            <div className="form-row">
              <label>
                收缩压 (mmHg)：
                <input
                  type="number"
                  required
                  value={newCheckup.systolicBP}
                  onChange={(e) => setNewCheckup({ ...newCheckup, systolicBP: e.target.value })}
                />
              </label>
              <label>
                舒张压 (mmHg)：
                <input
                  type="number"
                  required
                  value={newCheckup.diastolicBP}
                  onChange={(e) => setNewCheckup({ ...newCheckup, diastolicBP: e.target.value })}
                />
              </label>
            </div>
            <div className="form-row">
              <label>
                心率 (次/分)：
                <input
                  type="number"
                  value={newCheckup.heartRate}
                  onChange={(e) => setNewCheckup({ ...newCheckup, heartRate: e.target.value })}
                />
              </label>
            </div>
            <div className="form-row">
              <label>
                左眼视力：
                <input
                  type="number"
                  step="0.1"
                  value={newCheckup.visionLeft}
                  onChange={(e) => setNewCheckup({ ...newCheckup, visionLeft: e.target.value })}
                />
              </label>
              <label>
                右眼视力：
                <input
                  type="number"
                  step="0.1"
                  value={newCheckup.visionRight}
                  onChange={(e) => setNewCheckup({ ...newCheckup, visionRight: e.target.value })}
                />
              </label>
            </div>
            <button type="submit" className="btn-primary">保存体检记录</button>
          </form>
        </div>
      )}

      {resident && (
        <div className="records-panel">
          <h3>体检记录 ({checkupList.length} 条)</h3>
          {loading ? (
            <div className="loading">加载中...</div>
          ) : checkupList.length === 0 ? (
            <div className="empty-state">暂无体检记录</div>
          ) : (
            <div className="checkup-list">
              {checkupList.map((checkup) => {
                const bmiInfo = getBMICategory(checkup.bmi);
                const isHighBP = isHypertension(checkup.systolicBP, checkup.diastolicBP);

                return (
                  <div key={checkup.id} className="checkup-card">
                    <div className="checkup-header">
                      <span className="checkup-date">{formatDate(checkup.date)}</span>
                      {checkup.isInitial && (
                        <span className="badge initial">初始体检</span>
                      )}
                    </div>
                    <div className="checkup-body">
                      <div className="checkup-item">
                        <label>身高</label>
                        <span>{checkup.height} cm</span>
                      </div>
                      <div className="checkup-item">
                        <label>体重</label>
                        <span>{checkup.weight} kg</span>
                      </div>
                      <div className="checkup-item">
                        <label>BMI</label>
                        <span style={{ color: bmiInfo.color }}>
                          {checkup.bmi.toFixed(1)} ({bmiInfo.category})
                        </span>
                      </div>
                      <div className="checkup-item">
                        <label>血压</label>
                        <span style={{ color: isHighBP ? '#e74c3c' : '#333' }}>
                          {checkup.systolicBP}/{checkup.diastolicBP} mmHg
                          {isHighBP && <span className="badge hypertension">高血压</span>}
                        </span>
                      </div>
                      {checkup.heartRate > 0 && (
                        <div className="checkup-item">
                          <label>心率</label>
                          <span>{checkup.heartRate} 次/分</span>
                        </div>
                      )}
                      {checkup.visionLeft > 0 && (
                        <div className="checkup-item">
                          <label>左眼视力</label>
                          <span>{checkup.visionLeft}</span>
                        </div>
                      )}
                      {checkup.visionRight > 0 && (
                        <div className="checkup-item">
                          <label>右眼视力</label>
                          <span>{checkup.visionRight}</span>
                        </div>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default CheckupsPage;
