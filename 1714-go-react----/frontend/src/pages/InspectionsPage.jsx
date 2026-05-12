import { useState, useEffect } from 'react';
import { inspectionAPI, unitAPI } from '../api';

const INSPECTION_TYPES = ['日常监督', '专项检查', '投诉举报核查'];
const ITEM_RESULTS = ['合格', '不合格', '不适用'];

export default function InspectionsPage() {
  const [inspections, setInspections] = useState([]);
  const [units, setUnits] = useState([]);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showDetailModal, setShowDetailModal] = useState(false);
  const [selectedInspection, setSelectedInspection] = useState(null);
  const [templates, setTemplates] = useState([]);
  const [filterUnit, setFilterUnit] = useState('');
  const [filterType, setFilterType] = useState('');

  const [formData, setFormData] = useState({
    unit_id: '',
    inspection_date: new Date().toISOString().split('T')[0],
    inspectors: '',
    inspection_type: '日常监督',
    items: [],
  });

  useEffect(() => {
    loadInspections();
    loadUnits();
  }, [filterUnit, filterType]);

  const loadInspections = async () => {
    try {
      const params = {};
      if (filterUnit) params.unit_id = filterUnit;
      if (filterType) params.type = filterType;
      const res = await inspectionAPI.list(params);
      setInspections(res.data.data);
    } catch (err) {
      console.error('Failed to load inspections:', err);
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

  const handleUnitSelect = async (unitId) => {
    const unit = units.find(u => u.id === parseInt(unitId));
    if (unit) {
      const res = await inspectionAPI.getTemplates(unit.type);
      const templates = res.data.data;
      const items = templates.map(t => ({
        item_name: t.item_name,
        result: '合格',
      }));
      setFormData({ ...formData, unit_id: unitId, items });
      setTemplates(templates);
    }
  };

  const handleItemChange = (index, result) => {
    const newItems = [...formData.items];
    newItems[index].result = result;
    setFormData({ ...formData, items: newItems });
  };

  const calculatePassRate = (items) => {
    let applicable = 0;
    let pass = 0;
    items.forEach(item => {
      if (item.result !== '不适用') {
        applicable++;
        if (item.result === '合格') {
          pass++;
        }
      }
    });
    if (applicable === 0) return { rate: 0, isQualified: false };
    const rate = (pass / applicable) * 100;
    return { rate: rate.toFixed(1), isQualified: rate >= 80 };
  };

  const handleCreate = async (e) => {
    e.preventDefault();
    try {
      const inspectors = formData.inspectors.split(',').map(s => s.trim()).filter(s => s);
      if (inspectors.length < 2) {
        alert('检查人员必须两名以上');
        return;
      }
      await inspectionAPI.create(formData);
      setShowCreateModal(false);
      loadInspections();
      resetForm();
    } catch (err) {
      alert(err.response?.data?.error || '创建失败');
    }
  };

  const handleViewDetail = async (id) => {
    try {
      const res = await inspectionAPI.get(id);
      setSelectedInspection(res.data.data);
      setShowDetailModal(true);
    } catch (err) {
      console.error('Failed to load inspection:', err);
    }
  };

  const resetForm = () => {
    setFormData({
      unit_id: '',
      inspection_date: new Date().toISOString().split('T')[0],
      inspectors: '',
      inspection_type: '日常监督',
      items: [],
    });
    setTemplates([]);
  };

  const { rate: formRate, isQualified: formQualified } = calculatePassRate(formData.items);

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-800 mb-2">监督检查管理</h1>
      </div>

      <div className="flex flex-wrap gap-4 mb-6">
        <div className="flex gap-4 items-center">
          <label className="text-gray-600">单位:</label>
          <select
            value={filterUnit}
            onChange={(e) => setFilterUnit(e.target.value)}
            className="border rounded px-3 py-2"
          >
            <option value="">全部</option>
            {units.map(u => <option key={u.id} value={u.id}>{u.name}</option>)}
          </select>
        </div>
        <div className="flex gap-4 items-center">
          <label className="text-gray-600">检查类型:</label>
          <select
            value={filterType}
            onChange={(e) => setFilterType(e.target.value)}
            className="border rounded px-3 py-2"
          >
            <option value="">全部</option>
            {INSPECTION_TYPES.map(t => <option key={t}>{t}</option>)}
          </select>
        </div>
        <button
          onClick={() => { resetForm(); setShowCreateModal(true); }}
          className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
        >
          创建检查记录
        </button>
      </div>

      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">单位</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">检查日期</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">检查人员</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">检查类型</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">合格率</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">结果</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">操作</th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {inspections.map((ins) => (
              <tr key={ins.id}>
                <td className="px-4 py-3 whitespace-nowrap">{ins.unit?.name}</td>
                <td className="px-4 py-3 whitespace-nowrap">
                  {new Date(ins.inspection_date).toLocaleDateString()}
                </td>
                <td className="px-4 py-3 whitespace-nowrap">{ins.inspectors}</td>
                <td className="px-4 py-3 whitespace-nowrap">{ins.inspection_type}</td>
                <td className="px-4 py-3 whitespace-nowrap">{ins.pass_rate?.toFixed(1)}%</td>
                <td className="px-4 py-3 whitespace-nowrap">
                  <span className={`px-2 py-1 rounded text-xs font-medium ${
                    ins.is_qualified ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                  }`}>
                    {ins.is_qualified ? '合格' : '不合格'}
                  </span>
                </td>
                <td className="px-4 py-3 whitespace-nowrap text-sm">
                  <button
                    onClick={() => handleViewDetail(ins.id)}
                    className="text-blue-600 hover:text-blue-900"
                  >
                    查看详情
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {showCreateModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-lg p-6 w-full max-w-4xl max-h-[90vh] overflow-y-auto">
            <h2 className="text-xl font-bold mb-4">创建检查记录</h2>
            <form onSubmit={handleCreate} className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">被监督单位 *</label>
                  <select
                    value={formData.unit_id}
                    onChange={(e) => handleUnitSelect(e.target.value)}
                    className="w-full border rounded px-3 py-2"
                    required
                  >
                    <option value="">请选择单位</option>
                    {units.filter(u => u.status !== '停业整顿').map(u => (
                      <option key={u.id} value={u.id}>{u.name} ({u.type})</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">检查日期 *</label>
                  <input
                    type="date"
                    value={formData.inspection_date}
                    onChange={(e) => setFormData({...formData, inspection_date: e.target.value})}
                    className="w-full border rounded px-3 py-2"
                    required
                  />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">检查人员（用逗号分隔，至少2人）*</label>
                  <input
                    type="text"
                    value={formData.inspectors}
                    onChange={(e) => setFormData({...formData, inspectors: e.target.value})}
                    placeholder="张三,李四"
                    className="w-full border rounded px-3 py-2"
                    required
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">检查类型 *</label>
                  <select
                    value={formData.inspection_type}
                    onChange={(e) => setFormData({...formData, inspection_type: e.target.value})}
                    className="w-full border rounded px-3 py-2"
                  >
                    {INSPECTION_TYPES.map(t => <option key={t}>{t}</option>)}
                  </select>
                </div>
              </div>

              {formData.items.length > 0 && (
                <div>
                  <div className="flex justify-between items-center mb-2">
                    <label className="block text-sm font-medium text-gray-700">检查项目</label>
                    <div className="text-sm">
                      合格率: <span className={formQualified ? 'text-green-600 font-bold' : 'text-red-600 font-bold'}>
                        {formRate}%
                      </span>
                      <span className={`ml-2 px-2 py-1 rounded text-xs ${
                        formQualified ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                      }`}>
                        {formQualified ? '合格' : '不合格'}
                      </span>
                    </div>
                  </div>
                  <div className="border rounded overflow-hidden">
                    <table className="min-w-full">
                      <thead className="bg-gray-50">
                        <tr>
                          <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">序号</th>
                          <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">检查项</th>
                          <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">结果</th>
                        </tr>
                      </thead>
                      <tbody>
                        {formData.items.map((item, index) => (
                          <tr key={index} className="border-t">
                            <td className="px-4 py-2">{index + 1}</td>
                            <td className="px-4 py-2">{item.item_name}</td>
                            <td className="px-4 py-2">
                              <select
                                value={item.result}
                                onChange={(e) => handleItemChange(index, e.target.value)}
                                className={`border rounded px-2 py-1 ${
                                  item.result === '不合格' ? 'bg-red-50 border-red-300' :
                                  item.result === '不适用' ? 'bg-gray-50' : 'bg-green-50 border-green-300'
                                }`}
                              >
                                {ITEM_RESULTS.map(r => <option key={r}>{r}</option>)}
                              </select>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}

              <div className="flex gap-3 pt-4">
                <button
                  type="submit"
                  className="flex-1 bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
                >
                  创建
                </button>
                <button
                  type="button"
                  onClick={() => { setShowCreateModal(false); resetForm(); }}
                  className="flex-1 bg-gray-200 text-gray-700 px-4 py-2 rounded hover:bg-gray-300"
                >
                  取消
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {showDetailModal && selectedInspection && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-lg p-6 w-full max-w-3xl max-h-[90vh] overflow-y-auto">
            <h2 className="text-xl font-bold mb-4">检查记录详情</h2>
            <div className="grid grid-cols-2 gap-4 mb-4">
              <div><strong>单位:</strong> {selectedInspection.unit?.name}</div>
              <div><strong>检查日期:</strong> {new Date(selectedInspection.inspection_date).toLocaleDateString()}</div>
              <div><strong>检查人员:</strong> {selectedInspection.inspectors}</div>
              <div><strong>检查类型:</strong> {selectedInspection.inspection_type}</div>
              <div><strong>合格率:</strong> {selectedInspection.pass_rate?.toFixed(1)}%</div>
              <div>
                <strong>结果:</strong>
                <span className={`ml-2 px-2 py-1 rounded text-xs ${
                  selectedInspection.is_qualified ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                }`}>
                  {selectedInspection.is_qualified ? '合格' : '不合格'}
                </span>
              </div>
            </div>
            <h3 className="font-semibold mb-2">检查项目:</h3>
            <table className="min-w-full border">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">序号</th>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">检查项</th>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">结果</th>
                </tr>
              </thead>
              <tbody>
                {selectedInspection.items?.map((item, index) => (
                  <tr key={index} className="border-t">
                    <td className="px-4 py-2">{index + 1}</td>
                    <td className="px-4 py-2">{item.item_name}</td>
                    <td className="px-4 py-2">
                      <span className={`px-2 py-1 rounded text-xs ${
                        item.result === '不合格' ? 'bg-red-100 text-red-800' :
                        item.result === '不适用' ? 'bg-gray-100 text-gray-800' : 'bg-green-100 text-green-800'
                      }`}>
                        {item.result}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            <div className="flex justify-end mt-6">
              <button
                onClick={() => setShowDetailModal(false)}
                className="bg-gray-200 text-gray-700 px-4 py-2 rounded hover:bg-gray-300"
              >
                关闭
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
