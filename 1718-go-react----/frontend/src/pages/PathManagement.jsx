import { useState, useEffect } from 'react';
import { pathsApi } from '../api';

function PathManagement() {
  const [paths, setPaths] = useState([]);
  const [selectedPath, setSelectedPath] = useState(null);
  const [stages, setStages] = useState([]);
  const [items, setItems] = useState([]);
  const [showForm, setShowForm] = useState(false);
  const [formData, setFormData] = useState({
    code: '', name: '', icd_codes: '', min_days: 5, max_days: 7, min_cost: 0, max_cost: 0
  });
  const [newStage, setNewStage] = useState({ name: '', days: 1, order: 0 });
  const [newItem, setNewItem] = useState({ stage_id: '', name: '', category: '检查检验', is_required: true });
  const [message, setMessage] = useState({ type: '', text: '' });

  useEffect(() => {
    loadPaths();
  }, []);

  const loadPaths = async () => {
    try {
      const res = await pathsApi.list();
      setPaths(res.data.data || []);
    } catch (err) {
      showMessage('error', '加载路径列表失败');
    }
  };

  const loadPathDetail = async (path) => {
    setSelectedPath(path);
    try {
      const [stagesRes, itemsRes] = await Promise.all([
        pathsApi.listStages(path.id),
        pathsApi.listItems(path.id),
      ]);
      setStages(stagesRes.data.data || []);
      setItems(itemsRes.data.data || []);
    } catch (err) {
      showMessage('error', '加载路径详情失败');
    }
  };

  const showMessage = (type, text) => {
    setMessage({ type, text });
    setTimeout(() => setMessage({ type: '', text: '' }), 3000);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      if (selectedPath) {
        await pathsApi.update(selectedPath.id, formData);
        showMessage('success', '路径更新成功');
      } else {
        await pathsApi.create(formData);
        showMessage('success', '路径创建成功');
      }
      setShowForm(false);
      setFormData({ code: '', name: '', icd_codes: '', min_days: 5, max_days: 7, min_cost: 0, max_cost: 0 });
      loadPaths();
    } catch (err) {
      const errorMsg = err.response?.data?.error || '操作失败';
      showMessage('error', errorMsg);
    }
  };

  const handleEdit = (path) => {
    setSelectedPath(path);
    setFormData({
      code: path.code,
      name: path.name,
      icd_codes: path.icd_codes,
      min_days: path.min_days,
      max_days: path.max_days,
      min_cost: path.min_cost,
      max_cost: path.max_cost,
    });
    setShowForm(true);
  };

  const handleAddStage = async () => {
    if (!selectedPath || !newStage.name) return;
    try {
      await pathsApi.createStage(selectedPath.id, newStage);
      showMessage('success', '阶段添加成功');
      setNewStage({ name: '', days: 1, order: stages.length + 1 });
      loadPathDetail(selectedPath);
    } catch (err) {
      showMessage('error', '添加阶段失败');
    }
  };

  const handleAddItem = async () => {
    if (!selectedPath || !newItem.stage_id || !newItem.name) return;
    try {
      await pathsApi.createItem(selectedPath.id, newItem);
      showMessage('success', '医嘱项目添加成功');
      setNewItem({ stage_id: '', name: '', category: '检查检验', is_required: true });
      loadPathDetail(selectedPath);
    } catch (err) {
      showMessage('error', '添加医嘱项目失败');
    }
  };

  const getItemsByStage = (stageId) => items.filter(i => i.stage_id === stageId);

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold text-gray-800">路径管理</h1>
        <button
          onClick={() => { setShowForm(true); setSelectedPath(null); setFormData({ code: '', name: '', icd_codes: '', min_days: 5, max_days: 7, min_cost: 0, max_cost: 0 }); }}
          className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
        >
          + 新建路径
        </button>
      </div>

      {message.text && (
        <div className={`mb-4 p-3 rounded ${message.type === 'success' ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}`}>
          {message.text}
        </div>
      )}

      {showForm && (
        <div className="mb-6 p-6 bg-white rounded-lg shadow">
          <h2 className="text-lg font-semibold mb-4">{selectedPath ? '编辑路径' : '新建路径'}</h2>
          <form onSubmit={handleSubmit} className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">路径编号</label>
              <input
                type="text"
                value={formData.code}
                onChange={(e) => setFormData({ ...formData, code: e.target.value })}
                className="w-full border rounded px-3 py-2"
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">路径名称</label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                className="w-full border rounded px-3 py-2"
                required
              />
            </div>
            <div className="col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-1">适用ICD编码</label>
              <input
                type="text"
                value={formData.icd_codes}
                onChange={(e) => setFormData({ ...formData, icd_codes: e.target.value })}
                className="w-full border rounded px-3 py-2"
                placeholder="用逗号分隔，如: K35,K35.0,K35.1"
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">标准住院天数下限</label>
              <input
                type="number"
                value={formData.min_days}
                onChange={(e) => setFormData({ ...formData, min_days: parseInt(e.target.value) })}
                className="w-full border rounded px-3 py-2"
                min="1"
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">标准住院天数上限</label>
              <input
                type="number"
                value={formData.max_days}
                onChange={(e) => setFormData({ ...formData, max_days: parseInt(e.target.value) })}
                className="w-full border rounded px-3 py-2"
                min="1"
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">标准费用下限(元)</label>
              <input
                type="number"
                step="0.01"
                value={formData.min_cost}
                onChange={(e) => setFormData({ ...formData, min_cost: parseFloat(e.target.value) })}
                className="w-full border rounded px-3 py-2"
                min="0"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">标准费用上限(元)</label>
              <input
                type="number"
                step="0.01"
                value={formData.max_cost}
                onChange={(e) => setFormData({ ...formData, max_cost: parseFloat(e.target.value) })}
                className="w-full border rounded px-3 py-2"
                min="0"
              />
            </div>
            <div className="col-span-2 flex gap-2">
              <button type="submit" className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700">
                保存
              </button>
              <button type="button" onClick={() => setShowForm(false)} className="bg-gray-200 px-4 py-2 rounded hover:bg-gray-300">
                取消
              </button>
            </div>
          </form>
        </div>
      )}

      <div className="grid grid-cols-3 gap-6">
        <div className="col-span-1">
          <div className="bg-white rounded-lg shadow p-4">
            <h2 className="text-lg font-semibold mb-3">路径列表</h2>
            <div className="space-y-2">
              {paths.map((path) => (
                <div
                  key={path.id}
                  onClick={() => loadPathDetail(path)}
                  className={`p-3 rounded border cursor-pointer ${selectedPath?.id === path.id ? 'bg-blue-50 border-blue-300' : 'hover:bg-gray-50'}`}
                >
                  <div className="font-medium">{path.name}</div>
                  <div className="text-sm text-gray-500">{path.code}</div>
                  <div className="text-xs text-gray-400">
                    标准住院: {path.min_days}-{path.max_days}天
                  </div>
                </div>
              ))}
              {paths.length === 0 && <div className="text-gray-400 text-center py-4">暂无数据</div>}
            </div>
          </div>
        </div>

        {selectedPath && (
          <div className="col-span-2 space-y-4">
            <div className="bg-white rounded-lg shadow p-4">
              <div className="flex justify-between items-center mb-4">
                <h2 className="text-lg font-semibold">{selectedPath.name}</h2>
                <button onClick={() => handleEdit(selectedPath)} className="text-blue-600 hover:underline">编辑</button>
              </div>
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div><span className="text-gray-500">路径编号:</span> {selectedPath.code}</div>
                <div><span className="text-gray-500">ICD编码:</span> {selectedPath.icd_codes}</div>
                <div><span className="text-gray-500">标准住院:</span> {selectedPath.min_days}-{selectedPath.max_days}天</div>
                <div><span className="text-gray-500">标准费用:</span> {selectedPath.min_cost}-{selectedPath.max_cost}元</div>
              </div>
            </div>

            <div className="bg-white rounded-lg shadow p-4">
              <h3 className="font-semibold mb-3">阶段定义</h3>
              <div className="space-y-3 mb-4">
                {stages.sort((a, b) => a.order - b.order).map((stage) => (
                  <div key={stage.id} className="border rounded p-3">
                    <div className="font-medium">{stage.name} (持续{stage.days}天)</div>
                    <div className="mt-2 space-y-1">
                      {getItemsByStage(stage.id).map((item) => (
                        <div key={item.id} className="flex items-center text-sm">
                          <span className={`w-2 h-2 rounded-full mr-2 ${item.is_required ? 'bg-red-500' : 'bg-gray-400'}`}></span>
                          <span>{item.name}</span>
                          <span className="text-gray-400 ml-2">({item.category})</span>
                          <span className={`ml-2 text-xs px-1 rounded ${item.is_required ? 'bg-red-100 text-red-600' : 'bg-gray-100 text-gray-600'}`}>
                            {item.is_required ? '必选' : '可选'}
                          </span>
                        </div>
                      ))}
                      {getItemsByStage(stage.id).length === 0 && (
                        <div className="text-gray-400 text-sm">暂无医嘱项目</div>
                      )}
                    </div>
                  </div>
                ))}
              </div>
              <div className="border-t pt-4">
                <h4 className="font-medium mb-2">添加阶段</h4>
                <div className="flex gap-2">
                  <input
                    type="text"
                    placeholder="阶段名称"
                    value={newStage.name}
                    onChange={(e) => setNewStage({ ...newStage, name: e.target.value })}
                    className="border rounded px-3 py-2 flex-1"
                  />
                  <input
                    type="number"
                    placeholder="天数"
                    value={newStage.days}
                    onChange={(e) => setNewStage({ ...newStage, days: parseInt(e.target.value) || 1 })}
                    className="border rounded px-3 py-2 w-24"
                    min="1"
                  />
                  <button onClick={handleAddStage} className="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700">
                    添加
                  </button>
                </div>
              </div>
            </div>

            {stages.length > 0 && (
              <div className="bg-white rounded-lg shadow p-4">
                <h3 className="font-semibold mb-3">添加医嘱项目</h3>
                <div className="grid grid-cols-5 gap-2">
                  <select
                    value={newItem.stage_id}
                    onChange={(e) => setNewItem({ ...newItem, stage_id: e.target.value })}
                    className="border rounded px-3 py-2"
                  >
                    <option value="">选择阶段</option>
                    {stages.map((s) => (
                      <option key={s.id} value={s.id}>{s.name}</option>
                    ))}
                  </select>
                  <input
                    type="text"
                    placeholder="项目名称"
                    value={newItem.name}
                    onChange={(e) => setNewItem({ ...newItem, name: e.target.value })}
                    className="border rounded px-3 py-2"
                  />
                  <select
                    value={newItem.category}
                    onChange={(e) => setNewItem({ ...newItem, category: e.target.value })}
                    className="border rounded px-3 py-2"
                  >
                    <option value="检查检验">检查检验</option>
                    <option value="用药">用药</option>
                    <option value="护理">护理</option>
                    <option value="饮食">饮食</option>
                    <option value="手术">手术</option>
                    <option value="其他">其他</option>
                  </select>
                  <select
                    value={newItem.is_required ? 'true' : 'false'}
                    onChange={(e) => setNewItem({ ...newItem, is_required: e.target.value === 'true' })}
                    className="border rounded px-3 py-2"
                  >
                    <option value="true">必选</option>
                    <option value="false">可选</option>
                  </select>
                  <button onClick={handleAddItem} className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700">
                    添加
                  </button>
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

export default PathManagement;
