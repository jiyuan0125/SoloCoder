import React from 'react';
import { Routes, Route, NavLink } from 'react-router-dom';
import ProjectKanban from './pages/ProjectKanban';
import ProjectDetail from './pages/ProjectDetail';
import DataCenter from './pages/DataCenter';
import Achievements from './pages/Achievements';

export default function App() {
  return (
    <div className="app-container">
      <nav className="navbar">
        <div className="navbar-brand">科研协作管理平台</div>
        <div className="navbar-links">
          <NavLink to="/" end>项目看板</NavLink>
          <NavLink to="/data">数据共享中心</NavLink>
          <NavLink to="/achievements">成果管理</NavLink>
        </div>
      </nav>

      <main className="main-content">
        <Routes>
          <Route path="/" element={<ProjectKanban />} />
          <Route path="/projects/:id" element={<ProjectDetail />} />
          <Route path="/data" element={<DataCenter />} />
          <Route path="/achievements" element={<Achievements />} />
        </Routes>
      </main>
    </div>
  );
}
