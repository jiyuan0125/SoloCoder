import React, { useState, useEffect } from 'react'
import dayjs from 'dayjs'
import { drugApi, prescriptionApi } from '../api'

const statusMap = {
  pending: { label: '待审核', tag: 'tag-pending' },
  approved: { label: '已审核', tag: 'tag-approved' },
  cancelled: { label: '已取消', tag: 'tag-cancelled' },
}

const defaultPrescription = {
  patient_name: '',
  patient_id: '',
  diagnosis: '',
  items: [{ drug_id: '', quantity: 1, usage: '' }],
}

const defaultApprove = {
  operator1: '',
  operator2: '',
}

export default function PrescriptionManagement() {
  const [prescriptions, setPrescriptions] = useState([])
  const [drugs, setDrugs] = useState([])
  const [filterStatus, setFilterStatus] = useState('')
  const [loading, setLoading] = useState(false)
  const [modalType, setModalType] = useState(null)
  const [selectedPrescription, setSelectedPrescription] = useState(null)
  const [prescriptionData, setPrescriptionData] = useState(defaultPrescription)
  const [approveData, setApproveData] = useState(defaultApprove)
  const [rejectComment, setRejectComment] = useState('')
  const [error, setError] = useState('')
  const [confirmData, setConfirmData] = useState(null)

  const fetchPrescriptions = async () => {
    setLoading(true)
    try {
      const res = await prescriptionApi.list(filterStatus)
      setPrescriptions(res.data.data || [])
    } catch (err) {
      console.error('获取处方列表失败:', err)
    } finally {
      setLoading(false)
    }
  }

  const fetchDrugs = async () => {
    try {
      const res = await drugApi.list()
      setDrugs(res.data.data || [])
    } catch (err) {
      console.error('获取药品列表失败:', err)
    }
  }

  useEffect(() => {
    fetchPrescriptions()
    fetchDrugs()
  }, [filterStatus])

  const openCreate = () => {
    setPrescriptionData({ ...defaultPrescription, items: [{ drug_id: '', quantity: 1, usage: '' }] })
    setError('')
    setModalType('create')
  }

  const openDetail = async (prescription) => {
    try {
      const res = await prescriptionApi.get(prescription.id)
      setSelectedPrescription(res.data.data)
      setModalType('detail')
    } catch (err) {
      console.error('获取处方详情失败:', err)
    }
  }

  const openApprove = (prescription) => {
    setSelectedPrescription(prescription)
    setApproveData(defaultApprove)
    setError('')
    setModalType('approve')
  }

  const openReject = (prescription) => {
    setSelectedPrescription(prescription)
    setRejectComment('')
    setError('')
    setModalType('reject')
  }

  const openCancel = (prescription) => {
    if (prescription.status === 'approved') {
      alert('已审核通过的处方不能取消')
      return
    }
    setSelectedPrescription(prescription)
    setConfirmData({ type: 'cancel' })
    setModalType('confirm')
  }

  const handlePrescriptionChange = (e) => {
    const { name, value } = e.target
    setPrescriptionData(prev => ({ ...prev, [name]: value }))
  }

  const handleItemChange = (index, field, value) => {
    setPrescriptionData(prev => {
      const newItems = [...prev.items]
      newItems[index] = { ...newItems[index], [field]: value }
      return { ...prev, items: newItems }
    })
  }

  const addItem = () => {
    setPrescriptionData(prev => ({
      ...prev,
      items: [...prev.items, { drug_id: '', quantity: 1, usage: '' }]
    }))
  }

  const removeItem = (index) => {
    if (prescriptionData.items.length <= 1) return
    setPrescriptionData(prev => ({
      ...prev,
      items: prev.items.filter((_, i) => i !== index)
    }))
  }

  const handleApproveChange = (e) => {
    const { name, value } = e.target
    setApproveData(prev => ({ ...prev, [name]: value }))
  }

  const submitPrescription = async () => {
    if (!prescriptionData.patient_name || !prescriptionData.patient_id || !prescriptionData.diagnosis) {
      setError('请填写完整的患者信息')
      return
    }
    if (prescriptionData.items.some(item => !item.drug_id || item.quantity <= 0 || !item.usage)) {
      setError('请填写完整的药品明细')
      return
    }

    try {
      await prescriptionApi.create({
        ...prescriptionData,
        items: prescriptionData.items.map(item => ({
          ...item,
          drug_id: parseInt(item.drug_id),
          quantity: parseFloat(item.quantity),
        }))
      })
      closeModal()
      fetchPrescriptions()
    } catch (err) {
      setError(err.response?.data?.error || '创建失败')
    }
  }

  const executeAction = async () => {
    if (!confirmData || !selectedPrescription) return
    
    try {
      if (confirmData.type === 'approve') {
        await prescriptionApi.approve(selectedPrescription.id, approveData)
      } else if (confirmData.type === 'reject') {
        await prescriptionApi.reject(selectedPrescription.id, { comment: rejectComment })
      } else if (confirmData.type === 'cancel') {
        await prescriptionApi.cancel(selectedPrescription.id)
      }
      closeModal()
      fetchPrescriptions()
    } catch (err) {
      setError(err.response?.data?.error || '操作失败')
      setConfirmData(null)
    }
  }

  const confirmApprove = () => {
    if (!approveData.operator1) {
      setError('请输入操作人')
      return
    }
    setConfirmData({ type: 'approve' })
  }

  const confirmReject = () => {
    if (!rejectComment) {
      setError('请输入审核意见')
      return
    }
    setConfirmData({ type: 'reject' })
  }

  const closeModal = () => {
    setModalType(null)
    setSelectedPrescription(null)
    setConfirmData(null)
    setError('')
  }

  const getDrugName = (drugId) => {
    const drug = drugs.find(d => d.id === drugId)
    return drug ? drug.generic_name : ''
  }

  return (
    <div className="page-container">
      <h2 className="page-title">处方管理</h2>

      <div className="toolbar">
        <div className="status-filter">
          <button
            className={`filter-btn ${filterStatus === '' ? 'active' : ''}`}
            onClick={() => setFilterStatus('')}
          >
            全部
          </button>
          <button
            className={`filter-btn ${filterStatus === 'pending' ? 'active' : ''}`}
            onClick={() => setFilterStatus('pending')}
          >
            待审核
          </button>
          <button
            className={`filter-btn ${filterStatus === 'approved' ? 'active' : ''}`}
            onClick={() => setFilterStatus('approved')}
          >
            已审核
          </button>
          <button
            className={`filter-btn ${filterStatus === 'cancelled' ? 'active' : ''}`}
            onClick={() => setFilterStatus('cancelled')}
          >
            已取消
          </button>
        </div>
        <button className="btn btn-primary" onClick={openCreate}>
          + 新增处方
        </button>
      </div>

      <table className="data-table">
        <thead>
          <tr>
            <th>处方编号</th>
            <th>患者姓名</th>
            <th>患者ID</th>
            <th>诊断</th>
            <th>状态</th>
            <th>创建时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          {loading ? (
            <tr><td colSpan="7" className="empty-state">加载中...</td></tr>
          ) : prescriptions.length === 0 ? (
            <tr><td colSpan="7" className="empty-state">暂无处方</td></tr>
          ) : (
            prescriptions.map(p => (
              <tr key={p.id}>
                <td>{p.prescription_no}</td>
                <td>{p.patient_name}</td>
                <td>{p.patient_id}</td>
                <td>{p.diagnosis}</td>
                <td>
                  <span className={`tag ${statusMap[p.status]?.tag}`}>
                    {statusMap[p.status]?.label}
                  </span>
                </td>
                <td>{dayjs(p.created_at).format('YYYY-MM-DD HH:mm')}</td>
                <td>
                  <div className="action-buttons">
                    <button className="btn btn-default btn-small" onClick={() => openDetail(p)}>
                      详情
                    </button>
                    {p.status === 'pending' && (
                      <>
                        <button className="btn btn-success btn-small" onClick={() => openApprove(p)}>
                          审核通过
                        </button>
                        <button className="btn btn-danger btn-small" onClick={() => openReject(p)}>
                          拒绝
                        </button>
                        <button className="btn btn-warning btn-small" onClick={() => openCancel(p)}>
                          取消
                        </button>
                      </>
                    )}
                  </div>
                </td>
              </tr>
            ))
          )}
        </tbody>
      </table>

      {modalType === 'create' && (
        <div className="modal-overlay" onClick={closeModal}>
          <div className="modal" style={{ maxWidth: '700px' }} onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>新增处方</h3>
              <button className="modal-close" onClick={closeModal}>&times;</button>
            </div>
            <div className="modal-body">
              {error && <div className="error-message">{error}</div>}
              
              <div className="form-row">
                <div className="form-group">
                  <label className="form-label">患者姓名 *</label>
                  <input
                    type="text"
                    className="form-input"
                    name="patient_name"
                    value={prescriptionData.patient_name}
                    onChange={handlePrescriptionChange}
                    required
                  />
                </div>
                <div className="form-group">
                  <label className="form-label">患者ID *</label>
                  <input
                    type="text"
                    className="form-input"
                    name="patient_id"
                    value={prescriptionData.patient_id}
                    onChange={handlePrescriptionChange}
                    required
                  />
                </div>
              </div>

              <div className="form-group">
                <label className="form-label">诊断 *</label>
                <input
                  type="text"
                  className="form-input"
                  name="diagnosis"
                  value={prescriptionData.diagnosis}
                  onChange={handlePrescriptionChange}
                  required
                />
              </div>

              <h4 className="section-title" style={{ marginTop: '8px' }}>处方明细</h4>
              
              {prescriptionData.items.map((item, index) => (
                <div key={index} className="form-row" style={{ marginBottom: '12px' }}>
                  <div className="form-group" style={{ flex: 2 }}>
                    <label className="form-label">药品 *</label>
                    <select
                      className="form-select"
                      value={item.drug_id}
                      onChange={(e) => handleItemChange(index, 'drug_id', e.target.value)}
                      required
                    >
                      <option value="">请选择药品</option>
                      {drugs.map(d => (
                        <option key={d.id} value={d.id}>
                          {d.generic_name} ({d.specification}) - 库存: {d.current_stock}
                          {d.support_split ? ' [可拆零]' : ''}
                        </option>
                      ))}
                    </select>
                  </div>
                  <div className="form-group">
                    <label className="form-label">数量 *</label>
                    <input
                      type="number"
                      step="0.5"
                      min="0.5"
                      className="form-input"
                      value={item.quantity}
                      onChange={(e) => handleItemChange(index, 'quantity', e.target.value)}
                      required
                    />
                  </div>
                  <div className="form-group" style={{ flex: 2 }}>
                    <label className="form-label">用法用量 *</label>
                    <input
                      type="text"
                      className="form-input"
                      value={item.usage}
                      onChange={(e) => handleItemChange(index, 'usage', e.target.value)}
                      placeholder="如：每日3次，每次1片"
                      required
                    />
                  </div>
                  {prescriptionData.items.length > 1 && (
                    <button
                      type="button"
                      className="btn btn-danger btn-small"
                      style={{ marginTop: '24px', height: '40px' }}
                      onClick={() => removeItem(index)}
                    >
                      删除
                    </button>
                  )}
                </div>
              ))}

              <button
                type="button"
                className="btn btn-default btn-small"
                onClick={addItem}
              >
                + 添加药品
              </button>
            </div>
            <div className="modal-footer">
              <button className="btn btn-default" onClick={closeModal}>取消</button>
              <button className="btn btn-primary" onClick={submitPrescription}>创建处方</button>
            </div>
          </div>
        </div>
      )}

      {modalType === 'detail' && selectedPrescription && (
        <div className="modal-overlay" onClick={closeModal}>
          <div className="modal" style={{ maxWidth: '700px' }} onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>处方详情</h3>
              <button className="modal-close" onClick={closeModal}>&times;</button>
            </div>
            <div className="modal-body">
              <div className="detail-section">
                <h4>基本信息</h4>
                <div className="detail-row">
                  <span className="label">处方编号：</span>
                  <span className="value">{selectedPrescription.prescription_no}</span>
                </div>
                <div className="detail-row">
                  <span className="label">患者姓名：</span>
                  <span className="value">{selectedPrescription.patient_name}</span>
                </div>
                <div className="detail-row">
                  <span className="label">患者ID：</span>
                  <span className="value">{selectedPrescription.patient_id}</span>
                </div>
                <div className="detail-row">
                  <span className="label">诊断：</span>
                  <span className="value">{selectedPrescription.diagnosis}</span>
                </div>
                <div className="detail-row">
                  <span className="label">状态：</span>
                  <span className={`tag ${statusMap[selectedPrescription.status]?.tag}`}>
                    {statusMap[selectedPrescription.status]?.label}
                  </span>
                </div>
                <div className="detail-row">
                  <span className="label">创建时间：</span>
                  <span className="value">{dayjs(selectedPrescription.created_at).format('YYYY-MM-DD HH:mm:ss')}</span>
                </div>
                {selectedPrescription.review_comment && (
                  <div className="detail-row">
                    <span className="label">审核意见：</span>
                    <span className="value">{selectedPrescription.review_comment}</span>
                  </div>
                )}
              </div>

              <h4>处方明细</h4>
              <div className="prescription-items">
                <table>
                  <thead>
                    <tr>
                      <th>药品名称</th>
                      <th>数量</th>
                      <th>用法用量</th>
                    </tr>
                  </thead>
                  <tbody>
                    {selectedPrescription.items?.map((item, index) => (
                      <tr key={index}>
                        <td>{item.drug?.generic_name || item.drug_name}</td>
                        <td>{item.quantity}</td>
                        <td>{item.usage}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
            <div className="modal-footer">
              <button className="btn btn-default" onClick={closeModal}>关闭</button>
            </div>
          </div>
        </div>
      )}

      {modalType === 'approve' && selectedPrescription && (
        <div className="modal-overlay" onClick={closeModal}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>审核通过处方</h3>
              <button className="modal-close" onClick={closeModal}>&times;</button>
            </div>
            {confirmData ? (
              <div className="modal-body confirm-dialog">
                <div className="confirm-icon">✅</div>
                <div className="confirm-message">确认审核通过？</div>
                <div className="confirm-detail">
                  处方编号：{selectedPrescription.prescription_no}<br/>
                  审核通过后将自动扣减库存
                </div>
              </div>
            ) : (
              <div className="modal-body">
                {error && <div className="error-message">{error}</div>}
                <div className="form-group">
                  <label className="form-label">操作人1 *</label>
                  <input
                    type="text"
                    className="form-input"
                    name="operator1"
                    value={approveData.operator1}
                    onChange={handleApproveChange}
                    required
                  />
                </div>
                <div className="form-group">
                  <label className="form-label">操作人2（特殊药品需填写）</label>
                  <input
                    type="text"
                    className="form-input"
                    name="operator2"
                    value={approveData.operator2}
                    onChange={handleApproveChange}
                  />
                </div>
              </div>
            )}
            <div className="modal-footer">
              {confirmData ? (
                <>
                  <button className="btn btn-default" onClick={() => setConfirmData(null)}>取消</button>
                  <button className="btn btn-success" onClick={executeAction}>确认通过</button>
                </>
              ) : (
                <>
                  <button className="btn btn-default" onClick={closeModal}>取消</button>
                  <button className="btn btn-primary" onClick={confirmApprove}>提交</button>
                </>
              )}
            </div>
          </div>
        </div>
      )}

      {modalType === 'reject' && selectedPrescription && (
        <div className="modal-overlay" onClick={closeModal}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>拒绝处方</h3>
              <button className="modal-close" onClick={closeModal}>&times;</button>
            </div>
            {confirmData ? (
              <div className="modal-body confirm-dialog">
                <div className="confirm-icon">❌</div>
                <div className="confirm-message">确认拒绝此处方？</div>
                <div className="confirm-detail">
                  处方编号：{selectedPrescription.prescription_no}
                </div>
              </div>
            ) : (
              <div className="modal-body">
                {error && <div className="error-message">{error}</div>}
                <div className="form-group">
                  <label className="form-label">审核意见 *</label>
                  <textarea
                    className="form-textarea"
                    value={rejectComment}
                    onChange={(e) => setRejectComment(e.target.value)}
                    placeholder="请输入拒绝原因..."
                    required
                  />
                </div>
              </div>
            )}
            <div className="modal-footer">
              {confirmData ? (
                <>
                  <button className="btn btn-default" onClick={() => setConfirmData(null)}>取消</button>
                  <button className="btn btn-danger" onClick={executeAction}>确认拒绝</button>
                </>
              ) : (
                <>
                  <button className="btn btn-default" onClick={closeModal}>取消</button>
                  <button className="btn btn-primary" onClick={confirmReject}>提交</button>
                </>
              )}
            </div>
          </div>
        </div>
      )}

      {modalType === 'confirm' && selectedPrescription && confirmData && (
        <div className="modal-overlay" onClick={closeModal}>
          <div className="modal confirm-dialog" onClick={(e) => e.stopPropagation()}>
            <div className="modal-body">
              <div className="confirm-icon">⚠️</div>
              <div className="confirm-message">确认取消此处方？</div>
              <div className="confirm-detail">
                处方编号：{selectedPrescription.prescription_no}
              </div>
            </div>
            <div className="modal-footer">
              <button className="btn btn-default" onClick={closeModal}>取消</button>
              <button className="btn btn-warning" onClick={executeAction}>确认取消</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
