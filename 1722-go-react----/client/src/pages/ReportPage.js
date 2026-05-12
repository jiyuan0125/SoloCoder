import React, { useState, useEffect } from 'react';
import { api } from '../utils/api';

const MASTERY_LABELS = {
  mastered: { text: '已掌握', class: 'badge-mastered' },
  reinforce: { text: '需巩固', class: 'badge-reinforce' },
  weak: { text: '薄弱', class: 'badge-weak' },
};

function RadarChart({ data }) {
  const cx = 250;
  const cy = 250;
  const maxRadius = 180;
  const levels = 5;

  if (!data || data.length === 0) {
    return (
      <div className="radar-container">
        <div className="empty-state">
          <p>暂无知识点掌握数据</p>
        </div>
      </div>
    );
  }

  const displayData = data.slice(0, 8);
  const angleStep = (Math.PI * 2) / displayData.length;

  const getPoint = (index, radius) => {
    const angle = -Math.PI / 2 + index * angleStep;
    return {
      x: cx + radius * Math.cos(angle),
      y: cy + radius * Math.sin(angle),
    };
  };

  const gridPath = [];
  for (let level = 1; level <= levels; level++) {
    const radius = (maxRadius / levels) * level;
    const points = displayData.map((_, i) => {
      const p = getPoint(i, radius);
      return `${p.x},${p.y}`;
    });
    gridPath.push(<polygon key={level} points={points.join(' ')} fill="none" stroke="#e2e8f0" strokeWidth="1" />);
  }

  const axisLines = displayData.map((_, i) => {
    const outer = getPoint(i, maxRadius);
    return <line key={i} x1={cx} y1={cy} x2={outer.x} y2={outer.y} stroke="#e2e8f0" strokeWidth="1" />;
  });

  const dataPoints = displayData.map((item, i) => {
    const radius = (item.accuracy / 100) * maxRadius;
    return getPoint(i, radius);
  });

  const dataPath = dataPoints.map((p) => `${p.x},${p.y}`).join(' ');

  const labels = displayData.map((item, i) => {
    const pos = getPoint(i, maxRadius + 25);
    return (
      <text
        key={i}
        x={pos.x}
        y={pos.y}
        textAnchor="middle"
        dominantBaseline="middle"
        fontSize="11"
        fill="#4a5568"
      >
        {item.knowledge_point_name.length > 8
          ? item.knowledge_point_name.substring(0, 8) + '...'
          : item.knowledge_point_name}
      </text>
    );
  });

  return (
    <div className="radar-container">
      <svg className="radar-svg" viewBox="0 0 500 500">
        <defs>
          <linearGradient id="radarGradient" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stopColor="#667eea" stopOpacity="0.3" />
            <stop offset="100%" stopColor="#764ba2" stopOpacity="0.3" />
          </linearGradient>
        </defs>
        {gridPath}
        {axisLines}
        <polygon points={dataPath} fill="url(#radarGradient)" stroke="#667eea" strokeWidth="2" />
        {dataPoints.map((p, i) => (
          <circle key={i} cx={p.x} cy={p.y} r="4" fill="#667eea" />
        ))}
        {labels}
      </svg>
    </div>
  );
}

