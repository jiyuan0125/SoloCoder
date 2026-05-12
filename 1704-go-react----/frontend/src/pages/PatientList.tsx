import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { patientApi, planApi } from '../api';
import type { Patient, Plan } from '../types';

export default function PatientList() {
  const [patients, setPatients] = useState<Patient[]>([]);
  const [plans, setPlans] = useState<Record<number, Plan[]>>({});
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadPatients();
  }, []);

  const loadPatients = async () => {
    try {
      const res = await patientApi.list();
      setPatients(res.data);
      
      const planMap: Record<number, Plan[]> = {};
      for (const patient of res.data) {
        const planRes = await planApi.list(patient.id);
        planMap[patient.id] = planRes.data;
      }
      setPlans(planMap);
    } catch (error) {
      console.error('加载患者列表失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const getPlanProgress = (plan: Plan) => {
    if (!plan.exercises || plan.exercises.length === 0) return 0;
    const totalTasks = plan.exercises.reduce((sum, ex) => {
      if (ex.frequencyType === 'daily') {
        return sum + ex.frequencyCount * plan.durationWeeks * 7;
      }
      return sum + ex.frequencyCount * plan.durationWeeks;
    }, 0);
    
    const completedTasks = totalTasks * 0.5;
    return Math.round((completedTasks / totalTasks) * 100);
  };

  const getEffectivenessText = (effectiveness: string) => {
    if (effectiveness === 'poor') return '效果不佳';
    if (effectiveness === 'good') return '效果良好';
    return '评估中';
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="text-gray-500">加载中...</div>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6">
        <h2 className="text-2xl font-bold text-gray-900">患者列表</h2>
        <p className="text-gray-500 mt-1">查看所有康复中的患者和康复计划进度</p>
      </div>

      {patients.length === 0 ? (
        <div className="bg-white rounded-lg shadow p-8 text-center">
          <p className="text-gray-500">暂无患者数据</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {patients.map((patient) => {
            const patientPlans = plans[patient.id] || [];
            const activePlan = patientPlans.find(p => p.status === 'active');

            return (
              <div key={patient.id} className="bg-white rounded-lg shadow overflow-hidden">
                <div className="p-6">
                  <div className="flex items-center justify-between mb-4">
                    <h3 className="text-lg font-semibold text-gray-900">{patient.name}</h3>
                    <span className="px-2 py-1 bg-green-100 text-green-800 text-xs rounded">
                      康复中
                    </span>
                  </div>

                  <div className="space-y-2 text-sm text-gray-600">
                    <p>性别: {patient.gender || '-'}</p>
                    <p>电话: {patient.phone || '-'}</p>
                  </div>

                  {activePlan && (
                    <div className="mt-4 p-4 bg-blue-50 rounded-lg">
                      <h4 className="font-medium text-blue-900 mb-2">{activePlan.name}</h4>
                      <div className="flex items-center justify-between text-sm">
                        <span className="text-blue-700">
                          周期: {activePlan.durationWeeks}周
                        </span>
                        <span className="text-blue-700">
                          效果: {getEffectivenessText(activePlan.effectiveness)}
                        </span>
                      </div>
                      <div className="mt-3">
                        <div className="flex items-center justify-between text-sm mb-1">
                          <span className="text-gray-600">进度</span>
                          <span className="text-gray-900 font-medium">
                            {getPlanProgress(activePlan)}%
                          </span>
                        </div>
                        <div className="w-full bg-blue-200 rounded-full h-2">
                          <div
                            className="bg-blue-600 h-2 rounded-full transition-all"
                            style={{ width: `${getPlanProgress(activePlan)}%` }}
                          />
                        </div>
                      </div>
                    </div>
                  )}

                  <div className="mt-4 flex space-x-2">
                    <Link
                      to={`/patients/${patient.id}/training`}
                      className="flex-1 text-center px-3 py-2 bg-blue-600 text-white text-sm rounded hover:bg-blue-700 transition"
                    >
                      训练日历
                    </Link>
                    <Link
                      to={`/patients/${patient.id}/assessments`}
                      className="flex-1 text-center px-3 py-2 bg-purple-600 text-white text-sm rounded hover:bg-purple-700 transition"
                    >
                      评估记录
                    </Link>
                    <Link
                      to={`/patients/${patient.id}/summary`}
                      className="flex-1 text-center px-3 py-2 bg-green-600 text-white text-sm rounded hover:bg-green-700 transition"
                    >
                      汇总面板
                    </Link>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
