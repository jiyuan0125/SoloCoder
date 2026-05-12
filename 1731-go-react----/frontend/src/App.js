import React, { useState } from 'react';
import { Routes, Route, NavLink } from 'react-router-dom';
import OccupationPage from './pages/OccupationPage';
import ExamBatchPage from './pages/ExamBatchPage';
import CandidatePage from './pages/CandidatePage';
import CertificatePage from './pages/CertificatePage';
import DashboardPage from './pages/DashboardPage';

function App() {
  return (
    <div className="app-container">
      <nav className="navbar">
        <div className="navbar-content">
          <div className="navbar-brand">职业技能鉴定系统</div>
          <div className="navbar-links">
            <NavLink to="/" className="nav-link" end>统计看板</NavLink>
            <NavLink to="/occupations" className="nav-link">职业标准</NavLink>
            <NavLink to="/batches" className="nav-link">考试批次</NavLink>
            <NavLink to="/candidates" className="nav-link">考生报名</NavLink>
            <NavLink to="/certificates" className="nav-link">证书管理</NavLink>
          </div>
        </div>
      </nav>
      <main className="main-content">
        <Routes>
          <Route path="/" element={<DashboardPage />} />
          <Route path="/occupations" element={<OccupationPage />} />
          <Route path="/batches" element={<ExamBatchPage />} />
          <Route path="/candidates" element={<CandidatePage />} />
          <Route path="/certificates" element={<CertificatePage />} />
        </Routes>
      </main>
    </div>
  );
}

export default App;
