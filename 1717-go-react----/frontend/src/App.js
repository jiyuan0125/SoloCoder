import React from 'react';
import { Routes, Route, NavLink } from 'react-router-dom';
import InfectionCasesPage from './pages/InfectionCasesPage';
import PreventionMeasuresPage from './pages/PreventionMeasuresPage';
import TargetMonitoringPage from './pages/TargetMonitoringPage';
import StatisticsPage from './pages/StatisticsPage';

function App() {
  return (
    <div className="app-container">
      <aside className="sidebar">
        <h1>医院感染监测系统</h1>
        <nav>
          <NavLink to="/" end>
            📊 统计面板
          </NavLink>
          <NavLink to="/infections">
            🦠 感染病例
          </NavLink>
          <NavLink to="/measures">
            🛡️ 防控措施
          </NavLink>
          <NavLink to="/monitoring">
            🏥 目标性监测
          </NavLink>
        </nav>
      </aside>
      <main className="main-content">
        <Routes>
          <Route path="/" element={<StatisticsPage />} />
          <Route path="/infections" element={<InfectionCasesPage />} />
          <Route path="/measures" element={<PreventionMeasuresPage />} />
          <Route path="/monitoring" element={<TargetMonitoringPage />} />
        </Routes>
      </main>
    </div>
  );
}

export default App;
