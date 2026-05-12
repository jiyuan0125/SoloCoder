import { BrowserRouter as Router, Routes, Route, Link, useLocation } from 'react-router-dom';
import PathManagement from './pages/PathManagement';
import PatientEnrollment from './pages/PatientEnrollment';
import DailyExecution from './pages/DailyExecution';
import QualityPanel from './pages/QualityPanel';

function NavLink({ to, children }) {
  const location = useLocation();
  const isActive = location.pathname === to;
  return (
    <Link
      to={to}
      className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
        isActive
          ? 'bg-blue-600 text-white'
          : 'text-gray-600 hover:bg-gray-100'
      }`}
    >
      {children}
    </Link>
  );
}

function App() {
  return (
    <Router>
      <div className="min-h-screen bg-gray-100">
        <nav className="bg-white shadow-sm">
          <div className="max-w-7xl mx-auto px-4 py-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <div className="w-8 h-8 bg-blue-600 rounded flex items-center justify-center">
                  <span className="text-white font-bold">CP</span>
                </div>
                <h1 className="text-lg font-bold text-gray-800">临床路径管理系统</h1>
              </div>
              <div className="flex gap-1">
                <NavLink to="/">路径管理</NavLink>
                <NavLink to="/patients">患者入径</NavLink>
                <NavLink to="/execution">每日执行</NavLink>
                <NavLink to="/quality">质控面板</NavLink>
              </div>
            </div>
          </div>
        </nav>

        <main className="max-w-7xl mx-auto">
          <Routes>
            <Route path="/" element={<PathManagement />} />
            <Route path="/patients" element={<PatientEnrollment />} />
            <Route path="/execution" element={<DailyExecution />} />
            <Route path="/quality" element={<QualityPanel />} />
          </Routes>
        </main>
      </div>
    </Router>
  );
}

export default App;
