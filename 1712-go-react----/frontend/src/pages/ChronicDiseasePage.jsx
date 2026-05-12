import React, { useState, useEffect } from 'react';
import { followups, chronicPatients, residents } from '../api';
import { formatDate, diseaseNames } from '../utils';
import LineChart from '../components/LineChart';

const ChronicDiseasePage = () => {
  const [patients, setPatients] = useState([]);
  const [selectedPatient, setSelectedPatient] = useState(null);
  const [selectedDisease, setSelectedDisease] = useState('hypertension');
  const [followupRecords, setFollowupRecords] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [showForm, setShowForm] = useState(false);
  const [newFollowup, setNewFollowup] = useState({
    date: new Date().toISOString().split('T')[0],
    hypertension: { systolicBP: '', diastolicBP: '', medication: '' },
    diabetes: { fastingBloodSugar: '', postprandialBloodSugar: '', hba1c: '', medication: '' },
    coronary: { cardiacFunction: '', medication: '' },
  });

  const loadPatients = async () => {
    setLoading(true);
    setError('');
    try {
      const data = await chronicPatients.list();
      const patientsWithInfo = await Promise.all(
        data.map(async (p) => {
          try {
            const resident = await residents.search('', '', '');
            const matchingResident = resident.find((r) => r.id === p.residentID);
            return { ...p, resident: matchingResident };
          } catch {
            return p;
          }
        })
      );
      setPatients(patientsWithInfo);
    } catch (err) {
      setError('加载慢性病患者列表失败：' + err.message);
    } finally {
      setLoading(false);
    }
  };

  const loadFollowups = async (residentID, disease) => {
    if (!residentID) return;
    setLoading(true);
    setError('');
    try {
      const data = await followups.get(residentID, disease);
      setFollowupRecords(data);
    } catch (err) {
      setError('加载随访记录失败：' + err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadPatients();
  }, []);

  useEffect(() => {
    if (selectedPatient) {
      loadFollowups(selectedPatient.residentID, selectedDisease);
    }
  }, [selectedPatient, selectedDisease]);

  const handleAddFollowup = async (e) => {
    e.preventDefault();
    setError('');
    try {
      const followupData = {
        residentID: selectedPatient.residentID,
        disease: selectedDisease,
        date: new Date(newFollowup.date).toISOString(),
        hypertension: selectedDisease === 'hypertension' ? {
          systolicBP: parseInt(newFollowup.hypertension.systolicBP),
          diastolicBP: parseInt(newFollowup.hypertension.diastolicBP),
          medication: newFollowup.hypertension.medication,
        } : null,
        diabetes: selectedDisease === 'diabetes' ? {
          fastingBloodSugar: parseFloat(newFollowup.diabetes.fastingBloodSugar),
          postprandialBloodSugar: parseFloat(newFollowup.diabetes.postprandialBloodSugar),
          hba1c: parseFloat(newFollowup.diabetes.hba1c),
          medication: newFollowup.diabetes.medication,
        } : null,
        coronary: selectedDisease === 'coronary' ? {
          cardiacFunction: newFollowup.coronary.cardiacFunction,
          medication: newFollowup.coronary.medication,
        } : null,
      };

      await followups.create(followupData);
      setShowForm(false);
      loadFollowups(selectedPatient.residentID, selectedDisease);
      setNewFollowup({
        date: new Date().toISOString().split('T')[0],
        hypertension: { systolicBP: '', diastolicBP: '', medication: '' },
        diabetes: { fastingBloodSugar: '', postprandialBloodSugar: '', hba1c: '', medication: '' },
        coronary: { cardiacFunction: '', medication: '' },
      });
    } catch (err) {
      setError('添加随访记录失败：' + err.message);
    }
  };

  const getChartData = () => {
    if (!followupRecords || followupRecords.length === 0) return { data: [], config: {} };

    const sorted = [...followupRecords].sort(
      (a, b) => new Date(a.date) - new Date(b.date)
    );

    switch (selectedDisease) {
      case 'hypertension':
        return {
          data: sorted.map((f) => ({
            date: formatDate(f.date),
            systolic: f.hypertension?.systolicBP || 0,
            diastolic: f.hypertension?.diastolicBP || 0,
          })),
          config: {
            xKey: 'date',
            yKeys: ['systolic', 'diastolic'],
            yLabels: ['收缩压', '舒张压'],
            referenceLines: [
              { value: 140, label: '140 mmHg (收缩压参考)', color: '#e74c3c' },
              { value: 90, label: '90 mmHg (舒张压参考)', color: '#e67e22' },
            ],
          },
        };

      case 'diabetes':
        return {
          data: sorted.map((f) => ({
            date: formatDate(f.date),
            fasting: f.diabetes?.fastingBloodSugar || 0,
            postprandial: f.diabetes?.postprandialBloodSugar || 0,
          })),
          config: {
            xKey: 'date',
            yKeys: ['fasting', 'postprandial'],
            yLabels: ['空腹血糖', '餐后血糖'],
            referenceLines: [
              { value: 7.0, label: '7.0 mmol/L (空腹参考)', color: '#e74c3c' },
              { value: 11.1, label: '11.1 mmol/L (餐后参考)', color: '#e67e22' },
            ],
          },
        };

      default:
        return { data: [], config: {} };
    }
  };

  const { data: chartData, config: chartConfig } = getChartData();

  return (
    <div className="page chronic-page">
      <h1>慢性病管理</h1>

      {error && <div className="error-message">{error}</div>}

      <div className="patients-panel">
        <h3>慢性病患者列表</h3>
        {loading && patients.length === 0 ? (
          <div className="loading">加载中...</div>
        ) : patients.length === 0 ? (
          <div className="empty-state">暂无慢性病患者</div>
        ) : (
          <table className="results-table">
            <thead>
              <tr>
                <th>姓名</th>
                <th>慢性病类型</th>
                <th>最近随访日期</th>
                <th>下次随访日期</th>
                <th>状态</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {patients.map((patient) => (
                <tr
                  key={`${patient.residentID}-${patient.disease}`}
                  className={patient.isOverdue ? 'overdue' : ''}
                >
                  <td>{patient.resident?.name || '未知'}</td>
                  <td>{diseaseNames[patient.disease] || patient.disease}</td>
                  <td>{formatDate(patient.date)}</td>
                  <td>{formatDate(patient.nextDueDate)}</td>
                  <td>
                    {patient.isOverdue ? (
                      <span className="badge overdue">超期未随访</span>
                    ) : (
                      <span className="badge normal">正常</span>
                    )}
                  </td>
                  <td>
                    <button
                      onClick={() => {
                        setSelectedPatient(patient);
                        setSelectedDisease(patient.disease);
                      }}
                    >
                      查看详情
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {selectedPatient && (
        <div className="patient-detail">
          <div className="detail-header">
            <h3>
              {selectedPatient.resident?.name || '未知'} - {diseaseNames[selectedDisease]}
            </h3>
            <div className="disease-selector">
              <label>选择查看的慢性病：</label>
              <select
                value={selectedDisease}
                onChange={(e) => setSelectedDisease(e.target.value)}
              >
                <option value="hypertension">高血压</option>
                <option value="diabetes">糖尿病</option>
                <option value="coronary">冠心病</option>
                <option value="stroke">脑卒中</option>
                <option value="copd">慢性阻塞性肺疾病</option>
              </select>
            </div>
            <button onClick={() => setShowForm(!showForm)} className="btn-primary">
              {showForm ? '取消录入' : '录入随访'}
            </button>
          </div>

          {showForm && (
            <div className="create-form">
              <h3>录入随访记录</h3>
              <form onSubmit={handleAddFollowup}>
                <div className="form-row">
                  <label>
                    随访日期：
                    <input
                      type="date"
                      required
                      value={newFollowup.date}
                      onChange={(e) => setNewFollowup({ ...newFollowup, date: e.target.value })}
                    />
                  </label>
                </div>

                {selectedDisease === 'hypertension' && (
                  <>
                    <div className="form-row">
                      <label>
                        收缩压 (mmHg)：
                        <input
                          type="number"
                          required
                          value={newFollowup.hypertension.systolicBP}
                          onChange={(e) => setNewFollowup({
                            ...newFollowup,
                            hypertension: { ...newFollowup.hypertension, systolicBP: e.target.value },
                          })}
                        />
                      </label>
                      <label>
                        舒张压 (mmHg)：
                        <input
                          type="number"
                          required
                          value={newFollowup.hypertension.diastolicBP}
                          onChange={(e) => setNewFollowup({
                            ...newFollowup,
                            hypertension: { ...newFollowup.hypertension, diastolicBP: e.target.value },
                          })}
                        />
                      </label>
                    </div>
                    <div className="form-row">
                      <label>
                        用药情况：
                        <input
                          type="text"
                          value={newFollowup.hypertension.medication}
                          onChange={(e) => setNewFollowup({
                            ...newFollowup,
                            hypertension: { ...newFollowup.hypertension, medication: e.target.value },
                          })}
                        />
                      </label>
                    </div>
                  </>
                )}

                {selectedDisease === 'diabetes' && (
                  <>
                    <div className="form-row">
                      <label>
                        空腹血糖 (mmol/L)：
                        <input
                          type="number"
                          step="0.1"
                          value={newFollowup.diabetes.fastingBloodSugar}
                          onChange={(e) => setNewFollowup({
                            ...newFollowup,
                            diabetes: { ...newFollowup.diabetes, fastingBloodSugar: e.target.value },
                          })}
                        />
                      </label>
                      <label>
                        餐后两小时血糖 (mmol/L)：
                        <input
                          type="number"
                          step="0.1"
                          value={newFollowup.diabetes.postprandialBloodSugar}
                          onChange={(e) => setNewFollowup({
                            ...newFollowup,
                            diabetes: { ...newFollowup.diabetes, postprandialBloodSugar: e.target.value },
                          })}
                        />
                      </label>
                    </div>
                    <div className="form-row">
                      <label>
                        糖化血红蛋白 (%)：
                        <input
                          type="number"
                          step="0.1"
                          value={newFollowup.diabetes.hba1c}
                          onChange={(e) => setNewFollowup({
                            ...newFollowup,
                            diabetes: { ...newFollowup.diabetes, hba1c: e.target.value },
                          })}
                        />
                      </label>
                      <label>
                        用药情况：
                        <input
                          type="text"
                          value={newFollowup.diabetes.medication}
                          onChange={(e) => setNewFollowup({
                            ...newFollowup,
                            diabetes: { ...newFollowup.diabetes, medication: e.target.value },
                          })}
                        />
                      </label>
                    </div>
                  </>
                )}

                {selectedDisease === 'coronary' && (
                  <>
                    <div className="form-row">
                      <label>
                        心功能评估：
                        <input
                          type="text"
                          value={newFollowup.coronary.cardiacFunction}
                          onChange={(e) => setNewFollowup({
                            ...newFollowup,
                            coronary: { ...newFollowup.coronary, cardiacFunction: e.target.value },
                          })}
                        />
                      </label>
                      <label>
                        用药情况：
                        <input
                          type="text"
                          value={newFollowup.coronary.medication}
                          onChange={(e) => setNewFollowup({
                            ...newFollowup,
                            coronary: { ...newFollowup.coronary, medication: e.target.value },
                          })}
                        />
                      </label>
                    </div>
                  </>
                )}

                <button type="submit" className="btn-primary">保存随访记录</button>
              </form>
            </div>
          )}

          {chartData.length > 0 && chartConfig.xKey && (
            <div className="chart-panel">
              <LineChart
                data={chartData}
                xKey={chartConfig.xKey}
                yKeys={chartConfig.yKeys}
                yLabels={chartConfig.yLabels}
                referenceLines={chartConfig.referenceLines}
                title={`${diseaseNames[selectedDisease]}趋势图`}
              />
            </div>
          )}

          <div className="records-panel">
            <h3>随访记录 ({followupRecords.length} 条)</h3>
            {followupRecords.length === 0 ? (
              <div className="empty-state">暂无随访记录</div>
            ) : (
              <div className="followup-list">
                {followupRecords.map((record) => (
                  <div
                    key={record.id}
                    className={`followup-card ${record.isOverdue ? 'overdue' : ''}`}
                  >
                    <div className="followup-header">
                      <span className="followup-date">{formatDate(record.date)}</span>
                      {record.isOverdue && (
                        <span className="badge overdue">超期</span>
                      )}
                      <span className="next-due">下次随访：{formatDate(record.nextDueDate)}</span>
                    </div>
                    <div className="followup-body">
                      {record.disease === 'hypertension' && record.hypertension && (
                        <>
                          <div className="followup-item">
                            <label>收缩压</label>
                            <span>{record.hypertension.systolicBP} mmHg</span>
                          </div>
                          <div className="followup-item">
                            <label>舒张压</label>
                            <span>{record.hypertension.diastolicBP} mmHg</span>
                          </div>
                          {record.hypertension.medication && (
                            <div className="followup-item">
                              <label>用药</label>
                              <span>{record.hypertension.medication}</span>
                            </div>
                          )}
                        </>
                      )}
                      {record.disease === 'diabetes' && record.diabetes && (
                        <>
                          <div className="followup-item">
                            <label>空腹血糖</label>
                            <span>{record.diabetes.fastingBloodSugar} mmol/L</span>
                          </div>
                          <div className="followup-item">
                            <label>餐后血糖</label>
                            <span>{record.diabetes.postprandialBloodSugar} mmol/L</span>
                          </div>
                          <div className="followup-item">
                            <label>糖化血红蛋白</label>
                            <span>{record.diabetes.hba1c}%</span>
                          </div>
                          {record.diabetes.medication && (
                            <div className="followup-item">
                              <label>用药</label>
                              <span>{record.diabetes.medication}</span>
                            </div>
                          )}
                        </>
                      )}
                      {record.disease === 'coronary' && record.coronary && (
                        <>
                          <div className="followup-item">
                            <label>心功能评估</label>
                            <span>{record.coronary.cardiacFunction}</span>
                          </div>
                          {record.coronary.medication && (
                            <div className="followup-item">
                              <label>用药</label>
                              <span>{record.coronary.medication}</span>
                            </div>
                          )}
                        </>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default ChronicDiseasePage;
