import React, { useState } from 'react';

function BarChart({ data, width = 800, height = 300, threshold = 2.0 }) {
  const [tooltip, setTooltip] = useState({ show: false, x: 0, y: 0, content: '' });

  if (!data || data.length === 0) {
    return (
      <div className="chart-container">
        <div className="empty-state">暂无数据</div>
      </div>
    );
  }

  const padding = { top: 30, right: 30, bottom: 80, left: 60 };
  const chartWidth = width - padding.left - padding.right;
  const chartHeight = height - padding.top - padding.bottom;

  const maxValue = Math.max(...data.map(d => d.InfectionRate), threshold * 1.5);
  const minValue = 0;

  const barWidth = Math.min(60, (chartWidth / data.length) * 0.6);
  const gap = (chartWidth - barWidth * data.length) / (data.length + 1);

  const getY = (value) => padding.top + chartHeight - (value - minValue) / (maxValue - minValue) * chartHeight;
  const getX = (i) => padding.left + gap + i * (barWidth + gap);

  const gridLines = [0, 0.25, 0.5, 0.75, 1].map(ratio => {
    const y = padding.top + chartHeight * (1 - ratio);
    const value = minValue + (maxValue - minValue) * ratio;
    return { y, value: value.toFixed(1) };
  });

  const thresholdY = getY(threshold);

  const handleMouseEnter = (item, event) => {
    const rect = event.currentTarget.closest('svg').getBoundingClientRect();
    setTooltip({
      show: true,
      x: event.clientX - rect.left + 10,
      y: event.clientY - rect.top - 40,
      content: `${item.DepartmentName}\n${item.InfectionRate.toFixed(2)}%\n${item.InfectionCount}例/${item.DischargeCount}人`,
    });
  };

  const handleMouseLeave = () => {
    setTooltip({ show: false, x: 0, y: 0, content: '' });
  };

  return (
    <div className="chart-container" style={{ position: 'relative' }}>
      <svg viewBox={`0 0 ${width} ${height}`} className="svg-chart" preserveAspectRatio="xMidYMid meet">
        {gridLines.map((line, i) => (
          <g key={i}>
            <line
              x1={padding.left}
              y1={line.y}
              x2={width - padding.right}
              y2={line.y}
              className="line-chart-grid"
            />
            <text
              x={padding.left - 10}
              y={line.y + 4}
              textAnchor="end"
              fontSize="11"
              fill="#718096"
            >
              {line.value}%
            </text>
          </g>
        ))}

        {threshold <= maxValue && (
          <g>
            <line
              x1={padding.left}
              y1={thresholdY}
              x2={width - padding.right}
              y2={thresholdY}
              stroke="#e53e3e"
              strokeWidth="1.5"
              strokeDasharray="5,3"
            />
            <text
              x={width - padding.right - 5}
              y={thresholdY - 5}
              textAnchor="end"
              fontSize="10"
              fill="#e53e3e"
            >
              阈值 {threshold}%
            </text>
          </g>
        )}

        {data.map((item, i) => {
          const x = getX(i);
          const barHeight = chartHeight - (getY(item.InfectionRate) - padding.top);
          const exceeded = item.InfectionRate > threshold;

          return (
            <g key={i}>
              <rect
                x={x}
                y={getY(item.InfectionRate)}
                width={barWidth}
                height={barHeight}
                className={`bar-chart-bar ${exceeded ? 'exceeded' : ''}`}
                rx="4"
                onMouseEnter={(e) => handleMouseEnter(item, e)}
                onMouseLeave={handleMouseLeave}
                style={{ cursor: 'pointer' }}
              />
              <text
                x={x + barWidth / 2}
                y={getY(item.InfectionRate) - 8}
                textAnchor="middle"
                fontSize="10"
                fill={exceeded ? '#e53e3e' : '#3182ce'}
                fontWeight="600"
              >
                {item.InfectionRate.toFixed(2)}%
              </text>
              <text
                x={x + barWidth / 2}
                y={height - padding.bottom + 15}
                textAnchor="middle"
                fontSize="11"
                fill="#4a5568"
              >
                {item.DepartmentName}
              </text>
            </g>
          );
        })}

        <text x={width / 2} y={18} textAnchor="middle" fontSize="14" fontWeight="600" fill="#2d3748">
          各科室感染率排名
        </text>
      </svg>

      {tooltip.show && (
        <div
          className="tooltip"
          style={{ left: tooltip.x, top: tooltip.y, whiteSpace: 'pre-line' }}
        >
          {tooltip.content}
        </div>
      )}
    </div>
  );
}

export default BarChart;
