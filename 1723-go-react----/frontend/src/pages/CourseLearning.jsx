import React, { useState, useEffect } from 'react'
import { courseApi, studentApi, quizApi } from '../api'

function CourseLearning({ studentId }) {
  const [courses, setCourses] = useState([])
  const [selectedCourse, setSelectedCourse] = useState(null)
  const [progress, setProgress] = useState(null)
  const [showQuiz, setShowQuiz] = useState(false)
  const [quizQuestions, setQuizQuestions] = useState([])
  const [quizAnswers, setQuizAnswers] = useState({})
  const [quizResult, setQuizResult] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

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

  const selectCourse = async (courseId) => {
    setSelectedCourse(null)
    setProgress(null)
    setShowQuiz(false)
    setQuizResult(null)

    try {
      const courseRes = await courseApi.get(courseId)
      setSelectedCourse(courseRes.data.data)

      try {
        const progressRes = await studentApi.getProgress(studentId, courseId)
        if (progressRes.data.success) {
          setProgress(progressRes.data.data)
        }
      } catch (e) {
      }
    } catch (err) {
      setError(err.message)
    }
  }

  const startLearning = async (courseId) => {
    try {
      const res = await studentApi.startLearning({
        student_id: studentId,
        course_id: courseId,
      })
      if (res.data.success) {
        await selectCourse(courseId)
      }
    } catch (err) {
      setError(err.response?.data?.error || err.message)
    }
  }

  const completeUnit = async (unitId) => {
    if (!selectedCourse || !progress) return

    try {
      const res = await studentApi.completeUnit({
        student_id: studentId,
        course_id: selectedCourse.id,
        unit_id: unitId,
      })
      if (res.data.success) {
        const progressRes = await studentApi.getProgress(studentId, selectedCourse.id)
        if (progressRes.data.success) {
          setProgress(progressRes.data.data)
        }
      }
    } catch (err) {
      setError(err.response?.data?.error || err.message)
    }
  }

  const isUnitCompleted = (unitId) => {
    if (!progress) return false
    return progress.completed_unit_ids.includes(unitId)
  }

  const canTakeQuiz = () => {
    if (!selectedCourse || !progress) return false
    return progress.progress_percentage >= 100 || 
           progress.completed_unit_ids.length === selectedCourse.units?.length
  }

  const loadQuiz = async () => {
    if (!selectedCourse) return

    try {
      setLoading(true)
      const res = await quizApi.get(studentId, selectedCourse.id)
      if (res.data.success) {
        setQuizQuestions(res.data.data)
        setQuizAnswers({})
        setQuizResult(null)
        setShowQuiz(true)
      }
    } catch (err) {
      setError(err.response?.data?.error || err.message)
    } finally {
      setLoading(false)
    }
  }

  const submitQuiz = async () => {
    if (!selectedCourse || quizQuestions.length === 0) return

    try {
      setLoading(true)
      const answers = quizQuestions.map((q, i) => quizAnswers[q.id] ?? -1)

      const res = await quizApi.submit({
        student_id: studentId,
        course_id: selectedCourse.id,
        answers: answers,
        questions: quizQuestions,
      })

      if (res.data.success) {
        setQuizResult(res.data.data)
      }
    } catch (err) {
      setError(err.response?.data?.error || err.message)
    } finally {
      setLoading(false)
    }
  }

  const getProgressPercentage = () => {
    if (!progress || !selectedCourse) return 0
    return progress.progress_percentage
  }

  return (
    <div>
      <h1 className="page-title">课程学习</h1>

      {error && (
        <div className="card">
          <p style={{ color: '#e74c3c' }}>错误: {error}</p>
        </div>
      )}

      {!selectedCourse ? (
        <div className="card">
          <h3>选择课程</h3>
          <div className="grid grid-2">
            {courses.map(course => (
              <div 
                key={course.id} 
                className="path-node available"
                style={{ cursor: 'pointer' }}
                onClick={() => selectCourse(course.id)}
              >
                <h4 style={{ marginBottom: '0.5rem' }}>{course.name}</h4>
                <p style={{ fontSize: '0.875rem', color: '#666', marginBottom: '0.75rem' }}>
                  {course.description}
                </p>
                <div>
                  <span className={`tag tag-${course.difficulty}`}>{course.difficulty}</span>
                  <span className={`tag tag-${course.domain}`}>{course.domain}</span>
                  <span className="tag" style={{ background: '#f5f5f5', color: '#666' }}>
                    {course.expected_hours}小时
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>
      ) : (
        <div>
          <button 
            className="btn" 
            style={{ background: '#f5f5f5', color: '#333', marginBottom: '1rem' }}
            onClick={() => setSelectedCourse(null)}
          >
            ← 返回课程列表
          </button>

          <div className="card">
            <h3>{selectedCourse.name}</h3>
            <p style={{ color: '#666', marginBottom: '1rem' }}>{selectedCourse.description}</p>
            <div style={{ marginBottom: '1rem' }}>
              <span className={`tag tag-${selectedCourse.difficulty}`}>{selectedCourse.difficulty}</span>
              <span className={`tag tag-${selectedCourse.domain}`}>{selectedCourse.domain}</span>
              <span className="tag" style={{ background: '#f5f5f5', color: '#666' }}>
                {selectedCourse.expected_hours}小时
              </span>
            </div>

            {!progress ? (
              <button 
                className="btn btn-primary"
                onClick={() => startLearning(selectedCourse.id)}
              >
                开始学习
              </button>
            ) : (
              <div>
                <div style={{ marginBottom: '0.5rem' }}>
                  <span style={{ fontWeight: 500 }}>学习进度: </span>
                  <span style={{ color: '#667eea', fontWeight: 600 }}>
                    {getProgressPercentage().toFixed(1)}%
                  </span>
                </div>
                <div className="progress-bar" style={{ height: '16px' }}>
                  <div 
                    className="progress-fill blue"
                    style={{ width: `${getProgressPercentage()}%` }}
                  ></div>
                </div>
                <p style={{ fontSize: '0.875rem', color: '#666', marginTop: '0.5rem' }}>
                  已完成 {progress.completed_unit_ids.length} / {selectedCourse.units?.length || 0} 个单元
                </p>
              </div>
            )}
          </div>

          {progress && selectedCourse.units && selectedCourse.units.length > 0 && (
            <div className="card">
              <h3>学习单元</h3>
              <ul className="unit-list">
                {selectedCourse.units
                  .sort((a, b) => a.order - b.order)
                  .map((unit, index) => (
                    <li key={unit.id} className="unit-item">
                      <input
                        type="checkbox"
                        className="unit-checkbox"
                        checked={isUnitCompleted(unit.id)}
                        onChange={() => completeUnit(unit.id)}
                      />
                      <div style={{ flex: 1 }}>
                        <h4 style={{ marginBottom: '0.25rem' }}>
                          第 {index + 1} 章: {unit.name}
                        </h4>
                        <p style={{ fontSize: '0.875rem', color: '#666' }}>
                          {unit.description}
                        </p>
                      </div>
                      {isUnitCompleted(unit.id) && (
                        <span style={{ color: '#11998e', fontWeight: 500 }}>✓ 已完成</span>
                      )}
                    </li>
                  ))}
              </ul>
            </div>
          )}

          {progress && (
            <div className="card">
              <h3>课程测验</h3>
              {canTakeQuiz() ? (
                showQuiz ? (
                  <div className="quiz-container">
                    {!quizResult ? (
                      <div>
                        {quizQuestions.map((question, qIndex) => (
                          <div key={question.id} className="question">
                            <div className="question-text">
                              {qIndex + 1}. {question.text}
                            </div>
                            {question.options.map((option, oIndex) => (
                              <div 
                                key={oIndex}
                                className={`option ${quizAnswers[question.id] === oIndex ? 'selected' : ''}`}
                                onClick={() => setQuizAnswers(prev => ({ ...prev, [question.id]: oIndex }))}
                              >
                                <input
                                  type="radio"
                                  className="option-radio"
                                  name={`question-${question.id}`}
                                  checked={quizAnswers[question.id] === oIndex}
                                  onChange={() => setQuizAnswers(prev => ({ ...prev, [question.id]: oIndex }))}
                                />
                                <span>{String.fromCharCode(65 + oIndex)}. {option}</span>
                              </div>
                            ))}
                          </div>
                        ))}
                        <button
                          className="btn btn-primary"
                          onClick={submitQuiz}
                          disabled={loading || Object.keys(quizAnswers).length < quizQuestions.length}
                        >
                          {loading ? '提交中...' : '提交答案'}
                        </button>
                      </div>
                    ) : (
                      <div>
                        <div style={{ textAlign: 'center', marginBottom: '2rem' }}>
                          <h2 style={{ 
                            fontSize: '3rem', 
                            marginBottom: '1rem',
                            color: quizResult.passed ? '#11998e' : '#e74c3c'
                          }}>
                            {quizResult.passed ? '🎉 恭喜通过！' : '😔 未通过'}
                          </h2>
                          <p style={{ fontSize: '1.25rem' }}>
                            得分: {quizResult.score.toFixed(1)}分 
                            ({quizResult.correct_count}/{quizResult.total_count}题)
                          </p>
                          <p style={{ color: '#666', marginTop: '0.5rem' }}>
                            需要答对 8 题才能通过
                          </p>
                        </div>
                        {!quizResult.passed && (
                          <button
                            className="btn btn-primary"
                            onClick={() => {
                              setShowQuiz(false)
                              setQuizResult(null)
                              setQuizAnswers({})
                            }}
                          >
                            重新测验
                          </button>
                        )}
                      </div>
                    )}
                  </div>
                ) : (
                  <div>
                    <p style={{ marginBottom: '1rem' }}>
                      ✅ 所有单元已完成，可以参加测验
                    </p>
                    <p style={{ fontSize: '0.875rem', color: '#666', marginBottom: '1rem' }}>
                      💡 测验规则：10道单选题，答对8题通过。未通过可重考，每次重考至少30%题目不同。
                    </p>
                    <button
                      className="btn btn-success"
                      onClick={loadQuiz}
                      disabled={loading}
                    >
                      {loading ? '加载中...' : '开始测验'}
                    </button>
                  </div>
                )
              ) : (
                <div>
                  <p style={{ color: '#999' }}>
                    🔒 请先完成所有学习单元后解锁测验
                  </p>
                </div>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export default CourseLearning
