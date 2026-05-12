import { useState, useEffect } from 'react';
import { protocolAPI } from '../services/api';
import { useNavigate, useParams } from 'react-router-dom';

function ProtocolForm() {
  const navigate = useNavigate();
  const { id } = useParams();
  const isEdit = !!id;

  const [formData, setFormData] = useState({
    protocol_number: '',
    drug_name: '',
    indication: '',
    trial_phase: 'II期',
    planned_enrollment: 100,
    start_date: '',
    end_date: '',
    inclusion_criteria: '',
    exclusion_criteria: '',
    status: '筹备中',
    group_ratio: '1:1',
  });

  const [sites, setSites] = useState([]);
  const [visits, setVisits] = useState([]);
  const [newSite, setNewSite] = useState({
    site_code: '',
    site_name: '',
    principal_investigator: '',
    planned_enrollment: 50,
  });
  const [newVisit, setNewVisit] = useState({
    visit_name: '',
    visit_order: 0,
    window_days: 0,
    window_tolerance: 0,
  });

  useEffect(() => {
    if (isEdit) {
      loadProtocol();
    }
  }, [id]);

  const loadProtocol = async () => {
    try {
      const res = await protocolAPI.get(id);
      const p = res.data;
      setFormData({
        protocol_number: p.protocol_number,
        drug_name: p.drug_name,
        indication: p.indication,
        trial_phase: p.trial_phase,
        planned_enrollment: p.planned_enrollment,
        start_date: new Date(p.start_date).toISOString().split('T')[0],
        end_date: new Date(p.end_date).toISOString().split('T')[0],
        inclusion_criteria: p.inclusion_criteria || '',
        exclusion_criteria: p.exclusion_criteria || '',
        status: p.status,
        group_ratio: p.group_ratio || '1:1',
      });
      setSites(p.sites || []);
      setVisits(p.visits || []);
    } catch (error) {
      console.error('加载方案失败:', error);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      const data = {
        ...formData,
        start_date: new Date(formData.start_date),
        end_date: new Date(formData.end_date),
        sites,
        visits,
      };

      if (isEdit) {
        await protocolAPI.update(id, formData);
      } else {
        await protocolAPI.create(data);
      }
      navigate('/protocols');
    } catch (error) {
      alert('保存失败: ' + (error.response?.data?.error || error.message));
    }
  };

  const addSite = async () => {
    if (!newSite.site_code || !newSite.site_name || !newSite.principal_investigator) {
      alert('请填写完整的中心信息');
      return;
    }
    
    if (isEdit) {
      try {
        await protocolAPI.addSite(id, newSite);
        loadProtocol();
      } catch (error) {
        alert('添加中心失败: ' + (error.response?.data?.error || error.message));
        return;
      }
    } else {
      setSites([...sites, { ...newSite, id: Date.now() }]);
    }
    setNewSite({ site_code: '', site_name: '', principal_investigator: '', planned_enrollment: 50 });
  };

  const addVisit = async () => {
    if (!newVisit.visit_name) {
      alert('请填写访视名称');
      return;
    }

    if (isEdit) {
      try {
        await protocolAPI.addVisit(id, newVisit);
        loadProtocol();
      } catch (error) {
        alert('添加访视失败: ' + (error.response?.data?.error || error.message));
        return;
      }
    } else {
      setVisits([...visits, { ...newVisit, id: Date.now() }]);
    }
    setNewVisit({ visit_name: '', visit_order: visits.length + 1, window_days: 0, window_tolerance: 0 });
  };

  return (
    <div className="p-6 max-w-6xl mx-auto">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold text-gray-800">
          {isEdit ? '编辑试验方案' : '新建试验方案'}
        </h1>
        <button
          onClick={() => navigate('/protocols')}
          className="px-4 py-2 text-gray-600 hover:text-gray-800"
        >
          返回列表
        </button>
      </div>

      <form onSubmit={handleSubmit} className="space-y-8">
        <div className="bg-white shadow rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-4 text-gray-700">基本信息</h2>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">方案编号 *</label>
              <input
                type="text"
                value={formData.protocol_number}
                onChange={(e) => setFormData({ ...formData, protocol_number: e.target.value })}
                disabled={isEdit}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100"
                placeholder="CT-2026-001"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">药物名称 *</label>
              <input
                type="text"
                value={formData.drug_name}
                onChange={(e) => setFormData({ ...formData, drug_name: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">适应症 *</label>
              <input
                type="text"
                value={formData.indication}
                onChange={(e) => setFormData({ ...formData, indication: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">试验阶段</label>
              <select
                value={formData.trial_phase}
                onChange={(e) => setFormData({ ...formData, trial_phase: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="I期">I期</option>
                <option value="II期">II期</option>
                <option value="III期">III期</option>
                <option value="IV期">IV期</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">计划入组人数</label>
              <input
                type="number"
                value={formData.planned_enrollment}
                onChange={(e) => setFormData({ ...formData, planned_enrollment: parseInt(e.target.value) })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">分组比例</label>
              <input
                type="text"
                value={formData.group_ratio}
                onChange={(e) => setFormData({ ...formData, group_ratio: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="例如: 2:1 表示试验组:对照组"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">开始日期</label>
              <input
                type="date"
                value={formData.start_date}
                onChange={(e) => setFormData({ ...formData, start_date: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
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
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">状态</label>
              <select
                value={formData.status}
                onChange={(e) => setFormData({ ...formData, status: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="筹备中">筹备中</option>
                <option value="进行中">进行中</option>
                <option value="已完成">已完成</option>
                <option value="已终止">已终止</option>
              </select>
            </div>
          </div>
          <div className="mt-4">
            <label className="block text-sm font-medium text-gray-700 mb-1">纳入标准</label>
            <textarea
              value={formData.inclusion_criteria}
              onChange={(e) => setFormData({ ...formData, inclusion_criteria: e.target.value })}
              rows={3}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div className="mt-4">
            <label className="block text-sm font-medium text-gray-700 mb-1">排除标准</label>
            <textarea
              value={formData.exclusion_criteria}
              onChange={(e) => setFormData({ ...formData, exclusion_criteria: e.target.value })}
              rows={3}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
        </div>

        <div className="bg-white shadow rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-4 text-gray-700">研究中心</h2>
          <div className="grid grid-cols-4 gap-3 mb-4">
            <input
              type="text"
              value={newSite.site_code}
              onChange={(e) => setNewSite({ ...newSite, site_code: e.target.value })}
              placeholder="中心编码 (如S001)"
              className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <input
              type="text"
              value={newSite.site_name}
              onChange={(e) => setNewSite({ ...newSite, site_name: e.target.value })}
              placeholder="中心名称"
              className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <input
              type="text"
              value={newSite.principal_investigator}
              onChange={(e) => setNewSite({ ...newSite, principal_investigator: e.target.value })}
              placeholder="主要研究者(PI)"
              className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <div className="flex gap-2">
              <input
                type="number"
                value={newSite.planned_enrollment}
                onChange={(e) => setNewSite({ ...newSite, planned_enrollment: parseInt(e.target.value) })}
                placeholder="计划入组"
                className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <button
                type="button"
                onClick={addSite}
                className="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700"
              >
                添加
              </button>
            </div>
          </div>
          <div className="overflow-x-auto">
            <table className="min-w-full text-sm">
              <thead>
                <tr className="bg-gray-50">
                  <th className="px-4 py-2 text-left">中心编码</th>
                  <th className="px-4 py-2 text-left">中心名称</th>
                  <th className="px-4 py-2 text-left">PI</th>
                  <th className="px-4 py-2 text-left">计划入组</th>
                </tr>
              </thead>
              <tbody>
                {sites.map((site, idx) => (
                  <tr key={site.id || idx} className="border-t">
                    <td className="px-4 py-2">{site.site_code}</td>
                    <td className="px-4 py-2">{site.site_name}</td>
                    <td className="px-4 py-2">{site.principal_investigator}</td>
                    <td className="px-4 py-2">{site.planned_enrollment}</td>
                  </tr>
                ))}
                {sites.length === 0 && (
                  <tr><td colSpan={4} className="px-4 py-8 text-center text-gray-500">暂无中心</td></tr>
                )}
              </tbody>
            </table>
          </div>
        </div>

        <div className="bg-white shadow rounded-lg p-6">
          <h2 className="text-lg font-semibold mb-4 text-gray-700">访视计划</h2>
          <div className="grid grid-cols-5 gap-3 mb-4">
            <input
              type="text"
              value={newVisit.visit_name}
              onChange={(e) => setNewVisit({ ...newVisit, visit_name: e.target.value })}
              placeholder="访视名称 (如V0筛选期)"
              className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <input
              type="number"
              value={newVisit.visit_order}
              onChange={(e) => setNewVisit({ ...newVisit, visit_order: parseInt(e.target.value) })}
              placeholder="顺序"
              className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <input
              type="number"
              value={newVisit.window_days}
              onChange={(e) => setNewVisit({ ...newVisit, window_days: parseInt(e.target.value) })}
              placeholder="计划天数"
              className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <input
              type="number"
              value={newVisit.window_tolerance}
              onChange={(e) => setNewVisit({ ...newVisit, window_tolerance: parseInt(e.target.value) })}
              placeholder="窗口天数"
              className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <button
              type="button"
              onClick={addVisit}
              className="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700"
            >
              添加
            </button>
          </div>
          <div className="overflow-x-auto">
            <table className="min-w-full text-sm">
              <thead>
                <tr className="bg-gray-50">
                  <th className="px-4 py-2 text-left">顺序</th>
                  <th className="px-4 py-2 text-left">访视名称</th>
                  <th className="px-4 py-2 text-left">计划天数</th>
                  <th className="px-4 py-2 text-left">窗口天数</th>
                </tr>
              </thead>
              <tbody>
                {visits.map((visit, idx) => (
                  <tr key={visit.id || idx} className="border-t">
                    <td className="px-4 py-2">{visit.visit_order}</td>
                    <td className="px-4 py-2">{visit.visit_name}</td>
                    <td className="px-4 py-2">{visit.window_days}天</td>
                    <td className="px-4 py-2">±{visit.window_tolerance}天</td>
                  </tr>
                ))}
                {visits.length === 0 && (
                  <tr><td colSpan={4} className="px-4 py-8 text-center text-gray-500">暂无访视</td></tr>
                )}
              </tbody>
            </table>
          </div>
        </div>

        <div className="flex justify-end gap-4">
          <button
            type="button"
            onClick={() => navigate('/protocols')}
            className="px-6 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50"
          >
            取消
          </button>
          <button
            type="submit"
            className="px-6 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
          >
            {isEdit ? '保存修改' : '创建方案'}
          </button>
        </div>
      </form>
    </div>
  );
}

export default ProtocolForm;
