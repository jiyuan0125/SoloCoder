import React, { useState, useEffect } from 'react'
import { calibration, devices } from '../api'

const resultLabels = {
  pass: '合格',
  fail: '不合格',
  conditional: '条件合格',
}

const statusLabels = {
  pending: '待校准',
  in_progress: '校准中',
  completed: '已校准',
}

export default function Calibration() {
  const [records, setRecords] = useState([])
  const [agencies, setAgencies] = useState([])
  const [deviceList, setDeviceList] = useState([])
  const [reminders, setReminders] = useState([])
  const [activeTab, setActiveTab] = useState('reminders')
  const [showRecordModal, setShowRecordModal] = useState(false)
  const [showAgencyModal, setShowAgencyModal] = useState(false)
  const [showUpdateModal, setShowUpdateModal] = useState(false)
  const [selectedRecord, setSelectedRecord] = useState(null)
  const [newRecord, setNewRecord] = useState({
    device_id: '',
    agency_id: '',
    calibration_date: '',
    certificate_number: '',
  })
  const [newAgency, setNewAgency] = useState({
    name: '',
    certification_number: '',
    contact_info: '',
  })
  const [updateForm, setUpdateForm] = useState({
    status: '',
    result: '',
    limited_functions: '',
    certificate_number: '',
  })
  const [loading, setLoading] = useState(true)

  const loadData = async () => {
    setLoading(true)
    try {
      const [recs, ags, devs, rems] = await Promise.all([
        calibration.getRecords(),
        calibration.getAgencies(),
        devices.getAll(),
        calibration.getReminders(),
      ])
      setRecords(recs)
      setAgencies(ags)
      setDeviceList(devs)
      setReminders(rems)
    } catch (err) {
      console.error('加载数据失败:', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadData()
  }, [])

  const handleCreateRecord = async (e) => {
    e.preventDefault()
    try {
      await calibration.createRecord({
        ...newRecord,
        device_id: parseInt(newRecord.device_id),
        agency_id: parseInt(newRecord.agency_id),
      })
      setShowRecordModal(false)
      setNewRecord({ device_id: '', agency_id: '', calibration_date: '', certificate_number: '' })
      loadData()
    } catch (err) {
      alert(err.message)
    }
  }

  const handleCreateAgency = async (e) => {
    e.preventDefault()
    try {
      await calibration.createAgency(newAgency)
      setShowAgencyModal(false)
      setNewAgency({ name: '', certification_number: '', contact_info: '' })
      loadData()
    } catch (err) {
      alert(err.message)
    }
  }

  const handleUpdateStatus = async (e) => {
    e.preventDefault()
    if (!selectedRecord) return

    try {
      await calibration.updateStatus(selectedRecord.id, updateForm)
      setShowUpdateModal(false)
      setSelectedRecord(null)
      setUpdateForm({ status: '', result: '', limited_functions: '', certificate_number: '' })
      loadData()
    } catch (err) {
      alert(err.message)
    }
  }

  const openUpdateModal = (record) => {
    setSelectedRecord(record)
    if (record.status === 'pending') {
      setUpdateForm({ ...updateForm, status: 'in_progress' })
    } else if (record.status === 'in_progress') {
      setUpdateForm({ ...updateForm, status: 'completed' })
    }
    setShowUpdateModal(true)
  }

  if (loading) return <div>加载中...</div>

  return (
    <div className="page-container">
      <div className="page-header">
        <h2>校准管理</h2>
        <div style={{ display: 'flex', gap: '0.5rem' }}>
          <button className="btn btn-secondary" onClick={() => setShowAgencyModal(true)}>
            + 校准机构
          </button>
          <button className="btn btn-primary" onClick={() => setShowRecordModal(true)}>
            + 校准记录
          </button>
        </div>
      </div>

      <div className="tabs">
        <button
          className={`tab ${activeTab === 'reminders' ? 'active' : ''}`}
          onClick={() => setActiveTab('reminders')}
        >
          到期提醒 ({reminders.length})
        </button>
        <button
          className={`tab ${activeTab === 'records' ? 'active' : ''}`}
          onClick={() => setActiveTab('records')}
        >
          校准记录
        </button>
        <button
          className={`tab ${activeTab === 'agencies' ? 'active' : ''}`}
          onClick={() => setActiveTab('agencies')}
        >
          校准机构
        </button>
      </div>

      {activeTab === 'reminders' && (
        <>
          {reminders.length === 0 ? (
            <div className="empty-state">暂无即将到期或已过期的校准</div>
          ) : (
            <table>
              <thead>
                <tr>
                  <th>资产编号</th>
                  <th>设备名称</th>
                  <th>到期日期</th>
                  <th>状态</th>
                  <th>剩余天数</th>
                </tr>
              </thead>
              <tbody>
                {reminders.map((r) => (
                  <tr key={r.device_id} className={r.is_overdue ? 'cal-expired' : 'cal-warning'}>
                    <td>{r.asset_number}</td>
                    <td>{r.device_name}</td>
                    <td>{new Date(r.due_date).toLocaleDateString()}</td>
                    <td>
                      {r.is_overdue ? (
                        <span className="expired-alert">请立即校准</span>
                      ) : (
                        <span style={{ color: '#ff9800' }}>即将到期</span>
                      )}
                    </td>
                    <td style={{ color: r.is_overdue ? '#f44336' : '#ff9800', fontWeight: 'bold' }}>
                      {r.is_overdue ? `已过期 ${Math.abs(r.days_left)} 天` : `${r.days_left} 天`}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </>
      )}

      {activeTab === 'records' && (
        <>
          {records.length === 0 ? (
            <div className="empty-state">暂无校准记录</div>
          ) : (
            <table>
              <thead>
                <tr>
                  <th>ID</th>
                  <th>设备</th>
                  <th>校准机构</th>
                  <th>校准日期</th>
                  <th>状态</th>
                  <th>结果</th>
                  <th>下次校准</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {records.map((r) => (
                  <tr key={r.id}>
                    <td>#{r.id}</td>
                    <td>{r.device?.name || '-'}</td>
                    <td>{r.agency?.name || '-'}</td>
                    <td>{new Date(r.calibration_date).toLocaleDateString()}</td>
                    <td>
                      <span className={`status-badge ${r.status === 'completed' ? 'wo-completed' : r.status === 'in_progress' ? 'wo-pending' : 'wo-pending_acceptance'}`}>
                        {statusLabels[r.status]}
                      </span>
                    </td>
                    <td>
                      {r.result ? (
                        <span className={`status-badge ${r.result === 'pass' ? 'wo-completed' : r.result === 'fail' ? 'wo-closed' : 'priority-urgent'}`}>
                          {resultLabels[r.result]}
                        </span>
                      ) : '-'}
                    </td>
                    <td>
                      {r.next_calibration_date
                        ? new Date(r.next_calibration_date).toLocaleDateString()
                        : '-'}
                    </td>
                    <td>
                      {r.status !== 'completed' && (
                        <button className="btn btn-secondary" onClick={() => openUpdateModal(r)}>
                          更新状态
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </>
      )}

      {activeTab === 'agencies' && (
        <>
          {agencies.length === 0 ? (
            <div className="empty-state">暂无校准机构</div>
          ) : (
            <table>
              <thead>
                <tr>
                  <th>ID</th>
                  <th>名称</th>
                  <th>资质编号</th>
                  <th>联系方式</th>
                </tr>
              </thead>
              <tbody>
                {agencies.map((a) => (
                  <tr key={a.id}>
                    <td>#{a.id}</td>
                    <td>{a.name}</td>
                    <td>{a.certification_number}</td>
                    <td>{a.contact_info}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </>
      )}

      {showRecordModal && (
        <div className="modal-backdrop" onClick={() => setShowRecordModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>创建校准记录</h3>
              <button className="modal-close" onClick={() => setShowRecordModal(false)}>&times;</button>
            </div>
            <form onSubmit={handleCreateRecord}>
              <div className="form-group">
                <label>设备 *</label>
                <select
                  value={newRecord.device_id}
                  onChange={(e) => setNewRecord({ ...newRecord, device_id: e.target.value })}
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
                <label>校准机构 *</label>
                <select
                  value={newRecord.agency_id}
                  onChange={(e) => setNewRecord({ ...newRecord, agency_id: e.target.value })}
                  required
                >
                  <option value="">请选择机构</option>
                  {agencies.map((a) => (
                    <option key={a.id} value={a.id}>
                      {a.name}
                    </option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label>校准日期 *</label>
                <input
                  type="date"
                  value={newRecord.calibration_date}
                  onChange={(e) => setNewRecord({ ...newRecord, calibration_date: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label>校准证书编号</label>
                <input
                  type="text"
                  value={newRecord.certificate_number}
                  onChange={(e) => setNewRecord({ ...newRecord, certificate_number: e.target.value })}
                />
              </div>
              <div className="form-actions">
                <button type="button" className="btn btn-secondary" onClick={() => setShowRecordModal(false)}>
                  取消
                </button>
                <button type="submit" className="btn btn-primary">创建</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {showAgencyModal && (
        <div className="modal-backdrop" onClick={() => setShowAgencyModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>添加校准机构</h3>
              <button className="modal-close" onClick={() => setShowAgencyModal(false)}>&times;</button>
            </div>
            <form onSubmit={handleCreateAgency}>
              <div className="form-group">
                <label>机构名称 *</label>
                <input
                  type="text"
                  value={newAgency.name}
                  onChange={(e) => setNewAgency({ ...newAgency, name: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label>资质编号 *</label>
                <input
                  type="text"
                  value={newAgency.certification_number}
                  onChange={(e) => setNewAgency({ ...newAgency, certification_number: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label>联系方式 *</label>
                <input
                  type="text"
                  value={newAgency.contact_info}
                  onChange={(e) => setNewAgency({ ...newAgency, contact_info: e.target.value })}
                  required
                />
              </div>
              <div className="form-actions">
                <button type="button" className="btn btn-secondary" onClick={() => setShowAgencyModal(false)}>
                  取消
                </button>
                <button type="submit" className="btn btn-primary">添加</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {showUpdateModal && selectedRecord && (
        <div className="modal-backdrop" onClick={() => setShowUpdateModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>更新校准状态</h3>
              <button className="modal-close" onClick={() => setShowUpdateModal(false)}>&times;</button>
            </div>
            <form onSubmit={handleUpdateStatus}>
              <p style={{ marginBottom: '1rem' }}>
                <strong>设备:</strong> {selectedRecord.device?.name}<br />
                <strong>当前状态:</strong> {statusLabels[selectedRecord.status]}
              </p>
              <div className="form-group">
                <label>新状态</label>
                <select
                  value={updateForm.status}
                  onChange={(e) => setUpdateForm({ ...updateForm, status: e.target.value })}
                >
                  {selectedRecord.status === 'pending' && (
                    <option value="in_progress">校准中</option>
                  )}
                  {selectedRecord.status === 'in_progress' && (
                    <option value="completed">已校准</option>
                  )}
                </select>
              </div>
              {selectedRecord.status === 'in_progress' && (
                <>
                  <div className="form-group">
                    <label>校准结果</label>
                    <select
                      value={updateForm.result}
                      onChange={(e) => setUpdateForm({ ...updateForm, result: e.target.value })}
                    >
                      <option value="">请选择</option>
                      {Object.entries(resultLabels).map(([key, label]) => (
                        <option key={key} value={key}>{label}</option>
                      ))}
                    </select>
                  </div>
                  {updateForm.result === 'conditional' && (
                    <div className="form-group">
                      <label>受限功能</label>
                      <textarea
                        value={updateForm.limited_functions}
                        onChange={(e) => setUpdateForm({ ...updateForm, limited_functions: e.target.value })}
                        placeholder="请描述哪些功能受限"
                      />
                    </div>
                  )}
                  <div className="form-group">
                    <label>校准证书编号</label>
                    <input
                      type="text"
                      value={updateForm.certificate_number}
                      onChange={(e) => setUpdateForm({ ...updateForm, certificate_number: e.target.value })}
                    />
                  </div>
                </>
              )}
              <div className="form-actions">
                <button type="button" className="btn btn-secondary" onClick={() => setShowUpdateModal(false)}>
                  取消
                </button>
                <button type="submit" className="btn btn-primary">确认</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
