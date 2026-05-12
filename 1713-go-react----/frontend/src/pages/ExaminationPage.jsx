import { useState, useEffect } from 'react';
import { api } from '../api';

const EXAM_TYPES = ['上岗前体检', '在岗期间体检', '离岗时体检'];
const FLOW_STATES = ['登记', '等待', '进行中', '检查', '完成', '归档'];

export default function ExaminationPage() {
  const [exams, setExams] = useState([]);
  const [enterprises, setEnterprises] = useState([]);
  const [workers, setWorkers] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);

  const [activeTab, setActiveTab] = useState('all');
  const [showScheduleModal, setShowScheduleModal] = useState(false);
  const [showResultModal, setShowResultModal] = useState(false);
  const [selectedExam, setSelectedExam] = useState(null);

  const [scheduleForm, setScheduleForm] = useState({
    worker_id: '',
    enterprise_id: '',
    exam_type: '在岗期间体检',
    scheduled_date: '',
  });

  const [resultForm, setResultForm] = useState({
    exam_date: '',
    exam_items: [],
  });

  const [filterEnt, setFilterEnt] = useState('');
  const [filterAbnormal, setFilterAbnormal] = useState('');

  useEffect(() => {
    loadData();
  }, [activeTab, filterEnt, filterAbnormal]);

  useEffect(() => {
    loadEnterprises();
  }, []);

  async function loadEnterprises() {
    try {
      const res = await api.enterprises.list(1, 100);
      setEnterprises(res.data || []);
    } catch (e) {
      setError(e.message);
    }
  }

  async function loadData() {
    setLoading(true);
    try {
      const params = { page: 1, size: 100 };
      if (filterEnt) params.enterprise_id = filterEnt;
      if (activeTab === 'abnormal') params.has_abnormal = 'true';
      if (activeTab === 'suspected') params.is_suspected = 'true';

      const [examRes, workerRes] = await Promise.all([
        api.examinations.list(params),
        api.workers.list({ page: 1, size: 100 }),
      ]);
      setExams(examRes.data || []);
      setWorkers(workerRes.data || []);
    } catch (e) {
      setError(e.message);
    }
    setLoading(false);
  }

  async function handleScheduleSubmit(e) {
    e.preventDefault();
    setError(null);
    try {
      await api.examinations.schedule(scheduleForm);
      setShowScheduleModal(false);
      setScheduleForm({
        worker_id: '',
        enterprise_id: '',
        exam_type: '在岗期间体检',
        scheduled_date: '',
      });
      loadData();
      setSuccess('体检安排成功');
      setTimeout(() => setSuccess(null), 3000);
    } catch (e) {
      setError(e.message);
    }
  }

  async function openResultModal(exam) {
    setSelectedExam(exam);
    try {
      const detail = await api.examinations.get(exam.id);
      const items = (detail.exam_items || []).map((item) => ({
        item_code: item.item_code,
        item_name: item.item_name,
        result: item.result || '',
        is_abnormal: item.is_abnormal || false,
        is_suspected: item.is_suspected || false,
      }));
      setResultForm({
        exam_date: exam.exam_date ? new Date(exam.exam_date).toISOString().split('T')[0] : '',
        exam_items: items,
      });
      setShowResultModal(true);
    } catch (e) {
      setError(e.message);
    }
  }

  async function handleResultSubmit(e) {
    e.preventDefault();
    setError(null);
    try {
      await api.examinations.recordResult(selectedExam.id, resultForm);
      setShowResultModal(false);
      setSelectedExam(null);
      loadData();
      setSuccess('结果录入成功，已自动生成相应待办事项');
      setTimeout(() => setSuccess(null), 3000);
    } catch (e) {
      setError(e.message);
    }
  }

  async function updateFlow(exam, nextState) {
    try {
      await api.examinations.updateFlow(exam.id, nextState);
      loadData();
    } catch (e) {
      setError(e.message);
    }
  }

  function updateItemResult(index, field, value) {
    const newItems = [...resultForm.exam_items];
    newItems[index] = { ...newItems[index], [field]: value };
    setResultForm({ ...resultForm, exam_items: newItems });
  }

  function formatDate(d) {
    if (!d) return '-';
    return new Date(d).toLocaleDateString('zh-CN');
  }

  function getWorkersForEnterprise(entId) {
    return workers.filter((w) => w.enterprise_id === parseInt(entId));
  }

  return (
    <div>
      <div className="page-header">
        <h1>体检管理</h1>
        <button className="btn btn-primary" onClick={() => setShowScheduleModal(true)}>
          + 安排体检
        </button>
      </div>

      {error && <div className="alert alert-error">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}

      <div className="filter-bar">
        <select value={filterEnt} onChange={(e) => setFilterEnt(e.target.value)}>
          <option value="">全部企业</option>
          {enterprises.map((e) => (
            <option key={e.id} value={e.id}>{e.name}</option>
          ))}
        </select>
      </div>

      <div className="tabs">
        <button className={activeTab === 'all' ? 'active' : ''} onClick={() => setActiveTab('all')}>全部</button>
        <button className={activeTab === 'abnormal' ? 'active' : ''} onClick={() => setActiveTab('abnormal')}>异常结果</button>
        <button className={activeTab === 'suspected' ? 'active' : ''} onClick={() => setActiveTab('suspected')}>疑似职业病</button>
      </div>

      <div className="card">
        {loading ? (
          <div className="loading">加载中...</div>
        ) : exams.length === 0 ? (
          <div className="empty">暂无体检记录</div>
        ) : (
          <table>
            <thead>
              <tr>
                <th></th>
                <th>劳动者</th>
                <th>企业</th>
                <th>体检类型</th>
                <th>体检周期</th>
                <th>安排日期</th>
                <th>体检日期</th>
                <th>状态</th>
                <th>流程状态</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {exams.map((exam) => (
                <tr key={exam.id} className={exam.has_abnormal ? 'row-abnormal' : ''}>
                  <td>
                    {exam.is_suspected && <span className="icon-suspected" title="疑似职业病">!</span>}
                  </td>
                  <td>{exam.worker?.name || '-'}</td>
                  <td>{exam.enterprise?.name || '-'}</td>
                  <td>{exam.exam_type}</td>
                  <td>{exam.exam_cycle_months}个月</td>
                  <td>{formatDate(exam.scheduled_date)}</td>
                  <td>{formatDate(exam.exam_date)}</td>
                  <td>
                    <span className={`badge ${exam.is_suspected ? 'badge-suspected' : exam.has_abnormal ? 'badge-exceed' : 'badge-normal'}`}>
                      {exam.status}
                    </span>
                  </td>
                  <td><span className="badge badge-status">{exam.flow_status}</span></td>
                  <td>
                    {exam.flow_status !== '归档' && exam.status !== '已完成' && (
                      <button className="btn btn-sm btn-primary" onClick={() => openResultModal(exam)}>录入结果</button>
                    )}
                    {exam.flow_status === '等待' && (
                      <button className="btn btn-sm btn-secondary" onClick={() => updateFlow(exam, '进行中')}>开始</button>
                    )}
                    {exam.flow_status === '进行中' && (
                      <button className="btn btn-sm btn-secondary" onClick={() => updateFlow(exam, '检查')}>检查</button>
                    )}
                    {exam.flow_status === '检查' && (
                      <>
                        <button className="btn btn-sm btn-secondary" onClick={() => updateFlow(exam, '进行中')}>继续</button>
                        <button className="btn btn-sm btn-secondary" onClick={() => updateFlow(exam, '完成')}>完成</button>
                      </>
                    )}
                    {exam.flow_status === '完成' && (
                      <button className="btn btn-sm btn-primary" onClick={() => updateFlow(exam, '归档')}>归档</button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {showScheduleModal && (
        <div className="modal-overlay" onClick={() => setShowScheduleModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>安排体检</h3>
              <button className="modal-close" onClick={() => setShowScheduleModal(false)}>×</button>
            </div>
            <form onSubmit={handleScheduleSubmit}>
              <div className="form-group">
                <label>选择企业 *</label>
                <select required value={scheduleForm.enterprise_id} onChange={(e) => setScheduleForm({ ...scheduleForm, enterprise_id: e.target.value, worker_id: '' })}>
                  <option value="">请选择企业</option>
                  {enterprises.map((e) => <option key={e.id} value={e.id}>{e.name}</option>)}
                </select>
              </div>
              <div className="form-group">
                <label>选择劳动者 *</label>
                <select required value={scheduleForm.worker_id} onChange={(e) => setScheduleForm({ ...scheduleForm, worker_id: e.target.value })} disabled={!scheduleForm.enterprise_id}>
                  <option value="">请选择劳动者</option>
                  {getWorkersForEnterprise(scheduleForm.enterprise_id).map((w) => (
                    <option key={w.id} value={w.id}>{w.name} - {w.post_name || '未分配岗位'}</option>
                  ))}
                </select>
              </div>
              <div className="form-row">
                <div className="form-group">
                  <label>体检类型 *</label>
                  <select value={scheduleForm.exam_type} onChange={(e) => setScheduleForm({ ...scheduleForm, exam_type: e.target.value })}>
                    {EXAM_TYPES.map((t) => <option key={t} value={t}>{t}</option>)}
                  </select>
                </div>
                <div className="form-group">
                  <label>安排日期 *</label>
                  <input type="date" required value={scheduleForm.scheduled_date} onChange={(e) => setScheduleForm({ ...scheduleForm, scheduled_date: e.target.value })} />
                </div>
              </div>
              <div className="modal-footer">
                <button type="button" className="btn btn-secondary" onClick={() => setShowScheduleModal(false)}>取消</button>
                <button type="submit" className="btn btn-primary">安排</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {showResultModal && selectedExam && (
        <div className="modal-overlay" onClick={() => setShowResultModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()} style={{ maxWidth: 800 }}>
            <div className="modal-header">
              <h3>录入体检结果 - {selectedExam.worker?.name}</h3>
              <button className="modal-close" onClick={() => setShowResultModal(false)}>×</button>
            </div>
            <form onSubmit={handleResultSubmit}>
              <div className="form-group">
                <label>体检日期</label>
                <input type="date" value={resultForm.exam_date} onChange={(e) => setResultForm({ ...resultForm, exam_date: e.target.value })} />
              </div>
              <div className="card" style={{ marginBottom: 8, padding: 12 }}>
                <h4 style={{ marginBottom: 12 }}>体检项目</h4>
                <div style={{ fontSize: 12, color: '#666', marginBottom: 8 }}>
                  提示：勾选"疑似职业病"会自动生成职业病诊断待办，仅勾选"异常"会生成复查确认待办
                </div>
                <div style={{ display: 'grid', gridTemplateColumns: '2fr 2fr 80px 80px', gap: 12, fontWeight: 600, fontSize: 13, padding: '8px 0', borderBottom: '2px solid #eee' }}>
                  <div>项目名称</div>
                  <div>检查结果</div>
                  <div>异常</div>
                  <div>疑似职业病</div>
                </div>
                {resultForm.exam_items.map((item, idx) => (
                  <div key={item.item_code} style={{ display: 'grid', gridTemplateColumns: '2fr 2fr 80px 80px', gap: 12, alignItems: 'center', padding: '8px 0', borderBottom: '1px solid #eee' }}>
                    <div style={{ fontSize: 14 }}>{item.item_name}</div>
                    <input
                      value={item.result}
                      onChange={(e) => updateItemResult(idx, 'result', e.target.value)}
                      placeholder="输入结果"
                      style={{ padding: '6px 10px', border: '1px solid #ddd', borderRadius: 4, fontSize: 14 }}
                    />
                    <div style={{ textAlign: 'center' }}>
                      <input
                        type="checkbox"
                        checked={item.is_abnormal}
                        onChange={(e) => updateItemResult(idx, 'is_abnormal', e.target.checked)}
                      />
                    </div>
                    <div style={{ textAlign: 'center' }}>
                      <input
                        type="checkbox"
                        checked={item.is_suspected}
                        onChange={(e) => updateItemResult(idx, 'is_suspected', e.target.checked)}
                      />
                    </div>
                  </div>
                ))}
                {resultForm.exam_items.length === 0 && (
                  <div className="empty" style={{ margin: '16px 0' }}>暂无体检项目</div>
                )}
              </div>
              <div className="modal-footer">
                <button type="button" className="btn btn-secondary" onClick={() => setShowResultModal(false)}>取消</button>
                <button type="submit" className="btn btn-primary">保存结果</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
