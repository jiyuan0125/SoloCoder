import { useState, useEffect } from 'react'
import axios from 'axios'

export default function Dashboard() {
  const [data, setData] = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchDashboard()
  }, [])

  const fetchDashboard = async () => {
    try {
      const res = await axios.get('/api/stats/dashboard')
      setData(res.data)
    } catch (err) {
      console.error('Failed to fetch dashboard:', err)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return <div className="page"><p>加载中...</p></div>
  }

  return (
    <div className="page">
      <h1 className="page-title">📊 系统仪表盘</h1>
      
      <div className="card-grid">
        <div className="stat-card">
        <div className="stat-label">总献血者</div>
        <div className="stat-number">{data?.total_donors || 0}</div>
      </div>
        <div className="stat-card">
        <div className="stat-label">待检测</div>
        <div className="stat-number">{data?.pending_tests || 0}</div>
      </div>
        <div className="stat-card">
        <div className="stat-label">在库血液</div>
        <div className="stat-number">{data?.total_in_stock || 0}</div>
      </div>
        <div className="stat-card">
        <div className="stat-label">待处理申请</div>
        <div className="stat-number">{data?.pending_requests || 0}</div>
      </div>
    </div>

      {data?.low_stock_alerts?.length > 0 && (
        <div className="section">
        <h2 style={{ marginBottom: '1rem', color: '#dc3545' }}>⚠️ 低库存预警</h2>
        <table className="table">
        <thead>
          <tr>
            <th>血型</th>
            <th>血制品类型</th>
            <th>当前库存</th>
            <th>最低安全量</th>
            <th>缺口</th>
          </tr>
          </thead>
          <tbody>
            {data.low_stock_alerts.map((item => (
            <tr key={item.blood_type + item.product_type}>
            <td>{item.blood_type}</td>
              <td>{item.product_type}</td>
            <td>{item.current}</td>
              <td>{item.min_required}</td>
              <td style={{ color: '#dc3545', fontWeight: 'bold' }}>-{item.deficit}</td>
            </tr>
          ))}
        </tbody>
        </table>
      </div>
      )}
    </div>
  )
}
