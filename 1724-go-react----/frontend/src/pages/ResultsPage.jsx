import React, { useState, useEffect } from 'react'
import axios from 'axios'

const qualityLabels = {
  'insufficient': '数据不足',
  'reference': '参考',
  'valid': '有效'
}

const courseTypeLabels = {
  'theory': '理论课',
  'experiment': '实验课',
  'sport': '体育课'
}

function ResultsPage() {
  const [tasks, setTasks] = useState([])
  const [selectedTask, setSelectedTask] = useState(null)
  const [results, setResults] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [selectedResult, setSelectedResult] = useState(null)
  const [comments, setComments] = useState([])

  useEffect(() => {
    fetchTasks()
  }, [])

  const fetchTasks = async () => {
    try {
      const response = await axios.get('/api/tasks')
      setTasks(response.data)
    } catch (err) {
      setError('加载任务列表失败')
    }
  }

  const handleTaskSelect = async (taskId) => {
    try {
      setSelectedTask(taskId)
      setLoading(true)
      setError('')
      setSelectedResult(null)
      setComments([])

      const response = await axios.get(`/api/results/task/${taskId}`)
      setResults(response.data)
    } catch (err) {
      setError('加载统计结果失败')
    } finally {
      setLoading(false)
    }
  }

  const handleCalculate = async (taskId) => {
    try {
      setLoading(true)
      setError('')
      
      await axios.post(`/api/results/task/${taskId}/calculate`)
      handleTaskSelect(taskId)
    } catch (err) {
      setError('计算统计失败')
      setLoading(false)
    }
  }

  const handleExport = async (taskId) => {
    try {
      const response = await axios.get(`/api/results/export/${taskId}`, {
        responseType: 'blob'
      })
      
      const url = window.URL.createObjectURL(new Blob([response.data]))
      const link = document.createElement('a')
      link.href = url
      link.setAttribute('download', `statistics_task_${taskId}.json`)
      document.body.appendChild(link)
      link.click()
    } catch (err) {
      setError('导出失败')
    }
  }

  const handleViewComments = async (resultId) => {
    try {
      const response = await axios.get(`/api/results/${resultId}/comments`)
      setComments(response.data)
      setSelectedResult(resultId)
    } catch (err) {
      setError('加载评语失败')
    }
  }

  const getQualityClass = (quality) => {
    if (quality === 'insufficient') return 'quality-insufficient'
    if (quality === 'reference') return 'quality-reference'
    return 'quality-valid'
  }

  const validResults = results.filter(r => r.data_quality !== 'insufficient')
  const avgScore = validResults.length > 0
    ? (validResults.reduce((sum, r) => sum + r.overall_score, 0) / validResults.length).toFixed(2)
    : 0

  return (
    <div>
      <div className="card">
        <h2>结果统计</h2>
      </div>

      {error && <div className="error-message">{error}</div>}

      <div className="card">
        <h3>选择评估任务</h3>
        <div className="form-group">
          <select
            value={selectedTask || ''}
            onChange={(e) => handleTaskSelect(e.target.value ? parseInt(e.target.value) : null)}
          >
            <option value="">-- 请选择任务 --</option>
            {tasks.map(task => (
              <option key={task.id} value={task.id}>
                {task.semester} - ({task.status === 'completed' ? '已完成' : '进行中'})
              </option>
            ))}
          </select>
        </div>

        {selectedTask && results.length === 0 && !loading && (
          <div style={{ marginTop: '1rem' }}>
            <p>该任务暂无统计数据，请先计算统计结果。</p>
            <button 
              className="btn btn-primary" 
              style={{ marginTop: '1rem' }}
              onClick={() => handleCalculate(selectedTask)}
            >
              计算统计结果
            </button>
          </div>
        )}
      </div>

      {selectedTask && results.length > 0 && (
        <>
          <div className="stat-grid">
            <div className="stat-card">
              <div className="stat-value">{results.length}</div>
              <div className="stat-label">参评课程数</div>
            </div>
            <div className="stat-card">
              <div className="stat-value">{avgScore}</div>
              <div className="stat-label">平均综合得分</div>
            </div>
            <div className="stat-card">
              <div className="stat-value">{validResults.length}</div>
              <div className="stat-label">有效课程数</div>
            </div>
            <div className="stat-card">
              <div className="stat-value">
                {results.filter(r => r.is_low_score).length}
              </div>
              <div className="stat-label">需改进课程</div>
            </div>
          </div>

          <div className="card">
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <h3>统计详情</h3>
              <button 
                className="btn btn-primary"
                onClick={() => handleExport(selectedTask)}
              >
                导出数据
              </button>
            </div>

            {loading ? (
              <div className="loading">加载中...</div>
            ) : (
              <table>
                <thead>
                  <tr>
                    <th>排名</th>
                    <th>课程ID</th>
                    <th>课程类型</th>
                    <th>提交率</th>
                    <th>数据质量</th>
                    <th>通用均分</th>
                    <th>专项均分</th>
                    <th>综合得分</th>
                    <th>操作</th>
                  </tr>
                </thead>
                <tbody>
                  {results.map((result, index) => (
                    <tr key={result.id} style={result.is_low_score ? { backgroundColor: '#fff5f5' } : {}}>
                      <td>
                        {result.data_quality !== 'insufficient' ? index + 1 : '-'}
                      </td>
                      <td>{result.course_id}</td>
                      <td>{courseTypeLabels[result.course_type] || result.course_type}</td>
                      <td>{result.submission_rate.toFixed(1)}%</td>
                      <td>
                        <span className={`status-badge ${getQualityClass(result.data_quality)}`}>
                          {qualityLabels[result.data_quality]}
                        </span>
                      </td>
                      <td>{result.general_average.toFixed(2)}</td>
                      <td>{result.special_score.toFixed(2)}</td>
                      <td>
                        <strong style={{ color: result.is_low_score ? '#c53030' : '#38a169' }}>
                          {result.overall_score.toFixed(2)}
                        </strong>
                        {result.is_low_score && <span style={{ color: '#c53030', marginLeft: '0.5rem' }}>⚠️</span>}
                      </td>
                      <td>
                        <button
                          className="btn btn-secondary"
                          onClick={() => handleViewComments(result.id)}
                        >
                          查看评语
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>

          {selectedResult && (
            <div className="card">
              <h3>学生评语 (已过滤敏感词)</h3>
              {comments.length === 0 ? (
                <p>暂无学生评语</p>
              ) : (
                <div style={{ maxHeight: '300px', overflowY: 'auto' }}>
                  {comments.map((comment, index) => (
                    <div 
                      key={index} 
                      style={{ 
                        padding: '1rem', 
                        borderBottom: '1px solid #e2e8f0',
                        lineHeight: '1.6'
                      }}
                    >
                      {comment || '(无内容)'}
                    </div>
                  ))}
                </div>
              )}
              <button 
                className="btn btn-secondary" 
                style={{ marginTop: '1rem' }}
                onClick={() => { setSelectedResult(null); setComments([]); }}
              >
                关闭
              </button>
            </div>
          )}
        </>
      )}
    </div>
  )
}

export default ResultsPage
