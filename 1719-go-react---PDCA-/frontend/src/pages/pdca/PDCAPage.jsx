import React, { useState, useEffect } from 'react'
import { pdca, indicators } from '../../api'
import { PDCA_STATUS, PDCA_PHASES, getStatusLabel, getStatusColor } from '../../types'

const PDCAPage = () => {
  const [pdcaList, setPdcaList] = useState([])
  const [indicatorList, setIndicatorList] = useState([])
  const [selectedPDCA, setSelectedPDCA] = useState(null)
  const [phaseDetails, setPhaseDetails] = useState([])
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showPhaseModal, setShowPhaseModal] = useState(false)
  const [currentPhaseEdit, setCurrentPhaseEdit] = useState(null)
  const [formData, setFormData] = useState({
    name: '', indicator_id: '', responsible: '', start_date: ''
  })
  const [phaseForm, setPhaseForm] = useState({ content: '', complete_date: '', evidence: '' })

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    try {
      const [pdcaRes, indRes] = await Promise.all([pdca.list(), indicators.list()])
      setPdcaList(pdcaRes.data)
      setIndicatorList(indRes.data)
    } catch (err) {
      console.error('加载数据失败', err)
    }
  }

  const loadPDCADetail = async (item) => {
    setSelectedPDCA(item)
    try {
      const res = await pdca.get(item.id)
      setPhaseDetails(res.data.phases || [])
    } catch (err) {
      setPhaseDetails([])
    }
  }

  const handleCreatePDCA = async () => {
    try {
      const data = {
        ...formData,
        indicator_id: formData.indicator_id ? parseInt(formData.indicator_id) : null
      }
      await pdca.create(data)
      setShowCreateModal(false)
      setFormData({ name: '', indicator_id: '', responsible: '', start_date: '' })
      loadData()
      alert('PDCA项目创建成功')
    } catch (err) {
      alert('创建失败: ' + (err.response?.data?.error || err.message))
    }
  }

  const handleSavePhase = async () => {
    if (!selectedPDCA || !currentPhaseEdit) return
    try {
      const existing = phaseDetails.find(p => p.phase === currentPhaseEdit)
      if (existing) {
        await pdca.updatePhase(selectedPDCA.id, currentPhaseEdit, phaseForm)
      } else {
        await pdca.createPhase(selectedPDCA.id, { ...phaseForm, phase: currentPhaseEdit })
      }
      setShowPhaseModal(false)
      setPhaseForm({ content: '', complete_date: '', evidence: '' })
      loadPDCADetail(selectedPDCA)
      alert('阶段内容保存成功')
    } catch (err) {
      alert('保存失败: ' + (err.response?.data?.error || err.message))
    }
  }

  const handleNextPhase = async () => {
    if (!selectedPDCA) return
    try {
      await pdca.nextPhase(selectedPDCA.id)
      alert('阶段推进成功')
      loadData()
      loadPDCADetail(selectedPDCA)
    } catch (err) {
      alert('推进失败: ' + (err.response?.data?.error || err.message))
    }
  }

  const handleStartNextCycle = async () => {
    if (!selectedPDCA) return
    if (!confirm('确定要启动下一轮循环吗？当前项目将关闭。')) return
    try {
      await pdca.startNextCycle(selectedPDCA.id)
      alert('下一轮循环已启动')
      loadData()
    } catch (err) {
      alert('启动失败: ' + (err.response?.data?.error || err.message))
    }
  }

  const getPhaseIndex = (phase) => PDCA_PHASES.findIndex(p => p.value === phase)
  const getPhaseStatus = (phase, currentPhase) => {
    const phaseIdx = getPhaseIndex(phase)
    const currentIdx = getPhaseIndex(currentPhase)
    if (phaseIdx < currentIdx) return 'completed'
    if (phaseIdx === currentIdx) return 'current'
    return ''
  }

  const openPhaseEditor = (phase) => {
    const existing = phaseDetails.find(p => p.phase === phase)
    if (existing) {
      setPhaseForm({
        content: existing.content,
        complete_date: existing.complete_date || '',
        evidence: existing.evidence || ''
      })
    } else {
      setPhaseForm({ content: '', complete_date: '', evidence: '' })
    }
    setCurrentPhaseEdit(phase)
    setShowPhaseModal(true)
  }

  return (
    <div>
      <div className="page-header">
        <h1 className="page-title">PDCA管理</h1>
        <p className="page-subtitle">管理质量改进项目，追踪PDCA循环进度</p>
      </div>

      <div className="card">
        <div className="card-header">
          <h2 className="card-title">PDCA项目列表</h2>
          <button className="btn btn-primary" onClick={() => setShowCreateModal(true)}>
            + 新建PDCA项目
          </button>
        </div>

        <table>
          <thead>
            <tr>
              <th>项目名称</th>
              <th>关联指标</th>
              <th>负责人</th>
              <th>开始日期</th>
              <th>当前阶段</th>
              <th>状态</th>
              <th>改进效果</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {pdcaList.map(item => (
              <tr key={item.id} onClick={() => loadPDCADetail(item)} style={{ cursor: 'pointer' }}>
                <td>{item.name}</td>
                <td>{item.indicator?.name || '-'}</td>
                <td>{item.responsible}</td>
                <td>{item.start_date}</td>
                <td>{getStatusLabel(item.current_phase, PDCA_PHASES)}</td>
                <td>
                  <span className={`status-badge ${getStatusColor(item.status, PDCA_STATUS)}`}>
                    {getStatusLabel(item.status, PDCA_STATUS)}
                  </span>
                </td>
                <td>
                  {item.is_improved ? (
                    <span className="status-badge status-success">✓ 改进成功</span>
                  ) : '-'}
                </td>
                <td>
                  <button className="btn btn-sm btn-primary"
                    onClick={(e) => { e.stopPropagation(); loadPDCADetail(item); }}>
                    查看详情
                  </button>
                </td>
              </tr>
            ))}
            {pdcaList.length === 0 && (
              <tr><td colSpan="8" className="empty-state">暂无PDCA项目</td></tr>
            )}
          </tbody>
        </table>
      </div>

      {selectedPDCA && (
        <div className="card">
          <div className="card-header">
            <h2 className="card-title">{selectedPDCA.name} - 进度追踪</h2>
            <div style={{ display: 'flex', gap: '8px' }}>
              {selectedPDCA.current_phase !== 'act' && (
                <button className="btn btn-success" onClick={handleNextPhase}>
                  推进到下一阶段
                </button>
              )}
              {selectedPDCA.current_phase === 'act' && selectedPDCA.status !== 'closed' && (
                <button className="btn btn-primary" onClick={handleStartNextCycle}>
                  启动下一轮循环
                </button>
              )}
            </div>
          </div>

          <div className="pdca-flow">
            {PDCA_PHASES.map((phase, idx) => {
              const status = getPhaseStatus(phase.value, selectedPDCA.current_phase)
              const detail = phaseDetails.find(p => p.phase === phase.value)
              return (
                <React.Fragment key={phase.value}>
                  <div className={`pdca-phase ${status}`} onClick={() => openPhaseEditor(phase.value)} style={{ cursor: 'pointer' }}>
                    <div className="pdca-circle">
                      {status === 'completed' ? <span className="check-icon">✓</span> : phase.label}
                    </div>
                    <div className="pdca-label">{phase.fullLabel}</div>
                    {detail && <div style={{ fontSize: '11px', color: '#999', marginTop: '4px' }}>已填写</div>}
                  </div>
                  {idx < PDCA_PHASES.length - 1 && (
                    <div className={`pdca-arrow ${getPhaseIndex(selectedPDCA.current_phase) > idx ? 'completed' : ''}`}>
                      →
                    </div>
                  )}
                </React.Fragment>
              )
            })}
          </div>

          {phaseDetails.length > 0 && (
            <div style={{ marginTop: '20px' }}>
              <h3 style={{ marginBottom: '16px', fontSize: '15px', color: '#555' }}>阶段详情</h3>
              {phaseDetails.map(detail => {
                const phase = PDCA_PHASES.find(p => p.value === detail.phase)
                return (
                  <div key={detail.phase} className="card" style={{ marginBottom: '12px', background: '#fafbfc' }}>
                    <div className="card-header" style={{ marginBottom: '8px', paddingBottom: '8px' }}>
                      <h4 className="card-title">{phase?.fullLabel} ({phase?.label})</h4>
                      {detail.complete_date && <span style={{ fontSize: '13px', color: '#666' }}>完成日期: {detail.complete_date}</span>}
                    </div>
                    <div style={{ fontSize: '14px', lineHeight: '1.6', color: '#444' }}>
                      <p><strong>内容:</strong> {detail.content}</p>
                      {detail.evidence && <p style={{ marginTop: '8px' }}><strong>证据材料:</strong> {detail.evidence}</p>}
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </div>
      )}

      {showCreateModal && (
        <div className="modal-overlay" onClick={() => setShowCreateModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3 className="modal-title">新建PDCA项目</h3>
              <button className="modal-close" onClick={() => setShowCreateModal(false)}>×</button>
            </div>

            <div className="form-group">
              <label className="form-label">项目名称 *</label>
              <input className="form-input" value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })} />
            </div>

            <div className="form-group">
              <label className="form-label">关联质量指标</label>
              <select className="form-select" value={formData.indicator_id}
                onChange={(e) => setFormData({ ...formData, indicator_id: e.target.value })}>
                <option value="">不关联</option>
                {indicatorList.map(ind => (
                  <option key={ind.id} value={ind.id}>{ind.code} - {ind.name}</option>
                ))}
              </select>
            </div>

            <div className="form-row">
              <div className="form-group">
                <label className="form-label">负责人 *</label>
                <input className="form-input" value={formData.responsible}
                  onChange={(e) => setFormData({ ...formData, responsible: e.target.value })} />
              </div>
              <div className="form-group">
                <label className="form-label">开始日期 *</label>
                <input className="form-input" type="date" value={formData.start_date}
                  onChange={(e) => setFormData({ ...formData, start_date: e.target.value })} />
              </div>
            </div>

            <div className="modal-footer">
              <button className="btn btn-primary" onClick={handleCreatePDCA}>创建</button>
              <button className="btn btn-danger" onClick={() => setShowCreateModal(false)}>取消</button>
            </div>
          </div>
        </div>
      )}

      {showPhaseModal && currentPhaseEdit && (
        <div className="modal-overlay" onClick={() => setShowPhaseModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3 className="modal-title">
                编辑 {PDCA_PHASES.find(p => p.value === currentPhaseEdit)?.fullLabel}
              </h3>
              <button className="modal-close" onClick={() => setShowPhaseModal(false)}>×</button>
            </div>

            <div className="form-group">
              <label className="form-label">阶段内容描述 *</label>
              <textarea className="form-textarea" value={phaseForm.content}
                onChange={(e) => setPhaseForm({ ...phaseForm, content: e.target.value })}
                placeholder="请详细描述该阶段的工作内容..." />
            </div>

            <div className="form-row">
              <div className="form-group">
                <label className="form-label">完成日期</label>
                <input className="form-input" type="date" value={phaseForm.complete_date}
                  onChange={(e) => setPhaseForm({ ...phaseForm, complete_date: e.target.value })} />
              </div>
              <div className="form-group">
                <label className="form-label">相关证据材料</label>
                <input className="form-input" value={phaseForm.evidence}
                  onChange={(e) => setPhaseForm({ ...phaseForm, evidence: e.target.value })}
                  placeholder="简要描述证据材料" />
              </div>
            </div>

            <div className="modal-footer">
              <button className="btn btn-primary" onClick={handleSavePhase}>保存</button>
              <button className="btn btn-danger" onClick={() => setShowPhaseModal(false)}>取消</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default PDCAPage
