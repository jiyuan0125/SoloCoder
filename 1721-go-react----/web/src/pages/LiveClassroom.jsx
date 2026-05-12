import React, { useState, useEffect, useRef } from 'react'
import { useParams } from 'react-router-dom'
import { courseAPI } from '../services/api'

const LiveClassroom = ({ currentStudent }) => {
  const { id } = useParams()
  const [course, setCourse] = useState(null)
  const [danmakus, setDanmakus] = useState([])
  const [flyingDanmakus, setFlyingDanmakus] = useState([])
  const [input, setInput] = useState('')
  const [activeVotes, setActiveVotes] = useState([])
  const [onlineCount, setOnlineCount] = useState(0)
  const [canInteract, setCanInteract] = useState(false)
  const wsRef = useRef(null)
  const votedRef = useRef(new Set())

  useEffect(() => {
    loadCourse()
    return () => {
      if (wsRef.current) {
        wsRef.current.close()
      }
    }
  }, [id])

  const loadCourse = async () => {
    try {
      const res = await courseAPI.getById(id)
      setCourse(res.data)
      
      const now = new Date()
      const startTime = res.data.start_time ? new Date(res.data.start_time) : new Date()
      const timeUntilStart = startTime.getTime() - now.getTime()
      
      if (timeUntilStart < 0 && res.data.live_status === 'in_progress') {
        setCanInteract(true)
      } else {
        setCanInteract(false)
      }

      loadVotes()
      connectWebSocket()
    } catch (error) {
      console.error('Failed to load course:', error)
    }
  }

  const loadVotes = async () => {
    try {
      const res = await courseAPI.getVotes(id)
      setActiveVotes(res.data)
    } catch (error) {
      console.error('Failed to load votes:', error)
    }
  }

  const connectWebSocket = () => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/api/courses/${id}/ws?student_id=${currentStudent.id}`
    wsRef.current = new WebSocket(wsUrl)

    wsRef.current.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        
        if (data.type === 'danmaku') {
          const danmaku = {
            id: Date.now(),
            ...data.payload,
            top: Math.random() * 80 + 5
          }
          setDanmakus(prev => [...prev.slice(-50), { content: data.payload.content, time: new Date().toLocaleTimeString() }])
          setFlyingDanmakus(prev => [...prev, danmaku])
          setTimeout(() => {
            setFlyingDanmakus(prev => prev.filter(d => d.id !== danmaku.id))
          }, 8000)
        } else if (data.type === 'vote_result') {
          setActiveVotes(prev => prev.map(v => {
            if (v.id === data.payload.vote_id) {
              return { ...v, results: data.payload.results }
            }
            return v
          }))
        } else if (data.type === 'pong') {
          setOnlineCount(data.payload.online)
        } else if (data.type === 'error') {
          console.error('WebSocket error:', data.payload.message)
        }
      } catch (e) {
        console.error('Parse error:', e)
      }
    }

    wsRef.current.onopen = () => {
      setInterval(() => {
        if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
          wsRef.current.send(JSON.stringify({ type: 'ping' }))
        }
      }, 5000)
    }
  }

  const sendDanmaku = () => {
    if (!input.trim() || !canInteract) return
    
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({
        type: 'danmaku',
        payload: { content: input }
      }))
      setInput('')
    }
  }

  const handleVote = (voteId, optionId) => {
    if (!canInteract || votedRef.current.has(voteId)) return
    
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({
        type: 'vote',
        payload: { vote_id: voteId, option_id: optionId }
      }))
      votedRef.current.add(voteId)
    }
  }

  const createVote = async () => {
    const title = prompt('请输入投票主题：')
    if (!title) return

    const optionsStr = prompt('请输入选项（用逗号分隔，2-5个选项）：')
    if (!optionsStr) return

    const options = optionsStr.split(/[,，]/).filter(o => o.trim())
    if (options.length < 2 || options.length > 5) {
      alert('请输入2-5个选项')
      return
    }

    try {
      await courseAPI.createVote(id, { title, options })
      loadVotes()
    } catch (error) {
      console.error('Failed to create vote:', error)
    }
  }

  if (!course) {
    return <div className="card">加载中...</div>
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <div>
          <h1 className="text-xl font-bold">{course.name}</h1>
          <p className="text-sm text-gray-500">讲师：{course.instructor} | 在线人数：{onlineCount}</p>
        </div>
        <div className="flex space-x-2">
          {!canInteract && (
            <span className="text-sm text-gray-500">直播开始后可发弹幕和投票</span>
          )}
          <button className="btn btn-secondary" onClick={createVote}>
            创建投票
          </button>
        </div>
      </div>

      <div className="live-container">
        <div className="live-video">
          <div className="danmaku-layer">
            {flyingDanmakus.map(d => (
              <div 
                key={d.id} 
                className="danmaku-item"
                style={{ top: `${d.top}%` }}
              >
                {d.content}
              </div>
            ))}
          </div>
          <div className="text-center">
            <div className="text-4xl mb-4">📺</div>
            <div>直播画面区域</div>
            <div className="text-sm mt-2">
              状态：{course.live_status === 'in_progress' ? '直播中' : '未开始'}
            </div>
          </div>
        </div>

        <div className="sidebar">
          <div className="card">
            <h3 className="font-bold mb-2">弹幕</h3>
            <div className="danmaku-list mb-2">
              {danmakus.length === 0 ? (
                <p className="text-sm text-gray-500">暂无弹幕</p>
              ) : (
                danmakus.map((d, i) => (
                  <div key={i} className="danmaku-msg">
                    <span className="text-gray-400">[{d.time}]</span> {d.content}
                  </div>
                ))
              )}
            </div>
            <div className="flex space-x-2">
              <input
                type="text"
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && sendDanmaku()}
                placeholder={canInteract ? '输入弹幕...' : '直播开始后可发送'}
                disabled={!canInteract}
              />
              <button 
                className="btn btn-primary"
                onClick={sendDanmaku}
                disabled={!canInteract}
              >
                发送
              </button>
            </div>
          </div>

          <div className="card">
            <h3 className="font-bold mb-2">投票</h3>
            {activeVotes.length === 0 ? (
              <p className="text-sm text-gray-500">暂无投票</p>
            ) : (
              activeVotes.map(vote => (
                <div key={vote.id} className="vote-panel mb-4">
                  <h4 className="font-bold mb-2">{vote.title}</h4>
                  {vote.options.map(opt => {
                    const totalVotes = vote.options.reduce((sum, o) => sum + (o.votes || 0), 0)
                    const percentage = totalVotes > 0 ? (opt.votes / totalVotes * 100) : 0
                    const hasVoted = votedRef.current.has(vote.id)
                    
                    return (
                      <div 
                        key={opt.id} 
                        className={`vote-option ${hasVoted ? 'voted' : ''}`}
                        onClick={() => handleVote(vote.id, opt.id)}
                      >
                        <span>{opt.text}</span>
                        <span className="text-sm">
                          {opt.votes || 0} 票 ({percentage.toFixed(1)}%)
                        </span>
                      </div>
                    )
                  })}
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default LiveClassroom
