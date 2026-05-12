import { useState, useEffect, useCallback } from 'react';
import { statsService } from '../services/api';

function Statistics() {
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const fetchStats = useCallback(async () => {
    setLoading(true);
    try {
      const res = await statsService.getStats();
      setStats(res.data);
    } catch (err) {
      setError('加载统计数据失败');
      console.error(err);
    }
    setLoading(false);
  }, []);

  useEffect(() => {
    fetchStats();
    const interval = setInterval(fetchStats, 10000);
    return () => clearInterval(interval);
  }, [fetchStats]);

  if (loading && !stats) {
    return (
      <div className="page-content">
        <div className="empty-state">加载中...</div>
      </div>
    );
  }

  if (!stats) {
    return (
      <div className="page-content">
        <div className="empty-state">暂无统计数据</div>
      </div>
    );
  }

  return (
    <div className="page-content">
      {error && (
        <div style={{ color: '#dc3545', marginBottom: '15px', padding: '10px', backgroundColor: '#f8d7da', borderRadius: '4px' }}>
          {error}
          <button onClick={() => setError(null)} style={{ marginLeft: '15px', background: 'none', border: 'none', cursor: 'pointer' }}>×</button>
        </div>
      )}

      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
        <h2 style={{ fontSize: 20, fontWeight: 600 }}>统计面板</h2>
        <button 
          className="refresh-btn" 
          onClick={fetchStats}
          disabled={loading}
        >
          {loading ? '刷新中...' : '刷新'}
        </button>
      </div>

      <div className="stats-grid">
        <div className="stats-card">
          <div className="stats-card-title">今日接警数</div>
          <div className="stats-card-value">{stats.today_calls_count}</div>
          <div className="stats-card-subtitle">总计: {stats.total_calls_count}</div>
        </div>

        <div className="stats-card">
          <div className="stats-card-title">平均响应时间</div>
          <div className="stats-card-value">{stats.average_response_time_minutes.toFixed(1)}</div>
          <div className="stats-card-subtitle">单位: 分钟</div>
        </div>

        <div className="stats-card">
          <div className="stats-card-title">车辆利用率</div>
          <div className="stats-card-value">{stats.vehicle_utilization_rate.toFixed(1)}%</div>
          <div className="stats-card-subtitle">基于状态历史计算</div>
        </div>

        <div className="stats-card">
          <div className="stats-card-title">等待队列</div>
          <div className="stats-card-value">{stats.waiting_in_queue_count}</div>
          <div className="stats-card-subtitle">等待派车的求救</div>
        </div>

        <div className="stats-card">
          <div className="stats-card-title">一级/二级求救 (今日)</div>
          <div className="stats-card-value" style={{ color: '#dc3545' }}>{stats.level12_calls_today}</div>
          <div className="stats-card-subtitle">需要立即派车</div>
        </div>

        <div className="stats-card">
          <div className="stats-card-title">三级/四级求救 (今日)</div>
          <div className="stats-card-value" style={{ color: '#28a745' }}>{stats.level34_calls_today}</div>
          <div className="stats-card-subtitle">可排队等候</div>
        </div>
      </div>

      <div style={{ marginTop: 30, backgroundColor: 'white', padding: 25, borderRadius: 8, boxShadow: '0 2px 8px rgba(0,0,0,0.08)' }}>
        <h3 style={{ marginBottom: 20, fontSize: 16 }}>车辆状态概览</h3>
        <div style={{ display: 'flex', gap: 30 }}>
          <div style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 36, fontWeight: 700, color: '#28a745' }}>{stats.idle_vehicles_count}</div>
            <div style={{ color: '#6c757d', fontSize: 14 }}>空闲</div>
          </div>
          <div style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 36, fontWeight: 700, color: '#007bff' }}>{stats.in_transit_vehicles_count}</div>
            <div style={{ color: '#6c757d', fontSize: 14 }}>出车中</div>
          </div>
          <div style={{ textAlign: 'center' }}>
            <div style={{ fontSize: 36, fontWeight: 700, color: '#dc3545' }}>{stats.maintenance_vehicles_count}</div>
            <div style={{ color: '#6c757d', fontSize: 14 }}>维护中</div>
          </div>
        </div>
      </div>

      <div style={{ marginTop: 30, backgroundColor: 'white', padding: 25, borderRadius: 8, boxShadow: '0 2px 8px rgba(0,0,0,0.08)' }}>
        <h3 style={{ marginBottom: 15, fontSize: 16 }}>业务规则说明</h3>
        <ul style={{ fontSize: 14, color: '#6c757d', lineHeight: 2, paddingLeft: 20 }}>
          <li><strong>派车规则:</strong> 一级濒危、二级危重必须立即派车；三级急症5分钟内派车；四级非急症可以排队等候</li>
          <li><strong>车辆承载:</strong> 每辆救护车最多承载3名患者，超过需要增派车辆</li>
          <li><strong>状态流转:</strong> 空闲 → 出车中 → 返回途中 → 空闲</li>
          <li><strong>超时预警:</strong> 出车超过60分钟未更新状态产生预警</li>
          <li><strong>自动升级:</strong> 四级求救等待超过10分钟自动升级为三级</li>
          <li><strong>超时关闭:</strong> 超过48小时未接单自动关闭</li>
        </ul>
      </div>
    </div>
  );
}

export default Statistics;
