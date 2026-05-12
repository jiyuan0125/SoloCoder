import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter, Routes, Route, Link, Navigate } from 'react-router-dom'
import DrugCatalog from './pages/DrugCatalog'
import StockManagement from './pages/StockManagement'
import PrescriptionManagement from './pages/PrescriptionManagement'
import './styles/App.css'

function App() {
  return (
    <div className="app">
      <header className="app-header">
        <h1>医院药房管理系统</h1>
        <nav>
          <Link to="/drugs">药品目录</Link>
          <Link to="/stock">库存管理</Link>
          <Link to="/prescriptions">处方管理</Link>
        </nav>
      </header>
      <main className="app-main">
        <Routes>
          <Route path="/" element={<Navigate to="/stock" />} />
          <Route path="/drugs" element={<DrugCatalog />} />
          <Route path="/stock" element={<StockManagement />} />
          <Route path="/prescriptions" element={<PrescriptionManagement />} />
        </Routes>
      </main>
    </div>
  )
}

ReactDOM.createRoot(document.getElementById('root')).render(
  <BrowserRouter>
    <App />
  </BrowserRouter>
)
