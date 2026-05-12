import React, { useState, useEffect } from 'react'
import { courseApi, studentApi } from '../api'

function CourseGraph({ studentId }) {
  const [graphData, setGraphData] = useState(null)
  const [courses, setCourses] = useState([])
  const [studentProgress, setStudentProgress] = useState({})
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    loadData()
  }, [studentId])

  const loadData = async () => {
    try {
      setLoading(true)
      
      const [coursesRes, graphRes] = await Promise.all([
        courseApi.getAll(),
        courseApi.getGraph(),
      ])

      setCourses(coursesRes.data.data || [])
      setGraphData(graphRes.data.data)

      const progressMap = {}
      for (const course of (coursesRes.data.data || [])) {
        try {
          const progressRes = await studentApi.getProgress(studentId, course.id)
          if (progressRes.data.success && progressRes.data.data) {
            progressMap[course.id] = progressRes.data.data
          }
        } catch (e) {
        }
      }
      setStudentProgress(progressMap)
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  const getCourseStatus = (courseId) => {
    const progress = studentProgress[courseId]
    if (progress) {
      if (progress.progress_percentage >= 100) {
        return 'completed'
      }
      return 'learning'
    }
    return 'locked'
  }

  const renderDAG = () => {
    if (!graphData || !graphData.nodes) return null

    const nodes = Object.keys(graphData.nodes)
    const levels = graphData.levels || {}
    
    const levelGroups = {}
    nodes.forEach(id => {
      const level = levels[id] || 0
      if (!levelGroups[level]) levelGroups[level] = []
      levelGroups[level].push(id)
    })

    const maxLevel = Math.max(...Object.keys(levelGroups).map(Number), 0)
    const nodeWidth = 180
    const nodeHeight = 80
    const levelGap = 220
    const horizontalGap = 30
    const padding = 50

    const svgWidth = (maxLevel + 1) * levelGap + padding * 2
    const maxNodesInLevel = Math.max(...Object.values(levelGroups).map(arr => arr.length))
    const svgHeight = maxNodesInLevel * (nodeHeight + horizontalGap) + padding * 2

    const nodePositions = {}
    Object.entries(levelGroups).forEach(([level, ids]) => {
      const totalHeight = ids.length * (nodeHeight + horizontalGap) - horizontalGap
      const startY = (svgHeight - totalHeight) / 2
      ids.forEach((id, index) => {
        nodePositions[id] = {
          x: padding + Number(level) * levelGap,
          y: startY + index * (nodeHeight + horizontalGap)
        }
      })
    })

    const edges = []
    if (graphData.edges) {
      Object.entries(graphData.edges).forEach(([from, edgeList]) => {
        edgeList.forEach(edge => {
          if (nodePositions[from] && nodePositions[edge.To]) {
            edges.push({
              from: nodePositions[from],
              to: nodePositions[edge.To],
              fromId: from,
              toId: edge.To
            })
          }
        })
      })
    }

    const getNodeColor = (status) => {
      switch (status) {
        case 'completed': return '#38ef7d'
        case 'learning': return '#667eea'
        default: return '#9e9e9e'
      }
    }

    const getCourseById = (id) => {
      return courses.find(c => c.id === id) || { name: id, difficulty: 'beginner', domain: 'frontend' }
    }

    return (
      <svg
        width={Math.max(svgWidth, 800)}
        height={Math.max(svgHeight, 400)}
        style={{ minWidth: '800px' }}
      >
        <defs>
          <marker
            id="arrowhead"
            markerWidth="10"
            markerHeight="7"
            refX="9"
            refY="3.5"
            orient="auto"
          >
            <polygon points="0 0, 10 3.5, 0 7" fill="#ccc" />
          </marker>
        </defs>

        {edges.map((edge, index) => {
          const x1 = edge.from.x + nodeWidth
          const y1 = edge.from.y + nodeHeight / 2
          const x2 = edge.to.x
          const y2 = edge.to.y + nodeHeight / 2
          const midX = (x1 + x2) / 2

          return (
            <path
              key={index}
              d={`M ${x1} ${y1} C ${midX} ${y1}, ${midX} ${y2}, ${x2} ${y2}`}
              fill="none"
              stroke="#ccc"
              strokeWidth="2"
              markerEnd="url(#arrowhead)"
            />
          )
        })}

        {Object.entries(nodePositions).map(([id, pos]) => {
          const course = getCourseById(id)
          const status = getCourseStatus(id)
          const progress = studentProgress[id]

          return (
            <g key={id}>
              <rect
                x={pos.x}
                y={pos.y}
                width={nodeWidth}
                height={nodeHeight}
                rx="12"
                fill="white"
                stroke={getNodeColor(status)}
                strokeWidth="3"
                style={{ filter: 'drop-shadow(0 2px 4px rgba(0,0,0,0.1))' }}
              />
              
              <text
                x={pos.x + nodeWidth / 2}
                y={pos.y + 30}
                textAnchor="middle"
                fontSize="14"
                fontWeight="600"
                fill="#1a1a2e"
              >
                {course.name.length > 15 ? course.name.substring(0, 15) + '...' : course.name}
              </text>

              <text
                x={pos.x + nodeWidth / 2}
                y={pos.y + 50}
                textAnchor="middle"
                fontSize="11"
                fill="#666"
              >
                {course.difficulty} · {course.domain}
              </text>

              {progress && (
                <g>
                  <rect
                    x={pos.x + 10}
                    y={pos.y + nodeHeight - 20}
                    width={nodeWidth - 20}
                    height="6"
                    rx="3"
                    fill="#e0e0e0"
                  />
                  <rect
                    x={pos.x + 10}
                    y={pos.y + nodeHeight - 20}
                    width={(nodeWidth - 20) * (progress.progress_percentage / 100)}
                    height="6"
                    rx="3"
                    fill={getNodeColor(status)}
                  />
                </g>
              )}
            </g>
          )
        })}
      </svg>
    )
  }

  if (loading) {
    return (
      <div>
        <h1 className="page-title">课程图谱</h1>
        <div className="card">
          <p>加载中...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div>
        <h1 className="page-title">课程图谱</h1>
        <div className="card">
          <p style={{ color: '#e74c3c' }}>错误: {error}</p>
          <p>请确保后端服务已启动在 http://localhost:8080</p>
        </div>
      </div>
    )
  }

  return (
    <div>
      <h1 className="page-title">课程图谱</h1>

      <div className="card">
        <div className="dag-legend">
          <div className="legend-item">
            <div className="legend-circle completed"></div>
            <span>已完成</span>
          </div>
          <div className="legend-item">
            <div className="legend-circle learning"></div>
            <span>学习中</span>
          </div>
          <div className="legend-item">
            <div className="legend-circle locked"></div>
            <span>未解锁</span>
          </div>
        </div>
      </div>

      <div className="dag-container">
        {renderDAG()}
      </div>

      <div className="card">
        <h3>课程列表</h3>
        <div className="grid grid-2">
          {courses.map(course => {
            const status = getCourseStatus(course.id)
            const progress = studentProgress[course.id]
            
            return (
              <div key={course.id} className={`path-node ${status}`}>
                <h4 style={{ marginBottom: '0.5rem' }}>{course.name}</h4>
                <p style={{ fontSize: '0.875rem', color: '#666', marginBottom: '0.5rem' }}>
                  {course.description}
                </p>
                <div>
                  <span className={`tag tag-${course.difficulty}`}>{course.difficulty}</span>
                  <span className={`tag tag-${course.domain}`}>{course.domain}</span>
                  <span className="tag" style={{ background: '#f5f5f5', color: '#666' }}>
                    {course.expected_hours}小时
                  </span>
                </div>
                {progress && (
                  <div className="progress-bar" style={{ marginTop: '0.75rem' }}>
                    <div 
                      className={`progress-fill ${status === 'completed' ? 'green' : 'blue'}`}
                      style={{ width: `${progress.progress_percentage}%` }}
                    ></div>
                  </div>
                )}
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}

export default CourseGraph
