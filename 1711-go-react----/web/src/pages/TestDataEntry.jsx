import React, { useState, useEffect } from 'react'
import { api, STATUS_LABELS, SAMPLE_TYPE_LABELS } from '../utils/api.js'
import { useAuth } from '../context/AuthContext.jsx'

const RISK_OPTIONS = [
  { value: '高风险', label: '高风险' },
  { value: '中风险', label: '中风险' },
  { value: '低风险', label: '低风险' },
]

function TestDataEntry() {
  const { role, unitID } = useAuth()
  const [samples, setSamples] = useState([])
  const [loading, setLoading] = useState(false)
  const [selectedSample, setSelectedSample] = useState(null)
  const [editingItem, setEditingItem] = useState(null)
  const [message, setMessage] = useState(null)
  const [formData, setFormData] = useState({
    raw_data: '',
    result_summary: '',
    risk_level: '低风险',
    risk_conclusion: '',
  })

  const fetchData = async () => {
    setLoading(true)
    try {
      const data = await api.get('/samples?size=100', role, unitID)
      setSamples(data.items || [])
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

  const testableSamples = samples.filter(s => 
    s.status === 'testing' || 
    (s.test_items && s.test_items.some(item => !item.completed))
  )

  const handleSubmit = async (e) => {
    e.preventDefault()
    try {
      await api.patch(
        `/samples/${selectedSample.id}/test-items/${editingItem.code}`,
        formData,
        role,
        unitID
      )
      showMsg('success', '检测结果已保存！')
      setEditingItem(null)
      fetchData()
      setSelectedSample(null)
    } catch (err) {
      showMsg('error', '保存失败: ' + err.message)
    }
  }

  const getRiskClass = (level) => {
    if (level === '高风险') return 'risk-high'
    if (level === '中风险') return 'risk-medium'
    return 'risk-low'
  }

  return (
    <div className="space-y-6">
      {message && (
        <div className={`p-4 rounded-lg ${message.type === 'success' ? 'bg-green-50 text-green-800' : 'bg-red-50 text-red-800'}`}>
          {message.text}
        </div>
      )}

      <div>
        <h2 className="text-2xl font-bold text-gray-900">检测数据录入</h2>
        <p className="text-gray-600 mt-1">为样本录入检测结果数据</p>
      </div>

      {loading ? (
        <div className="text-center py-12 text-gray-500">加载中...</div>
      ) : testableSamples.length === 0 ? (
        <div className="bg-white rounded-lg shadow border border-gray-200 p-12 text-center">
          <div className="text-6xl mb-4">🔬</div>
          <h3 className="text-lg font-medium text-gray-900 mb-2">暂无待检测样本</h3>
          <p className="text-gray-500">所有样本的检测项目均已完成</p>
        </div>
      ) : (
        <div className="space-y-4">
          {testableSamples.map(sample => (
            <div key={sample.id} className="bg-white rounded-lg shadow border border-gray-200 overflow-hidden">
              <div className="p-4 border-b border-gray-100 flex justify-between items-center">
                <div className="flex items-center space-x-4">
                  <span className="font-mono font-bold text-lg">{sample.sample_code}</span>
                  <span className="status-badge status-testing">
                    {STATUS_LABELS[sample.status]}
                  </span>
                  <span className="text-gray-500 text-sm">
                    {SAMPLE_TYPE_LABELS[sample.sample_type]}
                  </span>
                </div>
                <button
                  onClick={() => setSelectedSample(selectedSample?.id === sample.id ? null : sample)}
                  className="text-primary-600 hover:text-primary-800"
                >
                  {selectedSample?.id === sample.id ? '收起' : '展开'}
                </button>
              </div>

              {selectedSample?.id === sample.id && (
                <div className="p-4 space-y-3">
                  <div className="text-sm text-gray-600 mb-3">
                    送检单位: {sample.unit_name} | 送检人: {sample.submitter}
                  </div>
                  {sample.test_items?.map((item, i) => (
                    <div key={i} className={`border rounded-lg p-4 ${
                      item.completed ? 'bg-gray-50 border-gray-200' : 'border-primary-200 bg-primary-50'
                    }`}>
                      <div className="flex justify-between items-start">
                        <div className="flex-1">
                          <div className="flex items-center space-x-2">
                            <span className="font-medium">{item.name}</span>
                            <span className="text-xs text-gray-500">({item.code})</span>
                            {item.completed ? (
                              <span className="status-badge status-reported">已完成</span>
                            ) : (
                              <span className="status-badge status-testing">待录入</span>
                            )}
                          </div>
                          <div className="text-sm text-gray-500 mt-1">{item.description}</div>
                          <div className="text-sm text-gray-500">¥{item.price} · {item.report_days}工作日</div>
                          
                          {item.completed && (
                            <div className="mt-3 pt-3 border-t border-gray-200 text-sm space-y-1">
                              <div><span className="text-gray-500">结果摘要:</span> {item.result_summary}</div>
                              <div className="flex items-center space-x-2">
                                <span className="text-gray-500">风险等级:</span>
                                <span className={`status-badge ${getRiskClass(item.risk_level)}`}>
                                  {item.risk_level}
                                </span>
                              </div>
                              <div><span className="text-gray-500">风险结论:</span> {item.risk_conclusion}</div>
                              <div><span className="text-gray-500">完成时间:</span> {item.completed_at}</div>
                            </div>
                          )}
                        </div>
                        
                        {!item.completed && (
                          <button
                            onClick={() => {
                              setEditingItem(item)
                              setFormData({
                                raw_data: '',
                                result_summary: '',
                                risk_level: '低风险',
                                risk_conclusion: '',
                              })
                            }}
                            className="ml-4 px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700 text-sm"
                          >
                            录入结果
                          </button>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {editingItem && selectedSample && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg max-w-lg w-full">
            <div className="p-6 border-b border-gray-200">
              <h3 className="text-lg font-semibold">
                录入检测结果 - {editingItem.name}
              </h3>
              <p className="text-sm text-gray-500 mt-1">
                样本: {selectedSample.sample_code}
              </p>
            </div>
            <form onSubmit={handleSubmit} className="p-6 space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  原始检测数据 (基因位点数值等)
                </label>
                <textarea
                  value={formData.raw_data}
                  onChange={(e) => setFormData({ ...formData, raw_data: e.target.value })}
                  rows={3}
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm font-mono"
                  placeholder="例如: BRCA1 c.5266dupC (杂合); TP53 阴性"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  结果摘要
                </label>
                <textarea
                  value={formData.result_summary}
                  onChange={(e) => setFormData({ ...formData, result_summary: e.target.value })}
                  rows={2}
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
                  placeholder="简要描述检测结果"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  风险等级
                </label>
                <select
                  value={formData.risk_level}
                  onChange={(e) => setFormData({ ...formData, risk_level: e.target.value })}
                  className="w-full border border-gray-300 rounded-md px-3 py-2"
                >
                  {RISK_OPTIONS.map(opt => (
                    <option key={opt.value} value={opt.value}>{opt.label}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  风险结论
                </label>
                <textarea
                  value={formData.risk_conclusion}
                  onChange={(e) => setFormData({ ...formData, risk_conclusion: e.target.value })}
                  rows={2}
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
                  placeholder="详细的风险评估结论"
                  required
                />
              </div>
              <div className="flex justify-end space-x-3 pt-4">
                <button
                  type="button"
                  onClick={() => setEditingItem(null)}
                  className="px-4 py-2 border border-gray-300 rounded-md hover:bg-gray-50"
                >
                  取消
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700"
                >
                  保存结果
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}

export default TestDataEntry
