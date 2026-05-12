import React, { useState, useEffect } from 'react'
import { Routes, Route, Navigate, useNavigate } from 'react-router-dom'
import SampleList from './pages/SampleList.jsx'
import TestDataEntry from './pages/TestDataEntry.jsx'
import ReportReview from './pages/ReportReview.jsx'
import SubmitterQuery from './pages/SubmitterQuery.jsx'
import { AuthContext } from './context/AuthContext.jsx'

const ROLES = {
  admin: { label: '系统管理员', value: 'admin' },
  reviewer: { label: '报告审核员', value: 'reviewer' },
  technician: { label: '实验室技术员', value: 'technician' },
  submitter: { label: '送检单位', value: 'submitter' },
}

const UNITS = [
  { id: 'unit001', name: '北京协和医院' },
  { id: 'unit002', name: '上海瑞金医院' },
  { id: 'unit003', name: '广州中山医院' },
]

function App() {
  const [role, setRole] = useState('admin')
  const [unitID, setUnitID] = useState('unit001')
  const navigate = useNavigate()

  useEffect(() => {
    fetch('/api/units').catch(() => {})
  }, [])

  const handleRoleChange = (newRole) => {
    setRole(newRole)
    if (newRole === 'submitter') {
      navigate('/submitter')
    } else if (newRole === 'technician') {
      navigate('/test-entry')
    } else if (newRole === 'reviewer') {
      navigate('/report-review')
    } else {
      navigate('/samples')
    }
  }

  const navItems = [
    { path: '/samples', label: '样本管理', roles: ['admin', 'reviewer', 'technician'] },
    { path: '/test-entry', label: '检测录入', roles: ['admin', 'technician'] },
    { path: '/report-review', label: '报告审核', roles: ['admin', 'reviewer'] },
    { path: '/submitter', label: '送检查询', roles: ['admin', 'submitter'] },
  ]

  const visibleNav = navItems.filter(item => item.roles.includes(role))

  return (
    <AuthContext.Provider value={{ role, setRole, unitID, setUnitID }}>
      <div className="min-h-screen bg-gray-50">
        <header className="bg-white shadow-sm border-b border-gray-200">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex justify-between items-center h-16">
              <div className="flex items-center space-x-8">
                <h1 className="text-xl font-bold text-primary-700">
                  🔬 GeneDeck - 基因检测样本管理系统
                </h1>
                {visibleNav.length > 0 && (
                  <nav className="hidden md:flex space-x-4">
                    {visibleNav.map(item => (
                      <button
                        key={item.path}
                        onClick={() => navigate(item.path)}
                        className="px-3 py-2 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-100 hover:text-gray-900 transition-colors"
                      >
                        {item.label}
                      </button>
                    ))}
                  </nav>
                )}
              </div>
              <div className="flex items-center space-x-4">
                {role === 'submitter' && (
                  <select
                    value={unitID}
                    onChange={(e) => setUnitID(e.target.value)}
                    className="border border-gray-300 rounded-md px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
                  >
                    {UNITS.map(u => (
                      <option key={u.id} value={u.id}>{u.name}</option>
                    ))}
                  </select>
                )}
                <div className="flex items-center space-x-2">
                  <span className="text-sm text-gray-500">当前角色:</span>
                  <select
                    value={role}
                    onChange={(e) => handleRoleChange(e.target.value)}
                    className="border border-gray-300 rounded-md px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 bg-white"
                  >
                    {Object.values(ROLES).map(r => (
                      <option key={r.value} value={r.value}>{r.label}</option>
                    ))}
                  </select>
                </div>
              </div>
            </div>
          </div>
        </header>

        <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <Routes>
            <Route path="/" element={<Navigate to="/samples" replace />} />
            <Route path="/samples" element={<SampleList />} />
            <Route path="/test-entry" element={<TestDataEntry />} />
            <Route path="/report-review" element={<ReportReview />} />
            <Route path="/submitter" element={<SubmitterQuery />} />
          </Routes>
        </main>
      </div>
    </AuthContext.Provider>
  )
}

export default App
