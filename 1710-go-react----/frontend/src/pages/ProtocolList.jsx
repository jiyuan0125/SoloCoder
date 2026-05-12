import { useState, useEffect } from 'react';
import { protocolAPI } from '../services/api';
import { Link } from 'react-router-dom';

function ProtocolList() {
  const [protocols, setProtocols] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadProtocols();
  }, []);

  const loadProtocols = async () => {
    try {
      const res = await protocolAPI.list();
      setProtocols(res.data);
    } catch (error) {
      console.error('加载方案列表失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const getStatusClass = (status) => {
    switch (status) {
      case '筹备中': return 'bg-yellow-100 text-yellow-800';
      case '进行中': return 'bg-blue-100 text-blue-800';
      case '已完成': return 'bg-green-100 text-green-800';
      case '已终止': return 'bg-red-100 text-red-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  if (loading) {
    return <div className="p-8 text-center">加载中...</div>;
  }

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold text-gray-800">试验方案管理</h1>
        <Link
          to="/protocols/new"
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition"
        >
          + 新建方案
        </Link>
      </div>

      <div className="bg-white shadow rounded-lg overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">方案编号</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">药物名称</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">试验阶段</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">状态</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">计划入组</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">试验周期</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">操作</th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {protocols.map((protocol) => (
              <tr key={protocol.id}>
                <td className="px-6 py-4 whitespace-nowrap font-medium text-blue-600">{protocol.protocol_number}</td>
                <td className="px-6 py-4 whitespace-nowrap">{protocol.drug_name}</td>
                <td className="px-6 py-4 whitespace-nowrap">{protocol.trial_phase}</td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className={`px-2 py-1 rounded-full text-xs ${getStatusClass(protocol.status)}`}>
                    {protocol.status}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">{protocol.planned_enrollment}人</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm">
                  {new Date(protocol.start_date).toLocaleDateString()} - {new Date(protocol.end_date).toLocaleDateString()}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm">
                  <Link to={`/protocols/${protocol.id}`} className="text-blue-600 hover:text-blue-900 mr-4">查看</Link>
                  <Link to={`/protocols/${protocol.id}/edit`} className="text-gray-600 hover:text-gray-900">编辑</Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {protocols.length === 0 && (
          <div className="text-center py-12 text-gray-500">
            暂无试验方案，点击上方按钮新建
          </div>
        )}
      </div>
    </div>
  );
}

export default ProtocolList;
