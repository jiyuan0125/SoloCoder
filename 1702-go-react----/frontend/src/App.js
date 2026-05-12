import { BrowserRouter as Router, Routes, Route, NavLink } from 'react-router-dom';
import AppointmentCalendar from './pages/AppointmentCalendar';
import ClinicWorkbench from './pages/ClinicWorkbench';
import FollowUpManagement from './pages/FollowupManagement';

export default function App() {
  return (
    <Router>
      <div className="app">
        <nav className="navbar">
          <ul>
            <li>
              <NavLink to="/" end className={({ isActive }) => isActive ? 'active' : ''}>
                预约日历
              </NavLink>
            </li>
            <li>
              <NavLink to="/workbench" className={({ isActive }) => isActive ? 'active' : ''}>
                诊疗工作台
              </NavLink>
            </li>
            <li>
              <NavLink to="/followup" className={({ isActive }) => isActive ? 'active' : ''}>
                回访管理
              </NavLink>
            </li>
          </ul>
        </nav>

        <Routes>
          <Route path="/" element={<AppointmentCalendar />} />
          <Route path="/workbench" element={<ClinicWorkbench />} />
          <Route path="/followup" element={<FollowUpManagement />} />
        </Routes>
      </div>
    </Router>
  );
}
