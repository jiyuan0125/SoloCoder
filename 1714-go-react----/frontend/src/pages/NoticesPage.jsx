import { useState, useEffect } from 'react';
import { noticeAPI, unitAPI } from '../api';

export default function NoticesPage() {
  const [notices, setNotices] = useState([]);
  const [units, setUnits] = useState([]);
  const [activeTab, setActiveTab] = useState('active');
  const [showModal, setShowModal] = useState(false);
  const [formData, setFormData] = useState({
    unit_id: '',
    title: '',
    content: '',
    notice_type: '检查结果',
    duration_days: 30,
  });

  useEffect(() => {
    loadNotices();
    loadUnits();
  }, [activeTab]);

  const loadNotices = async () => {
    try {
      const status = activeTab === 'active' ? '公示中' : '历史公示';
      const res = await noticeAPI.list({ status });
      setNotices(res.data.data);
    } catch (err) {
      console.error('Failed to load notices:', err);
    }
  };

  const loadUnits = async () => {
    try {
      const res = await unitAPI.list();
      setUnits(res.data.data);
    } catch (err) {
      console.error('Failed to load units:', err);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      await noticeAPI.create(formData);
      setShowModal(false);
      loadNotices();
      setFormData({
        unit_id: '',
        title: '',
        content: '',
        notice_type: '检查结果',
        duration_days: 30,
      });
    } catch (err) {
      alert(err.response?.data?.error || '发布公示失败');
    }
  };

  const handleCheckExpiry = async () => {
    try {
      await noticeAPI.checkExpiry();
      loadNotices();
      alert('已检查并更新过期公示');
    } catch (err) {
      alert('检查失败');
    }
  };

  const handleUpdateExpiry = async (notice) => {
    const newDays = prompt('请输入新的公示天数：', notice.duration_days || 30);
    if (newDays && parseInt(newDays) > 0) {
      try {
        await noticeAPI.updateExpiry(notice.id, { duration_days: parseInt(newDays) });
        loadNotices();
      } catch (err) {
        alert('更新失败');
      }
    }
  };

  const handleDelete = async (id) => {
    if (confirm('确定要删除此公示吗？')) {
      try {
        await noticeAPI.delete(id);
        loadNotices();
      } catch (err) {
        alert('删除失败');
      }
    }
  };

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-800 mb-2">公示管理</h1>
      </div>

      <div className="flex gap-2 mb-6">
        <button
          onClick={() => setActiveTab('active')}
          className={`px-4 py-2 rounded ${activeTab === 'active' ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-700'}`}
        >
          当前公示
        </button>
        <button
          onClick={() => setActiveTab('history')}
          className={`px-4 py-2 rounded ${activeTab === 'history' ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-700'}`}
        >
          历史公示
        </button>
        <button
          onClick={() => setShowModal(true)}
          className="ml-auto bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700"
        >
          发布公示
        </button>
        <button
          onClick={handleCheckExpiry}
          className="bg-yellow-600 text-white px-4 py-2 rounded hover:bg-yellow-700"
        >
          检查过期公示
        </button>
      </div>

      <div className="space-y-4">
        {notices.length === 0 ? (
          <div className="text-center py-12 text-gray-500">
            {activeTab === 'active' ? '暂无当前公示' : '暂无历史公示'}
          </div>
        ) : (
          notices.map((notice) => (
            <div key={notice.id} className="bg-white rounded-lg shadow p-6">
              <div className="flex justify-between items-start mb-4">
                <div>
                  <h3 className="text-lg font-semibold text-gray-800">{notice.title}</h3>
                  <div className="text-sm text-gray-500 mt-1 space-x-4">
                    <span>单位: {notice.unit?.name}</span>
                    <span>类型: {notice.notice_type}</span>
                    <span>发布时间: {new Date(notice.publish_date).toLocaleString()}</span>
                    <span>到期时间: {new Date(notice.expiry_date).toLocaleString()}</span>
                  </div>
                </div>
                <div className="flex gap-2">
                  {activeTab === 'active' && (
                    <button
                      onClick={() => handleUpdateExpiry(notice)}
                      className="text-blue-600 hover:text-blue-900 text-sm"
                    >
                      延长公示
                    </button>
                  )}
                  <button
                    onClick={() => handleDelete(notice.id)}
                    className="text-red-600 hover:text-red-900 text-sm"
                  >
                    删除
                  </button>
                </div>
              </div>
              <p className="text-gray-700 whitespace-pre-wrap">{notice.content}</p>
            </div>
          ))
        )}
      </div>

      {showModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-lg p-6 w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <h2 className="text-xl font-bold mb-4">发布公示</h2>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">被监督单位 *</label>
                  <select
                    value={formData.unit_id}
                    onChange={(e) => setFormData({...formData, unit_id: e.target.value})}
                    className="w-full border rounded px-3 py-2"
                    required
                  >
                    <option value="">请选择单位</option>
                    {units.map(u => (
                      <option key={u.id} value={u.id}>{u.name}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">公示类型</label>
                  <select
                    value={formData.notice_type}
                    onChange={(e) => setFormData({...formData, notice_type: e.target.value})}
                    className="w-full border rounded px-3 py-2"
                  >
                    <option value="检查结果">检查结果</option>
                    <option value="处罚决定">处罚决定</option>
                    <option value="投诉处理">投诉处理</option>
                    <option value="其他">其他</option>
                  </select>
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">公示标题 *</label>
                <input
                  type="text"
                  value={formData.title}
                  onChange={(e) => setFormData({...formData, title: e.target.value})}
                  className="w-full border rounded px-3 py-2"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">公示内容 *</label>
                <textarea
                  value={formData.content}
                  onChange={(e) => setFormData({...formData, content: e.target.value})}
                  className="w-full border rounded px-3 py-2"
                  rows={6}
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">公示天数</label>
                <input
                  type="number"
                  value={formData.duration_days}
                  onChange={(e) => setFormData({...formData, duration_days: parseInt(e.target.value) || 30})}
                  className="w-full border rounded px-3 py-2"
                  min="1"
                />
                <p className="text-xs text-gray-500 mt-1">默认30天</p>
              </div>
              <div className="flex gap-3 pt-4">
                <button
                  type="submit"
                  className="flex-1 bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
                >
                  发布
                </button>
                <button
                  type="button"
                  onClick={() => setShowModal(false)}
                  className="flex-1 bg-gray-200 text-gray-700 px-4 py-2 rounded hover:bg-gray-300"
                >
                  取消
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
