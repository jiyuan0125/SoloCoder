import React, { useState, useEffect } from 'react'
import { courseApi, studentApi } from '../api'

function LearningPath({ studentId }) {
  const [courses, setCourses] = useState([])
  const [selectedGoal, setSelectedGoal] = useState('')
  const [targetCourse, setTargetCourse] = useState('')
  const [learningPath, setLearningPath] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  const goals = [
    { id: 'fullstack', name: '全栈开发工程师', description: '掌握前端和后端开发技能' },
    { id: 'frontend', name: '前端开发专家', description: '精通前端框架和技术栈' },
    { id: 'backend', name: '后端开发专家', description: '深入学习服务器端开发' },
    { id: 'database', name: '数据库工程师', description: '专注于数据库设计和优化' },
  ]

  useEffect(() => {
    loadCourses()
  }, [])

  const loadCourses = async () => {
    try {
      const res = await courseApi.getAll()
      setCourses(res.data.data || [])
    } catch (err) {
      setError(err.message)
    }
  }

  const generatePath = async () => {
    if (!targetCourse || !selectedGoal) return

    try {
      setLoading(true)
      setError(null)
      
      const res = await studentApi.generatePath({
        student_id: studentId,
        target_course_id: targetCourse,
        goal: goals.find(g => g.id === selectedGoal)?.name || selectedGoal,
      })

      if (res.data.success) {
        setLearningPath(res.data.data)
      }
    } catch (err) {
      setError(err.response?.data?.error || err.message)
    } finally {
      setLoading(false)
    }
  }

  const getStatusBadge = (status) => {
    switch (status) {
      case 'completed':
        return <span className="tag" style={{ background: '#e8f5e9', color: '#2e7d32' }}>已完成</span>
      case 'in_progress':
        return <span className="tag" style={{ background: '#e3f2fd', color: '#1565c0' }}>学习中</span>
      case 'locked':
        return <span className="tag" style={{ background: '#f5f5f5', color: '#666' }}>未解锁</span>
      default:
        return <span className="tag" style={{ background: '#fff3e0', color: '#ef6c00' }}>可学习</span>
    }
  }

  const groupByLevel = (nodes) => {
    const groups = {}
    nodes.forEach(node => {
      if (!groups[node.level]) {
        groups[node.level] = []
      }
      groups[node.level].push(node)
    })
    return Object.entries(groups).sort((a, b) => Number(a[0]) - Number(b[0]))
  }

  const canReorder = (nodes) => {
    return nodes.length > 1
  }

  const handleDragStart = (e, level, index) => {
    e.dataTransfer.setData('level', level)
    e.dataTransfer.setData('index', index)
  }

  const handleDrop = (e, targetLevel, targetIndex) => {
    e.preventDefault()
    const sourceLevel = e.dataTransfer.getData('level')
    const sourceIndex = e.dataTransfer.getData('index')

    if (sourceLevel !== String(targetLevel) || !canReorder(learningPath.path_nodes)) {
      return
    }

    const newPath = { ...learningPath }
    const levelNodes = groupByLevel(newPath.path_nodes).find(([lvl]) => lvl === sourceLevel)
    
    if (levelNodes) {
      const nodes = levelNodes[1]
      if (nodes.length > 1) {
        const temp = nodes[sourceIndex]
        nodes[sourceIndex] = nodes[targetIndex]
        nodes[targetIndex] = temp
      }
    }

    setLearningPath(newPath)
  }

  const handleDragOver = (e) => {
    e.preventDefault()
  }

  return (
    <div>
      <h1 className="page-title">学习路径规划</h1>

      <div className="card">
        <h3>选择学习目标</h3>
        
        <div className="form-group">
          <label className="label">目标方向</label>
          <select
            className="select"
            value={selectedGoal}
            onChange={(e) => setSelectedGoal(e.target.value)}
          >
            <option value="">请选择目标方向</option>
            {goals.map(goal => (
              <option key={goal.id} value={goal.id}>
                {goal.name}
              </option>
            ))}
          </select>
        </div>

        <div className="form-group">
          <label className="label">目标课程</label>
          <select
            className="select"
            value={targetCourse}
            onChange={(e) => setTargetCourse(e.target.value)}
          >
            <option value="">请选择目标课程</option>
            {courses.map(course => (
              <option key={course.id} value={course.id}>
                {course.name} ({course.difficulty})
              </option>
            ))}
          </select>
        </div>

        <button
          className="btn btn-primary"
          onClick={generatePath}
          disabled={!targetCourse || !selectedGoal || loading}
        >
          {loading ? '生成中...' : '生成学习路径'}
        </button>
      </div>

      {error && (
        <div className="card">
          <p style={{ color: '#e74c3c' }}>错误: {error}</p>
        </div>
      )}

      {learningPath && (
        <div>
          <div className="card">
            <h3>学习路径: {learningPath.target_goal}</h3>
            <p>共 {learningPath.total_courses} 门课程需要学习</p>
            <p style={{ fontSize: '0.875rem', color: '#666', marginTop: '0.5rem' }}>
              💡 提示: 同一层级的课程可以拖拽调整顺序（无依赖关系）
            </p>
          </div>

          {groupByLevel(learningPath.path_nodes).map(([level, nodes], levelIndex) => (
            <div key={level} className="card">
              <div className="path-level">
                <h4 className="path-level-title">
                  📚 第 {Number(level) + 1} 阶段（{nodes.length} 门课）
                </h4>
                <div className="grid grid-2">
                  {nodes.map((node, nodeIndex) => (
                    <div
                      key={node.course_id}
                      className={`path-node ${node.status}`}
                      draggable={canReorder(nodes)}
                      onDragStart={(e) => handleDragStart(e, level, nodeIndex)}
                      onDragOver={handleDragOver}
                      onDrop={(e) => handleDrop(e, level, nodeIndex)}
                      style={{ cursor: canReorder(nodes) ? 'grab' : 'default' }}
                    >
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem' }}>
                        <h4 style={{ margin: 0 }}>{node.course_name}</h4>
                        {getStatusBadge(node.status)}
                      </div>
                      
                      <div className="progress-bar">
                        <div
                          className={`progress-fill ${node.status === 'completed' ? 'green' : 'blue'}`}
                          style={{ width: `${node.progress}%` }}
                        ></div>
                      </div>
                      <p style={{ fontSize: '0.875rem', color: '#666', marginTop: '0.5rem' }}>
                        进度: {node.progress.toFixed(1)}%
                      </p>
                    </div>
                  ))}
                </div>
              </div>
              {levelIndex < groupByLevel(learningPath.path_nodes).length - 1 && (
                <div className="arrow-down">⬇️</div>
              )}
            </div>
          ))}

          <div className="card">
            <h3>并行学习建议</h3>
            <p>
              同一阶段（层级）的课程之间没有依赖关系，可以并行学习。
              建议同时学习不超过 5 门课程，以保证学习质量。
            </p>
            <div style={{ marginTop: '1rem', display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
              {groupByLevel(learningPath.path_nodes).map(([level, nodes]) => (
                <span key={level} className="tag" style={{ background: '#e8eaf6', color: '#5c6bc0' }}>
                  阶段 {Number(level) + 1}: {nodes.length} 门课并行
                </span>
              ))}
            </div>
          </div>
        </div>
      )}

      {!learningPath && !loading && (
        <div className="card">
          <h3>路径说明</h3>
          <ul style={{ marginTop: '1rem', paddingLeft: '1.5rem', lineHeight: '1.8' }}>
            <li>选择一个学习目标方向和目标课程</li>
            <li>系统会根据 DAG 拓扑排序逆序生成学习路径</li>
            <li>路径只包含未学习或未完成的课程</li>
            <li>多个先修链汇聚时会并行展示</li>
            <li>同一层级的课程可以拖拽调整顺序</li>
          </ul>
        </div>
      )}
    </div>
  )
}

export default LearningPath
