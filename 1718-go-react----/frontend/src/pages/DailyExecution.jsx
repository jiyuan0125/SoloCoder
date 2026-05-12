import { useState, useEffect } from 'react';
import { enrollmentsApi, ordersApi, pathsApi } from '../api';

function DailyExecution() {
  const [enrollments, setEnrollments] = useState([]);
  const [paths, setPaths] = useState([]);
  const [selectedEnrollment, setSelectedEnrollment] = useState(null);
  const [orders, setOrders] = useState([]);
  const [showVariationModal, setShowVariationModal] = useState(false);
  const [selectedOrder, setSelectedOrder] = useState(null);
  const [variationForm, setVariationForm] = useState({
    date: new Date().toISOString().split('T')[0],
    content: '',
    reason: '',
    type: '可控变异',
    action: '',
  });
  const [message, setMessage] = useState({ type: '', text: '' });

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      const [enrRes, pathsRes] = await Promise.all([
        enrollmentsApi.list(),
        pathsApi.list(),
      ]);
      const activeEnrollments = (enrRes.data.data || []).filter(e => e.status === 'active');
      setEnrollments(activeEnrollments);
      setPaths(pathsRes.data.data || []);
    } catch (err) {
      showMessage('error', '加载数据失败');
    }
  };

  const showMessage = (type, text) => {
    setMessage({ type, text });
    setTimeout(() => setMessage({ type: '', text: '' }), 3000);
  };

  const loadOrders = async (enrollment) => {
    setSelectedEnrollment(enrollment);
    try {
      const res = await enrollmentsApi.listOrders(enrollment.id);
      setOrders(res.data.data || []);
    } catch (err) {
      showMessage('error', '加载医嘱失败');
    }
  };

  const handleExecuteOrder = async (order, executed) => {
    try {
      await ordersApi.execute(order.id, executed);
      showMessage('success', executed ? '医嘱已标记为已执行' : '医嘱已标记为未执行');
      if (selectedEnrollment) {
        loadOrders(selectedEnrollment);
      }
    } catch (err) {
      showMessage('error', '操作失败');
    }
  };

  const handleRecordVariation = async () => {
    if (!selectedEnrollment) return;
    try {
      await enrollmentsApi.createVariation(selectedEnrollment.id, variationForm);
      showMessage('success', '变异记录已保存');
      setShowVariationModal(false);
      setVariationForm({
        date: new Date().toISOString().split('T')[0],
        content: '',
        reason: '',
        type: '可控变异',
        action: '',
      });
      setSelectedOrder(null);
    } catch (err) {
      showMessage('error', '保存失败');
    }
  };

  const getPathName = (pathId) => {
    const path = paths.find(p => p.id === pathId);
    return path ? path.name : '未知路径';
  };

  const groupOrdersByDate = () => {
    const groups = {};
    orders.forEach(order => {
      const date = order.order_date;
      if (!groups[date]) {
        groups[date] = [];
      }
      groups[date].push(order);
    });
    return Object.entries(groups).sort((a, b) => new Date(a[0]) - new Date(b[0]));
  };

  const formatDate = (dateStr) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString('zh-CN', { year: 'numeric', month: 'long', day: 'numeric', weekday: 'long' });
  };

  const getOrderRowClass = (order) => {
    if (order.is_executed) {
      return 'bg-green-50';
    }
    if (order.is_required && !order.is_executed) {
      return 'bg-red-50';
    }
    return 'bg-gray-50';
  };

  const getOrderStatusIcon = (order) => {
    if (order.is_executed) {
      return <span className="text-green-600 text-lg">✓</span>;
    }
    if (order.is_variation) {
      return <span className="text-orange-600 text-lg">!</span>;
    }
    if (order.is_required) {
      return <span className="text-red-600 text-lg">*</span>;
    }
    return <span className="text-gray-400 text-lg">○</span>;
  };

  const getToday = () => {
    return new Date().toISOString().split('T')[0];
  };

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold text-gray-800">每日执行</h1>
        <button
          onClick={() => setShowVariationModal(true)}
          disabled={!selectedEnrollment}
          className="bg-orange-600 text-white px-4 py-2 rounded hover:bg-orange-700 disabled:bg-gray-300"
        >
          + 记录变异
        </button>
      </div>

      {message.text && (
        <div className={`mb-4 p-3 rounded ${message.type === 'success' ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}`}>
          {message.text}
        </div>
      )}

      <div className="grid grid-cols-4 gap-6">
        <div className="col-span-1">
          <div className="bg-white rounded-lg shadow p-4">
            <h2 className="font-semibold mb-3">当前路径患者</h2>
            <div className="space-y-2">
              {enrollments.map((e) => (
                <div
                  key={e.id}
                  onClick={() => loadOrders(e)}
                  className={`p-3 rounded border cursor-pointer ${selectedEnrollment?.id === e.id ? 'bg-blue-50 border-blue-300' : 'hover:bg-gray-50'}`}
                >
                  <div className="font-medium">{e.patient_name}</div>
                  <div className="text-sm text-gray-500">{e.patient_hospital_no}</div>
                  <div className="text-xs text-gray-400">{getPathName(e.path_id)}</div>
                  {e.suggest_exit && (
                    <div className="text-xs text-orange-600 mt-1">⚠️ 建议退径</div>
                  )}
                </div>
              ))}
              {enrollments.length === 0 && (
                <div className="text-gray-400 text-center py-4">暂无活动患者</div>
              )}
            </div>
          </div>
        </div>

        <div className="col-span-3">
          {selectedEnrollment ? (
            <div className="space-y-4">
              <div className="bg-white rounded-lg shadow p-4">
                <div className="flex justify-between items-center">
                  <div>
                    <h2 className="text-lg font-semibold">{selectedEnrollment.patient_name}</h2>
                    <div className="text-sm text-gray-500">
                      住院号: {selectedEnrollment.patient_hospital_no} | 
                      路径: {getPathName(selectedEnrollment.path_id)} |
                      入径日期: {new Date(selectedEnrollment.enroll_date).toLocaleDateString()}
                    </div>
                  </div>
                  <div className="text-sm">
                    <span className="text-gray-500">变异次数: </span>
                    <span className={selectedEnrollment.variation_count >= 3 ? 'text-orange-600 font-medium' : ''}>
                      {selectedEnrollment.variation_count}
                    </span>
                    {selectedEnrollment.suggest_exit && (
                      <span className="ml-2 text-orange-600 bg-orange-100 px-2 py-1 rounded text-xs">
                        建议退径
                      </span>
                    )}
                  </div>
                </div>
              </div>

              {groupOrdersByDate().map(([date, dateOrders]) => {
                const isToday = date.startsWith(getToday());
                return (
                  <div key={date} className="bg-white rounded-lg shadow">
                    <div className={`p-3 border-b font-medium ${isToday ? 'bg-blue-50 text-blue-800' : ''}`}>
                      {formatDate(date)}
                      {isToday && <span className="ml-2 text-xs bg-blue-600 text-white px-2 py-0.5 rounded">今天</span>}
                    </div>
                    <div className="divide-y">
                      {dateOrders.map((order) => (
                        <div key={order.id} className={`p-3 flex items-center gap-4 ${getOrderRowClass(order)}`}>
                          <div className="w-8 flex justify-center">
                            {getOrderStatusIcon(order)}
                          </div>
                          <div className="flex-1">
                            <div className="flex items-center gap-2">
                              <span className={`font-medium ${order.is_executed ? 'text-green-700' : order.is_required ? 'text-red-700' : 'text-gray-600'}`}>
                                {order.item_name}
                              </span>
                              <span className="text-xs bg-gray-200 px-2 py-0.5 rounded">{order.category}</span>
                              <span className={`text-xs px-2 py-0.5 rounded ${order.is_required ? 'bg-red-100 text-red-600' : 'bg-gray-100 text-gray-500'}`}>
                                {order.is_required ? '必选' : '可选'}
                              </span>
                              {order.is_variation && (
                                <span className="text-xs bg-orange-100 text-orange-600 px-2 py-0.5 rounded">变异</span>
                              )}
                            </div>
                            <div className="text-xs text-gray-400 mt-1">
                              阶段: {order.stage_name} | 路径第{order.path_day}天
                            </div>
                          </div>
                          <div className="flex gap-2">
                            {!order.is_executed ? (
                              <button
                                onClick={() => handleExecuteOrder(order, true)}
                                className="text-sm bg-green-600 text-white px-3 py-1 rounded hover:bg-green-700"
                              >
                                执行
                              </button>
                            ) : (
                              <button
                                onClick={() => handleExecuteOrder(order, false)}
                                className="text-sm bg-gray-200 text-gray-600 px-3 py-1 rounded hover:bg-gray-300"
                              >
                                撤销
                              </button>
                            )}
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                );
              })}

              {orders.length === 0 && (
                <div className="bg-white rounded-lg shadow p-8 text-center text-gray-400">
                  暂无医嘱计划
                </div>
              )}
            </div>
          ) : (
            <div className="bg-white rounded-lg shadow p-12 text-center text-gray-400">
              请从左侧选择一个患者查看医嘱计划
            </div>
          )}
        </div>
      </div>

      {showVariationModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-96">
            <h3 className="text-lg font-semibold mb-4">记录变异</h3>
            {selectedOrder && (
              <div className="mb-4 p-3 bg-yellow-50 rounded text-sm">
                <div className="font-medium">关联医嘱: {selectedOrder.item_name}</div>
              </div>
            )}
            <div className="space-y-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">日期</label>
                <input
                  type="date"
                  value={variationForm.date}
                  onChange={(e) => setVariationForm({ ...variationForm, date: e.target.value })}
                  className="w-full border rounded px-3 py-2"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">变异内容</label>
                <textarea
                  className="w-full border rounded px-3 py-2"
                  rows="2"
                  value={variationForm.content}
                  onChange={(e) => setVariationForm({ ...variationForm, content: e.target.value })}
                  placeholder="描述变异情况"
                ></textarea>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">变异原因</label>
                <input
                  type="text"
                  className="w-full border rounded px-3 py-2"
                  value={variationForm.reason}
                  onChange={(e) => setVariationForm({ ...variationForm, reason: e.target.value })}
                  placeholder="变异原因"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">变异类型</label>
                <select
                  className="w-full border rounded px-3 py-2"
                  value={variationForm.type}
                  onChange={(e) => setVariationForm({ ...variationForm, type: e.target.value })}
                >
                  <option value="可控变异">可控变异</option>
                  <option value="不可控变异">不可控变异</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">处理措施</label>
                <textarea
                  className="w-full border rounded px-3 py-2"
                  rows="2"
                  value={variationForm.action}
                  onChange={(e) => setVariationForm({ ...variationForm, action: e.target.value })}
                  placeholder="采取的处理措施"
                ></textarea>
              </div>
            </div>
            <div className="flex gap-2 mt-4">
              <button
                onClick={handleRecordVariation}
                disabled={!variationForm.content || !variationForm.reason || !variationForm.action}
                className="flex-1 bg-orange-600 text-white py-2 rounded hover:bg-orange-700 disabled:bg-gray-300"
              >
                保存
              </button>
              <button
                onClick={() => { setShowVariationModal(false); setSelectedOrder(null); }}
                className="flex-1 bg-gray-200 py-2 rounded hover:bg-gray-300"
              >
                取消
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default DailyExecution;
