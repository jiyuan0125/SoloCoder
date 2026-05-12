import { useState, useEffect } from 'react';
import { opinionAPI, penaltyAPI, inspectionAPI } from '../api';

export default function PenaltiesPage() {
  const [opinions, setOpinions] = useState([]);
  const [penalties, setPenalties] = useState([]);
  const [fineRanges, setFineRanges] = useState([]);
  const [activeTab, setActiveTab] = useState('opinions');
  const [showOpinionModal, setShowOpinionModal] = useState(false);
  const [showPenaltyModal, setShowPenaltyModal] = useState(false);
  const [selectedInspection, setSelectedInspection] = useState(null);
  const [filterStatus, setFilterStatus] = useState('');

  const [opinionForm, setOpinionForm] = useState({
    inspection_id: '',
    rectification_days: 7,
    remarks: '',
  });

  const [penaltyForm, setPenaltyForm] = useState({
    opinion_id: '',
    penalty_types: '',
    fine_amount: 0,
    fine_reason: '',
    remarks: '',
  });

  useEffect(() => {
    loadOpinions();
    loadPenalties();
    loadFineRanges();
  }, [filterStatus]);

  const loadOpinions = async () => {
    try {
      const params = {};
      if (filterStatus) params.status = filterStatus;
      const res = await opinionAPI.list(params);
      setOpinions(res.data.data);
    } catch (err) {
      console.error('Failed to load opinions:', err);
    }
  };

  const loadPenalties = async () => {
    try {
      const res = await penaltyAPI.list();
      setPenalties(res.data.data);
    } catch (err) {
      console.error('Failed to load penalties:', err);
    }
  };

  const loadFineRanges = async () => {
    try {
      const res = await penaltyAPI.getFineRanges();
      setFineRanges(res.data.data);
    } catch (err) {
      console.error('Failed to load fine ranges:', err);
    }
  };

  const handleIssueOpinion = async (e) => {
    e.preventDefault();
    try {
      await opinionAPI.create(opinionForm);
      setShowOpinionModal(false);
      loadOpinions();
      setOpinionForm({ inspection_id: '', rectification_days: 7, remarks: '' });
    } catch (err) {
      alert(err.response?.data?.error || '下达意见书失败');
    }
  };

  const handleReview = async (opinion, isPassed) => {
    if (!confirm(`确认复查结果为${isPassed ? '合格' : '不合格'}？`)) return;
    try {
      await opinionAPI.review(opinion.id, { is_passed: isPassed, remarks: '' });
      loadOpinions();
    } catch (err) {
      alert('复查操作失败');
    }
  };

  const handleIssuePenalty = async (e) => {
    e.preventDefault();
    try {
      await penaltyAPI.create(penaltyForm);
      setShowPenaltyModal(false);
      loadPenalties();
      loadOpinions();
      setPenaltyForm({
        opinion_id: '',
        penalty_types: '',
        fine_amount: 0,
        fine_reason: '',
        remarks: '',
      });
    } catch (err) {
      alert(err.response?.data?.error || '下达处罚失败');
    }
  };

  const openOpinionModal = async (inspectionId) => {
    try {
      const res = await inspectionAPI.get(inspectionId);
      setSelectedInspection(res.data.data);
      setOpinionForm({ ...opinionForm, inspection_id: inspectionId });
      setShowOpinionModal(true);
    } catch (err) {
      console.error('Failed to load inspection:', err);
    }
  };

  const getStatusColor = (status) => {
    switch (status) {
      case '待整改': return 'bg-yellow-100 text-yellow-800';
      case '待复查': return 'bg-blue-100 text-blue-800';
      case '整改合格': return 'bg-green-100 text-green-800';
      case '整改失败': return 'bg-red-100 text-red-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-800 mb-2">处罚管理</h1>
      </div>

      <div className="flex gap-2 mb-6">
        <button
          onClick={() => setActiveTab('opinions')}
          className={`px-4 py-2 rounded ${activeTab === 'opinions' ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-700'}`}
        >
          卫生监督意见书
        </button>
        <button
          onClick={() => setActiveTab('penalties')}
          className={`px-4 py-2 rounded ${activeTab === 'penalties' ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-700'}`}
        >
          行政处罚
        </button>
      </div>

      {activeTab === 'opinions' && (
        <div>
          <div className="flex flex-wrap gap-4 mb-6">
            <div className="flex gap-4 items-center">
              <label className="text-gray-600">状态:</label>
              <select
                value={filterStatus}
                onChange={(e) => setFilterStatus(e.target.value)}
                className="border rounded px-3 py-2"
              >
                <option value="">全部</option>
                <option value="待整改">待整改</option>
                <option value="待复查">待复查</option>
                <option value="整改合格">整改合格</option>
                <option value="整改失败">整改失败</option>
              </select>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow overflow-hidden">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">单位</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">下达日期</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">整改期限</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">截止日期</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">状态</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">操作</th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {opinions.map((op) => (
                  <tr key={op.id}>
                    <td className="px-4 py-3 whitespace-nowrap">
                      {op.inspection_record?.unit?.name}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap">
                      {new Date(op.issue_date).toLocaleDateString()}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap">{op.rectification_days}天</td>
                    <td className="px-4 py-3 whitespace-nowrap">
                      {new Date(op.deadline).toLocaleDateString()}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap">
                      <span className={`px-2 py-1 rounded text-xs font-medium ${getStatusColor(op.status)}`}>
                        {op.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap text-sm">
                      {op.status === '待整改' && (
                        <button
                          onClick={() => handleReview(op, true)}
                          className="text-green-600 hover:text-green-900 mr-3"
                        >
                          复查合格
                        </button>
                      )}
                      {op.status === '待整改' && (
                        <button
                          onClick={() => handleReview(op, false)}
                          className="text-red-600 hover:text-red-900"
                        >
                          复查不合格
                        </button>
                      )}
                      {op.status === '整改失败' && (
                        <button
                          onClick={() => {
                            setPenaltyForm({ ...penaltyForm, opinion_id: op.id });
                            setShowPenaltyModal(true);
                          }}
                          className="text-orange-600 hover:text-orange-900"
                        >
                          下达行政处罚
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {activeTab === 'penalties' && (
        <div>
          <div className="bg-white rounded-lg shadow overflow-hidden">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">单位</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">处罚类型</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">罚款金额</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">罚款原因</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">下达日期</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">备注</th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {penalties.map((p) => (
                  <tr key={p.id}>
                    <td className="px-4 py-3 whitespace-nowrap">
                      {p.opinion?.inspection_record?.unit?.name}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap">{p.penalty_types}</td>
                    <td className="px-4 py-3 whitespace-nowrap">
                      {p.fine_amount > 0 ? `¥${p.fine_amount.toLocaleString()}` : '-'}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap">{p.fine_reason || '-'}</td>
                    <td className="px-4 py-3 whitespace-nowrap">
                      {new Date(p.issue_date).toLocaleDateString()}
                    </td>
                    <td className="px-4 py-3 whitespace-nowrap">{p.remarks || '-'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {fineRanges.length > 0 && (
            <div className="mt-6 p-4 bg-gray-50 rounded-lg">
              <h3 className="font-semibold mb-2">罚款金额法定范围参考</h3>
              <ul className="list-disc list-inside text-sm text-gray-600 space-y-1">
                {fineRanges.map((r, idx) => (
                  <li key={idx}>
                    <strong>{r.reason}:</strong> ¥{r.min.toLocaleString()} - ¥{r.max.toLocaleString()}
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
      )}

      {showOpinionModal && selectedInspection && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-lg p-6 w-full max-w-md">
            <h2 className="text-xl font-bold mb-4">下达卫生监督意见书</h2>
            <div className="mb-4 p-3 bg-blue-50 rounded">
              <p><strong>单位:</strong> {selectedInspection.unit?.name}</p>
              <p><strong>检查日期:</strong> {new Date(selectedInspection.inspection_date).toLocaleDateString()}</p>
              <p><strong>合格率:</strong> {selectedInspection.pass_rate?.toFixed(1)}%</p>
            </div>
            <form onSubmit={handleIssueOpinion} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">整改期限</label>
                <select
                  value={opinionForm.rectification_days}
                  onChange={(e) => setOpinionForm({...opinionForm, rectification_days: parseInt(e.target.value)})}
                  className="w-full border rounded px-3 py-2"
                >
                  <option value={7}>7天（一般问题）</option>
                  <option value={15}>15天（严重问题）</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">备注</label>
                <textarea
                  value={opinionForm.remarks}
                  onChange={(e) => setOpinionForm({...opinionForm, remarks: e.target.value})}
                  className="w-full border rounded px-3 py-2"
                  rows={3}
                />
              </div>
              <div className="flex gap-3">
                <button
                  type="submit"
                  className="flex-1 bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
                >
                  下达
                </button>
                <button
                  type="button"
                  onClick={() => setShowOpinionModal(false)}
                  className="flex-1 bg-gray-200 text-gray-700 px-4 py-2 rounded hover:bg-gray-300"
                >
                  取消
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {showPenaltyModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-lg p-6 w-full max-w-md">
            <h2 className="text-xl font-bold mb-4">下达行政处罚</h2>
            <form onSubmit={handleIssuePenalty} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">处罚类型（可多选，用逗号分隔）*</label>
                <input
                  type="text"
                  value={penaltyForm.penalty_types}
                  onChange={(e) => setPenaltyForm({...penaltyForm, penalty_types: e.target.value})}
                  placeholder="警告,罚款,停业整顿,吊销许可证"
                  className="w-full border rounded px-3 py-2"
                  required
                />
                <p className="text-xs text-gray-500 mt-1">可选值: 警告、罚款、停业整顿、吊销许可证</p>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">罚款原因</label>
                <select
                  value={penaltyForm.fine_reason}
                  onChange={(e) => setPenaltyForm({...penaltyForm, fine_reason: e.target.value})}
                  className="w-full border rounded px-3 py-2"
                >
                  <option value="">请选择</option>
                  {fineRanges.map((r, idx) => (
                    <option key={idx} value={r.reason}>
                      {r.reason} (¥{r.min.toLocaleString()}-¥{r.max.toLocaleString()})
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">罚款金额</label>
                <input
                  type="number"
                  value={penaltyForm.fine_amount}
                  onChange={(e) => setPenaltyForm({...penaltyForm, fine_amount: parseFloat(e.target.value) || 0})}
                  className="w-full border rounded px-3 py-2"
                  min="0"
                  step="0.01"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">备注</label>
                <textarea
                  value={penaltyForm.remarks}
                  onChange={(e) => setPenaltyForm({...penaltyForm, remarks: e.target.value})}
                  className="w-full border rounded px-3 py-2"
                  rows={3}
                />
              </div>
              <div className="flex gap-3">
                <button
                  type="submit"
                  className="flex-1 bg-red-600 text-white px-4 py-2 rounded hover:bg-red-700"
                >
                  下达处罚
                </button>
                <button
                  type="button"
                  onClick={() => setShowPenaltyModal(false)}
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
