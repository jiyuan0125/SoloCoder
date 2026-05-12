import React, { useState, useEffect } from 'react'
import { donorsAPI, organsAPI, locationsAPI } from '../api'

const organTypes = [
  { value: 'heart', label: '心脏' },
  { value: 'liver', label: '肝脏' },
  { value: 'kidney_left', label: '左肾' },
  { value: 'kidney_right', label: '右肾' },
  { value: 'lung_left', label: '左肺' },
  { value: 'lung_right', label: '右肺' },
  { value: 'pancreas', label: '胰腺' },
  { value: 'cornea', label: '角膜' },
  { value: 'intestine', label: '小肠' }
]

const getOrganStatusBadge = (status) => {
  const mapping = {
    '待匹配': 'badge-pending',
    '已匹配': 'badge-matched',
    '已获取': 'badge-acquired',
    '已移植': 'badge-transplanted',
    '已超时废弃': 'badge-discarded'
  }
  return mapping[status] || 'badge-pending'
}

const getOrganTypeLabel = (type) => {
  const organ = organTypes.find(o => o.value === type)
  return organ ? organ.label : type
}

function DonorForm({ onSubmit, onClose, locations }) {
  const [form, setForm] = useState({
    donor_no: '',
    name: '',
    gender: '男',
    age: '',
    blood_type: 'A',
    height: '',
    weight: '',
    death_date: '',
    cause_of_death: '',
    organs: [],
    location_id: locations[0]?.id || ''
  })

  const handleOrganToggle = (type) => {
    setForm(prev => ({
      ...prev,
      organs: prev.organs.includes(type)
        ? prev.organs.filter(o => o !== type)
        : [...prev.organs, type]
    }))
  }

  const handleSubmit = (e) => {
    e.preventDefault()
    onSubmit({
      ...form,
      age: parseInt(form.age),
      height: parseFloat(form.height),
      weight: parseFloat(form.weight),
      death_date: new Date(form.death_date).toISOString(),
      organs: form.organs.map(type => ({ organ_type: type }))
    })
  }

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h2>登记捐献者</h2>
          <button className="close-btn" onClick={onClose}>×</button>
        </div>
        <form onSubmit={handleSubmit}>
          <div className="form-row">
            <div className="form-group">
              <label>捐献者编号</label>
              <input required value={form.donor_no}
                onChange={e => setForm({...form, donor_no: e.target.value})} />
            </div>
            <div className="form-group">
              <label>姓名</label>
              <input required value={form.name}
                onChange={e => setForm({...form, name: e.target.value})} />
            </div>
          </div>
          <div className="form-row">
            <div className="form-group">
              <label>性别</label>
              <select value={form.gender}
                onChange={e => setForm({...form, gender: e.target.value})}>
                <option value="男">男</option>
                <option value="女">女</option>
              </select>
            </div>
            <div className="form-group">
              <label>年龄</label>
              <input type="number" required min="0" value={form.age}
                onChange={e => setForm({...form, age: e.target.value})} />
            </div>
            <div className="form-group">
              <label>血型</label>
              <select value={form.blood_type}
                onChange={e => setForm({...form, blood_type: e.target.value})}>
                <option value="A">A型</option>
                <option value="B">B型</option>
                <option value="AB">AB型</option>
                <option value="O">O型</option>
              </select>
            </div>
          </div>
          <div className="form-row">
            <div className="form-group">
              <label>身高 (cm)</label>
              <input type="number" required min="0" value={form.height}
                onChange={e => setForm({...form, height: e.target.value})} />
            </div>
            <div className="form-group">
              <label>体重 (kg)</label>
              <input type="number" required min="0" value={form.weight}
                onChange={e => setForm({...form, weight: e.target.value})} />
            </div>
            <div className="form-group">
              <label>死亡日期</label>
              <input type="date" required value={form.death_date}
                onChange={e => setForm({...form, death_date: e.target.value})} />
            </div>
          </div>
          <div className="form-group">
            <label>死亡原因</label>
            <input required value={form.cause_of_death}
              onChange={e => setForm({...form, cause_of_death: e.target.value})} />
          </div>
          <div className="form-group">
            <label>位置</label>
            <select value={form.location_id}
              onChange={e => setForm({...form, location_id: e.target.value})}>
              {locations.map(loc => (
                <option key={loc.id} value={loc.id}>{loc.name}</option>
              ))}
            </select>
          </div>
          <div className="form-group">
            <label>捐献器官 (可多选)</label>
            <div style={{display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '10px'}}>
              {organTypes.map(organ => (
                <label key={organ.value} style={{display: 'flex', alignItems: 'center', gap: '8px', cursor: 'pointer'}}>
                  <input type="checkbox"
                    checked={form.organs.includes(organ.value)}
                    onChange={() => handleOrganToggle(organ.value)} />
                  {organ.label}
                </label>
              ))}
            </div>
          </div>
          <div className="modal-footer">
            <button type="button" className="btn" onClick={onClose}>取消</button>
            <button type="submit" className="btn btn-primary">登记</button>
          </div>
        </form>
      </div>
    </div>
  )
}

