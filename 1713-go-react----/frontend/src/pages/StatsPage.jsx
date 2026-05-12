import { useState, useEffect } from 'react';
import { api } from '../api';

function BarChart({ data, labelKey, valueKey, colorClass = '' }) {
  if (!data || data.length === 0) {
    return <div className="empty">暂无数据</div>;
  }

  const maxValue = Math.max(...data.map((d) => d[valueKey]), 1);

  return (
    <div className="bar-chart">
      {data.map((item, idx) => {
        const height = (item[valueKey] / maxValue) * 150;
        return (
          <div key={idx} className="bar-item">
            <div className={`bar ${colorClass}`} style={{ height: `${Math.max(height, 4)}px` }}>
              <span className="bar-value">{item[valueKey]?.toFixed ? item[valueKey].toFixed(1) : item[valueKey]}%</span>
            </div>
            <div className="bar-label">{item[labelKey]}</div>
          </div>
        );
      })}
    </div>
  );
}

export default function StatsPage() {
  const [metrics, setMetrics] = useState(null);
  const [todos, setTodos] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    loadData();
  }, []);

  async function loadData() {
    setLoading(true);
    try {
      const [metricsRes, todosRes] = await Promise.all([
        api.stats.metrics(),
        api.todos.list({ page: 1, size: 50 }),
      ]);
      setMetrics(metricsRes);
      setTodos(todosRes.data || []);
    } catch (e) {
      setError(e.message);
    }
    setLoading(false);
  }

  async function handleTodoStatus(todo, status) {
    try {
      await api.todos.update(todo.id, { status });
      loadData();
    } catch (e) {
      setError(e.message);
    }
  }

  function formatDate(d) {
    if (!d) return '-';
    return new Date(d).toLocaleDateString('zh-CN');
  }

  const pendingTodos = todos.filter((t) => t.status === '待处理');

  return (
    <div>
      <div className="page-header">
        <h1>统计面板</h1>
      </div>

      {error && <div className="alert alert-error">{error}</div>}

      {pendingTodos.length > 0 && (
        <div className="card" style={{ borderLeft: '4px solid #f39c12' }}>
          <h3 style={{ color: '#e67e22' }}>待处理事项 ({pendingTodos.length})</h3>
          <table>
            <thead>
              <tr>
                <th>类型</th>
                <th>标题</th>
                <th>负责人</th>
                <th>截止日期</th>
                <th>优先级</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {pendingTodos.map((t) => (
                <tr key={t.id}>
                  <td><span className="badge badge-status">{t.type}</span></td>
                  <td>{t.title}</td>
                  <td>{t.assignee}</td>
                  <td>{formatDate(t.due_date)}</td>
                  <td>
                    <span className={`badge ${t.priority === '高' ? 'badge-exceed' : t.priority === '中' ? 'badge-suspected' : 'badge-normal'}`}>
                      {t.priority}
                    </span>
                  </td>
                  <td>
                    <button className="btn btn-sm btn-primary" onClick={() => handleTodoStatus(t, '处理中')}>开始处理</button>
                    <button className="btn btn-sm btn-secondary" onClick={() => handleTodoStatus(t, '已完成')}>完成</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {loading ? (
        <div className="loading">加载中...</div>
      ) : metrics ? (
        <>
          <div className="stat-grid">
            <div className="stat-card">
              <div className="stat-label">行业分类数</div>
              <div className="stat-value">{metrics.industry_coverage?.length || 0}</div>
            </div>
            <div className="stat-card">
              <div className="stat-label">监测危害因素数</div>
              <div className="stat-value">{metrics.factor_exceed_rates?.length || 0}</div>
            </div>
            <div className="stat-card">
              <div className="stat-label">体检类型数</div>
              <div className="stat-value">{metrics.disease_detection_rates?.length || 0}</div>
            </div>
            <div className="stat-card">
              <div className="stat-label">待处理事项</div>
              <div className="stat-value" style={{ color: pendingTodos.length > 0 ? '#e74c3c' : undefined }}>
                {pendingTodos.length}
              </div>
            </div>
          </div>

          <div className="card chart-section">
            <h3>各行业体检覆盖率</h3>
            <BarChart
              data={metrics.industry_coverage?.map((i) => ({
                industry: i.industry,
                coverage: i.coverage || 0,
              }))}
              labelKey="industry"
              valueKey="coverage"
              colorClass="bar-green"
            />
          </div>

          <div className="card chart-section">
            <h3>各危害因素超标率</h3>
            <BarChart
              data={metrics.factor_exceed_rates?.map((f) => ({
                factor: f.factor,
                rate: f.rate || 0,
              }))}
              labelKey="factor"
              valueKey="rate"
              colorClass="bar-red"
            />
            <div style={{ marginTop: 16, fontSize: 13, color: '#666' }}>
              {metrics.factor_exceed_rates?.map((f, idx) => (
                <span key={idx} style={{ marginRight: 16 }}>
                  {f.factor}: {f.exceeded}/{f.total} 次超标 ({f.rate_str})
                </span>
              ))}
            </div>
          </div>

          <div className="card chart-section">
            <h3>各类型职业病检出率</h3>
            <BarChart
              data={metrics.disease_detection_rates?.map((d) => ({
                exam_type: d.exam_type,
                rate: d.rate || 0,
              }))}
              labelKey="exam_type"
              valueKey="rate"
              colorClass="bar-orange"
            />
          </div>
        </>
      ) : (
        <div className="empty">暂无统计数据</div>
      )}
    </div>
  );
}
