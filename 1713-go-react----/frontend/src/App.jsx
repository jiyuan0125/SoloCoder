import { Routes, Route, NavLink } from 'react-router-dom';
import EnterprisePage from './pages/EnterprisePage';
import ExaminationPage from './pages/ExaminationPage';
import ReportPage from './pages/ReportPage';
import StatsPage from './pages/StatsPage';

export default function App() {
  return (
    <div className="layout">
      <aside className="sidebar">
        <h2>职业病防治管理系统</h2>
        <nav>
          <NavLink to="/" end>企业管理</NavLink>
          <NavLink to="/examinations">体检管理</NavLink>
          <NavLink to="/reports">报告管理</NavLink>
          <NavLink to="/stats">统计面板</NavLink>
        </nav>
      </aside>
      <main className="main-content">
        <Routes>
          <Route path="/" element={<EnterprisePage />} />
          <Route path="/examinations" element={<ExaminationPage />} />
          <Route path="/reports" element={<ReportPage />} />
          <Route path="/stats" element={<StatsPage />} />
        </Routes>
      </main>
    </div>
  );
}
