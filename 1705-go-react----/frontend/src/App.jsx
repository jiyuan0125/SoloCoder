import React, { useState, useEffect } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import Login from './pages/Login.jsx'
import GrassrootDashboard from './pages/GrassrootDashboard.jsx'
import ExpertDashboard from './pages/ExpertDashboard.jsx'
import AdminDashboard from './pages/AdminDashboard.jsx'

function App() {
  const [user, setUser] = useState(null)

  useEffect(() => {
    const savedUser = localStorage.getItem('user')
    if (savedUser) {
      setUser(JSON.parse(savedUser))
    }
  }, [])

  const handleLogin = (userData) => {
    localStorage.setItem('user', JSON.stringify(userData.user))
    localStorage.setItem('token', userData.token)
    setUser(userData.user)
  }

  const handleLogout = () => {
    localStorage.removeItem('user')
    localStorage.removeItem('token')
    setUser(null)
  }

  if (!user) {
    return <Login onLogin={handleLogin} />
  }

  const getDashboard = () => {
    switch (user.role) {
      case 'grassroot':
        return <GrassrootDashboard user={user} onLogout={handleLogout} />
      case 'expert':
        return <ExpertDashboard user={user} onLogout={handleLogout} />
      case 'admin':
        return <AdminDashboard user={user} onLogout={handleLogout} />
      default:
        return <Navigate to="/login" />
    }
  }

  return (
    <Routes>
      <Route path="*" element={getDashboard()} />
    </Routes>
  )
}

export default App
