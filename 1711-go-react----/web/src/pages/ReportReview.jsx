import React, { useState, useEffect } from 'react'
import { api, REPORT_STATUS_LABELS, APPROVAL_LABELS } from '../utils/api.js'
import { useAuth } from '../context/AuthContext.jsx'

function ReportReview() {
  const { role, unitID } = useAuth()
  const [reports, setReports] = useState([])
  const [loading, setLoading] = useState(false)
  const [selectedReport, setSelectedReport] = useState(null)
  const [message, setMessage] = useState(null)
  const [reviewComment, setReviewComment] = useState('')

  const fetchData = async () => {
    setLoading(true)
    try {
      const data = await api.get('/reports?size=100', role, unitID)
      setReports(data.items || [])
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

  const handleSubmit = async (reportId) => {
    try {
      await api.post(`/reports/${reportId}/submit`, null, role, unitID)
      showMsg('success', '报告已提交审核！')
      fetchData()
      setSelectedReport(null)
    } catch (err) {
      showMsg('error', '提交失败: ' + err.message)
    }
  }

  const handleApprove = async (reportId) => {
    try {
      await api.post(`/reports/${reportId}/review`, { 
        action: 'approve', 
        comment: reviewComment 
      }, role, unitID)
      showMsg('success', '审核通过！')
      setReviewComment('')
      fetchData()
      setSelectedReport(null)
    } catch (err) {
      showMsg('error', '操作失败: ' + err.message)
    }
  }

  const handleReject = async (reportId) => {
    if (!reviewComment.trim()) {
      showMsg('error', '驳回时请填写审核意见')
      return
    }
    try {
      await api.post(`/reports/${reportId}/review`, { 
        action: 'reject', 
        comment: reviewComment 
      }, role, unitID)
      showMsg('success', '报告已驳回！')
      setReviewComment('')
      fetchData()
      setSelectedReport(null)
    } catch (err) {
      showMsg('error', '操作失败: ' + err.message)
    }
  }

  const handleSupplement = async (reportId) => {
    try {
      const result = await api.post(`/reports/${reportId}/supplement`, null, role, unitID)
      showMsg('success', `补充报告已创建，版本号: ${result.version}`)
      fetchData()
    } catch (err) {
      showMsg('error', '创建失败: ' + err.message)
    }
  }

  const getApprovalStage = (status) => {
    const stages = {
      first_review: '一审',
      second_review: '二审',
      final_review: '终审',
    }
    return stages[status] || status
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
        <h2 className="text-2xl font-bold text-gray-900">报告审核</h2>
        <p className="text-gray-600 mt-1">审核报告内容并管理报告发布</p>
      </div>

      {loading ? (
        <div className="text-center py-12 text-gray-500">加载中...</div>
      ) : reports.length === 0 ? (
        <div className="bg-white rounded-lg shadow border border-gray-200 p-12 text-center">
          <div className="text-6xl mb-4">📄</div>
          <h3 className="text-lg font-medium text-gray-900 mb-2">暂无报告</h3>
          <p className="text-gray-500">当前没有待审核的报告</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {reports.map(report => (
            <div key={report.id} className="bg-white rounded-lg shadow border border-gray-200 overflow-hidden">
              <div className="p-4 border-b border-gray-100">
                <div className="flex justify-between items-start">
                  <div>
                    <div className="flex items-center space-x-2">
                      <span className="font-mono font-bold">{report.sample_code}</span>
                      <span className="text-sm text-gray-500">v{report.version}</span>
                      {report.is_supplementary && (
                        <span className="status-badge bg-orange-100 text-orange-800">补充报告</span>
                      )}
                    </div>
                    <div className="flex items-center space-x-3 mt-2">
                      <span className={`status-badge ${
                        report.status === 'published' ? 'status-reported' :
                        report.status === 'pending' ? 'status-testing' : 'status-received'
                      }`}>
                        {REPORT_STATUS_LABELS[report.status]}
                      </span>
                      {report.status !== 'draft' && (
                        <span className={`status-badge ${
                          report.approval_status === 'approved' ? 'status-reported' :
                          report.approval_status === 'rejected' ? 'bg-red-100 text-red-800' : 
                          'status-testing'
                        }`}>
                          {APPROVAL_LABELS[report.approval_status]}
                        </span>
                      )}
                    </div>
                  </div>
                  <button
                    onClick={() => setSelectedReport(selectedReport?.id === report.id ? null : report)}
                    className="text-primary-600 hover:text-primary-800 text-sm"
                  >
                    {selectedReport?.id === report.id ? '收起' : '查看'}
                  </button>
                </div>
              </div>

              {selectedReport?.id === report.id && (
                <div className="p-4 space-y-4">
                  <div className="border border-gray-200 rounded-lg p-4 bg-gray-50">
                    <h4 className="font-medium text-gray-700 mb-3">报告内容摘要</h4>
                    
                    <div className="space-y-3">
                      <div className="grid grid-cols-2 gap-2 text-sm">
                        <div><span className="text-gray-500">样本编号:</span> {report.content.sample_info.sample_code}</div>
                        <div><span className="text-gray-500">送检单位:</span> {report.content.sample_info.unit_name}</div>
                        <div><span className="text-gray-500">送检人:</span> {report.content.sample_info.submitter}</div>
                        <div><span className="text-gray-500">生成时间:</span> {report.content.generated_at}</div>
                      </div>

                      <div>
                        <div className="text-sm text-gray-500 mb-2">检测项目:</div>
                        <div className="space-y-2">
                          {report.content.test_items.map((item, i) => (
                            <div key={i} className="flex justify-between items-center text-sm bg-white p-2 rounded">
                              <span>{item.name}</span>
                              <div className="flex items-center space-x-2">
                                <span className={`status-badge ${getRiskClass(item.risk_level)}`}>
                                  {item.risk_level}
                                </span>
                              </div>
                            </div>
                          ))}
                        </div>
                      </div>

                      <div>
                        <div className="text-sm text-gray-500 mb-1">综合风险评估:</div>
                        <div className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-medium ${getRiskClass(report.content.risk_assessment.overall_level)}`}>
                          {report.content.risk_assessment.overall_level}
                        </div>
                      </div>

                      {report.content.risk_assessment.high_risks?.length > 0 && (
                        <div>
                          <div className="text-sm text-gray-500 mb-1">高风险项目:</div>
                          <ul className="text-sm text-red-700 list-disc list-inside">
                            {report.content.risk_assessment.high_risks.map((r, i) => (
                              <li key={i}>{r}</li>
                            ))}
                          </ul>
                        </div>
                      )}

                      <div>
                        <div className="text-sm text-gray-500 mb-1">建议:</div>
                        <p className="text-sm text-gray-700">{report.content.recommendations}</p>
                      </div>
                    </div>
                  </div>

                  {report.review_comments?.length > 0 && (
                    <div>
                      <h4 className="font-medium text-gray-700 mb-2 text-sm">审核记录</h4>
                      <div className="space-y-2">
                        {report.review_comments.map((c, i) => (
                          <div key={i} className="bg-yellow-50 border border-yellow-200 rounded p-2 text-sm">
                            <div className="flex justify-between text-xs text-gray-500">
                              <span>{c.reviewer} - {APPROVAL_LABELS[c.stage] || c.stage}</span>
                              <span>{c.time}</span>
                            </div>
                            <p className="mt-1">{c.comment}</p>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  {report.status !== 'published' && (
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">审核意见 (驳回必填)</label>
                      <textarea
                        value={reviewComment}
                        onChange={(e) => setReviewComment(e.target.value)}
                        rows={2}
                        className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
                        placeholder="请输入审核意见..."
                      />
                    </div>
                  )}

                  <div className="flex flex-wrap gap-2 pt-2">
                    {report.status === 'draft' && (
                      <button
                        onClick={() => handleSubmit(report.id)}
                        className="px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700 text-sm"
                      >
                        提交审核
                      </button>
                    )}

                    {report.status === 'pending' && (
                      <>
                        <button
                          onClick={() => handleApprove(report.id)}
                          className="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 text-sm"
                        >
                          通过 ({getApprovalStage(report.approval_status)})
                        </button>
                        <button
                          onClick={() => handleReject(report.id)}
                          className="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 text-sm"
                        >
                          驳回
                        </button>
                      </>
                    )}

                    {report.status === 'published' && (
                      <button
                        onClick={() => handleSupplement(report.id)}
                        className="px-4 py-2 bg-orange-600 text-white rounded-md hover:bg-orange-700 text-sm"
                      >
                        创建补充报告
                      </button>
                    )}
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

export default ReportReview
