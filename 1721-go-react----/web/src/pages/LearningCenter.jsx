import React, { useState, useEffect } from 'react'
import { progressAPI, studentAPI, reviewAPI } from '../services/api'

const LearningCenter = ({ currentStudent }) => {
  const [progress, setProgress] = useState([])
  const [enrollments, setEnrollments] = useState([])
  const [showReviewModal, setShowReviewModal] = useState(false)
  const [selectedCourse, setSelectedCourse] = useState(null)
  const [reviewForm, setReviewForm] = useState({ rating: 5, comment: '' })

  useEffect(() => {
    loadData()
  }, [currentStudent.id])

  const loadData = async () => {
    try {
      const [progressRes, enrollRes] = await Promise.all([
        progressAPI.getByStudent(currentStudent.id),
        studentAPI.getEnrollments(currentStudent.id)
      ])
      setProgress(progressRes.data)
      setEnrollments(enrollRes.data)
    } catch (error) {
      console.error('Failed to load data:', error)
    }
  }

  const handleUpdateProgress = async (courseId, videoDuration, currentProgress) => {
    const newSeconds = Math.min(
      currentProgress.watched_seconds + 60,
      videoDuration
    )
    try {
      await progressAPI.update(currentStudent.id, courseId, newSeconds)
      loadData()
    } catch (error) {
      console.error('Failed to update progress:', error)
    }
  }

  const openReviewModal = (course) => {
    setSelectedCourse(course)
    setReviewForm({ rating: 5, comment: '' })
    setShowReviewModal(true)
  }

  const handleSubmitReview = async () => {
    if (!selectedCourse) return
    
    if (reviewForm.comment.length > 200) {
      alert('评价不能超过200字')
      return
    }

    try {
      await reviewAPI.create({
        student_id: currentStudent.id,
        course_id: selectedCourse.course_id,
        rating: reviewForm.rating,
        comment: reviewForm.comment
      })
      setShowReviewModal(false)
      alert('评价提交成功！')
    } catch (error) {
      alert(error.response?.data?.message || '提交失败')
    }
  }

  const getProgressItem = (courseId) => {
    return progress.find(p => p.course_id === courseId)
  }

  return (
    <div>
      <h1 className="text-xl font-bold mb-6">学习中心</h1>

      <div className="card mb-6">
        <h2 className="font-bold mb-4">我的课程</h2>
        {enrollments.length === 0 ? (
          <p className="text-gray-500">暂无已选课程，请前往课程大厅选课</p>
        ) : (
          <div className="space-y-4">
            {enrollments.map(enrollment => {
              const progressItem = getProgressItem(enrollment.course_id)
              const percent = progressItem ? Math.min(progressItem.progress_percent || 0, 100) : 0
              const isCompleted = progressItem?.is_completed || percent >= 90

              return (
                <div key={enrollment.id} className="border rounded-lg p-4">
                  <div className="flex justify-between items-start mb-3">
                    <div>
                      <h3 className="font-bold">{enrollment.course_name}</h3>
                      <p className="text-sm text-gray-500">
                        讲师：{enrollment.instructor} | 
                        类型：{enrollment.course_type === 'live' ? '直播课' : '录播课'} |
                        选报时间：{new Date(enrollment.enrolled_at).toLocaleDateString()}
                      </p>
                    </div>
                    <span className={`tag ${isCompleted ? 'bg-green-100 text-green-600' : 'bg-yellow-100 text-yellow-600'}`}>
                      {isCompleted ? '已完成' : '学习中'}
                    </span>
                  </div>

                  {enrollment.course_type === 'record' && (
                    <>
                      <div className="mb-3">
                        <div className="flex justify-between text-sm mb-1">
                          <span>学习进度</span>
                          <span>{percent.toFixed(1)}%</span>
                        </div>
                        <div className="progress-bar">
                          <div 
                            className="progress-bar-fill"
                            style={{ width: `${percent}%` }}
                          />
                        </div>
                      </div>
                      <div className="flex space-x-2">
                        <button 
                          className="btn btn-primary btn-sm"
                          onClick={() => handleUpdateProgress(
                            enrollment.course_id,
                            progressItem?.video_duration || 3600,
                            progressItem || { watched_seconds: 0 }
                          )}
                        >
                          模拟学习 (+60秒)
                        </button>
                        {isCompleted && (
                          <button 
                            className="btn btn-success btn-sm"
                            onClick={() => openReviewModal(enrollment)}
                          >
                            评价课程
                          </button>
                        )}
                      </div>
                    </>
                  )}
                </div>
              )
            })}
          </div>
        )}
      </div>

      <div className="card">
        <h2 className="font-bold mb-4">学习进度详情</h2>
        {progress.length === 0 ? (
          <p className="text-gray-500">暂无学习记录</p>
        ) : (
          <table>
            <thead>
              <tr>
                <th>课程名称</th>
                <th>已观看</th>
                <th>总时长</th>
                <th>进度</th>
                <th>状态</th>
              </tr>
            </thead>
            <tbody>
              {progress.map(p => (
                <tr key={p.id}>
                  <td>{p.course_name}</td>
                  <td>{Math.floor(p.watched_seconds / 60)}分钟</td>
                  <td>{p.video_duration ? Math.floor(p.video_duration / 60) : '-'}分钟</td>
                  <td>{(p.progress_percent || 0).toFixed(1)}%</td>
                  <td>
                    <span className={p.is_completed ? 'text-green-500' : 'text-gray-500'}>
                      {p.is_completed ? '已完成' : '学习中'}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {showReviewModal && (
        <div className="modal-overlay" onClick={() => setShowReviewModal(false)}>
          <div className="modal" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <h2 className="text-xl font-bold">评价课程</h2>
              <button onClick={() => setShowReviewModal(false)} className="text-2xl">&times;</button>
            </div>
            <div className="space-y-4">
              <div>
                <label className="block text-sm mb-2">课程：{selectedCourse?.course_name}</label>
              </div>
              <div>
                <label className="block text-sm mb-2">评分</label>
                <div className="flex space-x-2">
                  {[1, 2, 3, 4, 5].map(star => (
                    <button
                      key={star}
                      className={`text-2xl ${star <= reviewForm.rating ? 'text-yellow-400' : 'text-gray-300'}`}
                      onClick={() => setReviewForm({ ...reviewForm, rating: star })}
                    >
                      ★
                    </button>
                  ))}
                </div>
              </div>
              <div>
                <label className="block text-sm mb-2">评语（最多200字）</label>
                <textarea
                  value={reviewForm.comment}
                  onChange={(e) => setReviewForm({ ...reviewForm, comment: e.target.value.slice(0, 200) })}
                  rows={4}
                  placeholder="请输入您的评价..."
                />
                <p className="text-sm text-gray-500 mt-1">
                  {reviewForm.comment.length}/200
                </p>
              </div>
              <button 
                className="btn btn-primary w-full"
                onClick={handleSubmitReview}
              >
                提交评价
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default LearningCenter
