import React, { useState, useEffect } from 'react'
import { api, STATUS_LABELS, REPORT_STATUS_LABELS } from '../utils/api.js'
import { useAuth } from '../context/AuthContext.jsx'

function SubmitterQuery() {
  const { role, unitID } = useAuth()
  const [samples, setSamples] = useState([])
  const [reports, setReports] = useState([])
  const [loading, setLoading] = useState(false)
  const [activeTab, setActiveTab] = useState('samples')
  const [selectedReport, setSelectedReport] = useState(null)

  const fetchData = async () => {
    setLoading(true)
    try {
      const [samplesData, reportsData] = await Promise.all([
        api.get('/samples?size=100', role, unitID),
        api.get('/reports?size=100', role, unitID),
      ])
      setSamples(samplesData.items || [])
      setReports((reportsData.items || []).filter(r => r.status === 'published'))
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [role, unitID])

  const getRiskClass = (level) => {
    if (level === '高风险') return 'risk-high'
    if (level === '中风险') return 'risk-medium'
    return 'risk-low'
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-gray-900">送检单位查询</h2>
        <p className="text-gray-600 mt-1">查看本单位送检样本的状态和已发布报告</p>
      </div>

      <div className="bg-white rounded-lg shadow border border-gray-200">
        <div className="border-b border-gray-200">
          <nav className="flex">
            <button
              onClick={() => setActiveTab('samples')}
              className={`px-6 py-3 text-sm font-medium border-b-2 ${
                activeTab === 'samples'
                  ? 'border-primary-500 text-primary-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700'
              }`}
            >
              样本列表 ({samples.length})
            </button>
            <button
              onClick={() => setActiveTab('reports')}
              className={`px-6 py-3 text-sm font-medium border-b-2 ${
                activeTab === 'reports'
                  ? 'border-primary-500 text-primary-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700'
              }`}
            >
              已发布报告 ({reports.length})
            </button>
          </nav>
        </div>

        <div className="p-4">
          {loading ? (
            <div className="text-center py-12 text-gray-500">加载中...</div>
          ) : activeTab === 'samples' ? (
            samples.length === 0 ? (
              <div className="text-center py-12 text-gray-500">
                <div className="text-5xl mb-4">🧪</div>
                <p>暂无送检样本</p>
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">样本编号</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">患者姓名</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">送检人</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">接收日期</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">检测项目</th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">状态</th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {samples.map(sample => (
                      <tr key={sample.id} className="hover:bg-gray-50">
                        <td className="px-4 py-3 whitespace-nowrap font-mono text-sm">
                          {sample.sample_code}
                        </td>
                        <td className="px-4 py-3 whitespace-nowrap text-sm">
                          {sample.patient.name}
                        </td>
                        <td className="px-4 py-3 whitespace-nowrap text-sm">
                          {sample.submitter}
                        </td>
                        <td className="px-4 py-3 whitespace-nowrap text-sm">
                          {sample.received_date}
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
                        <td className="px-4 py-3 whitespace-nowrap">
                          <span className={`status-badge status-${sample.status}`}>
                            {STATUS_LABELS[sample.status]}
                          </span>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )
          ) : (
            reports.length === 0 ? (
              <div className="text-center py-12 text-gray-500">
                <div className="text-5xl mb-4">📄</div>
                <p>暂无已发布的报告</p>
              </div>
            ) : (
              <div className="space-y-3">
                {reports.map(report => (
                  <div
                    key={report.id}
                    className="border border-gray-200 rounded-lg overflow-hidden hover:border-primary-300 cursor-pointer"
                    onClick={() => setSelectedReport(selectedReport?.id === report.id ? null : report)}
                  >
                    <div className="p-4 flex justify-between items-center">
                      <div>
                        <div className="flex items-center space-x-3">
                          <span className="font-mono font-bold">{report.sample_code}</span>
                          <span className="text-sm text-gray-500">v{report.version}</span>
                          {report.is_supplementary && (
                            <span className="status-badge bg-orange-100 text-orange-800">补充报告</span>
                          )}
                          <span className="status-badge status-reported">已发布</span>
                        </div>
                        <div className="text-sm text-gray-500 mt-1">
                          生成时间: {report.content.generated_at}
                        </div>
                      </div>
                      <div className="text-right">
                        <div className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-medium ${getRiskClass(report.content.risk_assessment.overall_level)}`}>
                          {report.content.risk_assessment.overall_level}
                        </div>
                        <div className="text-sm text-gray-500 mt-1">
                          {selectedReport?.id === report.id ? '收起' : '查看详情'}
                        </div>
                      </div>
                    </div>

                    {selectedReport?.id === report.id && (
                      <div className="border-t border-gray-200 bg-gray-50 p-4">
                        <div className="bg-white rounded-lg p-4 space-y-4">
                          <div>
                            <h4 className="font-medium text-gray-700 mb-2">样本信息</h4>
                            <div className="grid grid-cols-2 md:grid-cols-3 gap-2 text-sm">
                              <div><span className="text-gray-500">样本编号:</span> {report.content.sample_info.sample_code}</div>
                              <div><span className="text-gray-500">送检单位:</span> {report.content.sample_info.unit_name}</div>
                              <div><span className="text-gray-500">送检人:</span> {report.content.sample_info.submitter}</div>
                              <div><span className="text-gray-500">接收日期:</span> {report.content.sample_info.received_date}</div>
                            </div>
                          </div>

                          <div>
                            <h4 className="font-medium text-gray-700 mb-2">检测项目结果</h4>
                            <div className="space-y-2">
                              {report.content.test_items.map((item, i) => (
                                <div key={i} className="border border-gray-200 rounded p-3">
                                  <div className="flex justify-between items-start">
                                    <div>
                                      <div className="font-medium">{item.name} <span className="text-gray-500 text-sm">({item.code})</span></div>
                                      <div className="text-sm text-gray-600 mt-1">{item.result_summary}</div>
                                    </div>
                                    <span className={`status-badge ${getRiskClass(item.risk_level)}`}>
                                      {item.risk_level}
                                    </span>
                                  </div>
                                </div>
                              ))}
                            </div>
                          </div>

                          <div className="grid md:grid-cols-2 gap-4">
                            <div>
                              <h4 className="font-medium text-gray-700 mb-2">综合风险评估</h4>
                              <div className="space-y-2">
                                <div className="flex items-center space-x-2">
                                  <span className="text-gray-500 text-sm">总体风险:</span>
                                  <span className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-medium ${getRiskClass(report.content.risk_assessment.overall_level)}`}>
                                    {report.content.risk_assessment.overall_level}
                                  </span>
                                </div>
                                {report.content.risk_assessment.high_risks?.length > 0 && (
                                  <div>
                                    <div className="text-sm text-gray-500 mb-1">高风险提示:</div>
                                    <ul className="text-sm text-red-700 list-disc list-inside space-y-1">
                                      {report.content.risk_assessment.high_risks.map((r, i) => (
                                        <li key={i}>{r}</li>
                                      ))}
                                    </ul>
                                  </div>
                                )}
                              </div>
                            </div>

                            <div>
                              <h4 className="font-medium text-gray-700 mb-2">建议</h4>
                              <p className="text-sm text-gray-700 leading-relaxed">
                                {report.content.recommendations}
                              </p>
                            </div>
                          </div>

                          {report.review_comments?.length > 0 && (
                            <div>
                              <h4 className="font-medium text-gray-700 mb-2">审核记录</h4>
                              <div className="space-y-2">
                                {report.review_comments.map((c, i) => (
                                  <div key={i} className="bg-yellow-50 border border-yellow-200 rounded p-2 text-sm">
                                    <div className="flex justify-between text-xs text-gray-500">
                                      <span>{c.reviewer}</span>
                                      <span>{c.time}</span>
                                    </div>
                                    <p className="mt-1">{c.comment}</p>
                                  </div>
                                ))}
                              </div>
                            </div>
                          )}
                        </div>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )
          )}
        </div>
      </div>

      <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
        <div className="flex items-start space-x-3">
          <div className="text-2xl">ℹ️</div>
          <div>
            <h4 className="font-medium text-yellow-800">隐私保护说明</h4>
            <p className="text-sm text-yellow-700 mt-1">
              作为送检单位，您可以查看本单位送检样本的基本信息和已发布报告的结论。
              原始基因检测数据已脱敏处理，仅展示"检出"或"未检出"和风险评估结论。
              患者身份信息也已进行部分脱敏处理。
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}

export default SubmitterQuery
