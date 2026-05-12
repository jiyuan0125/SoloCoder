import React, { useState } from 'react';

function LineChart({ data, width = 800, height = 300 }) {
  const [tooltip, setTooltip] = useState({ show: false, x: 0, y: 0, content: '' });

  if (!data || data.length === 0) {
    return (
      <div className="chart-container">
        <div className="empty-state">暂无数据</div>
      </div>
    );
  }

  const padding = { top: 30, right: 30, bottom: 50, left: 60 };
  const chartWidth = width - padding.left - padding.right;
  const chartHeight = height - padding.top - padding.bottom;

  const values = data.map(d => d.infection_rate);
  const maxValue = Math.max(...values, 5);
  const minValue = 0;

  const xStep = chartWidth / (data.length - 1);

  const getX = (i) => padding.left + i * xStep;
  const getY = (value) => padding.top + chartHeight - (value - minValue) / (maxValue - minValue) * chartHeight;

  const points = data.map((d, i) => ({
    x: getX(i),
    y: getY(d.infection_rate),
    month: d.month,
    rate: d.infection_rate,
  }));

  const pathData = points.map((p, i) => (i === 0 ? `M ${p.x} ${p.y}` : `L ${p.x} ${p.y}`)).join(' ');
  const areaData = `${pathData} L ${points[points.length - 1].x} ${padding.top + chartHeight} L ${points[0].x} ${padding.top + chartHeight} Z`;

  const gridLines = [0, 0.25, 0.5, 0.75, 1].map(ratio => {
    const y = padding.top + chartHeight * (1 - ratio);
    const value = minValue + (maxValue - minValue) * ratio;
    return { y, value: value.toFixed(1) };
  });

  const handleMouseEnter = (point, event) => {
    const rect = event.currentTarget.closest('svg').getBoundingClientRect();
    setTooltip({
      show: true,
      x: event.clientX - rect.left + 10,
      y: event.clientY - rect.top - 30,
      content: `${point.month}: ${point.rate.toFixed(2)}%`,
    });
  };

  const handleMouseLeave = () => {
    setTooltip({ show: false, x: 0, y: 0, content: '' });
  };

  return (
    <div className="chart-container" style={{ position: 'relative' }}>
      <svg viewBox={`0 0 ${width} ${height}`} className="svg-chart" preserveAspectRatio="xMidYMid meet">
        <defs>
          <linearGradient id="gradient" x1="0%" y1="0%" x2="0%" y2="100%">
            <stop offset="0%" stopColor="#3182ce" stopOpacity="0.3" />
            <stop offset="100%" stopColor="#3182ce" stopOpacity="0.05" />
          </linearGradient>
        </defs>

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

        {points.map((p, i) => (
          <text
            key={`x-label-${i}`}
            x={p.x}
            y={height - padding.bottom + 25}
            textAnchor="middle"
            fontSize="10"
            fill="#718096"
            transform={`rotate(-30, ${p.x}, ${height - padding.bottom + 20})`}
          >
            {p.month}
          </text>
        ))}

        <path d={areaData} className="line-chart-area" />
        <path d={pathData} className="line-chart-path" />

        {points.map((p, i) => (
          <circle
            key={i}
            cx={p.x}
            cy={p.y}
            r="5"
            className="line-chart-point"
            onMouseEnter={(e) => handleMouseEnter(p, e)}
            onMouseLeave={handleMouseLeave}
            style={{ cursor: 'pointer' }}
          />
        ))}

        <text x={width / 2} y={18} textAnchor="middle" fontSize="14" fontWeight="600" fill="#2d3748">
          近12个月全院感染率趋势
        </text>
      </svg>

      {tooltip.show && (
        <div
          className="tooltip"
          style={{ left: tooltip.x, top: tooltip.y }}
        >
          {tooltip.content}
        </div>
      )}
    </div>
  );
}

export default LineChart;
