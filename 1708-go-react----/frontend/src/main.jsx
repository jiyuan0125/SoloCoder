import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter, Routes, Route, Link } from 'react-router-dom'
import DonorRegistration from './pages/DonorRegistration'
import BloodTest from './pages/BloodTest'
import InventoryManagement from './pages/InventoryManagement'
import BloodRequest from './pages/BloodRequest'
import Dashboard from './pages/Dashboard'
import './App.css'

function App() {
  return (
    <BrowserRouter>
      <div className="app">
        <nav className="navbar">
          <h1 className="nav-title">🩸 血液管理系统</h1>
          <ul className="nav-links">
            <li><Link to="/">仪表盘</Link></li>
            <li><Link to="/donor-registration">献血者登记</Link></li>
            <li><Link to="/blood-test">血液检测</Link></li>
            <li><Link to="/inventory">库存管理</Link></li>
            <li><Link to="/blood-request">用血申请</Link></li>
          </ul>
        </nav>
        
        <main className="main-content">
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/donor-registration" element={<DonorRegistration />} />
            <Route path="/blood-test" element={<BloodTest />} />
            <Route path="/inventory" element={<InventoryManagement />} />
            <Route path="/blood-request" element={<BloodRequest />} />
          </Routes>
        </main>
      </div>
    </BrowserRouter>
  )
}

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
