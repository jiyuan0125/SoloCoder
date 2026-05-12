import { useState, useEffect } from 'react';
import { api } from '../api';

function formatTime(timeStr) {
  if (!timeStr) return '';
  const time = timeStr.split('T')[1]?.substring(0, 5);
  return time || '';
}

function formatDate(dateStr) {
  if (!dateStr) return '';
  const date = new Date(dateStr);
  const month = date.getMonth() + 1;
  const day = date.getDate();
  return `${month}月${day}日`;
}

function getStatusTag(status) {
  switch (status) {
    case 'not_started':
      return { text: '未开始', className: 'tag-gray' };
    case 'in_progress':
      return { text: '进行中', className: 'tag-blue' };
    case 'completed':
      return { text: '已完成', className: 'tag-green' };
    case 'canceled':
      return { text: '已取消', className: 'tag-red' };
    default:
      return { text: status, className: 'tag-gray' };
  }
}

function getStepDotClass(step) {
  if (step.is_overdue) return 'overdue';
  switch (step.status) {
    case 'completed': return 'completed';
    case 'in_progress': return 'in-progress';
    default: return 'not-started';
  }
}

function isLineCompleted(steps, idx) {
  if (idx >= steps.length - 1) return false;
  return steps[idx].status === 'completed' && steps[idx + 1].status === 'completed';
}