function LineChart({ data }) {
  const width = 800;
  const height = 300;
  const padding = { left: 60, right: 30, top: 30, bottom: 40 };
  const chartWidth = width - padding.left - padding.right;
  const chartHeight = height - padding.top - padding.bottom;

  if (!data || data.length === 0) {
    return (
      <div style={{ height: '300px', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <div className="empty-state">
          <p>暂无测试记录</p>
        </div>
      </div>
    );
  }

  const maxScore = Math.max(...data.map((d) => d.accuracy || 0), 100);
  const xStep = data.length > 1 ? chartWidth / (data.length - 1) : chartWidth;

  const points = data.map((d, i) => {
    const x = padding.left + i * xStep;
    const y = padding.top + chartHeight - ((d.accuracy || 0) / maxScore) * chartHeight;
    return { x, y, data: d };
  });

  const pathD = points
    .map((p, i) => (i === 0 ? `M ${p.x} ${p.y}` : `L ${p.x} ${p.y}`))
    .join(' ');

  const areaPath = `${pathD} L ${points[points.length - 1].x} ${padding.top + chartHeight} L ${points[0].x} ${padding.top + chartHeight} Z`;

  return (
    <svg className="line-chart" viewBox={`0 0 ${width} ${height}`}>
      <defs>
        <linearGradient id="lineGradient" x1="0%" y1="0%" x2="0%" y2="100%">
          <stop offset="0%" stopColor="#667eea" stopOpacity="0.3" />
          <stop offset="100%" stopColor="#667eea" stopOpacity="0" />
        </linearGradient>
      </defs>

      {[0, 0.25, 0.5, 0.75, 1].map((ratio) => {
        const y = padding.top + chartHeight * (1 - ratio);
        return (
          <g key={ratio}>
            <line x1={padding.left} y1={y} x2={width - padding.right} y2={y} stroke="#e2e8f0" strokeWidth="1" />
            <text x={padding.left - 10} y={y + 4} textAnchor="end" fontSize="10" fill="#718096">
              {Math.round(ratio * maxScore)}%
            </text>
          </g>
        );
      })}

      <path d={areaPath} fill="url(#lineGradient)" />
      <path d={pathD} fill="none" stroke="#667eea" strokeWidth="2" />

      {points.map((p, i) => (
        <g key={i}>
          <circle cx={p.x} cy={p.y} r="5" fill="#667eea" />
          <text x={p.x} y={padding.top + chartHeight + 20} textAnchor="middle" fontSize="10" fill="#718096">
            {i + 1}
          </text>
          <title>{`${p.data.date}: ${p.data.accuracy?.toFixed(1)}%`}</title>
        </g>
      ))}

      <text x={width / 2} y={height - 5} textAnchor="middle" fontSize="11" fill="#718096">
        测试次数
      </text>
    </svg>
  );
}

function ReportPage({ studentId }) {
  const [report, setReport] = useState(null);
  const [loading, setLoading] = useState(true);
  const [history, setHistory] = useState([]);

  const loadReport = async () => {
    try {
      setLoading(true);
      const data = await api.getStudentReport(studentId);
      setReport(data);
      const histData = await api.getExamHistory(studentId);
      setHistory(histData.data || []);
    } catch (error) {
      console.error('Failed to load report:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadReport();
  }, [studentId]);

  if (loading) {
    return (
      <div className="card">
        <div className="empty-state">
          <p>加载中...</p>
        </div>
      </div>
    );
  }

  return (
    <div>
      <div className="card">
        <h2>学习概览</h2>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '1rem', marginTop: '1rem' }}>
          <div style={{ textAlign: 'center', padding: '1.5rem', background: '#f8f9fa', borderRadius: '8px' }}>
            <div style={{ fontSize: '2rem', fontWeight: 'bold', color: '#667eea' }}>
              {report?.total_exams || 0}
            </div>
            <div style={{ color: '#666', fontSize: '0.9rem' }}>完成测试数</div>
          </div>
          <div style={{ textAlign: 'center', padding: '1.5rem', background: '#f8f9fa', borderRadius: '8px' }}>
            <div style={{ fontSize: '2rem', fontWeight: 'bold', color: '#28a745' }}>
              {report?.knowledge_mastery?.filter((m) => m.mastery_level === 'mastered').length || 0}
            </div>
            <div style={{ color: '#666', fontSize: '0.9rem' }}>已掌握知识点</div>
          </div>
          <div style={{ textAlign: 'center', padding: '1.5rem', background: '#f8f9fa', borderRadius: '8px' }}>
            <div style={{ fontSize: '2rem', fontWeight: 'bold', color: '#dc3545' }}>
              {report?.knowledge_mastery?.filter((m) => m.mastery_level === 'weak').length || 0}
            </div>
            <div style={{ color: '#666', fontSize: '0.9rem' }}>薄弱知识点</div>
          </div>
        </div>
      </div>

      <div className="card">
        <h2>成绩趋势</h2>
        <LineChart data={report?.score_trend || []} />
      </div>

      <div className="card">
        <h2>知识点掌握雷达图</h2>
        <RadarChart data={report?.knowledge_mastery || []} />
      </div>

      <div className="card">
        <h2>知识点掌握详情</h2>
        {report?.knowledge_mastery?.length === 0 ? (
          <div className="empty-state">
            <p>暂无知识点掌握数据</p>
          </div>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>学科</th>
                <th>知识点</th>
                <th>答题数</th>
                <th>正确数</th>
                <th>正确率</th>
                <th>掌握程度</th>
              </tr>
            </thead>
            <tbody>
              {report.knowledge_mastery.map((m) => (
                <tr key={m.knowledge_point_id}>
                  <td>{m.subject}</td>
                  <td>{m.knowledge_point_name}</td>
                  <td>{m.total_answered}</td>
                  <td>{m.correct_answered}</td>
                  <td>{m.accuracy?.toFixed(1) || 0}%</td>
                  <td>
                    <span className={`badge ${MASTERY_LABELS[m.mastery_level]?.class}`}>
                      {MASTERY_LABELS[m.mastery_level]?.text}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      <div className="card">
        <h2>测试历史</h2>
        {history.length === 0 ? (
          <div className="empty-state">
            <p>暂无测试记录</p>
          </div>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>测试ID</th>
                <th>状态</th>
                <th>总分</th>
                <th>正确数</th>
                <th>当前难度</th>
                <th>开始时间</th>
                <th>结束时间</th>
              </tr>
            </thead>
            <tbody>
              {history.map((exam) => (
                <tr key={exam.id}>
                  <td>#{exam.id}</td>
                  <td>
                    <span className={`badge ${exam.status === 'completed' ? 'badge-published' : exam.status === 'incomplete' ? 'badge-pending' : 'badge-draft'}`}>
                      {{ completed: '已完成', incomplete: '未完成', in_progress: '进行中' }[exam.status]}
                    </span>
                  </td>
                  <td>{exam.total_score}</td>
                  <td>{exam.correct_count}</td>
                  <td>L{exam.current_difficulty}</td>
                  <td>{exam.start_time ? new Date(exam.start_time).toLocaleString() : '-'}</td>
                  <td>{exam.end_time ? new Date(exam.end_time).toLocaleString() : '-'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}

export default ReportPage;
