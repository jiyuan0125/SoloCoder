import React, { useState } from 'react'
import { Routes, Route, useLocation, useNavigate } from 'react-router-dom'
import DonorsPage from './pages/DonorsPage'
import RecipientsPage from './pages/RecipientsPage'
import TransplantsPage from './pages/TransplantsPage'
import TodosPage from './pages/TodosPage'

const navItems = [
  { path: '/donors', label: '捐献者管理', icon: '❤️' },
  { path: '/recipients', label: '受体等待队列', icon: '🏥' },
  { path: '/transplants', label: '移植记录', icon: '🔬' },
  { path: '/todos', label: '待办事项', icon: '📋' }
]

function Sidebar() {
  const location = useLocation()
  const navigate = useNavigate()

  return (
    <div className="sidebar">
      <h2>器官捐献管理</h2>
      {navItems.map(item => (
        <div
          key={item.path}
          className={`nav-item ${location.pathname === item.path ? 'active' : ''}`}
          onClick={() => navigate(item.path)}
        >
          <span className="nav-icon">{item.icon}</span>
          {item.label}
        </div>
      ))}
    </div>
  )
}

export default function App() {
  const navigate = useNavigate()
  const [initialized, setInitialized] = useState(false)

  React.useEffect(() => {
    if (!initialized) {
      const path = window.location.pathname
      if (path === '/' || path === '') {
        navigate('/donors')
      }
      setInitialized(true)
    }
  }, [navigate, initialized])

  return (
    <div className="app">
      <Sidebar />
      <div className="main-content">
        <Routes>
          <Route path="/donors" element={<DonorsPage />} />
          <Route path="/recipients" element={<RecipientsPage />} />
          <Route path="/transplants" element={<TransplantsPage />} />
          <Route path="/todos" element={<TodosPage />} />
        </Routes>
      </div>
    </div>
  )
}