export default function ClinicWorkbench() {
  const [todayPatients, setTodayPatients] = useState([]);
  const [plans, setPlans] = useState([]);
  const [patients, setPatients] = useState([]);
  const [doctors, setDoctors] = useState([]);
  const [selectedPatient, setSelectedPatient] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  useEffect(() => {
    loadData();
  }, []);

  async function loadData() {
    try {
      const [patientsData, doctorsData, plansData, todayData] = await Promise.all([
        api.getPatients(),
        api.getDoctors(),
        api.getTreatmentPlans(),
        api.getTodayPatients(),
      ]);
      setPatients(patientsData);
      setDoctors(doctorsData);
      setPlans(plansData);
      setTodayPatients(todayData || []);
    } catch (err) {
      console.error(err);
    }
  }

  function getPatientPlans(patientId) {
    return plans.filter(p => p.patient_id === patientId);
  }

  async function handleStartStep(planId, stepId) {
    setLoading(true);
    setError('');
    try {
      await api.updateStepStatus(planId, stepId, 'in_progress');
      await loadData();
      setSuccess('步骤已开始');
      setTimeout(() => setSuccess(''), 2000);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  async function handleCompleteStep(planId, stepId) {
    setLoading(true);
    setError('');
    try {
      await api.updateStepStatus(planId, stepId, 'in_progress');
      await new Promise(resolve => setTimeout(resolve, 100));
      await api.updateStepStatus(planId, stepId, 'completed');
      await loadData();
      setSuccess('步骤已完成');
      setTimeout(() => setSuccess(''), 2000);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="container">
      <div className="page-header">
        <h1>诊疗工作台</h1>
        <p>查看今日就诊患者和治疗计划进度</p>
      </div>

      {error && <div className="alert alert-danger">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}

      <div style={{ display: 'grid', gridTemplateColumns: '350px 1fr', gap: '1.5rem' }}>
        <div>
          <div className="card">
            <div className="card-header">
              <h2 className="card-title">今日患者</h2>
              <span className="tag tag-blue">{todayPatients.length}人</span>
            </div>

            {todayPatients.length === 0 ? (
              <div className="empty-state">暂无今日预约患者</div>
            ) : (
              <div className="patient-list">
                {todayPatients.map((item, idx) => {
                  const appointment = item.appointment;
                  const patient = item.patient;
                  
                  return (
                    <div
                      key={idx}
                      className={`patient-card ${selectedPatient?.id === patient.id ? 'selected' : ''}`}
                      onClick={() => setSelectedPatient(patient)}
                    >
                      <div className="patient-card-header">
                        <div className="patient-name">{patient.name}</div>
                        <div className="patient-time">{formatTime(appointment.start_time)}</div>
                      </div>
                      <div className="patient-info">
                        <span>电话: {patient.phone}</span>
                      </div>
                      <div className="patient-info mt-1">
                        <span className="tag tag-blue">
                          {appointment.treatments?.join(', ') || ''}
                        </span>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>

          <div className="card">
            <div className="card-header">
              <h2 className="card-title">全部患者</h2>
            </div>
            <div className="patient-list">
              {patients.map(patient => (
                <div
                  key={patient.id}
                  className={`patient-card ${selectedPatient?.id === patient.id ? 'selected' : ''}`}
                  onClick={() => setSelectedPatient(patient)}
                >
                  <div className="patient-card-header">
                    <div className="patient-name">{patient.name}</div>
                  </div>
                  <div className="patient-info">
                    <span>电话: {patient.phone}</span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        <div>
          {selectedPatient ? (
            <>
              <div className="card">
                <div className="card-header">
                  <h2 className="card-title">患者信息</h2>
                </div>
                <div style={{ display: 'flex', gap: '2rem' }}>
                  <div>
                    <strong>姓名：</strong>{selectedPatient.name}
                  </div>
                  <div>
                    <strong>电话：</strong>{selectedPatient.phone}
                  </div>
                </div>
              </div>

              <div className="card">
                <div className="card-header">
                  <h2 className="card-title">治疗计划</h2>
                </div>

                {getPatientPlans(selectedPatient.id).length === 0 ? (
                  <div className="empty-state">暂无治疗计划</div>
                ) : (
                  getPatientPlans(selectedPatient.id).map(plan => {
                    const statusTag = getStatusTag(plan.status);
                    const totalStepsFee = plan.steps?.reduce((sum, s) => sum + (s.fee || 0), 0) || 0;
                    const totalFee = Math.round(totalStepsFee * (plan.discount_percent / 100));
                    
                    return (
                      <div key={plan.id} className="mb-3" style={{ border: '1px solid #eee', borderRadius: '8px', padding: '1rem' }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
                          <div>
                            <span className={`tag ${statusTag.className}`}>{statusTag.text}</span>
                            {plan.is_surgical && (
                              <span className="tag tag-red" style={{ marginLeft: '0.5rem' }}>手术类</span>
                            )}
                          </div>
                          <div className="text-sm text-muted">
                            折扣: {plan.discount_percent}% | 预估费用: ¥{(totalFee / 100).toFixed(2)}
                          </div>
                        </div>

                        <div className="step-tracker">
                          {plan.steps?.map((step, idx) => {
                            const dotClass = getStepDotClass(step);
                            const lineCompleted = isLineCompleted(plan.steps, idx);

                            return (
                              <div key={step.id} className="step-item">
                                {idx < plan.steps.length - 1 && (
                                  <div className={`step-line ${lineCompleted ? 'completed' : ''}`} />
                                )}
                                <div className={`step-dot ${dotClass}`}>
                                  {step.status === 'completed' ? '✓' : idx + 1}
                                </div>
                                <div className="step-label">
                                  <div style={{ fontWeight: '500' }}>{step.description}</div>
                                  <div className="text-muted">
                                    {formatDate(step.expected_date)}
                                  </div>
                                  {step.is_overdue && (
                                    <div style={{ color: '#ea4335', fontWeight: '500' }}>逾期</div>
                                  )}
                                </div>
                              </div>
                            );
                          })}
                        </div>

                        <div className="mt-2" style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem' }}>
                          {plan.steps?.map(step => (
                            <div key={step.id} style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                              <span className="text-sm">{step.description}</span>
                              {step.status === 'not_started' && (
                                <button
                                  className="btn btn-primary btn-sm"
                                  onClick={() => handleStartStep(plan.id, step.id)}
                                  disabled={loading}
                                >
                                  开始
                                </button>
                              )}
                              {step.status === 'in_progress' && (
                                <button
                                  className="btn btn-success btn-sm"
                                  onClick={() => handleCompleteStep(plan.id, step.id)}
                                  disabled={loading}
                                >
                                  完成
                                </button>
                              )}
                            </div>
                          ))}
                        </div>
                      </div>
                    );
                  })
                )}
              </div>

              <div className="card">
                <div className="card-header">
                  <h2 className="card-title">诊疗记录</h2>
                </div>
                <div className="empty-state">暂无历史记录</div>
              </div>
            </>
          ) : (
            <div className="card">
              <div className="empty-state">请选择一个患者查看详细信息</div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
