import React, { useState, useEffect } from 'react'
import { api, STATUS_LABELS, SAMPLE_TYPE_LABELS } from '../utils/api.js'
import { useAuth } from '../context/AuthContext.jsx'

const STATUS_FLOW = ['received', 'testing', 'test_complete', 'report_generating', 'reported']

function SampleList() {
  const { role, unitID } = useAuth()
  const [samples, setSamples] = useState([])
  const [testItems, setTestItems] = useState([])
  const [units, setUnits] = useState([])
  const [loading, setLoading] = useState(false)
  const [showCreate, setShowCreate] = useState(false)
  const [selectedSample, setSelectedSample] = useState(null)
  const [message, setMessage] = useState(null)

  const [formData, setFormData] = useState({
    sample_type: 'blood',
    collection_date: new Date().toISOString().split('T')[0],
    unit_id: 'unit001',
    submitter: '',
    patient: { name: '', id_card: '', phone: '' },
    test_item_codes: [],
  })

  const fetchData = async () => {
    setLoading(true)
    try {
      const [samplesData, testData, unitsData] = await Promise.all([
        api.get('/samples?size=100', role, unitID),
        api.get('/catalog/test-items', role, unitID),
        api.get('/units', role, unitID),
      ])
      setSamples(samplesData.items || [])
      setTestItems(testData || [])
      setUnits(unitsData || [])
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [role, unitID])

  const showMsg = (type, text) => {
    setMessage({ type, text })
    setTimeout(() => setMessage(null), 3000)
  }

  const handleCreateSample = async (e) => {
    e.preventDefault()
    try {
      await api.post('/samples', formData, role, unitID)
      showMsg('success', '样本创建成功！')
      setShowCreate(false)
      setFormData({
        sample_type: 'blood',
        collection_date: new Date().toISOString().split('T')[0],
        unit_id: 'unit001',
        submitter: '',
        patient: { name: '', id_card: '', phone: '' },
        test_item_codes: [],
      })
      fetchData()
    } catch (err) {
      showMsg('error', '创建失败: ' + err.message)
    }
  }

  const handleStatusTransition = async (sampleId, currentStatus) => {
    const idx = STATUS_FLOW.indexOf(currentStatus)
    if (idx >= STATUS_FLOW.length - 1) return
    const nextStatus = STATUS_FLOW[idx + 1]

    try {
      await api.patch(`/samples/${sampleId}/status`, { new_status: nextStatus }, role, unitID)
      showMsg('success', '状态已更新！')
      fetchData()
    } catch (err) {
      showMsg('error', '更新失败: ' + err.message)
    }
  }

  const toggleTestItem = (code) => {
    setFormData(prev => ({
      ...prev,
      test_item_codes: prev.test_item_codes.includes(code)
        ? prev.test_item_codes.filter(c => c !== code)
        : [...prev.test_item_codes, code],
    }))
  }

  const getPriceInfo = () => {
    const selected = formData.test_item_codes
    const total = selected.reduce((sum, code) => {
      const item = testItems.find(t => t.code === code)
      return sum + (item?.price || 0)
    }, 0)
    let discount = 1
    if (selected.length >= 5) discount = 0.85
    else if (selected.length >= 3) discount = 0.9
    const final = Math.round(total * discount * 100) / 100
    return { total, discount, final, count: selected.length }
  }

  const priceInfo = getPriceInfo()

  const canAdvanceStatus = (sample) => {
    if (role === 'submitter') return false
    const idx = STATUS_FLOW.indexOf(sample.status)
    if (idx >= STATUS_FLOW.length - 1) return false
    if (sample.status === 'received') return role === 'admin' || role === 'technician'
    if (sample.status === 'test_complete') return role === 'admin' || role === 'reviewer'
    return role === 'admin'
  }

  return (
    <div className="space-y-6">
      {message && (
        <div className={`p-4 rounded-lg ${message.type === 'success' ? 'bg-green-50 text-green-800' : 'bg-red-50 text-red-800'}`}>
          {message.text}
        </div>
      )}

      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">样本管理</h2>
          <p className="text-gray-600 mt-1">共 {samples.length} 个样本</p>
        </div>
        {(role === 'admin' || role === 'technician') && (
          <button
            onClick={() => setShowCreate(true)}
            className="px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700 transition-colors"
          >
            + 接收新样本
          </button>
        )}
      </div>

      {showCreate && (
        <div className="bg-white p-6 rounded-lg shadow border border-gray-200">
          <h3 className="text-lg font-semibold mb-4">接收新样本</h3>
          <form onSubmit={handleCreateSample} className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">样本类型</label>
                <select
                  value={formData.sample_type}
                  onChange={(e) => setFormData({ ...formData, sample_type: e.target.value })}
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                >
                  {Object.entries(SAMPLE_TYPE_LABELS).map(([key, label]) => (
                    <option key={key} value={key}>{label}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">采集日期</label>
                <input
                  type="date"
                  value={formData.collection_date}
                  onChange={(e) => setFormData({ ...formData, collection_date: e.target.value })}
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">送检单位</label>
                <select
                  value={formData.unit_id}
                  onChange={(e) => setFormData({ ...formData, unit_id: e.target.value })}
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                >
                  {units.map(u => (
                    <option key={u.id} value={u.id}>{u.name}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">送检人</label>
                <input
                  type="text"
                  required
                  value={formData.submitter}
                  onChange={(e) => setFormData({ ...formData, submitter: e.target.value })}
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                  placeholder="送检人姓名"
                />
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">患者姓名</label>
                <input
                  type="text"
                  required
                  value={formData.patient.name}
                  onChange={(e) => setFormData({ ...formData, patient: { ...formData.patient, name: e.target.value } })}
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">身份证号</label>
                <input
                  type="text"
                  required
                  value={formData.patient.id_card}
                  onChange={(e) => setFormData({ ...formData, patient: { ...formData.patient, id_card: e.target.value } })}
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">联系电话</label>
                <input
                  type="text"
                  required
                  value={formData.patient.phone}
                  onChange={(e) => setFormData({ ...formData, patient: { ...formData.patient, phone: e.target.value } })}
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                />
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">检测项目</label>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                {testItems.map(item => (
                  <label
                    key={item.code}
                    className={`flex items-start p-3 border rounded-lg cursor-pointer transition-colors ${
                      formData.test_item_codes.includes(item.code)
                        ? 'border-primary-500 bg-primary-50'
                        : 'border-gray-200 hover:border-gray-300'
                    }`}
                  >
                    <input
                      type="checkbox"
                      checked={formData.test_item_codes.includes(item.code)}
                      onChange={() => toggleTestItem(item.code)}
                      className="mt-1 mr-3"
                    />
                    <div>
                      <div className="font-medium text-sm">{item.name}</div>
                      <div className="text-xs text-gray-500">{item.code} · ¥{item.price} · {item.report_days}工作日</div>
                    </div>
                  </label>
                ))}
              </div>
            </div>

            {priceInfo.count > 0 && (
              <div className="bg-gray-50 p-4 rounded-lg">
                <div className="flex items-center justify-between">
                  <div>
                    <span className="text-gray-600">已选 {priceInfo.count} 项</span>
                    {priceInfo.discount < 1 && (
                      <span className="ml-3 text-green-600 font-medium">
                        {priceInfo.count >= 5 ? '5项及以上 85折' : '3项及以上 9折'}
                      </span>
                    )}
                  </div>
                  <div className="text-right">
                    <div className="text-sm text-gray-500">
                      原价: <span className={priceInfo.discount < 1 ? 'line-through' : ''}>¥{priceInfo.total}</span>
                    </div>
                    <div className="text-xl font-bold text-primary-600">¥{priceInfo.final}</div>
                  </div>
                </div>
              </div>
            )}

            <div className="flex justify-end space-x-3">
              <button
                type="button"
                onClick={() => setShowCreate(false)}
                className="px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50"
              >
                取消
              </button>
              <button
                type="submit"
                disabled={formData.test_item_codes.length === 0}
                className="px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                确认接收
              </button>
            </div>
          </form>
        </div>
      )}

      {loading ? (
        <div className="text-center py-12 text-gray-500">加载中...</div>
      ) : (
        <div className="bg-white rounded-lg shadow overflow-hidden border border-gray-200">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">样本编号</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">类型</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">状态</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">送检单位</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">患者</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">检测项目</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">费用</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">操作</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {samples.length === 0 ? (
                <tr>
                  <td colSpan="8" className="px-4 py-12 text-center text-gray-500">
                    暂无样本数据
                  </td>
                </tr>
              ) : (
                samples.map(sample => (
                  <tr key={sample.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 whitespace-nowrap">
                      <span className="font-mono text-sm">{sample.sample_code}</span>
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-sm">
                      {SAMPLE_TYPE_LABELS[sample.sample_type]}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap">
                      <span className={`status-badge status-${sample.status}`}>
                        {STATUS_LABELS[sample.status]}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-sm">{sample.unit_name}</td>
                    <td className="px-4 py-3 text-sm">
                      <div>{sample.patient.name}</div>
                      <div className="text-gray-500 text-xs">{sample.patient.phone}</div>
                    </td>
                    <td className="px-4 py-3 text-sm">
                      <div className="flex flex-wrap gap-1">
                        {sample.test_items?.slice(0, 2).map(item => (
                          <span key={item.code} className="text-xs bg-gray-100 px-2 py-0.5 rounded">
                            {item.name}
                          </span>
                        ))}
                        {sample.test_items?.length > 2 && (
                          <span className="text-xs text-gray-500">+{sample.test_items.length - 2}</span>
                        )}
                      </div>
                    </td>
                    <td className="px-4 py-3 text-sm">
                      <div className="font-medium">¥{sample.final_price}</div>
                      {sample.discount < 1 && (
                        <div className="text-xs text-gray-500">折扣 {Math.round(sample.discount * 100)}%</div>
                      )}
                    </td>
                    <td className="px-4 py-3 text-sm space-x-2">
                      <button
                        onClick={() => setSelectedSample(sample)}
                        className="text-primary-600 hover:text-primary-800"
                      >
                        详情
                      </button>
                      {canAdvanceStatus(sample) && (
                        <button
                          onClick={() => handleStatusTransition(sample.id, sample.status)}
                          className="text-green-600 hover:text-green-800"
                        >
                          推进状态
                        </button>
                      )}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      )}

      {selectedSample && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg max-w-2xl w-full max-h-[90vh] overflow-y-auto">
            <div className="p-6 border-b border-gray-200 flex justify-between items-center">
              <h3 className="text-lg font-semibold">样本详情 - {selectedSample.sample_code}</h3>
              <button onClick={() => setSelectedSample(null)} className="text-gray-400 hover:text-gray-600 text-2xl">&times;</button>
            </div>
            <div className="p-6 space-y-6">
              <div>
                <h4 className="font-medium text-gray-700 mb-2">基本信息</h4>
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div><span className="text-gray-500">样本编号:</span> {selectedSample.sample_code}</div>
                  <div><span className="text-gray-500">样本类型:</span> {SAMPLE_TYPE_LABELS[selectedSample.sample_type]}</div>
                  <div><span className="text-gray-500">接收日期:</span> {selectedSample.received_date}</div>
                  <div><span className="text-gray-500">采集日期:</span> {selectedSample.collection_date}</div>
                  <div><span className="text-gray-500">送检单位:</span> {selectedSample.unit_name}</div>
                  <div><span className="text-gray-500">送检人:</span> {selectedSample.submitter}</div>
                  <div><span className="text-gray-500">当前状态:</span>
                    <span className={`ml-2 status-badge status-${selectedSample.status}`}>
                      {STATUS_LABELS[selectedSample.status]}
                    </span>
                  </div>
                </div>
              </div>

              <div>
                <h4 className="font-medium text-gray-700 mb-2">患者信息</h4>
                <div className="grid grid-cols-3 gap-4 text-sm">
                  <div><span className="text-gray-500">姓名:</span> {selectedSample.patient.name}</div>
                  <div><span className="text-gray-500">身份证:</span> {selectedSample.patient.id_card}</div>
                  <div><span className="text-gray-500">电话:</span> {selectedSample.patient.phone}</div>
                </div>
              </div>

              <div>
                <h4 className="font-medium text-gray-700 mb-2">检测项目</h4>
                <div className="space-y-2">
                  {selectedSample.test_items?.map((item, i) => (
                    <div key={i} className="border border-gray-200 rounded-lg p-3">
                      <div className="flex justify-between items-start">
                        <div>
                          <div className="font-medium">{item.name} <span className="text-gray-500 text-sm">({item.code})</span></div>
                          <div className="text-sm text-gray-500">{item.description}</div>
                        </div>
                        <div className="text-right">
                          <div className="text-sm">¥{item.price}</div>
                          {item.completed ? (
                            <span className="status-badge status-reported">已完成</span>
                          ) : (
                            <span className="status-badge status-testing">待检测</span>
                          )}
                        </div>
                      </div>
                      {item.completed && (
                        <div className="mt-2 pt-2 border-t border-gray-100 text-sm space-y-1">
                          <div><span className="text-gray-500">结果摘要:</span> {item.result_summary}</div>
                          <div><span className="text-gray-500">风险等级:</span> {item.risk_level}</div>
                          <div><span className="text-gray-500">结论:</span> {item.risk_conclusion}</div>
                          <div><span className="text-gray-500">完成时间:</span> {item.completed_at}</div>
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              </div>

              <div className="bg-gray-50 p-4 rounded-lg">
                <div className="flex justify-between items-center">
                  <div>
                    <span className="text-gray-600">总金额: ¥{selectedSample.total_price}</span>
                    {selectedSample.discount < 1 && (
                      <span className="ml-3 text-green-600">折扣 {Math.round(selectedSample.discount * 100)}%</span>
                    )}
                  </div>
                  <div className="text-2xl font-bold text-primary-600">¥{selectedSample.final_price}</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default SampleList