function OrganAssessmentModal({ organ, onSubmit, onClose }) {
  const [form, setForm] = useState({
    function_score: '',
    has_vessel_abnormality: false,
    has_other_lesions: false
  })

  const handleSubmit = (e) => {
    e.preventDefault()
    onSubmit({
      ...form,
      function_score: parseInt(form.function_score)
    })
  }

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h2>器官评估 - {getOrganTypeLabel(organ.organ_type)}</h2>
          <button className="close-btn" onClick={onClose}>×</button>
        </div>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>器官功能评分 (0-100，低于60分不能用于移植)</label>
            <input type="number" required min="0" max="100" value={form.function_score}
              onChange={e => setForm({...form, function_score: e.target.value})} />
          </div>
          <div className="form-group">
            <label style={{display: 'flex', alignItems: 'center', gap: '8px'}}>
              <input type="checkbox"
                checked={form.has_vessel_abnormality}
                onChange={e => setForm({...form, has_vessel_abnormality: e.target.checked})} />
              是否有血管异常
            </label>
          </div>
          <div className="form-group">
            <label style={{display: 'flex', alignItems: 'center', gap: '8px'}}>
              <input type="checkbox"
                checked={form.has_other_lesions}
                onChange={e => setForm({...form, has_other_lesions: e.target.checked})} />
              是否有其他病变
            </label>
          </div>
          <div className="modal-footer">
            <button type="button" className="btn" onClick={onClose}>取消</button>
            <button type="submit" className="btn btn-primary">提交评估</button>
          </div>
        </form>
      </div>
    </div>
  )
}

