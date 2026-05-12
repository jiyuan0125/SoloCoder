import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { subjectAPI, aeAPI, todoAPI } from '../services/api';

function AEPage() {
  const { subjectId } = useParams();
  const navigate = useNavigate();

  const [subject, setSubject] = useState(null);
  const [aes, setAes] = useState([]);
  const [todos, setTodos] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [formData, setFormData] = useState({
    event_name: '',
    start_date: new Date().toISOString().split('T')[0],
    end_date: '',
    severity: '轻度',
    relationship: '可能有关',
    is_sae: false,
    treatment: '',
    outcome: '',
  });

  useEffect(() => {
    loadData();
  }, [subjectId]);

  const loadData = async () => {
    try {
      if (subjectId) {
        const subjectRes = await subjectAPI.get(subjectId);
        setSubject(subjectRes.data);
        const aesRes = await aeAPI.list({ subject_id: subjectId });
        setAes(aesRes.data);
      } else {
        const aesRes = await aeAPI.list({});
        setAes(aesRes.data);
      }
      const todosRes = await todoAPI.list();
      setTodos(todosRes.data);
    } catch (error) {
      console.error('加载数据失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    const data = {
      ...formData,
      subject_id: subjectId,
      start_date: new Date(formData.start_date),
      end_date: formData.end_date ? new Date(formData.end_date) : null,
    };
    try {
      await aeAPI.create(data);
      setShowModal(false);
      setFormData({
        event_name: '',
        start_date: new Date().toISOString().split('T')[0],
        end_date: '',
        severity: '轻度',
        relationship: '可能有关',
        is_sae: false,
        treatment: '',
        outcome: '',
      });
      loadData();
    } catch (error) {
      alert('保存失败: ' + (error.response?.data?.error || error.message));
    }
  };

  const handleSubmitReport = async (ae) => {
    try {
      await aeAPI.submitReport(ae.id);
      loadData();
      alert('SAE报告已提交');
    } catch (error) {
      alert('提交失败: ' + (error.response?.data?.error || error.message));
    }
  };

  const handleCompleteTodo = async (todo) => {
    try {
      await todoAPI.updateStatus(todo.id, '已完成');
      loadData();
    } catch (error) {
      alert('操作失败: ' + (error.response?.data?.error || error.message));
    }
  };

  const getSeverityClass = (severity) => {
    switch (severity) {
      case '轻度': return 'bg-yellow-100 text-yellow-800';
      case '中度': return 'bg-orange-100 text-orange-800';
      case '重度': return 'bg-red-100 text-red-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const getTodoStatusClass = (status) => {
    switch (status) {
      case '待办': return 'bg-yellow-100 text-yellow-800';
      case '已完成': return 'bg-green-100 text-green-800';
      case '逾期': return 'bg-red-100 text-red-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  if (loading) {
    return <div className="p-8 text-center">加载中...</div>;
  }

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-800">不良事件管理</h1>
          {subject && (
            <p className="text-sm text-gray-500 mt-1">
              受试者: {subject.randomization_id} ({subject.name_initials})
            </p>
          )}
        </div>
        <div className="flex gap-3">
          <button
            onClick={() => navigate('/subjects')}
            className="px-4 py-2 text-gray-600 hover:text-gray-800"
          >
            返回列表
          </button>
          {subjectId && (
            <button
              onClick={() => setShowModal(true)}
              className="px-4 py-2 bg-orange-600 text-white rounded-md hover:bg-orange-700"
            >
              + 记录AE
            </button>
          )}
        </div>
      </div>

      {todos.length > 0 && (
        <div className="mb-6 bg-yellow-50 border border-yellow-200 rounded-lg p-4">
          <h3 className="font-semibold text-yellow-800 mb-3">待办事项</h3>
          <div className="space-y-2">
            {todos.map((todo) => (
              <div key={todo.id} className="flex justify-between items-center bg-white rounded p-3">
                <div>
                  <div className="font-medium">{todo.title}</div>
                  <div className="text-sm text-gray-500">{todo.description}</div>
                  {todo.due_date && (
                    <div className="text-xs text-gray-400 mt-1">
                      截止: {new Date(todo.due_date).toLocaleString()}
                    </div>
                  )}
                </div>
                <div className="flex items-center gap-3">
                  <span className={`px-2 py-1 rounded text-xs ${getTodoStatusClass(todo.status)}`}>
                    {todo.status}
                  </span>
                  {todo.status !== '已完成' && (
                    <button
                      onClick={() => handleCompleteTodo(todo)}
                      className="text-blue-600 hover:text-blue-900 text-sm"
                    >
                      标记完成
                    </button>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="bg-white shadow rounded-lg overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">事件名称</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">发生日期</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">严重程度</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">与药物关系</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">SAE</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">转归</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">操作</th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {aes.map((ae) => (
              <tr key={ae.id}>
                <td className="px-6 py-4 whitespace-nowrap font-medium">{ae.event_name}</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm">
                  {new Date(ae.start_date).toLocaleDateString()}
                  {ae.end_date && ` - ${new Date(ae.end_date).toLocaleDateString()}`}
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className={`px-2 py-1 rounded-full text-xs ${getSeverityClass(ae.severity)}`}>
                    {ae.severity}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm">{ae.relationship}</td>
                <td className="px-6 py-4 whitespace-nowrap">
                  {ae.is_sae ? (
                    <span className="px-2 py-1 rounded-full text-xs bg-red-100 text-red-800">SAE</span>
                  ) : (
                    <span className="text-gray-400 text-sm">-</span>
                  )}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm">{ae.outcome || '-'}</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm">
                  {ae.is_sae && !ae.report_submitted && (
                    <button
                      onClick={() => handleSubmitReport(ae)}
                      className="text-red-600 hover:text-red-900"
                    >
                      提交SAE报告
                    </button>
                  )}
                  {ae.is_sae && ae.report_submitted && (
                    <span className="text-green-600">报告已提交</span>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {aes.length === 0 && (
          <div className="text-center py-12 text-gray-500">
            暂无不良事件记录
          </div>
        )}
      </div>

      {showModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-lg">
            <h2 className="text-xl font-bold mb-4">记录不良事件</h2>
            <form onSubmit={handleSubmit} className="space-y-4 max-h-[70vh] overflow-y-auto">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">事件名称 *</label>
                <input
                  type="text"
                  value={formData.event_name}
                  onChange={(e) => setFormData({ ...formData, event_name: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  required
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">发生日期 *</label>
                  <input
                    type="date"
                    value={formData.start_date}
                    onChange={(e) => setFormData({ ...formData, start_date: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    required
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">结束日期</label>
                  <input
                    type="date"
                    value={formData.end_date}
                    onChange={(e) => setFormData({ ...formData, end_date: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">严重程度</label>
                  <select
                    value={formData.severity}
                    onChange={(e) => setFormData({ ...formData, severity: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="轻度">轻度</option>
                    <option value="中度">中度</option>
                    <option value="重度">重度</option>
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">与药物关系</label>
                  <select
                    value={formData.relationship}
                    onChange={(e) => setFormData({ ...formData, relationship: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="肯定有关">肯定有关</option>
                    <option value="可能有关">可能有关</option>
                    <option value="可能无关">可能无关</option>
                    <option value="无关">无关</option>
                  </select>
                </div>
              </div>
              <div className="flex items-center">
                <input
                  type="checkbox"
                  id="is_sae"
                  checked={formData.is_sae}
                  onChange={(e) => setFormData({ ...formData, is_sae: e.target.checked })}
                  className="mr-2"
                />
                <label htmlFor="is_sae" className="text-sm font-medium text-gray-700">
                  严重不良事件 (SAE)
                </label>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">处理措施</label>
                <textarea
                  value={formData.treatment}
                  onChange={(e) => setFormData({ ...formData, treatment: e.target.value })}
                  rows={2}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">转归</label>
                <select
                  value={formData.outcome}
                  onChange={(e) => setFormData({ ...formData, outcome: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="">请选择</option>
                  <option value="恢复">恢复</option>
                  <option value="好转">好转</option>
                  <option value="未恢复">未恢复</option>
                  <option value="死亡">死亡</option>
                  <option value="导致退出试验">导致退出试验</option>
                </select>
              </div>
              <div className="flex justify-end gap-3 pt-4">
                <button
                  type="button"
                  onClick={() => setShowModal(false)}
                  className="px-4 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50"
                >
                  取消
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-orange-600 text-white rounded-md hover:bg-orange-700"
                >
                  保存
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

export default AEPage;
