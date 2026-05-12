import { BrowserRouter as Router, Routes, Route, NavLink, Navigate } from 'react-router-dom';
import ProtocolList from './pages/ProtocolList';
import ProtocolForm from './pages/ProtocolForm';
import SubjectList from './pages/SubjectList';
import DataEntry from './pages/DataEntry';
import AEPage from './pages/AEPage';

function Navbar() {
  return (
    <nav className="bg-gradient-to-r from-blue-700 to-blue-900 text-white shadow-lg">
      <div className="max-w-7xl mx-auto px-4">
        <div className="flex items-center justify-between h-16">
          <div className="flex items-center">
            <span className="text-xl font-bold">临床试验管理系统</span>
          </div>
          <div className="flex items-center space-x-1">
            <NavLink
              to="/protocols"
              className={({ isActive }) =>
                `px-4 py-2 rounded-md transition ${isActive ? 'bg-blue-600' : 'hover:bg-blue-600'}`
              }
            >
              方案管理
            </NavLink>
            <NavLink
              to="/subjects"
              className={({ isActive }) =>
                `px-4 py-2 rounded-md transition ${isActive ? 'bg-blue-600' : 'hover:bg-blue-600'}`
              }
            >
              受试者管理
            </NavLink>
            <NavLink
              to="/aes"
              className={({ isActive }) =>
                `px-4 py-2 rounded-md transition ${isActive ? 'bg-blue-600' : 'hover:bg-blue-600'}`
              }
            >
              AE管理
            </NavLink>
          </div>
        </div>
      </div>
    </nav>
  );
}

function HomePage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-gray-50 to-blue-50">
      <div className="text-center max-w-2xl px-8">
        <h1 className="text-4xl font-bold text-gray-800 mb-4">
          临床试验管理系统
        </h1>
        <p className="text-lg text-gray-600 mb-8">
          完整的CRO临床试验解决方案，支持方案管理、受试者入组、数据采集和AE管理
        </p>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <NavLink
            to="/protocols"
            className="p-6 bg-white rounded-xl shadow-md hover:shadow-lg transition border border-gray-100"
          >
            <div className="text-4xl mb-3">📋</div>
            <div className="font-semibold text-gray-800">方案管理</div>
            <div className="text-sm text-gray-500 mt-1">创建和管理试验方案</div>
          </NavLink>
          <NavLink
            to="/subjects"
            className="p-6 bg-white rounded-xl shadow-md hover:shadow-lg transition border border-gray-100"
          >
            <div className="text-4xl mb-3">👥</div>
            <div className="font-semibold text-gray-800">受试者管理</div>
            <div className="text-sm text-gray-500 mt-1">入组、退组、状态跟踪</div>
          </NavLink>
          <NavLink
            to="/aes"
            className="p-6 bg-white rounded-xl shadow-md hover:shadow-lg transition border border-gray-100"
          >
            <div className="text-4xl mb-3">⚠️</div>
            <div className="font-semibold text-gray-800">AE管理</div>
            <div className="text-sm text-gray-500 mt-1">记录和跟踪不良事件</div>
          </NavLink>
        </div>
      </div>
    </div>
  );
}

function App() {
  return (
    <Router>
      <div className="min-h-screen bg-gray-50">
        <Navbar />
        <main>
          <Routes>
            <Route path="/" element={<HomePage />} />
            <Route path="/protocols" element={<ProtocolList />} />
            <Route path="/protocols/new" element={<ProtocolForm />} />
            <Route path="/protocols/:id" element={<ProtocolForm />} />
            <Route path="/protocols/:id/edit" element={<ProtocolForm />} />
            <Route path="/subjects" element={<SubjectList />} />
            <Route path="/data-entry/:subjectId" element={<DataEntry />} />
            <Route path="/ae" element={<AEPage />} />
            <Route path="/ae/:subjectId" element={<AEPage />} />
            <Route path="/aes" element={<AEPage />} />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </main>
      </div>
    </Router>
  );
}

export default App;
