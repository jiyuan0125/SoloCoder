import React from 'react'
import { Routes, Route, NavLink } from 'react-router-dom'
import IndicatorsPage from './pages/indicators/IndicatorsPage.jsx'
import PDCAPage from './pages/pdca/PDCAPage.jsx'
import TodosPage from './pages/todos/TodosPage.jsx'
import DashboardPage from './pages/dashboard/DashboardPage.jsx'

function App() {
  return (
    <div className="app-container">
      <nav className="navbar">
        <div className="navbar-content">
          <div className="navbar-brand">医疗质量管理系统</div>
          <div className="navbar-nav">
            <NavLink to="/" className="nav-link" end>指标管理</NavLink>
            <NavLink to="/pdca" className="nav-link">PDCA管理</NavLink>
            <NavLink to="/todos" className="nav-link">待办事项</NavLink>
            <NavLink to="/dashboard" className="nav-link">聚合面板</NavLink>
          </div>
        </div>
      </nav>

      <main className="main-content">
        <Routes>
          <Route path="/" element={<IndicatorsPage />} />
          <Route path="/pdca" element={<PDCAPage />} />
          <Route path="/todos" element={<TodosPage />} />
          <Route path="/dashboard" element={<DashboardPage />} />
        </Routes>
      </main>
    </div>
  )
}

export default App
