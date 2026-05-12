import React, { useState, useEffect } from 'react'
import { studentApi } from '../api'

function Profile({ studentId }) {
  const [analytics, setAnalytics] = useState(null)
  const [achievements, setAchievements] = useState([])
  const [recommendations, setRecommendations] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  const achievementIcons = {
    'first_course': '🎓',
    'consecutive_7_days': '🔥',
    'total_100_hours': '⏰',
    'perfect_score': '🏆',
  }

  useEffect(() => {
    loadAnalytics()
  }, [studentId])

  const loadAnalytics = async () => {
    try {
      setLoading(true)
      const res = await studentApi.getAnalytics(studentId)
      if (res.data.success) {
        setAnalytics(res.data.data.analytics)
        setAchievements(res.data.data.achievements || [])
        setRecommendations(res.data.data.recommendations || [])
      }
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div>
        <h1 className="page-title">个人中心</h1>
        <div className="card">
          <p>加载中...</p>
        </div>
      </div>
    )
  }

  return (
    <div>
      <h1 className="page-title">个人中心</h1>

      {error && (
        <div className="card">
          <p style={{ color: '#e74c3c' }}>错误: {error}</p>
        </div>
      )}

      {analytics && (
        <div className="card">
          <h3>学习统计</h3>
          <div className="grid grid-3">
            <div className="stat-card">
              <div className="stat-value">{analytics.total_learning_hours.toFixed(1)}</div>
              <div className="stat-label">总学习时长（小时）</div>
            </div>
            <div className="stat-card">
              <div className="stat-value" style={{ color: analytics.learning_speed < 0.5 ? '#e74c3c' : '#11998e' }}>
                {analytics.learning_speed}
              </div>
              <div className="stat-label">学习速度</div>
            </div>
            <div className="stat-card">
              <div className="stat-value" style={{ color: analytics.mastery_level < 60 ? '#e74c3c' : '#11998e' }}>
                {analytics.mastery_level}%
              </div>
              <div className="stat-label">掌握程度</div>
            </div>
            <div className="stat-card">
              <div className="stat-value">{analytics.average_quiz_score.toFixed(1)}%</div>
              <div className="stat-label">平均测验成绩</div>
            </div>
            <div className="stat-card">
              <div className="stat-value">{analytics.completed_courses}</div>
              <div className="stat-label">已完成课程</div>
            </div>
            <div className="stat-card">
              <div className="stat-value">{analytics.current_active_courses}</div>
              <div className="stat-label">正在学习</div>
            </div>
          </div>
        </div>
      )}

      {recommendations.length > 0 && (
        <div className="card">
          <h3>💡 学习建议</h3>
          <ul style={{ paddingLeft: '1.5rem', lineHeight: '2' }}>
            {recommendations.map((rec, index) => (
              <li key={index} style={{ color: '#e74c3c' }}>
                {rec}
              </li>
            ))}
          </ul>
        </div>
      )}

      {analytics && (
        <div className="card">
          <h3>学习指标解读</h3>
          <div style={{ lineHeight: '1.8' }}>
            <p><strong>学习速度：</strong>
              {analytics.learning_speed < 0.5 ? (
                <span style={{ color: '#e74c3c' }}>
                  偏低（{analytics.learning_speed}），建议减少同时学习的课程数量
                </span>
              ) : (
                <span style={{ color: '#11998e' }}>
                  正常（{analytics.learning_speed}），继续保持良好的学习节奏
                </span>
              )}
            </p>
            <p><strong>掌握程度：</strong>
              {analytics.mastery_level < 60 ? (
                <span style={{ color: '#e74c3c' }}>
                  偏低（{analytics.mastery_level}%），建议复习先修课程以加强基础
                </span>
              ) : (
                <span style={{ color: '#11998e' }}>
                  良好（{analytics.mastery_level}%），知识掌握扎实
                </span>
              )}
            </p>
            <p><strong>同时学习：</strong>
              {analytics.current_active_courses > 5 ? (
                <span style={{ color: '#e74c3c' }}>
                  过多（{analytics.current_active_courses}门），建议同时学习不超过5门
                </span>
              ) : (
                <span style={{ color: '#11998e' }}>
                  适中（{analytics.current_active_courses}门）
                </span>
              )}
            </p>
          </div>
        </div>
      )}

      <div className="card">
        <h3>🏅 成就徽章</h3>
        {achievements.length > 0 ? (
          <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap' }}>
            {achievements.map(achievement => (
              <div key={achievement.id} className="achievement-badge">
                <span className="achievement-icon">
                  {achievementIcons[achievement.type] || '🎖️'}
                </span>
                <div>
                  <div style={{ fontWeight: 600 }}>{achievement.name}</div>
                  <div style={{ fontSize: '0.75rem', opacity: 0.9 }}>
                    {new Date(achievement.achieved_at).toLocaleDateString()}
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div style={{ textAlign: 'center', padding: '2rem', color: '#999' }}>
            <p style={{ fontSize: '3rem', marginBottom: '1rem' }}>🎯</p>
            <p>还没有获得任何成就</p>
            <p style={{ fontSize: '0.875rem', marginTop: '0.5rem' }}>
              开始学习第一门课程，解锁你的第一个成就！
            </p>
          </div>
        )}
      </div>

      <div className="card">
        <h3>可解锁的成就</h3>
        <div className="grid grid-2">
          {[
            { icon: '🎓', name: '第一门课', desc: '开始学习第一门课程' },
            { icon: '🔥', name: '连续学习7天', desc: '连续7天每天都有学习记录' },
            { icon: '⏰', name: '累计100小时', desc: '累计学习时间达到100小时' },
            { icon: '🏆', name: '满分通过测验', desc: '在任意测验中获得满分' },
          ].map((item, index) => {
            const unlocked = achievements.some(a => 
              (a.name === item.name) || 
              (a.type === 'first_course' && item.name === '第一门课') ||
              (a.type === 'consecutive_7_days' && item.name === '连续学习7天') ||
              (a.type === 'total_100_hours' && item.name === '累计100小时') ||
              (a.type === 'perfect_score' && item.name === '满分通过测验')
            )
            return (
              <div 
                key={index} 
                style={{ 
                  padding: '1rem', 
                  borderRadius: '8px',
                  background: unlocked ? '#e8f5e9' : '#f5f5f5',
                  opacity: unlocked ? 1 : 0.6
                }}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                  <span style={{ fontSize: '1.5rem' }}>{item.icon}</span>
                  <div>
                    <h4 style={{ marginBottom: '0.25rem' }}>
                      {item.name}
                      {unlocked && <span style={{ color: '#11998e', marginLeft: '0.5rem' }}>✓</span>}
                    </h4>
                    <p style={{ fontSize: '0.875rem', color: '#666' }}>{item.desc}</p>
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      </div>

      <div className="card">
        <h3>学习提示</h3>
        <ul style={{ paddingLeft: '1.5rem', lineHeight: '2' }}>
          <li>💡 建议同时学习不超过 5 门课程</li>
          <li>📚 进度 100% 且通过测验后，课程才算完成</li>
          <li>🧠 学习速度低于 0.5 时，考虑减少课程数量</li>
          <li>📖 掌握程度低于 60% 时，建议复习先修课程</li>
          <li>🎯 坚持每天学习，解锁更多成就徽章！</li>
        </ul>
      </div>
    </div>
  )
}

export default Profile
