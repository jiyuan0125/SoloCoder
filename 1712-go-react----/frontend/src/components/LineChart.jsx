import React from 'react';

const LineChart = ({ data, xKey, yKeys, yLabels, referenceLines = [], title }) => {
  if (!data || data.length === 0) {
    return <div className="chart-empty">暂无数据</div>;
  }

  const width = 800;
  const height = 400;
  const padding = { top: 40, right: 40, bottom: 60, left: 60 };
  const chartWidth = width - padding.left - padding.right;
  const chartHeight = height - padding.top - padding.bottom;

  const maxY = Math.max(
    ...data.flatMap(d => yKeys.map(k => d[k] || 0)),
    ...referenceLines.map(r => r.value)
  );
  const minY = Math.min(
    ...data.flatMap(d => yKeys.map(k => d[k] || 0)),
    ...referenceLines.map(r => r.value)
  );
  const yRange = maxY - minY || 1;
  const yScale = (value) => chartHeight - ((value - minY) / yRange) * chartHeight + padding.top;
  const xScale = (index) => padding.left + (index / Math.max(data.length - 1, 1)) * chartWidth;

  const colors = ['#3498db', '#e74c3c', '#2ecc71'];

  return (
    <div className="line-chart">
      {title && <h3 className="chart-title">{title}</h3>}
      <svg width={width} height={height}>
        <defs>
          {yKeys.map((_, i) => (
            <linearGradient key={i} id={`gradient-${i}`} x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor={colors[i % colors.length]} stopOpacity="0.3" />
              <stop offset="100%" stopColor={colors[i % colors.length]} stopOpacity="0" />
            </linearGradient>
          ))}
        </defs>

        <line x1={padding.left} y1={padding.top} x2={padding.left} y2={height - padding.bottom} stroke="#ddd" />
        <line x1={padding.left} y1={height - padding.bottom} x2={width - padding.right} y2={height - padding.bottom} stroke="#ddd" />

        {[0, 0.25, 0.5, 0.75, 1].map((ratio, i) => {
          const y = padding.top + chartHeight * ratio;
          const value = maxY - (maxY - minY) * ratio;
          return (
            <g key={i}>
              <line x1={padding.left} y1={y} x2={width - padding.right} y2={y} stroke="#eee" strokeDasharray="5,5" />
              <text x={padding.left - 10} y={y + 4} textAnchor="end" fill="#666" fontSize="12">
                {value.toFixed(1)}
              </text>
            </g>
          );
        })}

        {referenceLines.map((ref, i) => (
          <g key={`ref-${i}`}>
            <line
              x1={padding.left}
              y1={yScale(ref.value)}
              x2={width - padding.right}
              y2={yScale(ref.value)}
              stroke={ref.color || '#e74c3c'}
              strokeWidth="2"
              strokeDasharray="10,5"
            />
            <text
              x={width - padding.right + 5}
              y={yScale(ref.value) + 4}
              fill={ref.color || '#e74c3c'}
              fontSize="12"
            >
              {ref.label}
            </text>
          </g>
        ))}

        {yKeys.map((yKey, keyIndex) => {
          const pathData = data
            .map((d, i) => `${i === 0 ? 'M' : 'L'} ${xScale(i)} ${yScale(d[yKey] || 0)}`)
            .join(' ');

          const areaData = pathData + 
            ` L ${xScale(data.length - 1)} ${yScale(minY)}` +
            ` L ${padding.left} ${yScale(minY)} Z`;

          return (
            <g key={yKey}>
              <path
                d={areaData}
                fill={`url(#gradient-${keyIndex})`}
                opacity="0.5"
              />
              <path
                d={pathData}
                fill="none"
                stroke={colors[keyIndex % colors.length]}
                strokeWidth="2"
              />
              {data.map((d, i) => (
                <circle
                  key={i}
                  cx={xScale(i)}
                  cy={yScale(d[yKey] || 0)}
                  r="4"
                  fill={colors[keyIndex % colors.length]}
                />
              ))}
            </g>
          );
        })}

        {data.map((d, i) => (
          <text
            key={i}
            x={xScale(i)}
            y={height - padding.bottom + 20}
            textAnchor="middle"
            fill="#666"
            fontSize="10"
            transform={`rotate(-45, ${xScale(i)}, ${height - padding.bottom + 20})`}
          >
            {String(d[xKey]).substring(0, 10)}
          </text>
        ))}

        <g transform={`translate(${padding.left + chartWidth / 2 - 60}, ${height - padding.bottom + 35})`}>
          {yKeys.map((yKey, i) => (
            <g key={yKey} transform={`translate(${i * 80}, 0)`}>
              <rect x="0" y="0" width="12" height="12" fill={colors[i % colors.length]} />
              <text x="18" y="10" fill="#333" fontSize="12">{yLabels?.[i] || yKey}</text>
            </g>
          ))}
        </g>
      </svg>
    </div>
  );
};

export default LineChart;
