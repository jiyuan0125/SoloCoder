import { useState, useEffect } from 'react';
import { unitAPI } from '../api';

const UNIT_TYPES = [
  '医院', '诊所', '美容美发店', '宾馆酒店', '游泳馆', '学校', '集中式供水单位'
];

const UNIT_STATUSES = [
  '正常', '整改中', '停业整顿', '注销'
];

export default function UnitsPage() {
  const [units, setUnits] = useState([]);
  const [warnings, setWarnings] = useState([]);
  const [filterType, setFilterType] = useState('');
  const [filterStatus, setFilterStatus] = useState('');
  const [showModal, setShowModal] = useState(false);
  const [editingUnit, setEditingUnit] = useState(null);
  const [formData, setFormData] = useState({
    credit_code: '',
    name: '',
    type: '医院',
    address: '',
    legal_representative: '',
    phone: '',
    health_license_number: '',
    health_license_validity: '',
  });

  useEffect(() => {
    loadUnits();
    loadWarnings();
  }, [filterType, filterStatus]);

  const loadUnits = async () => {
    try {
      const params = {};
      if (filterType) params.type = filterType;
      if (filterStatus) params.status = filterStatus;
      const res = await unitAPI.list(params);
      setUnits(res.data.data);
    } catch (err) {
      console.error('Failed to load units:', err);
    }
  };

  const loadWarnings = async () => {
    try {
      const res = await unitAPI.getWarnings();
      setWarnings(res.data.data);
    } catch (err) {
      console.error('Failed to load warnings:', err);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      if (editingUnit) {
        await unitAPI.update(editingUnit.id, formData);
      } else {
        await unitAPI.create(formData);
      }
      setShowModal(false);
      loadUnits();
      resetForm();
    } catch (err) {
      alert(err.response?.data?.error || '操作失败');
    }
  };

  const handleEdit = (unit) => {
    setEditingUnit(unit);
    setFormData({
      credit_code: unit.credit_code,
      name: unit.name,
      type: unit.type,
      address: unit.address,
      legal_representative: unit.legal_representative,
      phone: unit.phone,
      health_license_number: unit.health_license_number,
      health_license_validity: unit.health_license_validity?.split('T')[0] || '',
    });
    setShowModal(true);
  };

  const handleDelete = async (id) => {
    if (confirm('确定要删除吗？')) {
      try {
        await unitAPI.delete(id);
        loadUnits();
      } catch (err) {
        alert('删除失败');
      }
    }
  };

  const handleStatusChange = async (unit, newStatus) => {
    try {
      await unitAPI.update(unit.id, { status: newStatus });
      loadUnits();
    } catch (err) {
      alert('状态更新失败');
    }
  };

  const resetForm = () => {
    setEditingUnit(null);
    setFormData({
      credit_code: '',
      name: '',
      type: '医院',
      address: '',
      legal_representative: '',
      phone: '',
      health_license_number: '',
      health_license_validity: '',
    });
  };

  const getStatusColor = (status) => {
    switch (status) {
      case '正常': return 'bg-green-100 text-green-800';
      case '整改中': return 'bg-yellow-100 text-yellow-800';
      case '停业整顿': return 'bg-red-100 text-red-800';
      case '注销': return 'bg-gray-100 text-gray-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-800 mb-2">被监督单位管理</h1>
        {warnings.length > 0 && (
          <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-4">
            <h3 className="font-semibold text-yellow-800 mb-2">许可证即将到期预警</h3>
            <ul className="list-disc list-inside text-yellow-700">
              {warnings.map(w => (
                <li key={w.id}>{w.name} - 有效期至: {new Date(w.health_license_validity).toLocaleDateString()}</li>
              ))}
            </ul>
          </div>
        )}
      </div>

      <div className="flex flex-wrap gap-4 mb-6">
        <div className="flex gap-4 items-center">
          <label className="text-gray-600">单位类型:</label>
          <select
            value={filterType}
            onChange={(e) => setFilterType(e.target.value)}
            className="border rounded px-3 py-2"
          >
            <option value="">全部</option>
            {UNIT_TYPES.map(t => <option key={t}>{t}</option>)}
          </select>
        </div>
        <div className="flex gap-4 items-center">
          <label className="text-gray-600">状态:</label>
          <select
            value={filterStatus}
            onChange={(e) => setFilterStatus(e.target.value)}
            className="border rounded px-3 py-2"
          >
            <option value="">全部</option>
            {UNIT_STATUSES.map(s => <option key={s}>{s}</option>)}
          </select>
        </div>
        <button
          onClick={() => { resetForm(); setShowModal(true); }}
          className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
        >
          添加单位
        </button>
      </div>

      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">单位名称</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">类型</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">地址</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">法定代表人</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">许可证号</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">许可证有效期</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">状态</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">操作</th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {units.map((unit) => (
              <tr key={unit.id}>
                <td className="px-4 py-3 whitespace-nowrap">{unit.name}</td>
                <td className="px-4 py-3 whitespace-nowrap">{unit.type}</td>
                <td className="px-4 py-3 whitespace-nowrap">{unit.address}</td>
                <td className="px-4 py-3 whitespace-nowrap">{unit.legal_representative}</td>
                <td className="px-4 py-3 whitespace-nowrap">{unit.health_license_number}</td>
                <td className="px-4 py-3 whitespace-nowrap">
                  {new Date(unit.health_license_validity).toLocaleDateString()}
                </td>
                <td className="px-4 py-3 whitespace-nowrap">
                  <span className={`px-2 py-1 rounded text-xs font-medium ${getStatusColor(unit.status)}`}>
                    {unit.status}
                  </span>
                </td>
                <td className="px-4 py-3 whitespace-nowrap text-sm">
                  <button onClick={() => handleEdit(unit)} className="text-blue-600 hover:text-blue-900 mr-3">编辑</button>
                  <button onClick={() => handleDelete(unit.id)} className="text-red-600 hover:text-red-900 mr-3">删除</button>
                  {unit.status === '停业整顿' && (
                    <button
                      onClick={() => handleStatusChange(unit, '正常')}
                      className="text-green-600 hover:text-green-900">恢复</button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {showModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-lg p-6 w-full max-w-md max-h-[90vh] overflow-y-auto">
            <h2 className="text-xl font-bold mb-4">{editingUnit ? '编辑单位' : '添加单位'}</h2>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">统一社会信用代码 *</label>
                <input
                  type="text"
                  value={formData.credit_code}
                  onChange={(e) => setFormData({...formData, credit_code: e.target.value})}
                  className="w-full border rounded px-3 py-2"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">单位名称 *</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({...formData, name: e.target.value})}
                  className="w-full border rounded px-3 py-2"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">单位类型 *</label>
                <select
                  value={formData.type}
                  onChange={(e) => setFormData({...formData, type: e.target.value})}
                  className="w-full border rounded px-3 py-2"
                >
                  {UNIT_TYPES.map(t => <option key={t}>{t}</option>)}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">地址</label>
                <input
                  type="text"
                  value={formData.address}
                  onChange={(e) => setFormData({...formData, address: e.target.value})}
                  className="w-full border rounded px-3 py-2"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">法定代表人</label>
                <input
                  type="text"
                  value={formData.legal_representative}
                  onChange={(e) => setFormData({...formData, legal_representative: e.target.value})}
                  className="w-full border rounded px-3 py-2"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">联系电话</label>
                <input
                  type="text"
                  value={formData.phone}
                  onChange={(e) => setFormData({...formData, phone: e.target.value})}
                  className="w-full border rounded px-3 py-2"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">卫生许可证编号 *</label>
                <input
                  type="text"
                  value={formData.health_license_number}
                  onChange={(e) => setFormData({...formData, health_license_number: e.target.value})}
                  className="w-full border rounded px-3 py-2"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">许可证有效期 *</label>
                <input
                  type="date"
                  value={formData.health_license_validity}
                  onChange={(e) => setFormData({...formData, health_license_validity: e.target.value})}
                  className="w-full border rounded px-3 py-2"
                  required
                />
              </div>
              <div className="flex gap-3 pt-4">
                <button
                  type="submit"
                  className="flex-1 bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
                >
                  {editingUnit ? '保存' : '创建'}
                </button>
                <button
                  type="button"
                  onClick={() => { setShowModal(false); resetForm(); }}
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
