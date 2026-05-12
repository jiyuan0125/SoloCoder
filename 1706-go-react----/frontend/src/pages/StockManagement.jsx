import React, { useState, useEffect } from 'react'
import dayjs from 'dayjs'
import { drugApi, stockApi } from '../api'

function getExpiryStatus(expiryDate) {
  const expiry = dayjs(expiryDate)
  const now = dayjs()
  const daysLeft = expiry.diff(now, 'day')

  if (daysLeft < 0) return 'expired'
  if (daysLeft <= 90) return 'critical'
  if (daysLeft <= 180) return 'near'
  return 'normal'
}

function getStatusTag(status) {
  switch (status) {
    case 'near': return <span className="tag tag-near">近效期</span>
    case 'critical': return <span className="tag tag-critical">临期</span>
    case 'expired': return <span className="tag tag-expired">已过期</span>
    default: return <span className="tag tag-normal">正常</span>
  }
}

const defaultStockIn = {
  drug_id: '',
  quantity: '',
  production_date: '',
  expiry_date: '',
  supplier: '',
  supplier_code: '',
  operator1: '',
  operator2: '',
}

const defaultStockOut = {
  drug_id: '',
  quantity: '',
  operator1: '',
  operator2: '',
  remark: '',
}

export default function StockManagement() {
  const [drugs, setDrugs] = useState([])
  const [search, setSearch] = useState('')
  const [alerts, setAlerts] = useState([])
  const [stockValue, setStockValue] = useState({ cost_value: 0, retail_value: 0 })
  const [nearExpiryItems, setNearExpiryItems] = useState([])
  const [modalType, setModalType] = useState(null)
  const [selectedDrug, setSelectedDrug] = useState(null)
  const [stockItems, setStockItems] = useState([])
  const [stockInData, setStockInData] = useState(defaultStockIn)
  const [stockOutData, setStockOutData] = useState(defaultStockOut)
  const [error, setError] = useState('')
  const [confirmData, setConfirmData] = useState(null)

  const fetchAll = async () => {
    try {
      const [drugsRes, alertsRes, valueRes, expiryRes] = await Promise.all([
        drugApi.list(search),
        stockApi.alerts(),
        stockApi.value(),
        stockApi.nearExpiry(),
      ])
      setDrugs(drugsRes.data.data || [])
      setAlerts(alertsRes.data.data || [])
      setStockValue(valueRes.data.data || { cost_value: 0, retail_value: 0 })
      setNearExpiryItems(expiryRes.data.data || [])
    } catch (err) {
      console.error('获取数据失败:', err)
    }
  }

  useEffect(() => {
    fetchAll()
    const interval = setInterval(fetchAll, 30000)
    return () => clearInterval(interval)
  }, [])

  useEffect(() => {
    const timer = setTimeout(() => {
      fetchAll()
    }, 300)
    return () => clearTimeout(timer)
  }, [search])

  const openStockIn = (drug = null) => {
    setStockInData({
      ...defaultStockIn,
      drug_id: drug?.id || '',
    })
    setError('')
    setModalType('stockIn')
  }

  const openStockOut = (drug) => {
    setStockOutData({
      ...defaultStockOut,
      drug_id: drug?.id || '',
    })
    setError('')
    setModalType('stockOut')
  }

  const viewStockItems = async (drug) => {
    setSelectedDrug(drug)
    try {
      const res = await stockApi.items(drug.id)
      setStockItems(res.data.data || [])
      setModalType('stockItems')
    } catch (err) {
      console.error('获取库存批次失败:', err)
    }
  }

  const handleStockInChange = (e) => {
    const { name, value } = e.target
    setStockInData(prev => ({ ...prev, [name]: value }))
  }

  const handleStockOutChange = (e) => {
    const { name, value } = e.target
    setStockOutData(prev => ({ ...prev, [name]: value }))
  }

  const confirmStockIn = () => {
    if (!stockInData.drug_id) {
      setError('请选择药品')
      return
    }
    if (!stockInData.quantity || parseInt(stockInData.quantity) <= 0) {
      setError('请输入有效的入库数量')
      return
    }
    if (!stockInData.production_date || !stockInData.expiry_date) {
      setError('请填写生产日期和有效期')
      return
    }
    if (!stockInData.supplier || !stockInData.supplier_code) {
      setError('请填写供应商信息')
      return
    }
    setConfirmData({ type: 'stockIn', data: stockInData })
  }

  const confirmStockOut = () => {
    if (!stockOutData.drug_id) {
      setError('请选择药品')
      return
    }
    if (!stockOutData.quantity || parseInt(stockOutData.quantity) <= 0) {
      setError('请输入有效的出库数量')
      return
    }
    setConfirmData({ type: 'stockOut', data: stockOutData })
  }

  const executeAction = async () => {
    if (!confirmData) return
    try {
      if (confirmData.type === 'stockIn') {
        await stockApi.in({
          ...confirmData.data,
          quantity: parseInt(confirmData.data.quantity),
        })
      } else if (confirmData.type === 'stockOut') {
        await stockApi.out({
          ...confirmData.data,
          quantity: parseInt(confirmData.data.quantity),
        })
      }
      setConfirmData(null)
      setModalType(null)
      fetchAll()
    } catch (err) {
      setError(err.response?.data?.error || '操作失败')
      setConfirmData(null)
    }
  }

  const closeModal = () => {
    setModalType(null)
    setConfirmData(null)
    setSelectedDrug(null)
    setError('')
  }

  const markAlertRead = async (id) => {
    try {
      await stockApi.markAlertRead(id)
      fetchAll()
    } catch (err) {
      console.error('标记已读失败:', err)
    }
  }

  return (
    <div>
      {alerts.length > 0 && (
        <div className="alert-bar">
          <div className="alert-marquee">
            {alerts.map(alert => (
              <span key={alert.id} className="alert-item">
                ⚠️ {alert.message}
              </span>
            ))}
          </div>
        </div>
      )}

      <div className="page-container">
        <h2 className="page-title">库存管理</h2>

        <div className="stock-value">
          <div className="value-card">
            <div className="value-label">成本金额（按进货价）</div>
            <div className="value-amount">¥{stockValue.cost_value.toFixed(2)}</div>
          </div>
          <div className="value-card retail">
            <div className="value-label">库存金额（按零售价）</div>
            <div className="value-amount">¥{stockValue.retail_value.toFixed(2)}</div>
          </div>
        </div>

        <div className="toolbar">
          <div className="search-box">
            <input
              type="text"
              className="search-input"
              placeholder="搜索药品名称或编码..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>
          <button className="btn btn-primary" onClick={() => openStockIn()}>
            + 入库
          </button>
        </div>

        <table className="data-table">
          <thead>
            <tr>
              <th>药品编码</th>
              <th>药品名称</th>
              <th>规格</th>
              <th>单位</th>
              <th>当前库存</th>
              <th>预警阈值</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {drugs.length === 0 ? (
              <tr><td colSpan="8" className="empty-state">暂无数据</td></tr>
            ) : (
              drugs.map(drug => {
                const stockStatus = drug.current_stock < drug.min_stock ? 'critical' :
                                    drug.current_stock > drug.max_stock ? 'near' : 'normal'
                return (
                  <tr key={drug.id}>
                    <td>{drug.drug_code}</td>
                    <td>{drug.generic_name}</td>
                    <td>{drug.specification}</td>
                    <td>{drug.unit}</td>
                    <td style={{
                      color: stockStatus === 'critical' ? '#c62828' :
                             stockStatus === 'near' ? '#e65100' : '#333',
                      fontWeight: 600
                    }}>
                      {drug.current_stock}
                    </td>
                    <td>{drug.min_stock} - {drug.max_stock}</td>
                    <td>
                      {drug.current_stock < drug.min_stock && <span className="tag tag-expired">库存不足</span>}
                      {drug.current_stock > drug.max_stock && <span className="tag tag-near">库存积压</span>}
                      {drug.current_stock >= drug.min_stock && drug.current_stock <= drug.max_stock && <span className="tag tag-normal">正常</span>}
                    </td>
                    <td>
                      <div className="action-buttons">
                        <button className="btn btn-default btn-small" onClick={() => viewStockItems(drug)}>
                          批次
                        </button>
                        <button className="btn btn-success btn-small" onClick={() => openStockIn(drug)}>
                          入库
                        </button>
                        <button className="btn btn-warning btn-small" onClick={() => openStockOut(drug)}>
                          出库
                        </button>
                      </div>
                    </td>
                  </tr>
                )
              })
            )}
          </tbody>
        </table>

        {nearExpiryItems.length > 0 && (
          <>
            <h3 className="section-title">近效期药品</h3>
            <table className="data-table">
              <thead>
                <tr>
                  <th>药品名称</th>
                  <th>批号</th>
                  <th>库存数量</th>
                  <th>生产日期</th>
                  <th>有效期</th>
                  <th>状态</th>
                </tr>
              </thead>
              <tbody>
                {nearExpiryItems.map(item => (
                  <tr key={item.id}>
                    <td>{item.drug?.generic_name}</td>
                    <td>{item.batch_number}</td>
                    <td>{item.quantity}</td>
                    <td>{dayjs(item.production_date).format('YYYY-MM-DD')}</td>
                    <td>{dayjs(item.expiry_date).format('YYYY-MM-DD')}</td>
                    <td>{getStatusTag(getExpiryStatus(item.expiry_date))}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </>
        )}
      </div>

      {modalType === 'stockIn' && (
        <div className="modal-overlay" onClick={closeModal}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>药品入库</h3>
              <button className="modal-close" onClick={closeModal}>&times;</button>
            </div>
            {confirmData ? (
              <div className="modal-body confirm-dialog">
                <div className="confirm-icon">⚠️</div>
                <div className="confirm-message">确认入库操作</div>
                <div className="confirm-detail">
                  药品：{drugs.find(d => d.id === parseInt(confirmData.data.drug_id))?.generic_name}<br/>
                  数量：{confirmData.data.quantity}
                </div>
              </div>
            ) : (
              <div className="modal-body">
                {error && <div className="error-message">{error}</div>}
                <div className="form-group">
                  <label className="form-label">药品 *</label>
                  <select
                    className="form-select"
                    name="drug_id"
                    value={stockInData.drug_id}
                    onChange={handleStockInChange}
                    required
                  >
                    <option value="">请选择药品</option>
                    {drugs.map(d => (
                      <option key={d.id} value={d.id}>{d.generic_name} - {d.drug_code}</option>
                    ))}
                  </select>
                </div>
                <div className="form-row">
                  <div className="form-group">
                    <label className="form-label">入库数量 *</label>
                    <input
                      type="number"
                      className="form-input"
                      name="quantity"
                      value={stockInData.quantity}
                      onChange={handleStockInChange}
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="form-label">供应商 *</label>
                    <input
                      type="text"
                      className="form-input"
                      name="supplier"
                      value={stockInData.supplier}
                      onChange={handleStockInChange}
                      required
                    />
                  </div>
                </div>
                <div className="form-row">
                  <div className="form-group">
                    <label className="form-label">供应商编号 *</label>
                    <input
                      type="text"
                      className="form-input"
                      name="supplier_code"
                      value={stockInData.supplier_code}
                      onChange={handleStockInChange}
                      placeholder="批号前缀"
                      required
                    />
                  </div>
                  <div className="form-group"></div>
                </div>
                <div className="form-row">
                  <div className="form-group">
                    <label className="form-label">生产日期 *</label>
                    <input
                      type="date"
                      className="form-input"
                      name="production_date"
                      value={stockInData.production_date}
                      onChange={handleStockInChange}
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="form-label">有效期 *</label>
                    <input
                      type="date"
                      className="form-input"
                      name="expiry_date"
                      value={stockInData.expiry_date}
                      onChange={handleStockInChange}
                      required
                    />
                  </div>
                </div>
                <div className="form-row">
                  <div className="form-group">
                    <label className="form-label">操作人1</label>
                    <input
                      type="text"
                      className="form-input"
                      name="operator1"
                      value={stockInData.operator1}
                      onChange={handleStockInChange}
                    />
                  </div>
                  <div className="form-group">
                    <label className="form-label">操作人2</label>
                    <input
                      type="text"
                      className="form-input"
                      name="operator2"
                      value={stockInData.operator2}
                      onChange={handleStockInChange}
                    />
                  </div>
                </div>
              </div>
            )}
            <div className="modal-footer">
              {confirmData ? (
                <>
                  <button className="btn btn-default" onClick={() => setConfirmData(null)}>取消</button>
                  <button className="btn btn-success" onClick={executeAction}>确认入库</button>
                </>
              ) : (
                <>
                  <button className="btn btn-default" onClick={closeModal}>关闭</button>
                  <button className="btn btn-primary" onClick={confirmStockIn}>提交</button>
                </>
              )}
            </div>
          </div>
        </div>
      )}

      {modalType === 'stockOut' && (
        <div className="modal-overlay" onClick={closeModal}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>药品出库</h3>
              <button className="modal-close" onClick={closeModal}>&times;</button>
            </div>
            {confirmData ? (
              <div className="modal-body confirm-dialog">
                <div className="confirm-icon">⚠️</div>
                <div className="confirm-message">确认出库操作</div>
                <div className="confirm-detail">
                  药品：{drugs.find(d => d.id === parseInt(confirmData.data.drug_id))?.generic_name}<br/>
                  数量：{confirmData.data.quantity}
                </div>
              </div>
            ) : (
              <div className="modal-body">
                {error && <div className="error-message">{error}</div>}
                <div className="form-group">
                  <label className="form-label">药品 *</label>
                  <select
                    className="form-select"
                    name="drug_id"
                    value={stockOutData.drug_id}
                    onChange={handleStockOutChange}
                    required
                  >
                    <option value="">请选择药品</option>
                    {drugs.map(d => (
                      <option key={d.id} value={d.id}>{d.generic_name} - 当前库存：{d.current_stock}</option>
                    ))}
                  </select>
                </div>
                <div className="form-row">
                  <div className="form-group">
                    <label className="form-label">出库数量 *</label>
                    <input
                      type="number"
                      className="form-input"
                      name="quantity"
                      value={stockOutData.quantity}
                      onChange={handleStockOutChange}
                      required
                    />
                  </div>
                  <div className="form-group">
                    <label className="form-label">备注</label>
                    <input
                      type="text"
                      className="form-input"
                      name="remark"
                      value={stockOutData.remark}
                      onChange={handleStockOutChange}
                    />
                  </div>
                </div>
                <div className="form-row">
                  <div className="form-group">
                    <label className="form-label">操作人1</label>
                    <input
                      type="text"
                      className="form-input"
                      name="operator1"
                      value={stockOutData.operator1}
                      onChange={handleStockOutChange}
                    />
                  </div>
                  <div className="form-group">
                    <label className="form-label">操作人2（特殊药品必填）</label>
                    <input
                      type="text"
                      className="form-input"
                      name="operator2"
                      value={stockOutData.operator2}
                      onChange={handleStockOutChange}
                    />
                  </div>
                </div>
              </div>
            )}
            <div className="modal-footer">
              {confirmData ? (
                <>
                  <button className="btn btn-default" onClick={() => setConfirmData(null)}>取消</button>
                  <button className="btn btn-warning" onClick={executeAction}>确认出库</button>
                </>
              ) : (
                <>
                  <button className="btn btn-default" onClick={closeModal}>关闭</button>
                  <button className="btn btn-primary" onClick={confirmStockOut}>提交</button>
                </>
              )}
            </div>
          </div>
        </div>
      )}

      {modalType === 'stockItems' && selectedDrug && (
        <div className="modal-overlay" onClick={closeModal}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>库存批次 - {selectedDrug.generic_name}</h3>
              <button className="modal-close" onClick={closeModal}>&times;</button>
            </div>
            <div className="modal-body">
              {stockItems.length === 0 ? (
                <div className="empty-state">暂无库存批次</div>
              ) : (
                <table className="data-table">
                  <thead>
                    <tr>
                      <th>批号</th>
                      <th>数量</th>
                      <th>生产日期</th>
                      <th>有效期</th>
                      <th>供应商</th>
                      <th>状态</th>
                    </tr>
                  </thead>
                  <tbody>
                    {stockItems.map(item => (
                      <tr key={item.id}>
                        <td>{item.batch_number}</td>
                        <td>{item.quantity}</td>
                        <td>{dayjs(item.production_date).format('YYYY-MM-DD')}</td>
                        <td>{dayjs(item.expiry_date).format('YYYY-MM-DD')}</td>
                        <td>{item.supplier}</td>
                        <td>{getStatusTag(getExpiryStatus(item.expiry_date))}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
            <div className="modal-footer">
              <button className="btn btn-default" onClick={closeModal}>关闭</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
