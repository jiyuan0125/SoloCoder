import React, { useState, useEffect } from 'react'
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer,
  LineChart, Line, PieChart, Pie, Cell
} from 'recharts'
import { aggregations } from '../../api'

const COLORS = ['#1e3c72', '#2a5298', '#27ae60', '#f39c12', '#e74c3c', '#9b59b6']

const DashboardPage = () => {
  const [activeTab, setActiveTab] = useState('department')
  const [deptData, setDeptData] = useState([])
  const [categoryData, setCategoryData] = useState([])
  const [timeData, setTimeData] = useState([])
  const [periodType, setPeriodType] = useState('month')
  const [selectedMonth, setSelectedMonth] = useState('')
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (activeTab === 'department') loadDepartmentData()
    if (activeTab === 'category') loadCategoryData()
    if (activeTab === 'time') loadTimeData()
  }, [activeTab, periodType, selectedMonth])

  const loadDepartmentData = async () => {
    setLoading(true)
    try {
      const res = await aggregations.byDepartment(selectedMonth)
      setDeptData(res.data)
    } catch (err) {
      console.error('加载科室聚合失败', err)
    } finally {
      setLoading(false)
    }
  }

  const loadCategoryData = async () => {
    setLoading(true)
    try {
      const res = await aggregations.byCategory()
      setCategoryData(res.data)
    } catch (err) {
      console.error('加载类别聚合失败', err)
    } finally {
      setLoading(false)
    }
  }

  const loadTimeData = async () => {
    setLoading(true)
    try {
      const res = await aggregations.byTime(periodType)
      setTimeData(res.data)
    } catch (err) {
      console.error('加载时间聚合失败', err)
    } finally {
      setLoading(false)
    }
  }

  const allIndicators = []
  deptData.forEach(d => {
    d.indicators.forEach(ind => {
      if (!allIndicators.find(i => i.indicator_code === ind.indicator_code)) {
        allIndicators.push(ind)
      }
    })
  })

  const chartData = deptData.map(d => ({
    name: d.department.length > 8 ? d.department.substring(0, 8) + '...' : d.department,
    fullName: d.department,
    达标率: parseFloat(d.compliance_rate.toFixed(1)),
    指标数: d.total_count,
    达标数: d.met_count
  }))

  const categoryChartData = categoryData.map(c => ({
    name: c.category_name,
    达标率: parseFloat(c.compliance_rate.toFixed(1)),
    数据量: c.data_count,
    达标数: c.met_count
  }))

  return (
    <div>
      <div className="page-header">
        <h1 className="page-title">聚合面板</h1>
        <p className="page-subtitle">多维度的指标聚合展示与分析</p>
      </div>

      <div className="tabs">
        <button className={`tab ${activeTab === 'department' ? 'active' : ''}`} onClick={() => setActiveTab('department')}>
          科室维度
        </button>
        <button className={`tab ${activeTab === 'category' ? 'active' : ''}`} onClick={() => setActiveTab('category')}>
          类别维度
        </button>
        <button className={`tab ${activeTab === 'time' ? 'active' : ''}`} onClick={() => setActiveTab('time')}>
          时间维度
        </button>
      </div>

      {activeTab === 'department' && (
        <>
          <div className="card">
            <div className="card-header">
              <h2 className="card-title">科室质量指标统计</h2>
              <div className="filter-item">
                <span className="filter-label">月份:</span>
                <input className="form-input" type="month" value={selectedMonth}
                  onChange={(e) => setSelectedMonth(e.target.value)}
                  style={{ width: '160px' }} />
              </div>
            </div>

            {deptData.length > 0 ? (
              <>
                <div className="chart-container">
                  <ResponsiveContainer width="100%" height="100%">
                    <BarChart data={chartData}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="name" />
                      <YAxis yAxisId="left" />
                      <YAxis yAxisId="right" orientation="right" />
                      <Tooltip formatter={(value, name) => [value, name]} />
                      <Legend />
                      <Bar yAxisId="left" dataKey="达标率" fill="#1e3c72" name="达标率(%)" />
                      <Bar yAxisId="right" dataKey="指标数" fill="#27ae60" name="指标总数" />
                    </BarChart>
                  </ResponsiveContainer>
                </div>

                <div className="card" style={{ marginTop: '20px' }}>
                  <h3 style={{ marginBottom: '16px', fontSize: '15px', color: '#555' }}>科室-指标矩阵</h3>
                  <div className="department-table">
                    <table>
                      <thead>
                        <tr>
                          <th>科室</th>
                          {allIndicators.map(ind => (
                            <th key={ind.indicator_code}>{ind.indicator_code}</th>
                          ))}
                          <th>达标率</th>
                        </tr>
                      </thead>
                      <tbody>
                        {deptData.map(dept => (
                          <tr key={dept.department}>
                            <td><strong>{dept.department}</strong></td>
                            {allIndicators.map(ind => {
                              const found = dept.indicators.find(i => i.indicator_code === ind.indicator_code)
                              return (
                                <td key={ind.indicator_code} style={{ textAlign: 'center' }}>
                                  {found ? (
                                    <div>
                                      <div className={found.is_target_met ? 'metric-success' : 'metric-danger'}>
                                        {found.value.toFixed(2)}
                                        <span>{found.is_target_met ? '✓' : '✗'}</span>
                                      </div>
                                    </div>
                                  ) : (
                                    <span style={{ color: '#ccc' }}>-</span>
                                  )}
                                </td>
                              )
                            })}
                            <td style={{ textAlign: 'center', fontWeight: '600' }}>
                              <span className={dept.compliance_rate >= 80 ? 'metric-success' : 'metric-danger'}>
                                {dept.compliance_rate.toFixed(1)}%
                              </span>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              </>
            ) : (
              <div className="empty-state">
                {loading ? '加载中...' : '暂无数据，请先录入指标数据'}
              </div>
            )}
          </div>
        </>
      )}

      {activeTab === 'category' && (
        <div className="card">
          <div className="card-header">
            <h2 className="card-title">指标类别达标率统计</h2>
          </div>

          {categoryData.length > 0 ? (
            <>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '20px' }}>
                <div className="chart-container" style={{ height: '300px' }}>
                  <ResponsiveContainer width="100%" height="100%">
                    <PieChart>
                      <Pie
                        data={categoryChartData}
                        dataKey="达标率"
                        nameKey="name"
                        cx="50%"
                        cy="50%"
                        outerRadius={100}
                        label={({ name, value }) => `${name}: ${value}%`}
                      >
                        {categoryChartData.map((_, index) => (
                          <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                        ))}
                      </Pie>
                      <Tooltip />
                      <Legend />
                    </PieChart>
                  </ResponsiveContainer>
                </div>

                <div className="chart-container" style={{ height: '300px' }}>
                  <ResponsiveContainer width="100%" height="100%">
                    <BarChart data={categoryChartData} layout="vertical">
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis type="number" />
                      <YAxis dataKey="name" type="category" width={100} />
                      <Tooltip />
                      <Legend />
                      <Bar dataKey="达标率" fill="#1e3c72" name="达标率(%)" />
                    </BarChart>
                  </ResponsiveContainer>
                </div>
              </div>

              <table style={{ marginTop: '20px' }}>
                <thead>
                  <tr>
                    <th>指标类别</th>
                    <th>指标总数</th>
                    <th>数据记录数</th>
                    <th>达标数</th>
                    <th>达标率</th>
                    <th>平均值</th>
                  </tr>
                </thead>
                <tbody>
                  {categoryData.map(cat => (
                    <tr key={cat.category}>
                      <td><strong>{cat.category_name}</strong></td>
                      <td>{cat.total_indicators}</td>
                      <td>{cat.data_count}</td>
                      <td>{cat.met_count}</td>
                      <td>
                        <span className={cat.compliance_rate >= 80 ? 'metric-success' : 'metric-danger'}>
                          {cat.compliance_rate.toFixed(1)}%
                        </span>
                      </td>
                      <td>{cat.average_value.toFixed(2)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </>
          ) : (
            <div className="empty-state">
              {loading ? '加载中...' : '暂无数据'}
            </div>
          )}
        </div>
      )}

      {activeTab === 'time' && (
        <div className="card">
          <div className="card-header">
            <h2 className="card-title">时间维度达标率趋势</h2>
            <div style={{ display: 'flex', gap: '8px' }}>
              {['month', 'quarter', 'year'].map(p => (
                <button
                  key={p}
                  className={`btn ${periodType === p ? 'btn-primary' : 'btn-sm btn-primary'}`}
                  style={{ opacity: periodType === p ? 1 : 0.6 }}
                  onClick={() => setPeriodType(p)}
                >
                  {p === 'month' ? '月度' : p === 'quarter' ? '季度' : '年度'}
                </button>
              ))}
            </div>
          </div>

          {timeData.length > 0 ? (
            <div className="chart-container">
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={timeData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="period" />
                  <YAxis />
                  <Tooltip />
                  <Legend />
                  <Line type="monotone" dataKey="达标率" stroke="#1e3c72" strokeWidth={2} name="达标率(%)"
                    dot={{ fill: '#1e3c72', r: 5 }} />
                  <Line type="monotone" dataKey="TotalCount" stroke="#27ae60" name="指标总数" />
                </LineChart>
              </ResponsiveContainer>
            </div>
          ) : (
            <div className="empty-state">
              {loading ? '加载中...' : '暂无数据'}
            </div>
          )}

          {timeData.length > 0 && (
            <table style={{ marginTop: '20px' }}>
              <thead>
                <tr>
                  <th>{periodType === 'month' ? '月份' : periodType === 'quarter' ? '季度' : '年度'}</th>
                  <th>指标记录数</th>
                  <th>达标数</th>
                  <th>达标率</th>
                </tr>
              </thead>
              <tbody>
                {timeData.map(t => (
                  <tr key={t.period}>
                    <td><strong>{t.period}</strong></td>
                    <td>{t.total_count}</td>
                    <td>{t.met_count}</td>
                    <td>
                      <span className={t.compliance_rate >= 80 ? 'metric-success' : 'metric-danger'}>
                        {t.compliance_rate.toFixed(1)}%
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}
    </div>
  )
}

export default DashboardPage
