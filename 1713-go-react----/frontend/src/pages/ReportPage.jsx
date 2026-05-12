import { useState, useEffect } from 'react';
import { api } from '../api';

export default function ReportPage() {
  const [reports, setReports] = useState([]);
  const [enterprises, setEnterprises] = useState([]);
  const [regulatoryStats, setRegulatoryStats] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  const [activeTab, setActiveTab] = useState('enterprise');

  const [genForm, setGenForm] = useState({
    enterprise_id: '',
    year: new Date().getFullYear(),
  });

  const [filterRegion, setFilterRegion] = useState('');
  const [filterYear, setFilterYear] = useState(new Date().getFullYear());

  useEffect(() => {
    loadData();
  }, [activeTab]);

  useEffect(() => {
    if (activeTab === 'regulatory') {
      loadRegulatoryStats();
    }
  }, [activeTab, filterRegion, filterYear]);

  async function loadData() {
    setLoading(true);
    try {
      const [reportRes, entRes] = await Promise.all([
        api.reports.list({ page: 1, size: 100 }),
        api.enterprises.list(1, 100),
      ]);
      setReports(reportRes.data || []);
      setEnterprises(entRes.data || []);
    } catch (e) {
      setError(e.message);
    }
    setLoading(false);
  }

  async function loadRegulatoryStats() {
    try {
      const params = { year: filterYear };
      if (filterRegion) params.region = filterRegion;
      const res = await api.stats.regulatory(params);
      setRegulatoryStats(res);
    } catch (e) {
      setError(e.message);
    }
  }

  async function handleGenerate(e) {
    e.preventDefault();
    setError(null);
    try {
      await api.reports.generate(genForm);
      loadData();
      setSuccess('报告生成成功');
      setTimeout(() => setSuccess(null), 3000);
    } catch (e) {
      setError(e.message);
    }
  }

  function formatDate(d) {
    if (!d) return '-';
    return new Date(d).toLocaleString('zh-CN');
  }

  const regions = [...new Set(enterprises.map((e) => e.region).filter(Boolean))];

  return (
    <div>
      <div className="page-header">
        <h1>报告管理</h1>
      </div>

      {error && <div className="alert alert-error">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}

      <div className="tabs">
        <button className={activeTab === 'enterprise' ? 'active' : ''} onClick={() => setActiveTab('enterprise')}>企业报告</button>
        <button className={activeTab === 'regulatory' ? 'active' : ''} onClick={() => setActiveTab('regulatory')}>监管统计</button>
      </div>

      {activeTab === 'enterprise' && (
        <>
          <div className="card">
            <h3>生成企业年度报告</h3>
            <form onSubmit={handleGenerate}>
              <div className="form-row">
                <div className="form-group">
                  <label>选择企业</label>
                  <select required value={genForm.enterprise_id} onChange={(e) => setGenForm({ ...genForm, enterprise_id: e.target.value })}>
                    <option value="">请选择企业</option>
                    {enterprises.map((e) => <option key={e.id} value={e.id}>{e.name}</option>)}
                  </select>
                </div>
                <div className="form-group">
                  <label>年份</label>
                  <input type="number" required value={genForm.year} onChange={(e) => setGenForm({ ...genForm, year: parseInt(e.target.value) })} min={2000} max={2100} />
                </div>
                <div className="form-group" style={{ alignSelf: 'flex-end' }}>
                  <button type="submit" className="btn btn-primary">生成报告</button>
                </div>
              </div>
            </form>
          </div>

          <div className="card">
            <h3>报告列表</h3>
            {loading ? (
              <div className="loading">加载中...</div>
            ) : reports.length === 0 ? (
              <div className="empty">暂无报告，请先生成</div>
            ) : (
              <table>
                <thead>
                  <tr>
                    <th>企业名称</th>
                    <th>年份</th>
                    <th>生成时间</th>
                    <th>操作</th>
                  </tr>
                </thead>
                <tbody>
                  {reports.map((r) => (
                    <tr key={r.id}>
                      <td>{r.enterprise?.name || '-'}</td>
                      <td>{r.year}</td>
                      <td>{formatDate(r.generated_at)}</td>
                      <td>
                        <a href={api.reports.exportUrl(r.id)} className="btn btn-sm btn-primary" target="_blank" rel="noreferrer">
                          导出
                        </a>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </>
      )}

      {activeTab === 'regulatory' && (
        <>
          <div className="card">
            <h3>筛选条件</h3>
            <div className="form-row">
              <div className="form-group">
                <label>地区</label>
                <select value={filterRegion} onChange={(e) => setFilterRegion(e.target.value)}>
                  <option value="">全部地区</option>
                  {regions.map((r) => <option key={r} value={r}>{r}</option>)}
                </select>
              </div>
              <div className="form-group">
                <label>年份</label>
                <input type="number" value={filterYear} onChange={(e) => setFilterYear(parseInt(e.target.value))} min={2000} max={2100} />
              </div>
            </div>
          </div>

          {regulatoryStats && (
            <div className="stat-grid">
              <div className="stat-card">
                <div className="stat-label">企业数量</div>
                <div className="stat-value">{regulatoryStats.total_enterprises}</div>
              </div>
              <div className="stat-card">
                <div className="stat-label">接触危害因素人数</div>
                <div className="stat-value">{regulatoryStats.exposed_workers}</div>
              </div>
              <div className="stat-card">
                <div className="stat-label">体检覆盖率</div>
                <div className="stat-value">{regulatoryStats.coverage_rate}</div>
              </div>
              <div className="stat-card">
                <div className="stat-label">检出疑似职业病人数</div>
                <div className="stat-value" style={{ color: regulatoryStats.suspected_count > 0 ? '#e74c3c' : undefined }}>
                  {regulatoryStats.suspected_count}
                </div>
              </div>
            </div>
          )}

          {regulatoryStats && (
            <div className="card">
              <h3>监管统计详情</h3>
              <table>
                <thead>
                  <tr>
                    <th>指标</th>
                    <th>数值</th>
                  </tr>
                </thead>
                <tbody>
                  <tr><td>地区</td><td>{regulatoryStats.region || '全部'}</td></tr>
                  <tr><td>统计年度</td><td>{regulatoryStats.year}</td></tr>
                  <tr><td>企业总数</td><td>{regulatoryStats.total_enterprises} 家</td></tr>
                  <tr><td>接触危害因素人数</td><td>{regulatoryStats.exposed_workers} 人</td></tr>
                  <tr><td>应体检人数</td><td>{regulatoryStats.should_examine} 人</td></tr>
                  <tr><td>已体检人数</td><td>{regulatoryStats.examined_workers} 人</td></tr>
                  <tr><td>体检覆盖率</td><td><strong>{regulatoryStats.coverage_rate}</strong></td></tr>
                  <tr><td>体检异常人数</td><td>{regulatoryStats.abnormal_count} 人</td></tr>
                  <tr><td>检出疑似职业病人数</td><td style={{ color: '#e74c3c', fontWeight: 600 }}>{regulatoryStats.suspected_count} 人</td></tr>
                </tbody>
              </table>
            </div>
          )}
        </>
      )}
    </div>
  );
}
