import { BrowserRouter, Routes, Route, NavLink } from 'react-router-dom';
import Dashboard from './pages/Dashboard';
import VehicleManagement from './pages/VehicleManagement';
import Statistics from './pages/Statistics';

function App() {
  return (
    <BrowserRouter>
      <div className="app-container">
        <div className="sidebar">
          <div className="sidebar-logo">
            🚑 120急救调度系统
          </div>
          <nav className="sidebar-nav">
            <NavLink 
              to="/" 
              end
              className={({ isActive }) => isActive ? 'active' : ''}
            >
              调度台
            </NavLink>
            <NavLink 
              to="/vehicles" 
              className={({ isActive }) => isActive ? 'active' : ''}
            >
              车辆管理
            </NavLink>
            <NavLink 
              to="/statistics" 
              className={({ isActive }) => isActive ? 'active' : ''}
            >
              统计面板
            </NavLink>
          </nav>
        </div>

        <div className="main-content">
          <div className="page-header">
            <PageTitle />
          </div>
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/vehicles" element={<VehicleManagement />} />
            <Route path="/statistics" element={<Statistics />} />
          </Routes>
        </div>
      </div>
    </BrowserRouter>
  );
}

function PageTitle() {
  const pathname = window.location.pathname;
  let title = '调度台';
  
  if (pathname === '/vehicles') {
    title = '车辆管理';
  } else if (pathname === '/statistics') {
    title = '统计面板';
  }

  return <h1>{title}</h1>;
}

export default App;
