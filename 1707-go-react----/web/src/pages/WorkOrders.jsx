import React, { useState, useEffect } from 'react'
import { workorders, devices } from '../api'

const priorityLabels = {
  normal: '普通',
  urgent: '紧急',
  critical: '特急',
}

const statusLabels = {
  pending: '待接单',
  in_progress: '维修中',
  pending_acceptance: '待验收',
  completed: '已完成',
  closed: '已关闭',
}

const nextStatusMap = {
  pending: { status: 'in_progress', label: '接单维修' },
  in_progress: { status: 'pending_acceptance', label: '申请验收' },
  pending_acceptance: { status: 'completed', label: '验收通过' },
  completed: { status: 'closed', label: '关闭工单' },
}

export default function WorkOrders() {
  const [orderList, setOrderList] = useState([])
  const [deviceList, setDeviceList] = useState([])
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showDetailModal, setShowDetailModal] = useState(false)
  const [selectedOrder, setSelectedOrder] = useState(null)
  const [newWorkOrder, setNewWorkOrder] = useState({
    device_id: '',
    description: '',
    priority: 'normal',
    assigned_to: '',
  })
  const [statusUpdateForm, setStatusUpdateForm] = useState({
    repair_content: '',
    replaced_parts: '',
  })
  const [loading, setLoading] = useState(true)

  const loadData = async () => {
    setLoading(true)
    try {
      const [orders, devs] = await Promise.all([
        workorders.getAll(),
        devices.getAll(),
      ])
      setOrderList(orders)
      setDeviceList(devs)
    } catch (err) {
      alert(err.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadData()
  }, [])

  const stats = {
    pending: orderList.filter((o) => o.status === 'pending').length,
    inProgress: orderList.filter((o) => o.status === 'in_progress').length,
    pendingAcceptance: orderList.filter((o) => o.status === 'pending_acceptance').length,
    critical: orderList.filter((o) => o.priority === 'critical' && o.status !== 'closed').length,
  }

  const handleCreate = async (e) => {
    e.preventDefault()
    try {
      await workorders.create({
        ...newWorkOrder,
        device_id: parseInt(newWorkOrder.device_id),
      })
      setShowCreateModal(false)
      setNewWorkOrder({ device_id: '', description: '', priority: 'normal', assigned_to: '' })
      loadData()
    } catch (err) {
      alert(err.message)
    }
  }

  const handleUpdateStatus = async (e) => {
    e.preventDefault()
    if (!selectedOrder) return

    const nextStatus = nextStatusMap[selectedOrder.status]
    if (!nextStatus) return

    try {
      await workorders.updateStatus(selectedOrder.id, {
        status: nextStatus.status,
        ...statusUpdateForm,
      })
      setShowDetailModal(false)
      setSelectedOrder(null)
      setStatusUpdateForm({ repair_content: '', replaced_parts: '' })
      loadData()
    } catch (err) {
      alert(err.message)
    }
  }

  const handleViewDetail = async (order) => {
    setSelectedOrder(order)
    setShowDetailModal(true)
  }

  if (loading) return <div>加载中...</div>

  return (
    <div className="page-container">
      <div className="page-header">
        <h2>工单管理</h2>
        <button className="btn btn-primary" onClick={() => setShowCreateModal(true)}>
          + 新建报修
        </button>
      </div>

      <div className="stats-cards">
        <div className="stat-card">
          <div className="stat-value">{stats.pending}</div>
          <div className="stat-label">待接单</div>
        </div>
        <div className="stat-card warning">
          <div className="stat-value">{stats.inProgress}</div>
          <div className="stat-label">维修中</div>
        </div>
        <div className="stat-card">
          <div className="stat-value">{stats.pendingAcceptance}</div>
          <div className="stat-label">待验收</div>
        </div>
        <div className="stat-card danger">
          <div className="stat-value">{stats.critical}</div>
          <div className="stat-label">特急工单</div>
        </div>
      </div>

      <table>
        <thead>
          <tr>
            <th>工单ID</th>
            <th>设备</th>
            <th>描述</th>
            <th>优先级</th>
            <th>状态</th>
            <th>报修时间</th>
            <th>升级次数</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          {orderList.map((order) => (
            <tr key={order.id}>
              <td>#{order.id}</td>
              <td>{order.device?.name || '-'}</td>
              <td style={{ maxWidth: '200px', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                {order.description}
              </td>
              <td>
                <span className={`status-badge priority-${order.priority}`}>
                  {priorityLabels[order.priority]}
                </span>
              </td>
              <td>
                <span className={`status-badge wo-${order.status}`}>
                  {statusLabels[order.status]}
                </span>
              </td>
              <td>{new Date(order.report_time).toLocaleString()}</td>
              <td>{order.upgrade_count || 0}</td>
              <td>
                <button className="btn btn-secondary" onClick={() => handleViewDetail(order)}>
                  详情
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {showCreateModal && (
        <div className="modal-backdrop" onClick={() => setShowCreateModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>故障报修</h3>
              <button className="modal-close" onClick={() => setShowCreateModal(false)}>&times;</button>
            </div>
            <form onSubmit={handleCreate}>
              <div className="form-group">
                <label>设备 *</label>
                <select
                  value={newWorkOrder.device_id}
                  onChange={(e) => setNewWorkOrder({ ...newWorkOrder, device_id: e.target.value })}
                  required
                >
                  <option value="">请选择设备</option>
                  {deviceList.map((d) => (
                    <option key={d.id} value={d.id}>
                      {d.asset_number} - {d.name}
                    </option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label>故障描述 *</label>
                <textarea
                  value={newWorkOrder.description}
                  onChange={(e) => setNewWorkOrder({ ...newWorkOrder, description: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label>紧急程度 *</label>
                <select
                  value={newWorkOrder.priority}
                  onChange={(e) => setNewWorkOrder({ ...newWorkOrder, priority: e.target.value })}
                >
                  {Object.entries(priorityLabels).map(([key, label]) => (
                    <option key={key} value={key}>{label}</option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label>分配给</label>
                <input
                  type="text"
                  value={newWorkOrder.assigned_to}
                  onChange={(e) => setNewWorkOrder({ ...newWorkOrder, assigned_to: e.target.value })}
                  placeholder="维修工程师姓名"
                />
              </div>
              <div className="form-actions">
                <button type="button" className="btn btn-secondary" onClick={() => setShowCreateModal(false)}>
                  取消
                </button>
                <button type="submit" className="btn btn-primary">提交报修</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {showDetailModal && selectedOrder && (
        <div className="modal-backdrop" onClick={() => setShowDetailModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>工单详情 #{selectedOrder.id}</h3>
              <button className="modal-close" onClick={() => setShowDetailModal(false)}>&times;</button>
            </div>
            <div style={{ lineHeight: '2' }}>
              <p><strong>设备:</strong> {selectedOrder.device?.name || '-'}</p>
              <p><strong>描述:</strong> {selectedOrder.description}</p>
              <p>
                <strong>优先级:</strong>{' '}
                <span className={`status-badge priority-${selectedOrder.priority}`}>
                  {priorityLabels[selectedOrder.priority]}
                </span>
              </p>
              <p>
                <strong>状态:</strong>{' '}
                <span className={`status-badge wo-${selectedOrder.status}`}>
                  {statusLabels[selectedOrder.status]}
                </span>
              </p>
              <p><strong>报修时间:</strong> {new Date(selectedOrder.report_time).toLocaleString()}</p>
              {selectedOrder.response_start_time && (
                <p><strong>响应开始时间:</strong> {new Date(selectedOrder.response_start_time).toLocaleString()}</p>
              )}
              {selectedOrder.accepted_time && (
                <p><strong>接单时间:</strong> {new Date(selectedOrder.accepted_time).toLocaleString()}</p>
              )}
              {selectedOrder.repair_content && (
                <p><strong>维修内容:</strong> {selectedOrder.repair_content}</p>
              )}
              {selectedOrder.upgrade_count > 0 && (
                <p><strong>紧急升级次数:</strong> {selectedOrder.upgrade_count}</p>
              )}
            </div>

            {nextStatusMap[selectedOrder.status] && (
              <form onSubmit={handleUpdateStatus} style={{ marginTop: '1.5rem' }}>
                {(selectedOrder.status === 'in_progress') && (
                  <>
                    <div className="form-group">
                      <label>维修内容 *</label>
                      <textarea
                        value={statusUpdateForm.repair_content}
                        onChange={(e) => setStatusUpdateForm({ ...statusUpdateForm, repair_content: e.target.value })}
                        required
                      />
                    </div>
                    <div className="form-group">
                      <label>更换配件</label>
                      <input
                        type="text"
                        value={statusUpdateForm.replaced_parts}
                        onChange={(e) => setStatusUpdateForm({ ...statusUpdateForm, replaced_parts: e.target.value })}
                      />
                    </div>
                  </>
                )}
                <div className="form-actions">
                  <button type="submit" className="btn btn-primary">
                    {nextStatusMap[selectedOrder.status].label}
                  </button>
                </div>
              </form>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
