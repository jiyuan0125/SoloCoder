import React, { useState, useEffect } from 'react'
import { devices } from '../api'

const categoryLabels = {
  diagnostic: '诊断设备',
  therapeutic: '治疗设备',
  auxiliary: '辅助设备',
  monitoring: '监护设备',
}

const statusLabels = {
  in_use: '使用中',
  idle: '闲置',
  maintenance: '维护中',
  calibrating: '校准中',
  disabled: '停用',
  scrapped: '已报废',
}

const departments = ['内科', '外科', '急诊科', '放射科', '手术室', 'ICU', '门诊']

export default function DeviceList() {
  const [deviceList, setDeviceList] = useState([])
  const [loading, setLoading] = useState(true)
  const [filters, setFilters] = useState({
    department: '',
    category: '',
    status: '',
  })
  const [showModal, setShowModal] = useState(false)
  const [selectedDevice, setSelectedDevice] = useState(null)
  const [newDevice, setNewDevice] = useState({
    asset_number: '',
    name: '',
    brand_model: '',
    serial_number: '',
    category: 'diagnostic',
    department: '',
    location: '',
    purchase_date: '',
    purchase_price: '',
    warranty_expiry_date: '',
  })
  const [error, setError] = useState('')

  const loadDevices = async () => {
    setLoading(true)
    try {
      const params = {}
      if (filters.department) params.department = filters.department
      if (filters.category) params.category = filters.category
      if (filters.status) params.status = filters.status
      const data = await devices.getAll(params)
      setDeviceList(data)
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadDevices()
  }, [filters])

  const getCalibrationStatus = (device) => {
    if (!device.next_calibration_date) return 'normal'
    const now = new Date()
    const dueDate = new Date(device.next_calibration_date)
    const daysLeft = Math.floor((dueDate - now) / (1000 * 60 * 60 * 24))
    
    if (daysLeft < 0) return 'expired'
    if (daysLeft <= 30) return 'warning'
    return 'normal'
  }

  const handleCreateDevice = async (e) => {
    e.preventDefault()
    try {
      await devices.create(newDevice)
      setShowModal(false)
      setNewDevice({
        asset_number: '',
        name: '',
        brand_model: '',
        serial_number: '',
        category: 'diagnostic',
        department: '',
        location: '',
        purchase_date: '',
        purchase_price: '',
        warranty_expiry_date: '',
      })
      loadDevices()
    } catch (err) {
      alert(err.message)
    }
  }

  const handleViewDetail = async (id) => {
    try {
      const device = await devices.get(id)
      setSelectedDevice(device)
    } catch (err) {
      alert(err.message)
    }
  }

  if (loading) return <div>加载中...</div>

  return (
    <div className="page-container">
      <div className="page-header">
        <h2>设备台账</h2>
        <button className="btn btn-primary" onClick={() => setShowModal(true)}>
          + 注册新设备
        </button>
      </div>

      <div className="filters">
        <div className="filter-group">
          <label>科室</label>
          <select
            value={filters.department}
            onChange={(e) => setFilters({ ...filters, department: e.target.value })}
          >
            <option value="">全部</option>
            {departments.map((d) => (
              <option key={d} value={d}>{d}</option>
            ))}
          </select>
        </div>
        <div className="filter-group">
          <label>分类</label>
          <select
            value={filters.category}
            onChange={(e) => setFilters({ ...filters, category: e.target.value })}
          >
            <option value="">全部</option>
            {Object.entries(categoryLabels).map(([key, label]) => (
              <option key={key} value={key}>{label}</option>
            ))}
          </select>
        </div>
        <div className="filter-group">
          <label>状态</label>
          <select
            value={filters.status}
            onChange={(e) => setFilters({ ...filters, status: e.target.value })}
          >
            <option value="">全部</option>
            {Object.entries(statusLabels).map(([key, label]) => (
              <option key={key} value={key}>{label}</option>
            ))}
          </select>
        </div>
      </div>

      {error && <div style={{ color: 'red', marginBottom: '1rem' }}>{error}</div>}

      <table>
        <thead>
          <tr>
            <th>资产编号</th>
            <th>设备名称</th>
            <th>分类</th>
            <th>科室</th>
            <th>位置</th>
            <th>状态</th>
            <th>下次校准</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          {deviceList.map((device) => {
            const calStatus = getCalibrationStatus(device)
            return (
              <tr
                key={device.id}
                className={calStatus === 'expired' ? 'cal-expired' : calStatus === 'warning' ? 'cal-warning' : ''}
              >
                <td>{device.asset_number}</td>
                <td>{device.name}</td>
                <td>{categoryLabels[device.category]}</td>
                <td>{device.department}</td>
                <td>{device.location}</td>
                <td>
                  <span className={`status-badge status-${device.status}`}>
                    {statusLabels[device.status]}
                  </span>
                </td>
                <td>
                  {device.next_calibration_date ? (
                    <>
                      {new Date(device.next_calibration_date).toLocaleDateString()}
                      {calStatus === 'expired' && (
                        <span className="expired-alert" style={{ marginLeft: '0.5rem' }}>
                          请立即校准
                        </span>
                      )}
                    </>
                  ) : '-'}
                </td>
                <td>
                  <button className="btn btn-secondary" onClick={() => handleViewDetail(device.id)}>
                    详情
                  </button>
                </td>
              </tr>
            )
          })}
        </tbody>
      </table>

      {showModal && (
        <div className="modal-backdrop" onClick={() => setShowModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>注册新设备</h3>
              <button className="modal-close" onClick={() => setShowModal(false)}>&times;</button>
            </div>
            <form onSubmit={handleCreateDevice}>
              <div className="form-group">
                <label>资产编号 *</label>
                <input
                  type="text"
                  value={newDevice.asset_number}
                  onChange={(e) => setNewDevice({ ...newDevice, asset_number: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label>设备名称 *</label>
                <input
                  type="text"
                  value={newDevice.name}
                  onChange={(e) => setNewDevice({ ...newDevice, name: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label>品牌型号 *</label>
                <input
                  type="text"
                  value={newDevice.brand_model}
                  onChange={(e) => setNewDevice({ ...newDevice, brand_model: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label>序列号 *</label>
                <input
                  type="text"
                  value={newDevice.serial_number}
                  onChange={(e) => setNewDevice({ ...newDevice, serial_number: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label>分类 *</label>
                <select
                  value={newDevice.category}
                  onChange={(e) => setNewDevice({ ...newDevice, category: e.target.value })}
                >
                  {Object.entries(categoryLabels).map(([key, label]) => (
                    <option key={key} value={key}>{label}</option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label>科室 *</label>
                <select
                  value={newDevice.department}
                  onChange={(e) => setNewDevice({ ...newDevice, department: e.target.value })}
                  required
                >
                  <option value="">请选择</option>
                  {departments.map((d) => (
                    <option key={d} value={d}>{d}</option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label>安装位置 *</label>
                <input
                  type="text"
                  value={newDevice.location}
                  onChange={(e) => setNewDevice({ ...newDevice, location: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label>购入日期 *</label>
                <input
                  type="date"
                  value={newDevice.purchase_date}
                  onChange={(e) => setNewDevice({ ...newDevice, purchase_date: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label>购入价格 *</label>
                <input
                  type="number"
                  step="0.01"
                  value={newDevice.purchase_price}
                  onChange={(e) => setNewDevice({ ...newDevice, purchase_price: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label>保修截止日期</label>
                <input
                  type="date"
                  value={newDevice.warranty_expiry_date}
                  onChange={(e) => setNewDevice({ ...newDevice, warranty_expiry_date: e.target.value })}
                />
              </div>
              <div className="form-actions">
                <button type="button" className="btn btn-secondary" onClick={() => setShowModal(false)}>
                  取消
                </button>
                <button type="submit" className="btn btn-primary">注册</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {selectedDevice && (
        <div className="modal-backdrop" onClick={() => setSelectedDevice(null)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>设备详情</h3>
              <button className="modal-close" onClick={() => setSelectedDevice(null)}>&times;</button>
            </div>
            <div style={{ lineHeight: '2' }}>
              <p><strong>资产编号:</strong> {selectedDevice.asset_number}</p>
              <p><strong>设备名称:</strong> {selectedDevice.name}</p>
              <p><strong>品牌型号:</strong> {selectedDevice.brand_model}</p>
              <p><strong>序列号:</strong> {selectedDevice.serial_number}</p>
              <p><strong>分类:</strong> {categoryLabels[selectedDevice.category]}</p>
              <p><strong>科室:</strong> {selectedDevice.department}</p>
              <p><strong>位置:</strong> {selectedDevice.location}</p>
              <p><strong>购入日期:</strong> {new Date(selectedDevice.purchase_date).toLocaleDateString()}</p>
              <p><strong>购入价格:</strong> ¥{selectedDevice.purchase_price.toLocaleString()}</p>
              <p><strong>保修截止:</strong> {new Date(selectedDevice.warranty_expiry_date).toLocaleDateString()}</p>
              <p><strong>状态:</strong> <span className={`status-badge status-${selectedDevice.status}`}>{statusLabels[selectedDevice.status]}</span></p>
              <p><strong>上次校准:</strong> {selectedDevice.last_calibration_date ? new Date(selectedDevice.last_calibration_date).toLocaleDateString() : '-'}</p>
              <p><strong>下次校准:</strong> {selectedDevice.next_calibration_date ? new Date(selectedDevice.next_calibration_date).toLocaleDateString() : '-'}</p>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
