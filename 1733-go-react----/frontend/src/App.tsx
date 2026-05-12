import { BrowserRouter, Routes, Route, Link, useLocation } from 'react-router-dom';
import TrainingCenter from './pages/TrainingCenter';
import EvaluationManagement from './pages/EvaluationManagement';
import TitleReview from './pages/TitleReview';
import Dashboard from './pages/Dashboard';
import './App.css';

function NavLink({ to, children }: { to: string; children: React.ReactNode }) {
  const location = useLocation();
  const active = location.pathname === to;
  return (
    <Link to={to} className={active ? 'nav-link active' : 'nav-link'}>
      {children}
    </Link>
  );
}

function Layout() {
  return (
    <div className="app">
      <header className="header">
        <h1>教师发展管理系统</h1>
        <nav>
          <NavLink to="/">统计看板</NavLink>
          <NavLink to="/training">培训中心</NavLink>
          <NavLink to="/evaluation">考核管理</NavLink>
          <NavLink to="/review">职称评审</NavLink>
        </nav>
      </header>
      <main className="main">
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/training" element={<TrainingCenter />} />
          <Route path="/evaluation" element={<EvaluationManagement />} />
          <Route path="/review" element={<TitleReview />} />
        </Routes>
      </main>
    </div>
  );
}

export default function App() {
  return (
    <BrowserRouter>
      <Layout />
    </BrowserRouter>
  );
}
