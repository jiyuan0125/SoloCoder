import { useState, useEffect } from 'react';
import { subjectAPI, protocolAPI } from '../services/api';
import { Link } from 'react-router-dom';

function SubjectList() {
  const [subjects, setSubjects] = useState([]);
  const [protocols, setProtocols] = useState([]);
  const [loading, setLoading] = useState(true);
  const [selectedProtocol, setSelectedProtocol] = useState('');
  const [showEnrollModal, setShowEnrollModal] = useState(false);
  const [enrollForm, setEnrollForm] = useState({
    protocol_id: '',
    site_id: '',
    screening_number: '',
    name_initials: '',
    gender: '男',
    birth_date: '',
    initial_status: '筛选中',
  });

  useEffect(() => {
    loadProtocols();
    loadSubjects();
  }, []);

  useEffect(() => {
    loadSubjects();
  }, [selectedProtocol]);

  const loadProtocols = async () => {
    try {
      const res = await protocolAPI.list();
      setProtocols(res.data);
    } catch (error) {
      console.error('加载方案失败:', error);
    }
  };

  const loadSubjects = async () => {
    try {
      const params = selectedProtocol ? { protocol_id: selectedProtocol } : {};
      const res = await subjectAPI.list(params);
      setSubjects(res.data);
    } catch (error) {
      console.error('加载受试者失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const getStatusClass = (status) => {
    switch (status) {
      case '筛选中': return 'bg-yellow-100 text-yellow-800';
      case '入组': return 'bg-blue-100 text-blue-800';
      case '治疗中': return 'bg-purple-100 text-purple-800';
      case '随访中': return 'bg-cyan-100 text-cyan-800';
      case '已完成': return 'bg-green-100 text-green-800';
      case '退出': return 'bg-red-100 text-red-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const getSelectedProtocolSites = () => {
    if (!enrollForm.protocol_id) return [];
    const p = protocols.find(p => p.id === enrollForm.protocol_id);
    return p?.sites || [];
  };

  const handleEnroll = async (e) => {
    e.preventDefault();
    try {
      await subjectAPI.enroll(enrollForm);
      setShowEnrollModal(false);
      setEnrollForm({
        protocol_id: '',
        site_id: '',
        screening_number: '',
        name_initials: '',
        gender: '男',
        birth_date: '',
        initial_status: '筛选中',
      });
      loadSubjects();
    } catch (error) {
      alert('入组失败: ' + (error.response?.data?.error || error.message));
    }
  };

  const handleWithdraw = async (subject) => {
    if (!confirm(`确定要让受试者 ${subject.randomization_id} 退出试验吗？退出后该编号将不能再使用。`)) {
      return;
    }
    try {
      await subjectAPI.withdraw(subject.id);
      loadSubjects();
    } catch (error) {
      alert('退组失败: ' + (error.response?.data?.error || error.message));
    }
  };

  const handleUpdateStatus = async (subject, newStatus) => {
    try {
      await subjectAPI.updateStatus(subject.id, newStatus);
      loadSubjects();
    } catch (error) {
      alert('状态更新失败: ' + (error.response?.data?.error || error.message));
    }
  };

  if (loading) {
    return <div className="p-8 text-center">加载中...</div>;
  }

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold text-gray-800">受试者管理</h1>
        <div className="flex items-center gap-4">
          <select
            value={selectedProtocol}
            onChange={(e) => setSelectedProtocol(e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">全部方案</option>
            {protocols.map(p => (
              <option key={p.id} value={p.id}>{p.protocol_number} - {p.drug_name}</option>
            ))}
          </select>
          <button
            onClick={() => setShowEnrollModal(true)}
            className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition"
          >
            + 入组受试者
          </button>
        </div>
      </div>

      <div className="bg-white shadow rounded-lg overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">筛选编号</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">随机编号</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">姓名缩写</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">性别</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">出生日期</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">入组日期</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">状态</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">操作</th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {subjects.map((subject) => (
              <tr key={subject.id}>
                <td className="px-6 py-4 whitespace-nowrap font-medium">{subject.screening_number}</td>
                <td className="px-6 py-4 whitespace-nowrap text-blue-600">{subject.randomization_id}</td>
                <td className="px-6 py-4 whitespace-nowrap">{subject.name_initials}</td>
                <td className="px-6 py-4 whitespace-nowrap">{subject.gender}</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm">{new Date(subject.birth_date).toLocaleDateString()}</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm">
                  {subject.enrollment_date ? new Date(subject.enrollment_date).toLocaleDateString() : '-'}
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className={`px-2 py-1 rounded-full text-xs ${getStatusClass(subject.status)}`}>
                    {subject.status}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm space-x-2">
                  <Link to={`/data-entry/${subject.id}`} className="text-blue-600 hover:text-blue-900">录入数据</Link>
                  <Link to={`/ae/${subject.id}`} className="text-orange-600 hover:text-orange-900">AE记录</Link>
                  {subject.status === '筛选中' && (
                    <button onClick={() => handleUpdateStatus(subject, '入组')} className="text-green-600 hover:text-green-900">入组</button>
                  )}
                  {subject.status === '入组' && (
                    <button onClick={() => handleUpdateStatus(subject, '治疗中')} className="text-green-600 hover:text-green-900">开始治疗</button>
                  )}
                  {subject.status === '治疗中' && (
                    <button onClick={() => handleUpdateStatus(subject, '随访中')} className="text-green-600 hover:text-green-900">进入随访</button>
                  )}
                  {subject.status === '随访中' && (
                    <button onClick={() => handleUpdateStatus(subject, '已完成')} className="text-green-600 hover:text-green-900">完成</button>
                  )}
                  {!['已完成', '退出'].includes(subject.status) && (
                    <button onClick={() => handleWithdraw(subject)} className="text-red-600 hover:text-red-900">退组</button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {subjects.length === 0 && (
          <div className="text-center py-12 text-gray-500">
            暂无受试者
          </div>
        )}
      </div>

      {showEnrollModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-md">
            <h2 className="text-xl font-bold mb-4">入组受试者</h2>
            <form onSubmit={handleEnroll} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">选择方案 *</label>
                <select
                  value={enrollForm.protocol_id}
                  onChange={(e) => setEnrollForm({ ...enrollForm, protocol_id: e.target.value, site_id: '' })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  required
                >
                  <option value="">请选择</option>
                  {protocols.map(p => (
                    <option key={p.id} value={p.id}>{p.protocol_number} - {p.drug_name}</option>
                  ))}
                </select>
              </div>
              {enrollForm.protocol_id && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">选择中心 *</label>
                  <select
                    value={enrollForm.site_id}
                    onChange={(e) => setEnrollForm({ ...enrollForm, site_id: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    required
                  >
                    <option value="">请选择</option>
                    {getSelectedProtocolSites().map(s => (
                      <option key={s.id} value={s.id}>{s.site_code} - {s.site_name}</option>
                    ))}
                  </select>
                </div>
              )}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">筛选编号 *</label>
                <input
                  type="text"
                  value={enrollForm.screening_number}
                  onChange={(e) => setEnrollForm({ ...enrollForm, screening_number: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">姓名缩写 *</label>
                <input
                  type="text"
                  value={enrollForm.name_initials}
                  onChange={(e) => setEnrollForm({ ...enrollForm, name_initials: e.target.value })}
                  placeholder="如: 张XX"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  required
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">性别 *</label>
                  <select
                    value={enrollForm.gender}
                    onChange={(e) => setEnrollForm({ ...enrollForm, gender: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="男">男</option>
                    <option value="女">女</option>
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">出生日期 *</label>
                  <input
                    type="date"
                    value={enrollForm.birth_date}
                    onChange={(e) => setEnrollForm({ ...enrollForm, birth_date: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    required
                  />
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">初始状态</label>
                <select
                  value={enrollForm.initial_status}
                  onChange={(e) => setEnrollForm({ ...enrollForm, initial_status: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="筛选中">筛选中</option>
                  <option value="入组">入组</option>
                </select>
              </div>
              <div className="flex justify-end gap-3 pt-4">
                <button
                  type="button"
                  onClick={() => setShowEnrollModal(false)}
                  className="px-4 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50"
                >
                  取消
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
                >
                  确认入组
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

export default SubjectList;
