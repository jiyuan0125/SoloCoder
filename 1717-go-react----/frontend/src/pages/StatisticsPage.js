import React, { useState, useEffect } from 'react';
import { statisticsAPI } from '../api';
import LineChart from '../components/LineChart';
import BarChart from '../components/BarChart';

const INFECTION_SITES = {
  respiratory: '呼吸道',
  surgical_incision: '手术切口',
  urinary_tract: '泌尿道',
  bloodstream: '血液',
  digestive: '消化系统',
  skin_soft_tissue: '皮肤软组织',
};

function StatisticsPage() {
  const [trendData, setTrendData] = useState([]);
  const [stats, setStats] = useState(null);
  const [alerts, setAlerts] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [selectedMonth, setSelectedMonth] = useState('');

  useEffect(() => {
    loadData();
  }, [selectedMonth]);

  const loadData = async () => {
    setLoading(true);
    setError('');
    try {
      const [trendRes, statsRes, alertsRes] = await Promise.all([
        statisticsAPI.getTrend(),
        statisticsAPI.getRates(selectedMonth),
        statisticsAPI.getAlerts({ status: 'active' }),
      ]);
      setTrendData(trendRes.data);
      setStats(statsRes.data);
      setAlerts(alertsRes.data);
    } catch (err) {
      setError(err.response?.data?.error || '加载数据失败');
    } finally {
      setLoading(false);
    }
  };

  const getCurrentMonth = () => {
    const now = new Date();
    return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`;
  };

  return (
    <div>
      <div className="page-header">
        <h2>统计面板</h2>
        <div style={{ display: 'flex', gap: '12px' }}>
          <input
            type="month"
            value={selectedMonth}
            onChange={(e) => setSelectedMonth(e.target.value)}
          />
          <button className="btn btn-secondary" onClick={() => setSelectedMonth('')}>
            重置
          </button>
        </div>
      </div>

      {error && <div className="alert-box alert-error">{error}</div>}

      {stats && (
        <div className="stats-grid">
          <div className="stat-card">
            <div className="label">全院感染率</div>
            <div className="value">{stats.hospital_rate.toFixed(2)}%</div>
            <div className="trend" style={{ color: stats.hospital_rate > 2 ? '#e53e3e' : '#38a169' }}>
              预警阈值: 2.00%
            </div>
          </div>
          <div className="stat-card">
            <div className="label">活跃预警</div>
            <div className="value" style={{ color: alerts.length > 0 ? '#e53e3e' : '#38a169' }}>
              {alerts.length}
            </div>
            <div className="trend">
              {alerts.filter(a => a.alert_type === 'need_intervention').length} 个需要干预
            </div>
          </div>
          <div className="stat-card">
            <div className="label">统计月份</div>
            <div className="value" style={{ fontSize: '20px' }}>{stats.month || getCurrentMonth()}</div>
          </div>
          <div className="stat-card">
            <div className="label">科室数量</div>
            <div className="value">{stats.department_rates?.length || 0}</div>
          </div>
        </div>
      )}

      <div className="card">
        <LineChart data={trendData} />
      </div>

      <div className="card">
        {stats?.department_rates ? (
          <BarChart data={stats.department_rates} />
        ) : (
          <div className="empty-state">加载中...</div>
        )}
      </div>

      {alerts.length > 0 && (
        <div className="card">
          <h3>预警通知</h3>
          <ul className="ranking-list">
            {alerts.map((alert, i) => (
              <li key={alert.id}>
                <div className={`rank-number ${alert.alert_type === 'need_intervention' ? 'top-1' : ''}`}>
                  {i + 1}
                </div>
                <div className="rank-name">
                  <div style={{ fontWeight: 600 }}>
                    {alert.department?.name || '未知科室'} - {alert.month}
                  </div>
                  <div style={{ fontSize: '12px', color: '#718096', marginTop: '4px' }}>
                    {alert.message}
                  </div>
                </div>
                <span className={`badge ${alert.alert_type === 'need_intervention' ? 'badge-danger' : 'badge-warning'}`}>
                  {alert.alert_type === 'need_intervention' ? '需要干预' : '感染率超标'}
                </span>
              </li>
            ))}
          </ul>
        </div>
      )}

      {stats && (
        <div className="distribution-grid">
          <div className="card">
            <h3>感染部位分布</h3>
            {stats.site_distribution?.length > 0 ? (
              <div>
                {stats.site_distribution.map((item, i) => (
                  <div key={i} style={{ marginBottom: '16px' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '4px' }}>
                      <span style={{ fontSize: '14px', color: '#4a5568' }}>
                        {INFECTION_SITES[item.site] || item.site}
                      </span>
                      <span style={{ fontSize: '14px', fontWeight: 600, color: '#2d3748' }}>
                        {item.count}例 ({item.ratio.toFixed(1)}%)
                      </span>
                    </div>
                    <div className="progress-bar">
                      <div
                        className="progress-bar-fill"
                        style={{ width: `${Math.min(item.ratio, 100)}%` }}
                      />
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="empty-state" style={{ padding: '20px' }}>暂无数据</div>
            )}
          </div>

          <div className="card">
            <h3>病原体分布</h3>
            {stats.pathogen_distribution?.length > 0 ? (
              <div>
                {stats.pathogen_distribution.slice(0, 8).map((item, i) => (
                  <div key={i} style={{ marginBottom: '16px' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '4px' }}>
                      <span style={{ fontSize: '14px', color: '#4a5568' }}>{item.pathogen}</span>
                      <span style={{ fontSize: '14px', fontWeight: 600, color: '#2d3748' }}>
                        {item.count}例 ({item.ratio.toFixed(1)}%)
                      </span>
                    </div>
                    <div className="progress-bar">
                      <div
                        className="progress-bar-fill"
                        style={{ width: `${Math.min(item.ratio, 100)}%` }}
                      />
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="empty-state" style={{ padding: '20px' }}>暂无数据</div>
            )}
          </div>
        </div>
      )}

      {stats?.department_rates && (
        <div className="card">
          <h3>各科室感染率排名</h3>
          <ul className="ranking-list">
            {stats.department_rates.map((dept, i) => (
              <li key={dept.department_id}>
                <div className={`rank-number ${i === 0 ? 'top-1' : i === 1 ? 'top-2' : i === 2 ? 'top-3' : ''}`}>
                  {i + 1}
                </div>
                <div className="rank-name">{dept.department_name}</div>
                <div style={{ marginRight: '16px', fontSize: '13px', color: '#718096' }}>
                  {dept.infection_count}例 / {dept.discharge_count}人
                </div>
                <div className="rank-value" style={{ color: dept.exceeded ? '#e53e3e' : '#38a169' }}>
                  {dept.infection_rate.toFixed(2)}%
                  {dept.exceeded && <span style={{ fontSize: '10px', marginLeft: '4px' }}>超标</span>}
                </div>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}

export default StatisticsPage;
