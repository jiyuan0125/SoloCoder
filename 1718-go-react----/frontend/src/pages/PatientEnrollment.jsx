import { useState, useEffect } from 'react';
import { pathsApi, patientsApi, enrollmentsApi } from '../api';

function PatientEnrollment() {
  const [patients, setPatients] = useState([]);
  const [paths, setPaths] = useState([]);
  const [enrollments, setEnrollments] = useState([]);
  const [activeTab, setActiveTab] = useState('eligible');
  const [selectedPatient, setSelectedPatient] = useState(null);
  const [selectedPath, setSelectedPath] = useState(null);
  const [exitReason, setExitReason] = useState('');
  const [showExitModal, setShowExitModal] = useState(false);
  const [selectedEnrollment, setSelectedEnrollment] = useState(null);
  const [message, setMessage] = useState({ type: '', text: '' });
  const [activeEnrollmentsMap, setActiveEnrollmentsMap] = useState({});

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      const [patientsRes, pathsRes, enrollmentsRes] = await Promise.all([
        patientsApi.list(),
        pathsApi.list(),
        enrollmentsApi.list(),
      ]);
      setPatients(patientsRes.data.data || []);
      setPaths(pathsRes.data.data || []);
      const enrData = enrollmentsRes.data.data || [];
      setEnrollments(enrData);
      
      const activeMap = {};
      enrData.forEach(e => {
        if (e.status === 'active') {
          activeMap[e.patient_id] = e;
        }
      });
      setActiveEnrollmentsMap(activeMap);
    } catch (err) {
      showMessage('error', '加载数据失败');
    }
  };

  const showMessage = (type, text) => {
    setMessage({ type, text });
    setTimeout(() => setMessage({ type: '', text: '' }), 3000);
  };

  const isEligibleForPath = (patient, path) => {
    const icdCodes = path.icd_codes.split(',').map(c => c.trim()).filter(c => c);
    return icdCodes.some(code => patient.diagnosis.includes(code));
  };

  const getEligiblePaths = (patient) => {
    return paths.filter(p => isEligibleForPath(patient, p));
  };

  const getActiveEnrollment = (patientId) => {
    return activeEnrollmentsMap[patientId];
  };

  const handleEnroll = async () => {
    if (!selectedPatient || !selectedPath) return;
    try {
      await enrollmentsApi.create({
        patient_id: selectedPatient.id,
        path_id: selectedPath.id,
      });
      showMessage('success', '患者入径成功');
      setSelectedPatient(null);
      setSelectedPath(null);
      loadData();
    } catch (err) {
      const errorMsg = err.response?.data?.error || '入径失败';
      showMessage('error', errorMsg);
    }
  };

  const handleExit = async () => {
    if (!selectedEnrollment || !exitReason) return;
    try {
      await enrollmentsApi.exit(selectedEnrollment.id, { exit_reason: exitReason });
      showMessage('success', '退径成功');
      setShowExitModal(false);
      setExitReason('');
      setSelectedEnrollment(null);
      loadData();
    } catch (err) {
      showMessage('error', '退径失败');
    }
  };

  const getStatusBadge = (status) => {
    const styles = {
      active: 'bg-green-100 text-green-700',
      completed: 'bg-blue-100 text-blue-700',
      exited: 'bg-gray-100 text-gray-700',
    };
    const labels = {
      active: '进行中',
      completed: '已完成',
      exited: '已退径',
    };
    return (
      <span className={`text-xs px-2 py-1 rounded ${styles[status] || 'bg-gray-100'}`}>
        {labels[status] || status}
      </span>
    );
  };

  const getPathName = (pathId) => {
    const path = paths.find(p => p.id === pathId);
    return path ? path.name : '未知路径';
  };

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold text-gray-800 mb-6">患者入径管理</h1>

      {message.text && (
        <div className={`mb-4 p-3 rounded ${message.type === 'success' ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}`}>
          {message.text}
        </div>
      )}

      <div className="mb-4">
        <div className="flex gap-4 border-b">
          <button
            onClick={() => setActiveTab('eligible')}
            className={`px-4 py-2 font-medium ${activeTab === 'eligible' ? 'text-blue-600 border-b-2 border-blue-600' : 'text-gray-500'}`}
          >
            待入径患者
          </button>
          <button
            onClick={() => setActiveTab('enrolled')}
            className={`px-4 py-2 font-medium ${activeTab === 'enrolled' ? 'text-blue-600 border-b-2 border-blue-600' : 'text-gray-500'}`}
          >
            已入径患者
          </button>
        </div>
      </div>

      {activeTab === 'eligible' && (
        <div className="grid grid-cols-3 gap-6">
          <div className="col-span-2">
            <div className="bg-white rounded-lg shadow">
              <div className="p-4 border-b font-semibold">患者列表</div>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-4 py-3 text-left text-sm font-medium text-gray-600">住院号</th>
                      <th className="px-4 py-3 text-left text-sm font-medium text-gray-600">姓名</th>
                      <th className="px-4 py-3 text-left text-sm font-medium text-gray-600">诊断</th>
                      <th className="px-4 py-3 text-left text-sm font-medium text-gray-600">状态</th>
                      <th className="px-4 py-3 text-left text-sm font-medium text-gray-600">操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    {patients.filter(p => !getActiveEnrollment(p.id)).map((patient) => {
                      const eligiblePaths = getEligiblePaths(patient);
                      return (
                        <tr key={patient.id} className={`border-t hover:bg-gray-50 cursor-pointer ${selectedPatient?.id === patient.id ? 'bg-blue-50' : ''}`}
                            onClick={() => setSelectedPatient(patient)}>
                          <td className="px-4 py-3 text-sm">{patient.hospital_no}</td>
                          <td className="px-4 py-3 text-sm">{patient.name}</td>
                          <td className="px-4 py-3 text-sm">{patient.diagnosis}</td>
                          <td className="px-4 py-3 text-sm">
                            {eligiblePaths.length > 0 ? (
                              <span className="text-green-600">符合{eligiblePaths.length}个路径</span>
                            ) : (
                              <span className="text-gray-400">暂无匹配路径</span>
                            )}
                          </td>
                          <td className="px-4 py-3 text-sm">
                            {eligiblePaths.length > 0 && (
                              <span className="text-blue-600 hover:underline">选择入径</span>
                            )}
                          </td>
                        </tr>
                      );
                    })}
                    {patients.filter(p => !getActiveEnrollment(p.id)).length === 0 && (
                      <tr><td colSpan="5" className="px-4 py-8 text-center text-gray-400">暂无待入径患者</td></tr>
                    )}
                  </tbody>
                </table>
              </div>
            </div>
          </div>

          <div className="col-span-1">
            <div className="bg-white rounded-lg shadow p-4">
              <h3 className="font-semibold mb-4">办理入径</h3>
              {selectedPatient ? (
                <div className="space-y-4">
                  <div className="p-3 bg-gray-50 rounded">
                    <div className="font-medium">{selectedPatient.name}</div>
                    <div className="text-sm text-gray-500">{selectedPatient.hospital_no}</div>
                    <div className="text-sm text-gray-600">{selectedPatient.diagnosis}</div>
                  </div>
                  
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">选择路径</label>
                    <select
                      className="w-full border rounded px-3 py-2"
                      value={selectedPath?.id || ''}
                      onChange={(e) => setSelectedPath(paths.find(p => p.id === parseInt(e.target.value)))}
                    >
                      <option value="">请选择路径</option>
                      {getEligiblePaths(selectedPatient).map((path) => (
                        <option key={path.id} value={path.id}>{path.name} ({path.code})</option>
                      ))}
                    </select>
                  </div>

                  <button
                    onClick={handleEnroll}
                    disabled={!selectedPath}
                    className="w-full bg-blue-600 text-white py-2 rounded hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed"
                  >
                    确认入径
                  </button>
                </div>
              ) : (
                <div className="text-gray-400 text-center py-8">请选择患者</div>
              )}
            </div>
          </div>
        </div>
      )}

      {activeTab === 'enrolled' && (
        <div className="bg-white rounded-lg shadow">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left text-sm font-medium text-gray-600">住院号</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-gray-600">姓名</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-gray-600">路径</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-gray-600">入径日期</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-gray-600">变异次数</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-gray-600">状态</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-gray-600">提示</th>
                  <th className="px-4 py-3 text-left text-sm font-medium text-gray-600">操作</th>
                </tr>
              </thead>
              <tbody>
                {enrollments.map((enrollment) => (
                  <tr key={enrollment.id} className="border-t hover:bg-gray-50">
                    <td className="px-4 py-3 text-sm">{enrollment.patient_hospital_no}</td>
                    <td className="px-4 py-3 text-sm">{enrollment.patient_name}</td>
                    <td className="px-4 py-3 text-sm">{getPathName(enrollment.path_id)}</td>
                    <td className="px-4 py-3 text-sm">{new Date(enrollment.enroll_date).toLocaleDateString()}</td>
                    <td className="px-4 py-3 text-sm">{enrollment.variation_count}</td>
                    <td className="px-4 py-3 text-sm">{getStatusBadge(enrollment.status)}</td>
                    <td className="px-4 py-3 text-sm">
                      {enrollment.suggest_exit && enrollment.status === 'active' ? (
                        <span className="text-orange-600 font-medium">建议退径</span>
                      ) : (
                        <span className="text-gray-400">-</span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-sm">
                      {enrollment.status === 'active' && (
                        <button
                          onClick={() => { setSelectedEnrollment(enrollment); setShowExitModal(true); }}
                          className="text-red-600 hover:underline"
                        >
                          退径
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
                {enrollments.length === 0 && (
                  <tr><td colSpan="8" className="px-4 py-8 text-center text-gray-400">暂无入径记录</td></tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {showExitModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-96">
            <h3 className="text-lg font-semibold mb-4">办理退径</h3>
            <div className="mb-4">
              <div className="text-sm text-gray-500 mb-1">患者</div>
              <div className="font-medium">{selectedEnrollment?.patient_name} ({selectedEnrollment?.patient_hospital_no})</div>
            </div>
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-1">退径原因</label>
              <textarea
                className="w-full border rounded px-3 py-2"
                rows="3"
                value={exitReason}
                onChange={(e) => setExitReason(e.target.value)}
                placeholder="请输入退径原因"
              ></textarea>
            </div>
            <div className="flex gap-2">
              <button
                onClick={handleExit}
                disabled={!exitReason}
                className="flex-1 bg-red-600 text-white py-2 rounded hover:bg-red-700 disabled:bg-gray-300"
              >
                确认退径
              </button>
              <button
                onClick={() => { setShowExitModal(false); setExitReason(''); }}
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

export default PatientEnrollment;
