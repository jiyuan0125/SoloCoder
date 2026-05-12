import React, { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import axios from 'axios'

function QuestionnairePage() {
  const { taskId, courseId } = useParams()
  const navigate = useNavigate()
  
  const [questions, setQuestions] = useState([])
  const [currentIndex, setCurrentIndex] = useState(0)
  const [answers, setAnswers] = useState({})
  const [comment, setComment] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [showComment, setShowComment] = useState(false)

  useEffect(() => {
    fetchQuestions()
  }, [courseId])

  const fetchQuestions = async () => {
    try {
      setLoading(true)
      const courseTypes = ['theory', 'experiment', 'sport']
      const type = courseTypes[parseInt(courseId) % 3]
      
      const response = await axios.get(`/api/questions/${type}`)
      setQuestions(response.data)
    } catch (err) {
      setError('加载问卷失败')
    } finally {
      setLoading(false)
    }
  }

  const handleStarClick = (score) => {
    const question = questions[currentIndex]
    setAnswers({
      ...answers,
      [question.id]: score
    })
  }

  const handleNext = () => {
    if (currentIndex < questions.length - 1) {
      setCurrentIndex(currentIndex + 1)
    } else {
      setShowComment(true)
    }
  }

  const handlePrev = () => {
    if (showComment) {
      setShowComment(false)
    } else if (currentIndex > 0) {
      setCurrentIndex(currentIndex - 1)
    }
  }

  const handleSubmit = async () => {
    try {
      setSubmitting(true)
      setError('')

      const answersArray = Object.entries(answers).map(([questionId, score]) => ({
        question_id: parseInt(questionId),
        score: score
      }))

      await axios.post('/api/evaluations', {
        student_id: 1,
        course_id: parseInt(courseId),
        task_id: parseInt(taskId),
        answers: answersArray,
        comment: comment
      })

      alert('评估提交成功！')
      navigate('/')
    } catch (err) {
      if (err.response?.status === 409) {
        setError('错误 (409): 您已对该课程完成评估')
      } else if (err.response?.status === 400) {
        setError(`错误 (400): ${err.response?.data?.error}`)
      } else {
        setError(err.response?.data?.error || '提交失败')
      }
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) return <div className="loading">加载中...</div>
  if (questions.length === 0) return <div className="card">暂无问卷数据</div>

  const currentQuestion = questions[currentIndex]
  const progress = showComment 
    ? 100 
    : ((currentIndex + 1) / questions.length) * 100
  const currentScore = answers[currentQuestion?.id] || 0

  return (
    <div className="questionnaire-container">
      <div className="card">
        <h2>课程评估问卷</h2>
        
        <div className="progress-bar">
          <div className="progress-fill" style={{ width: `${progress}%` }}></div>
        </div>

        {error && <div className="error-message">{error}</div>}

        {!showComment ? (
          <div className="question-card">
            <div className="question-number">
              第 {currentIndex + 1} / {questions.length} 题
              <span style={{ marginLeft: '1rem', fontSize: '0.85rem', color: '#718096' }}>
                {currentQuestion.question_type === 'general' ? '通用指标' : '专项指标'}
              </span>
            </div>
            <div className="question-text">{currentQuestion.text}</div>
            
            <div className="star-rating">
              {[1, 2, 3, 4, 5].map(score => (
                <button
                  key={score}
                  className={`star-btn ${score <= currentScore ? 'selected' : ''}`}
                  onClick={() => handleStarClick(score)}
                >
                  ★
                </button>
              ))}
            </div>

            <div style={{ marginBottom: '1rem', color: '#718096' }}>
              {currentScore > 0 && (
                <span>您的评分: {currentScore}分 ({['', '很差', '较差', '一般', '较好', '很好'][currentScore]})</span>
              )}
            </div>

            <div className="navigation-buttons">
              <button
                className="btn btn-secondary"
                onClick={handlePrev}
                disabled={currentIndex === 0}
              >
                上一题
              </button>
              <button
                className="btn btn-primary"
                onClick={handleNext}
                disabled={!currentScore}
              >
                {currentIndex === questions.length - 1 ? '下一步' : '下一题'}
              </button>
            </div>
          </div>
        ) : (
          <div>
            <div className="comment-section">
              <div className="form-group">
                <label>文字评语 (可选，最多500字)</label>
                <textarea
                  value={comment}
                  onChange={(e) => setComment(e.target.value.slice(0, 500))}
                  placeholder="请输入您对本课程的建议和意见..."
                />
                <div style={{ textAlign: 'right', color: '#718096', fontSize: '0.85rem' }}>
                  {comment.length}/500
                </div>
              </div>
            </div>

            <div className="navigation-buttons">
              <button className="btn btn-secondary" onClick={handlePrev}>
                返回修改
              </button>
              <button
                className="btn btn-primary"
                onClick={handleSubmit}
                disabled={submitting}
              >
                {submitting ? '提交中...' : '提交评估'}
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export default QuestionnairePage
