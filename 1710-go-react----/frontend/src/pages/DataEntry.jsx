import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { subjectAPI, protocolAPI, visitAPI } from '../services/api';

function DataEntry() {
  const { subjectId } = useParams();
  const navigate = useNavigate();

  const [subject, setSubject] = useState(null);
  const [protocol, setProtocol] = useState(null);
  const [selectedVisit, setSelectedVisit] = useState(null);
  const [records, setRecords] = useState([]);
  const [loading, setLoading] = useState(true);
  const [formData, setFormData] = useState({
    actual_date: new Date().toISOString().split('T')[0],
    vital_signs: '',
    lab_tests: '',
    other_data: '',
  });

  useEffect(() => {
    loadData();
  }, [subjectId]);

  const loadData = async () => {
    try {
      const subjectRes = await subjectAPI.get(subjectId);
      const subj = subjectRes.data;
      setSubject(subj);

      const protocolRes = await protocolAPI.get(subj.protocol_id);
      setProtocol(protocolRes.data);

      const recordsRes = await visitAPI.list({ subject_id: subjectId });
      setRecords(recordsRes.data);
    } catch (error) {
      console.error('加载数据失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const getVisitStatus = (visit) => {
    const record = records.find(r => r.visit_id === visit.id);
    if (record) {
      return { recorded: true, record };
    }

    if (!subject || !subject.enrollment_date) {
      return { recorded: false, outOfWindow: false };
    }

    const enrollmentDate = new Date(subject.enrollment_date);
    const expectedDate = new Date(enrollmentDate);
    expectedDate.setDate(expectedDate.getDate() + visit.window_days);
    const minDate = new Date(expectedDate);
    minDate.setDate(minDate.getDate() - visit.window_tolerance);
    const maxDate = new Date(expectedDate);
    maxDate.setDate(maxDate.getDate() + visit.window_tolerance);
    const today = new Date();

    return {
      recorded: false,
      outOfWindow: today > maxDate,
      minDate,
      maxDate,
    };
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!selectedVisit) {
      alert('请选择访视');
      return;
    }

    try {
      await visitAPI.record({
        protocol_id: protocol.id,
        subject_id: subjectId,
        visit_id: selectedVisit.id,
        actual_date: new Date(formData.actual_date),
        vital_signs: formData.vital_signs,
        lab_tests: formData.lab_tests,
        other_data: formData.other_data,
      });
      alert('数据录入成功');
      loadData();
      setFormData({
        actual_date: new Date().toISOString().split('T')[0],
        vital_signs: '',
        lab_tests: '',
        other_data: '',
      });
    } catch (error) {
      alert('录入失败: ' + (error.response?.data?.error || error.message));
    }
  };

  if (loading) {
    return <div className="p-8 text-center">加载中...</div>;
  }

  if (!subject || !protocol) {
    return <div className="p-8 text-center text-red-500">数据加载失败</div>;
  }

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-800">数据录入</h1>
          <p className="text-sm text-gray-500 mt-1">
            受试者: {subject.randomization_id} ({subject.name_initials}) | 
            方案: {protocol.protocol_number}
          </p>
        </div>
        <button
          onClick={() => navigate('/subjects')}
          className="px-4 py-2 text-gray-600 hover:text-gray-800"
        >
          返回列表
        </button>
      </div>

      <div className="flex gap-6">
        <div className="w-1/3 bg-white shadow rounded-lg overflow-hidden">
          <div className="px-4 py-3 bg-gray-50 border-b font-semibold text-gray-700">
            访视列表
          </div>
          <div className="divide-y">
            {protocol.visits?.sort((a, b) => a.visit_order - b.visit_order).map((visit) => {
              const status = getVisitStatus(visit);
              const isSelected = selectedVisit?.id === visit.id;
              return (
                <div
                  key={visit.id}
                  onClick={() => !status.recorded && setSelectedVisit(visit)}
                  className={`p-4 cursor-pointer transition ${
                    status.outOfWindow ? 'bg-red-50' : 
                    isSelected ? 'bg-blue-50' : 'hover:bg-gray-50'
                  } ${status.recorded ? 'opacity-60 cursor-not-allowed' : ''}`}
                >
                  <div className="flex justify-between items-start">
                    <div>
                      <div className="font-medium text-gray-900">{visit.visit_name}</div>
                      <div className="text-sm text-gray-500 mt-1">
                        计划: 入组后{visit.window_days}天 ±{visit.window_tolerance}天
                      </div>
                    </div>
                    <div>
                      {status.recorded ? (
                        <span className="text-xs bg-green-100 text-green-800 px-2 py-1 rounded">已录入</span>
                      ) : status.outOfWindow ? (
                        <span className="text-xs bg-red-100 text-red-800 px-2 py-1 rounded">超窗</span>
                      ) : (
                        <span className="text-xs bg-blue-100 text-blue-800 px-2 py-1 rounded">待录入</span>
                      )}
                    </div>
                  </div>
                </div>
              );
            })}
            {(!protocol.visits || protocol.visits.length === 0) && (
              <div className="p-8 text-center text-gray-500">
                该方案暂无访视计划
              </div>
            )}
          </div>
        </div>

        <div className="w-2/3 bg-white shadow rounded-lg p-6">
          {selectedVisit ? (
            <div>
              <h2 className="text-lg font-semibold text-gray-800 mb-4">
                录入访视数据: {selectedVisit.visit_name}
              </h2>
              <form onSubmit={handleSubmit} className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">实际访视日期 *</label>
                  <input
                    type="date"
                    value={formData.actual_date}
                    onChange={(e) => setFormData({ ...formData, actual_date: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    required
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">生命体征</label>
                  <textarea
                    value={formData.vital_signs}
                    onChange={(e) => setFormData({ ...formData, vital_signs: e.target.value })}
                    rows={3}
                    placeholder="体温、心率、血压、呼吸频率等"
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">实验室检查</label>
                  <textarea
                    value={formData.lab_tests}
                    onChange={(e) => setFormData({ ...formData, lab_tests: e.target.value })}
                    rows={3}
                    placeholder="血常规、肝肾功能等"
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">其他数据</label>
                  <textarea
                    value={formData.other_data}
                    onChange={(e) => setFormData({ ...formData, other_data: e.target.value })}
                    rows={3}
                    placeholder="其他需要记录的信息"
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div className="flex justify-end gap-3 pt-4">
                  <button
                    type="button"
                    onClick={() => setSelectedVisit(null)}
                    className="px-4 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50"
                  >
                    取消
                  </button>
                  <button
                    type="submit"
                    className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
                  >
                    保存数据
                  </button>
                </div>
              </form>
            </div>
          ) : (
            <div className="text-center py-20 text-gray-500">
              请从左侧列表选择一个待录入的访视
            </div>
          )}

          {records.length > 0 && (
            <div className="mt-8">
              <h3 className="text-lg font-semibold text-gray-800 mb-4">已录入的记录</h3>
              <div className="space-y-3">
                {records.map((record) => {
                  const visit = protocol.visits?.find(v => v.id === record.visit_id);
                  return (
                    <div key={record.id} className="border rounded-lg p-4">
                      <div className="flex justify-between items-start">
                        <div>
                          <div className="font-medium">{visit?.visit_name || '未知访视'}</div>
                          <div className="text-sm text-gray-500 mt-1">
                            录入日期: {new Date(record.actual_date).toLocaleDateString()}
                          </div>
                        </div>
                        {record.is_out_of_window && (
                          <span className="text-xs bg-red-100 text-red-800 px-2 py-1 rounded">超窗</span>
                        )}
                      </div>
                      {record.vital_signs && (
                        <div className="mt-2 text-sm"><strong>生命体征:</strong> {record.vital_signs}</div>
                      )}
                      {record.lab_tests && (
                        <div className="mt-1 text-sm"><strong>实验室检查:</strong> {record.lab_tests}</div>
                      )}
                      {record.other_data && (
                        <div className="mt-1 text-sm"><strong>其他:</strong> {record.other_data}</div>
                      )}
                    </div>
                  );
                })}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default DataEntry;