function DonorDetail({ donor, onClose, onAssess, onStartMatching }) {
  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" style={{maxWidth: '800px'}} onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h2>捐献者详情 - {donor.name}</h2>
          <button className="close-btn" onClick={onClose}>×</button>
        </div>
        <div className="detail-section">
          <h3>基本信息</h3>
          <div className="detail-grid">
            <div className="detail-item"><label>编号</label><div className="value">{donor.donor_no}</div></div>
            <div className="detail-item"><label>姓名</label><div className="value">{donor.name}</div></div>
            <div className="detail-item"><label>性别</label><div className="value">{donor.gender}</div></div>
            <div className="detail-item"><label>年龄</label><div className="value">{donor.age}岁</div></div>
            <div className="detail-item"><label>血型</label><div className="value">{donor.blood_type}型</div></div>
            <div className="detail-item"><label>身高/体重</label><div className="value">{donor.height}cm / {donor.weight}kg</div></div>
          </div>
        </div>
        <div className="detail-section">
          <h3>捐献器官</h3>
          <div className="table-container">
            <table>
              <thead>
                <tr>
                  <th>器官类型</th>
                  <th>状态</th>
                  <th>功能评分</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {donor.organs?.map(organ => (
                  <tr key={organ.id}>
                    <td>{getOrganTypeLabel(organ.organ_type)}</td>
                    <td><span className={`badge ${getOrganStatusBadge(organ.status)}`}>{organ.status}</span></td>
                    <td>{organ.assessment?.function_score ?? '未评估'}</td>
                    <td>
                      {organ.status === '待匹配' && !organ.assessment && (
                        <button className="btn btn-sm btn-primary" onClick={() => onAssess(organ)}>评估</button>
                      )}
                      {organ.status === '待匹配' && organ.assessment && organ.assessment.function_score >= 60 && (
                        <button className="btn btn-sm btn-success" onClick={() => onStartMatching(organ.id)}>开始匹配</button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
        <div className="modal-footer">
          <button className="btn" onClick={onClose}>关闭</button>
        </div>
      </div>
    </div>
  )
}

export default function DonorsPage() {
  const [donors, setDonors] = useState([])
  const [locations, setLocations] = useState([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [selectedDonor, setSelectedDonor] = useState(null)
  const [assessingOrgan, setAssessingOrgan] = useState(null)
  const [message, setMessage] = useState(null)

  const loadData = async () => {
    try {
      setLoading(true)
      const [donorsData, locationsData] = await Promise.all([
        donorsAPI.list(1, 50),
        locationsAPI.list()
      ])
      setDonors(donorsData.data || [])
      if (locationsData.length === 0) {
        const newLoc = await locationsAPI.create({ name: '默认位置', address: '默认地址' })
        setLocations([newLoc])
      } else {
        setLocations(locationsData)
      }
    } catch (e) {
      console.error(e)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadData()
  }, [])

  const handleCreateDonor = async (data) => {
    try {
      await donorsAPI.create(data)
      setShowForm(false)
      setMessage({ type: 'success', text: '捐献者登记成功' })
      loadData()
    } catch (e) {
      setMessage({ type: 'error', text: e.message })
    }
    setTimeout(() => setMessage(null), 3000)
  }

  const handleAssess = async (data) => {
    try {
      await organsAPI.assess(assessingOrgan.id, data)
      setAssessingOrgan(null)
      setMessage({ type: 'success', text: '评估完成' })
      if (selectedDonor) {
        const donor = await donorsAPI.get(selectedDonor.id)
        setSelectedDonor(donor)
      }
      loadData()
    } catch (e) {
      setMessage({ type: 'error', text: e.message })
    }
    setTimeout(() => setMessage(null), 3000)
  }

  const handleStartMatching = async (organId) => {
    try {
      await organsAPI.startMatching(organId)
      setMessage({ type: 'success', text: '匹配已启动' })
      if (selectedDonor) {
        const donor = await donorsAPI.get(selectedDonor.id)
        setSelectedDonor(donor)
      }
      loadData()
    } catch (e) {
      setMessage({ type: 'error', text: e.message })
    }
    setTimeout(() => setMessage(null), 3000)
  }

  return (
    <div>
      <div className="page-header">
        <h1>捐献者管理</h1>
        <p>管理器官捐献者信息和器官评估</p>
      </div>

      {message && (
        <div className={`alert alert-${message.type}`}>{message.text}</div>
      )}

      <div className="card">
        <div className="card-header">
          <h2>捐献者列表</h2>
          <button className="btn btn-primary" onClick={() => setShowForm(true)}>
            + 登记捐献者
          </button>
        </div>
        
        {loading ? (
          <div className="loading">加载中...</div>
        ) : donors.length === 0 ? (
          <div className="empty-state">
            <h3>暂无捐献者数据</h3>
            <p>点击上方按钮登记新的捐献者</p>
          </div>
        ) : (
          <div className="table-container">
            <table>
              <thead>
                <tr>
                  <th>编号</th>
                  <th>姓名</th>
                  <th>性别</th>
                  <th>年龄</th>
                  <th>血型</th>
                  <th>器官数量</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {donors.map(donor => (
                  <tr key={donor.id}>
                    <td>{donor.donor_no}</td>
                    <td>{donor.name}</td>
                    <td>{donor.gender}</td>
                    <td>{donor.age}岁</td>
                    <td>{donor.blood_type}型</td>
                    <td>{donor.organs?.length || 0}个</td>
                    <td>
                      <button className="btn btn-sm btn-primary" onClick={async () => {
                        const detail = await donorsAPI.get(donor.id)
                        setSelectedDonor(detail)
                      }}>详情</button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {showForm && (
        <DonorForm
          locations={locations}
          onSubmit={handleCreateDonor}
          onClose={() => setShowForm(false)}
        />
      )}

      {selectedDonor && (
        <DonorDetail
          donor={selectedDonor}
          onClose={() => setSelectedDonor(null)}
          onAssess={(organ) => setAssessingOrgan(organ)}
          onStartMatching={handleStartMatching}
        />
      )}

      {assessingOrgan && (
        <OrganAssessmentModal
          organ={assessingOrgan}
          onSubmit={handleAssess}
          onClose={() => setAssessingOrgan(null)}
        />
      )}
    </div>
  )
}
