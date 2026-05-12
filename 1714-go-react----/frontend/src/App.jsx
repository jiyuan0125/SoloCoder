import { useState, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Link, useLocation } from 'react-router-dom';
import { auditAPI } from './api';
import UnitsPage from './pages/UnitsPage';
import InspectionsPage from './pages/InspectionsPage';
import PenaltiesPage from './pages/PenaltiesPage';
import NoticesPage from './pages/NoticesPage';
import './App.css';

const navItems = [
  { path: '/units', label: '单位管理', icon: '🏢' },
  { path: '/inspections', label: '检查管理', icon: '📋' },
  { path: '/penalties', label: '处罚管理', icon: '⚖️' },
  { path: '/notices', label: '公示管理', icon: '📢' },
  { path: '/audit', label: '审计日志', icon: '📝' },
];

function AuditPage() {
  const [logs, setLogs] = useState([]);
  const [filterAction, setFilterAction] = useState('');
  const [filterResource, setFilterResource] = useState('');

  useEffect(() => {
    loadLogs();
  }, [filterAction, filterResource]);

  const loadLogs = async () => {
    try {
      const params = {};
      if (filterAction) params.action = filterAction;
      if (filterResource) params.resource = filterResource;
      const res = await auditAPI.list(params);
      setLogs(res.data.data);
    } catch (err) {
      console.error('Failed to load audit logs:', err);
    }
  };

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-800 mb-2">审计日志</h1>
        <p className="text-sm text-gray-500">审计日志只能查看，不可修改或删除</p>
      </div>

      <div className="flex flex-wrap gap-4 mb-6">
        <div className="flex gap-4 items-center">
          <label className="text-gray-600">操作类型:</label>
          <select
            value={filterAction}
            onChange={(e) => setFilterAction(e.target.value)}
            className="border rounded px-3 py-2"
          >
            <option value="">全部</option>
            <option value="创建">创建</option>
            <option value="更新">更新</option>
            <option value="删除">删除</option>
            <option value="下达">下达</option>
            <option value="复查">复查</option>
            <option value="发布">发布</option>
          </select>
        </div>
        <div className="flex gap-4 items-center">
          <label className="text-gray-600">资源类型:</label>
          <select
            value={filterResource}
            onChange={(e) => setFilterResource(e.target.value)}
            className="border rounded px-3 py-2"
          >
            <option value="">全部</option>
            <option value="被监督单位">被监督单位</option>
            <option value="检查记录">检查记录</option>
            <option value="卫生监督意见书">卫生监督意见书</option>
            <option value="行政处罚">行政处罚</option>
            <option value="公示">公示</option>
          </select>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">时间</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">用户</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">操作</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">资源</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">描述</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">IP</th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {logs.map((log) => (
              <tr key={log.id}>
                <td className="px-4 py-3 whitespace-nowrap text-sm">
                  {new Date(log.created_at).toLocaleString()}
                </td>
                <td className="px-4 py-3 whitespace-nowrap text-sm">{log.username || 'system'}</td>
                <td className="px-4 py-3 whitespace-nowrap text-sm">
                  <span className="px-2 py-1 rounded text-xs font-medium bg-blue-100 text-blue-800">
                    {log.action}
                  </span>
                </td>
                <td className="px-4 py-3 whitespace-nowrap text-sm">{log.resource}</td>
                <td className="px-4 py-3 text-sm">{log.description}</td>
                <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-500">{log.ip_address}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function Sidebar() {
  const location = useLocation();

  return (
    <aside className="w-64 bg-slate-800 text-white min-h-screen">
      <div className="p-4 border-b border-slate-700">
        <h1 className="text-xl font-bold">卫生监督管理系统</h1>
      </div>
      <nav className="p-4">
        <ul className="space-y-2">
          {navItems.map((item) => (
            <li key={item.path}>
              <Link
                to={item.path}
                className={`flex items-center gap-3 px-4 py-3 rounded-lg transition-colors ${
                  location.pathname.startsWith(item.path)
                    ? 'bg-blue-600 text-white'
                    : 'text-slate-300 hover:bg-slate-700'
                }`}
              >
                <span>{item.icon}</span>
                <span>{item.label}</span>
              </Link>
            </li>
          ))}
        </ul>
      </nav>
    </aside>
  );
}

function App() {
  return (
    <Router>
      <div className="flex min-h-screen bg-gray-100">
        <Sidebar />
        <main className="flex-1 overflow-auto">
          <Routes>
            <Route path="/" element={<UnitsPage />} />
            <Route path="/units" element={<UnitsPage />} />
            <Route path="/inspections" element={<InspectionsPage />} />
            <Route path="/penalties" element={<PenaltiesPage />} />
            <Route path="/notices" element={<NoticesPage />} />
            <Route path="/audit" element={<AuditPage />} />
          </Routes>
        </main>
      </div>
    </Router>
  );
}

export default App;
