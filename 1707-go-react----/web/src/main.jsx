import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter, Routes, Route, Link } from 'react-router-dom'
import DeviceList from './pages/DeviceList'
import MaintenancePlans from './pages/MaintenancePlans'
import WorkOrders from './pages/WorkOrders'
import Calibration from './pages/Calibration'
import './index.css'

function App() {
  return (
    <BrowserRouter>
      <div className="app">
        <header className="app-header">
          <h1>医疗设备管理系统</h1>
          <nav>
            <Link to="/">设备台账</Link>
            <Link to="/maintenance">维护计划</Link>
            <Link to="/workorders">工单管理</Link>
            <Link to="/calibration">校准管理</Link>
          </nav>
        </header>
        <main>
          <Routes>
            <Route path="/" element={<DeviceList />} />
            <Route path="/maintenance" element={<MaintenancePlans />} />
            <Route path="/workorders" element={<WorkOrders />} />
            <Route path="/calibration" element={<Calibration />} />
          </Routes>
        </main>
      </div>
    </BrowserRouter>
  )
}

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
)
