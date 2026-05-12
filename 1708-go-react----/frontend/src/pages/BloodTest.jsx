import { useState, useEffect } from 'react'
import axios from 'axios'

const testItems = [
  { key: 'abo_front', label: 'ABO血型正定型' },
  { key: 'abo_back', label: 'ABO血型反定型' },
  { key: 'rhd', label: 'RhD血型' },
  { key: 'alt', label: '谷丙转氨酶' },
  { key: 'hbsag', label: '乙肝表面抗原' },
  { key: 'anti_hcv', label: '丙肝抗体' },
  { key: 'anti_hiv', label: '艾滋病抗体' },
  { key: 'anti_syphilis', label: '梅毒抗体' },
  { key: 'nat', label: '核酸检测' }
]

const resultOptions = [
  { value: 'Negative', label: '阴性' },
  { value: 'Positive', label: '阳性' },
  { value: 'Invalid', label: '无效' }
]

export default function BloodTest() {
  const [collections, setCollections] = useState([])
  const [selectedCollection, setSelectedCollection] = useState(null)
  const [testData, setTestData] = useState(null)
  const [results, setResults] = useState({ 1: {}, 2: {}, 3: {} })
  const [message, setMessage] = useState(null)
  const [error, setError] = useState(null)
  const [vendor, setVendor] = useState('')
  const [operator, setOperator] = useState('')

  useEffect(() => {
    fetchCollections()
  }, [])

  const fetchCollections = async () => {
    try {
      const res = await axios.get('/api/collections')
      const pending = (res.data || []).filter(c => 
        c.status === 'Pending_Test' || c.status === 'Testing'
      )
      setCollections(pending)
    } catch (err) {
      console.error('Failed to fetch collections:', err)
    }
  }

  const selectCollection = async (id) => {
    try {
      setSelectedCollection(id)
      const res = await axios.get(`/api/tests/${id}`)
      setTestData(res.data)
      
      const records = res.data.test_records || []
      const newResults = { 1: {}, 2: {}, 3: {} }
      
      records.forEach(record => {
        const round = record.test_round
        newResults[round] = {
          abo_front: record.abo_front,
          abo_back: record.abo_back,
          rhd: record.rhd,
          alt: record.alt,
          hbsag: record.hbsag,
          anti_hcv: record.anti_hcv,
          anti_hiv: record.anti_hiv,
          anti_syphilis: record.anti_syphilis,
          nat: record.nat
        }
      })
      setResults(newResults)
    } catch (err) {
      console.error('Failed to fetch test data:', err)
    }
  }

  const handleResultChange = (round, key, value) => {
    setResults(prev => ({
      ...prev,
      [round]: { ...prev[round], [key]: value }
    }))
  }

  const submitTest = async (round) => {
    setError(null)
    setMessage(null)

    const currentResults = results[round]
    const incomplete = testItems.some(item => !currentResults[item.key])
    if (incomplete) {
      setError('请完成本轮所有检测项目结果')
      return
    }

    try {
      await axios.post('/api/tests', {
        collection_id: selectedCollection,
        test_round: round,
        reagent_vendor: vendor,
        operator_id: operator,
        results: currentResults
      })
      
      setMessage(`第 ${round} 轮检测录入成功！`)
      await selectCollection(selectedCollection)
      await fetchCollections()
    } catch (err) {
      if (err.response?.data?.error) {
        setError(err.response.data.error)
      } else {
        setError('提交失败，请重试')
      }
    }
  }

  const isRoundComplete = (round) => {
    return testItems.every(item => results[round][item.key])
  }

  const getProgressPercent = () => {
    if (!testData) return 0
    const progress = testData.progress || {}
    const round = progress.current_round || 0
    const status = progress.status || ''
    
    if (status === 'Scrapped' || status === 'Disqualified' || 
        status === 'Qualified' || status === 'Pending_Storage') {
      return 100
    }
    return (round / 2) * 100
  }

  const getStatusBadge = (status) => {
    switch (status) {
      case 'Pending_Test': return <span className="status-badge status-pending">待检测</span>
      case 'Testing': return <span className="status-badge status-testing">检测中</span>
      case 'Qualified': return <span className="status-badge status-qualified">合格</span>
      case 'Scrapped': return <span className="status-badge status-scrapped">报废</span>
      case 'Disqualified': return <span className="status-badge status-scrapped">不合格</span>
      default: return <span className="status-badge status-pending">{status}</span>
    }
  }

  const hasRoundRecord = (round) => {
    return testData?.test_records?.some(r => r.test_round === round) || false
  }

  return (
    <div className="page">
      <h1 className="page-title">🔬 血液检测</h1>

      {error && <div className="alert alert-danger">{error}</div>}
      {message && <div className="alert alert-success">{message}</div>}

      {!selectedCollection && (
        <div className="section">
          <h2 style={{ marginBottom: '1rem' }}>选择待检测血液</h2>
          {collections.length === 0 ? (
            <p>暂无待检测血液</p>
          ) : (
            <table className="table">
              <thead>
                <tr>
                  <th>条码号</th>
                  <th>采集类型</th>
                  <th>状态</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {collections.map(col => (
                  <tr key={col.id}>
                    <td>{col.barcode}</td>
                    <td>{col.collection_type}</td>
                    <td>{getStatusBadge(col.status)}</td>
                    <td>
                      <button className="btn btn-primary" onClick={() => selectCollection(col.id)}>
                        开始检测
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {selectedCollection && testData && (
        <>
          <div className="test-section">
            <h2 className="test-section-title">血液信息</h2>
            <div className="form-row">
              <div className="form-group">
                <label className="form-label">条码号</label>
                <div className="form-input" style={{ padding: '0.75rem', backgroundColor: '#f8f9fa' }}>
                  {testData.collection?.barcode || ''}
                </div>
              </div>
              <div className="form-group">
                <label className="form-label">当前状态</label>
                <div style={{ padding: '0.75rem' }}>
                  {testData.progress && getStatusBadge(testData.progress.status)}
                </div>
              </div>
            </div>

            <div className="form-group">
              <label className="form-label">检测进度</label>
              <div className="progress-bar">
                <div className="progress-fill" style={{ width: `${getProgressPercent()}%` }} />
              </div>
              <p style={{ textAlign: 'right', marginTop: '0.5rem', color: '#666' }}>
                {testData.progress && `当前完成: 第 ${testData.progress.current_round || 0} 轮`}
              </p>
            </div>

            {testData.progress?.is_scrapped && (
              <div className="alert alert-danger">
                ⚠️ 该血液已报废，无法继续检测
              </div>
            )}

            <button
              className="btn btn-primary"
              onClick={() => {
                setSelectedCollection(null)
                setTestData(null)
              }}
              style={{ marginTop: '1rem' }}
            >
              返回列表
            </button>
          </div>

          {!testData.progress?.is_scrapped && (
            <>
              <div className="form-row" style={{ marginBottom: '1rem' }}>
                <div className="form-group">
                  <label className="form-label">试剂厂家</label>
                  <input
                    type="text"
                    className="form-input"
                    value={vendor}
                    onChange={(e) => setVendor(e.target.value)}
                    placeholder="请输入试剂厂家"
                  />
                </div>
                <div className="form-group">
                  <label className="form-label">操作人员</label>
                  <input
                    type="text"
                    className="form-input"
                    value={operator}
                    onChange={(e) => setOperator(e.target.value)}
                    placeholder="请输入操作人员ID"
                  />
                </div>
              </div>

              {[1, 2, 3].map(round => {
                const alreadyDone = hasRoundRecord(round)
                const allComplete = isRoundComplete(round)

                return (
                  <div className="test-section" key={round}>
                    <h3 style={{ marginBottom: '1rem' }}>
                      第 {round} 轮检测
                      {alreadyDone && (
                        <span style={{ marginLeft: '1rem', color: '#28a745', fontWeight: 'bold' }}>已完成</span>
                      )}
                    </h3>
                    <table className="table">
                      <thead>
                        <tr>
                          <th>检测项目</th>
                          <th>结果</th>
                        </tr>
                      </thead>
                      <tbody>
                        {testItems.map(item => (
                          <tr key={item.key}>
                            <td>{item.label}</td>
                            <td>
                              <div className="result-options">
                                {resultOptions.map(option => (
                                  <button
                                    type="button"
                                    key={option.value}
                                    className={`result-btn ${
                                      results[round][item.key] === option.value ? 'selected' : ''
                                    } ${option.value === 'Positive' ? 'positive' : ''}
                                    ${option.value === 'Invalid' ? 'invalid' : ''}
                                    `}
                                    onClick={() => handleResultChange(round, item.key, option.value)}
                                    disabled={alreadyDone}
                                  >
                                    {option.label}
                                  </button>
                                ))}
                              </div>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>

                    {!alreadyDone && (
                      <button
                        className="btn btn-success"
                        onClick={() => submitTest(round)}
                        disabled={!allComplete}
                      >
                        提交第 {round} 轮检测
                      </button>
                    )}
                  </div>
                )
              })}
            </>
          )}
        </>
      )}
    </div>
  )
}
