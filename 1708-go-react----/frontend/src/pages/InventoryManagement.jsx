import { useState, useEffect } from 'react'
import axios from 'axios'

export default function InventoryManagement() {
  const [inventory, setInventory] = useState([])
  const [loading, setLoading] = useState(true)
  const [message, setMessage] = useState(null)
  const [error, setError] = useState(null)

  useEffect(() => {
    fetchInventory()
  }, [])

  const fetchInventory = async () => {
    try {
      setLoading(true)
      const res = await axios.get('/api/inventory')
      setInventory(res.data || [])
    } catch (err) {
      console.error('Failed to fetch inventory:', err)
    } finally {
      setLoading(false)
    }
  }

  const freezeInventory = async (barcode) => {
    try {
      await axios.post('/api/inventory/freeze', { barcode })
      setMessage('库存已冻结')
      fetchInventory()
    } catch (err) {
      if (err.response?.data?.error) {
        setError(err.response.data.error)
      } else {
        setError('冻结失败')
      }
    }
  }

  const unfreezeInventory = async (barcode) => {
    try {
      await axios.post('/api/inventory/unfreeze', { barcode })
      setMessage('库存已解冻')
      fetchInventory()
    } catch (err) {
      if (err.response?.data?.error) {
        setError(err.response.data.error)
      } else {
        setError('解冻失败')
      }
    }
  }

  const getStatusBadge = (status) => {
    switch (status) {
      case 'In_Stock': return <span className="status-badge status-instock">在库</span>
      case 'Frozen': return <span className="status-badge status-pending">已冻结</span>
      case 'Issued': return <span className="status-badge status-issued">已出库</span>
      case 'Expired': return <span className="status-badge status-scrapped">已过期</span>
      default: return <span className="status-badge status-pending">{status}</span>
    }
  }

  const getProductTypeName = (type) => {
    const names = {
      'Whole_Blood': '全血',
      'Red_Blood_Cells': '红细胞悬液',
      'Fresh_Frozen_Plasma': '新鲜冰冻血浆',
      'Platelets': '机采血小板'
    }
    return names[type] || type
  }

  if (loading) {
    return <div className="page"><p>加载中...</p></div>
  }

  return (
    <div className="page">
      <h1 className="page-title">📦 库存管理</h1>

      {error && <div className="alert alert-danger">{error}</div>}
      {message && <div className="alert alert-success">{message}</div>}

      <div className="section">
        <h2 style={{ marginBottom: '1rem' }}>库存汇总</h2>
        
        {inventory.length === 0 ? (
          <p>暂无库存</p>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>血型</th>
                <th>血制品类型</th>
                <th>当前库存</th>
                <th>总容量 (ml)</th>
                <th>最低安全量</th>
                <th>状态</th>
              </tr>
            </thead>
            <tbody>
              {inventory.map(item => (
                <tr key={item.blood_type + item.product_type} className={item.below_min ? 'low-stock' : ''}>
                  <td style={{ fontWeight: 'bold' }}>{item.blood_type}</td>
                  <td>{getProductTypeName(item.product_type)}</td>
                  <td>{item.count} 袋</td>
                  <td>{item.total_volume}</td>
                  <td>{item.min_stock} 袋</td>
                  <td>
                    {item.below_min ? (
                      <span className="status-badge status-scrapped">⚠️ 低于安全库存</span>
                    ) : (
                      <span className="status-badge status-qualified">正常</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      <div className="section">
        <h2 style={{ marginBottom: '1rem' }}>库存明细</h2>
        
        {inventory.length === 0 ? (
          <p>暂无库存</p>
        ) : (
          inventory.map(group => (
            <div key={group.blood_type + group.product_type} style={{ marginBottom: '2rem' }}>
              <h3 style={{ marginBottom: '1rem', padding: '0.5rem', backgroundColor: '#f8f9fa', borderRadius: '5px' }}>
                {group.blood_type} - {getProductTypeName(group.product_type)}
                {group.below_min && (
                  <span style={{ marginLeft: '1rem', color: '#dc3545' }}>⚠️ 低库存</span>
                )}
              </h3>
              
              {group.items && group.items.length > 0 && (
                <table className="table">
                  <thead>
                    <tr>
                      <th>条码号</th>
                      <th>入库时间</th>
                      <th>有效期</th>
                      <th>容量 (ml)</th>
                      <th>状态</th>
                      <th>操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    {group.items.map(item => (
                      <tr key={item.id}>
                        <td>{item.barcode}</td>
                        <td>{new Date(item.storage_date).toLocaleDateString()}</td>
                        <td>
                          {new Date(item.expiry_date).toLocaleDateString()}
                          {new Date(item.expiry_date) < new Date() && (
                            <span style={{ color: '#dc3545', marginLeft: '0.5rem' }}>(已过期)</span>
                          )}
                        </td>
                        <td>{item.volume_ml}</td>
                        <td>{getStatusBadge(item.status)}</td>
                        <td>
                          {item.status === 'In_Stock' && (
                            <button
                              className="btn btn-danger"
                              onClick={() => freezeInventory(item.barcode)}
                              style={{ padding: '0.25rem 0.75rem', fontSize: '0.875rem' }}
                            >
                              冻结
                            </button>
                          )}
                          {item.status === 'Frozen' && (
                            <button
                              className="btn btn-success"
                              onClick={() => unfreezeInventory(item.barcode)}
                              style={{ padding: '0.25rem 0.75rem', fontSize: '0.875rem' }}
                            >
                              解冻
                            </button>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          ))
        )}
      </div>
    </div>
  )
}
