import React, { useState, useEffect } from 'react'
import { statsAPI, courseAPI, studentAPI } from '../services/api'

const AdminPanel = () => {
  const [subjectStats, setSubjectStats] = useState([])
  const [instructorStats, setInstructorStats] = useState([])
  const [liveStats, setLiveStats] = useState([])
  const [courses, setCourses] = useState([])
  const [students, setStudents] = useState([])
  const [activeTab, setActiveTab] = useState('overview')

  useEffect(() => {
    loadAllData()
  }, [])

  const loadAllData = async () => {
    try {
      const [subjects, instructors, lives, courseList, studentList] = await Promise.all([
        statsAPI.getSubjects(),
        statsAPI.getInstructors(),
        statsAPI.getLiveStats(),
        courseAPI.getAll(),
        studentAPI.getAll()
      ])
      setSubjectStats(subjects.data)
      setInstructorStats(instructors.data)
      setLiveStats(lives.data)
      setCourses(courseList.data)
      setStudents(studentList.data)
    } catch (error) {
      console.error('Failed to load data:', error)
    }
  }

  const updateLiveStatus = async (courseId, status) => {
    try {
      await courseAPI.updateStatus(courseId, status)
      loadAllData()
    } catch (error) {
      console.error('Failed to update status:', error)
    }
  }

  const getStatusText = (status) => {
    const map = {
      'not_started': '未开始',
      'in_progress': '进行中',
      'ended': '已结束',
      'has_replay': '已生成回放'
    }
    return map[status] || status
  }

  const getStatusColor = (status) => {
    const map = {
      'not_started': 'text-gray-500',
      'in_progress': 'text-green-500',
      'ended': 'text-yellow-500',
      'has_replay': 'text-blue-500'
    }
    return map[status] || 'text-gray-500'
  }

  return (
    <div>
      <h1 className="text-xl font-bold mb-6">管理面板</h1>

      <div className="flex space-x-2 mb-6">
        {['overview', 'courses', 'students', 'live'].map(tab => (
          <button
            key={tab}
            className={`btn ${activeTab === tab ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setActiveTab(tab)}
          >
            {tab === 'overview' && '总览'}
            {tab === 'courses' && '课程管理'}
            {tab === 'students' && '学生管理'}
            {tab === 'live' && '直播数据'}
          </button>
        ))}
      </div>

      {activeTab === 'overview' && (
        <div className="grid grid-cols-2">
          <div className="card">
            <h2 className="font-bold mb-4">学科课程和选课人数排名</h2>
            {subjectStats.length === 0 ? (
              <p className="text-gray-500">暂无数据</p>
            ) : (
              <table>
                <thead>
                  <tr>
                    <th>排名</th>
                    <th>学科</th>
                    <th>课程数</th>
                    <th>选课人数</th>
                  </tr>
                </thead>
                <tbody>
                  {subjectStats.map((s, i) => (
                    <tr key={s.subject}>
                      <td>{i + 1}</td>
                      <td>{s.subject_name}</td>
                      <td>{s.course_count}</td>
                      <td>{s.enrollment_count}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>

          <div className="card">
            <h2 className="font-bold mb-4">讲师课程数和平均评分</h2>
            {instructorStats.length === 0 ? (
              <p className="text-gray-500">暂无数据</p>
            ) : (
              <table>
                <thead>
                  <tr>
                    <th>讲师</th>
                    <th>课程数</th>
                    <th>平均评分</th>
                  </tr>
                </thead>
                <tbody>
                  {instructorStats.map(inst => (
                    <tr key={inst.instructor_name}>
                      <td>{inst.instructor_name}</td>
                      <td>{inst.course_count}</td>
                      <td>
                        {inst.avg_rating ? (
                          <span>{inst.avg_rating.toFixed(1)} ★</span>
                        ) : (
                          <span className="text-gray-400">暂无评价</span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>

          <div className="card">
            <h2 className="font-bold mb-4">直播课参与数据</h2>
            {liveStats.length === 0 ? (
              <p className="text-gray-500">暂无已结束的直播</p>
            ) : (
              <table>
                <thead>
                  <tr>
                    <th>课程</th>
                    <th>峰值人数</th>
                    <th>平均在线</th>
                    <th>弹幕数</th>
                    <th>投票参与率</th>
                  </tr>
                </thead>
                <tbody>
                  {liveStats.map(live => (
                    <tr key={live.id}>
                      <td>{live.course_name}</td>
                      <td>{live.peak_online}</td>
                      <td>{live.average_online_time.toFixed(0)}秒</td>
                      <td>{live.danmaku_count}</td>
                      <td>{live.vote_participation.toFixed(1)}%</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>

          <div className="card">
            <h2 className="font-bold mb-4">统计概览</h2>
            <div className="grid grid-cols-2 gap-4">
              <div className="bg-blue-100 p-4 rounded text-center">
                <div className="text-3xl font-bold text-blue-600">{courses.length}</div>
                <div className="text-sm text-gray-600">课程总数</div>
              </div>
              <div className="bg-green-100 p-4 rounded text-center">
                <div className="text-3xl font-bold text-green-600">{students.length}</div>
                <div className="text-sm text-gray-600">学生总数</div>
              </div>
              <div className="bg-yellow-100 p-4 rounded text-center">
                <div className="text-3xl font-bold text-yellow-600">
                  {courses.filter(c => c.course_type === 'live').length}
                </div>
                <div className="text-sm text-gray-600">直播课程</div>
              </div>
              <div className="bg-red-100 p-4 rounded text-center">
                <div className="text-3xl font-bold text-red-600">
                  {courses.filter(c => c.course_type === 'record').length}
                </div>
                <div className="text-sm text-gray-600">录播课程</div>
              </div>
            </div>
          </div>
        </div>
      )}

      {activeTab === 'courses' && (
        <div className="card">
          <h2 className="font-bold mb-4">课程管理</h2>
          <table>
            <thead>
              <tr>
                <th>ID</th>
                <th>课程名称</th>
                <th>类型</th>
                <th>讲师</th>
                <th>价格</th>
                <th>状态</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {courses.map(course => (
                <tr key={course.id}>
                  <td>{course.id}</td>
                  <td>{course.name}</td>
                  <td>
                    <span className={`tag ${course.course_type === 'live' ? 'tag-live' : 'tag-record'}`}>
                      {course.course_type === 'live' ? '直播' : '录播'}
                    </span>
                  </td>
                  <td>{course.instructor}</td>
                  <td>
                    {course.price === 0 ? '免费' : `¥${(course.price / 100).toFixed(2)}`}
                  </td>
                  <td>
                    {course.live_status ? (
                      <span className={getStatusColor(course.live_status)}>
                        {getStatusText(course.live_status)}
                      </span>
                    ) : (
                      <span className="text-gray-400">-</span>
                    )}
                  </td>
                  <td>
                    {course.course_type === 'live' && (
                      <div className="flex space-x-1">
                        <button 
                          className="btn btn-success"
                          onClick={() => updateLiveStatus(course.id, 'in_progress')}
                          disabled={course.live_status === 'in_progress'}
                        >
                          开始
                        </button>
                        <button 
                          className="btn btn-secondary"
                          onClick={() => updateLiveStatus(course.id, 'ended')}
                          disabled={course.live_status !== 'in_progress'}
                        >
                          结束
                        </button>
                        <button 
                          className="btn btn-primary"
                          onClick={() => updateLiveStatus(course.id, 'has_replay')}
                          disabled={course.live_status !== 'ended'}
                        >
                          生成回放
                        </button>
                      </div>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {activeTab === 'students' && (
        <div className="card">
          <h2 className="font-bold mb-4">学生管理</h2>
          <table>
            <thead>
              <tr>
                <th>ID</th>
                <th>昵称</th>
                <th>手机号</th>
                <th>年级</th>
                <th>余额</th>
                <th>注册时间</th>
              </tr>
            </thead>
            <tbody>
              {students.map(student => (
                <tr key={student.id}>
                  <td>{student.id}</td>
                  <td>{student.nickname}</td>
                  <td>{student.phone}</td>
                  <td>{student.grade}</td>
                  <td>¥{(student.balance / 100).toFixed(2)}</td>
                  <td>{new Date(student.created_at).toLocaleDateString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {activeTab === 'live' && (
        <div className="card">
          <h2 className="font-bold mb-4">直播数据统计</h2>
          {liveStats.length === 0 ? (
            <p className="text-gray-500">暂无已结束的直播数据。请先在"课程管理"中开始并结束一个直播。</p>
          ) : (
            <table>
              <thead>
                <tr>
                  <th>课程</th>
                  <th>峰值在线人数</th>
                  <th>平均在线时长</th>
                  <th>弹幕总数</th>
                  <th>投票参与率</th>
                </tr>
              </thead>
              <tbody>
                {liveStats.map(live => (
                  <tr key={live.id}>
                    <td>{live.course_name}</td>
                    <td>{live.peak_online}</td>
                    <td>{live.average_online_time.toFixed(1)}秒</td>
                    <td>{live.danmaku_count}</td>
                    <td>{live.vote_participation.toFixed(1)}%</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}
    </div>
  )
}

export default AdminPanel
