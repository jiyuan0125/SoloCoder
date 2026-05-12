import React, { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { courseAPI, paymentAPI, subjects, courseTypes } from '../services/api'

const CourseHall = ({ currentStudent, setCurrentStudent }) => {
  const [courses, setCourses] = useState([])
  const [filters, setFilters] = useState({ subject: '', type: '' })
  const [showRechargeModal, setShowRechargeModal] = useState(false)
  const [rechargeForm, setRechargeForm] = useState({ cardNumber: '', password: '' })
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState('')

  useEffect(() => {
    loadCourses()
  }, [filters])

  const loadCourses = async () => {
    try {
      const params = {}
      if (filters.subject) params.subject = filters.subject
      if (filters.type) params.type = filters.type
      const res = await courseAPI.getAll(params)
      setCourses(res.data)
    } catch (error) {
      console.error('Failed to load courses:', error)
    }
  }

  const handleEnroll = async (course) => {
    if (course.price > currentStudent.balance) {
      setMessage('余额不足，请先充值')
      setTimeout(() => setMessage(''), 3000)
      return
    }

    try {
      await paymentAPI.enroll(currentStudent.id, course.id)
      setCurrentStudent({
        ...currentStudent,
        balance: currentStudent.balance - course.price
      })
      setMessage(`成功选报课程：${course.name}`)
      setTimeout(() => setMessage(''), 3000)
    } catch (error) {
      setMessage(error.response?.data?.message || '选课失败')
      setTimeout(() => setMessage(''), 3000)
    }
  }

  const handleRecharge = async () => {
    if (!rechargeForm.cardNumber || !rechargeForm.password) {
      setMessage('请填写完整信息')
      return
    }

    setLoading(true)
    try {
      const res = await paymentAPI.recharge(currentStudent.id, rechargeForm.cardNumber, rechargeForm.password)
      const amount = res.data.amount
      setCurrentStudent({
        ...currentStudent,
        balance: currentStudent.balance + amount
      })
      setMessage(`充值成功：¥${(amount / 100).toFixed(2)}`)
      setShowRechargeModal(false)
      setRechargeForm({ cardNumber: '', password: '' })
      setTimeout(() => setMessage(''), 3000)
    } catch (error) {
      setMessage(error.response?.data?.message || '充值失败')
    }
    setLoading(false)
  }

  const getSubjectName = (subject) => {
    const found = subjects.find(s => s.id === subject)
    return found ? found.name : subject
  }

  const formatPrice = (price) => {
    if (price === 0) return '免费'
    return `¥${(price / 100).toFixed(2)}`
  }

  return (
    <div>
      {message && (
        <div className="card bg-blue-100 mb-4">
          <p className="text-sm">{message}</p>
        </div>
      )}

      <div className="flex justify-between items-center mb-6">
        <h1 className="text-xl font-bold">课程大厅</h1>
        <button 
          className="btn btn-success"
          onClick={() => setShowRechargeModal(true)}
        >
          充值
        </button>
      </div>

      <div className="card mb-6">
        <div className="flex space-x-4">
          <select 
            className="flex-1"
            value={filters.type}
            onChange={(e) => setFilters({ ...filters, type: e.target.value })}
          >
            <option value="">全部类型</option>
            {courseTypes.map(t => (
              <option key={t.id} value={t.id}>{t.name}</option>
            ))}
          </select>
          <select 
            className="flex-1"
            value={filters.subject}
            onChange={(e) => setFilters({ ...filters, subject: e.target.value })}
          >
            <option value="">全部学科</option>
            {subjects.map(s => (
              <option key={s.id} value={s.id}>{s.name}</option>
            ))}
          </select>
        </div>
      </div>

      <div className="grid grid-cols-3">
        {courses.map(course => (
          <div key={course.id} className="card">
            <div className="flex items-center mb-2">
              <span className={`tag ${course.course_type === 'live' ? 'tag-live' : 'tag-record'}`}>
                {course.course_type === 'live' ? '直播课' : '录播课'}
              </span>
              {course.price === 0 && <span className="tag tag-free">免费</span>}
            </div>
            <h3 className="text-lg font-bold mb-2">{course.name}</h3>
            <p className="text-sm text-gray-500 mb-2">讲师：{course.instructor}</p>
            {course.subject && (
              <p className="text-sm text-gray-500 mb-2">学科：{getSubjectName(course.subject)}</p>
            )}
            {course.course_type === 'live' && course.start_time && (
              <p className="text-sm text-gray-500 mb-2">
                开始时间：{new Date(course.start_time).toLocaleString()}
              </p>
            )}
            {course.course_type === 'record' && course.video_duration && (
              <p className="text-sm text-gray-500 mb-2">
                视频时长：{Math.floor(course.video_duration / 60)}分钟
              </p>
            )}
            <div className="flex justify-between items-center mt-4">
              <span className="text-lg font-bold text-red-500">{formatPrice(course.price)}</span>
              {course.course_type === 'live' ? (
                <Link to={`/live/${course.id}`} className="btn btn-primary">
                  进入课堂
                </Link>
              ) : (
                <button 
                  className="btn btn-primary"
                  onClick={() => handleEnroll(course)}
                >
                  立即选报
                </button>
              )}
            </div>
          </div>
        ))}
      </div>

      {courses.length === 0 && (
        <div className="card text-center py-12 text-gray-500">
          暂无课程
        </div>
      )}

      {showRechargeModal && (
        <div className="modal-overlay" onClick={() => setShowRechargeModal(false)}>
          <div className="modal" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <h2 className="text-xl font-bold">充值</h2>
              <button onClick={() => setShowRechargeModal(false)} className="text-2xl">&times;</button>
            </div>
            <div className="space-y-4">
              <div>
                <label className="block text-sm mb-1">卡号</label>
                <input
                  type="text"
                  value={rechargeForm.cardNumber}
                  onChange={(e) => setRechargeForm({ ...rechargeForm, cardNumber: e.target.value })}
                  placeholder="请输入卡号，例如：CARD001"
                />
              </div>
              <div>
                <label className="block text-sm mb-1">密码</label>
                <input
                  type="password"
                  value={rechargeForm.password}
                  onChange={(e) => setRechargeForm({ ...rechargeForm, password: e.target.value })}
                  placeholder="请输入密码，例如：PASS123"
                />
              </div>
              <div className="bg-gray-50 p-3 rounded text-sm text-gray-600">
                <p>测试卡号：</p>
                <p>CARD001 / PASS123 (50元)</p>
                <p>CARD002 / PASS456 (100元)</p>
                <p>CARD003 / PASS789 (200元)</p>
              </div>
              <button 
                className="btn btn-primary w-full"
                onClick={handleRecharge}
                disabled={loading}
              >
                {loading ? '充值中...' : '确认充值'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default CourseHall
