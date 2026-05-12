import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import { format } from 'date-fns';
import { assessmentApi, patientApi } from '../api';
import type { Assessment, Patient } from '../types';

export default function AssessmentRecords() {
  const { id } = useParams<{ id: string }>();
  const patientId = parseInt(id || '0');

  const [patient, setPatient] = useState<Patient | null>(null);
  const [assessments, setAssessments] = useState<Assessment[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadData();
  }, [patientId]);

  const loadData = async () => {
    try {
      const [patientRes, assessmentRes] = await Promise.all([
        patientApi.get(patientId),
        assessmentApi.list(patientId)
      ]);
      setPatient(patientRes.data);
      setAssessments(assessmentRes.data);
    } catch (error) {
      console.error('加载数据失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const chartData = assessments
    .filter(a => a.assessmentDate)
    .map(a => ({
      date: format(new Date(a.assessmentDate!), 'MM-dd'),
      score: a.currentScore,
      improvement: a.improvementPercent
    }));

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
        <div className="flex items-center space-x-4">
          <Link to="/" className="text-blue-600 hover:text-blue-800">
            &larr; 返回列表
          </Link>
        </div>
        <h2 className="text-2xl font-bold text-gray-900 mt-2">
          {patient?.name} - 评估记录
        </h2>
        <p className="text-gray-500 mt-1">
          查看评估历史和指标趋势
        </p>
      </div>

      {chartData.length > 0 && (
        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <h3 className="text-lg font-semibold mb-4">评估分数趋势</h3>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="date" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Line type="monotone" dataKey="score" stroke="#3b82f6" name="评分" />
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}

      {assessments.length === 0 ? (
        <div className="bg-white rounded-lg shadow p-8 text-center">
          <p className="text-gray-500">暂无评估记录</p>
        </div>
      ) : (
        <div className="space-y-4">
          {assessments.map((assessment) => (
            <div key={assessment.id} className="bg-white rounded-lg shadow p-6">
              <div className="flex items-start justify-between mb-4">
                <div>
                  <h3 className="text-lg font-semibold">
                    评估日期: {format(new Date(assessment.scheduledDate), 'yyyy年MM月dd日')}
                  </h3>
                  <div className="mt-2 text-sm text-gray-600">
                    <p>评估类型: {assessment.assessmentType}</p>
                    {assessment.assessmentDate && (
                      <>
                        <p>实际评估日期: {format(new Date(assessment.assessmentDate), 'yyyy年MM月dd日')}</p>
                        <p>当前评分: {assessment.currentScore} 分</p>
                        <p>改善百分比: {assessment.improvementPercent.toFixed(1)}%</p>
                      </>
                    )}
                  </div>
                </div>
              </div>

              {assessment.indicators && assessment.indicators.length > 0 && (
                <div className="mt-4">
                  <h4 className="font-medium mb-2">评估指标:</h4>
                  <div className="grid grid-cols-2 gap-2">
                    {assessment.indicators.map((ind, idx) => (
                      <div key={idx} className="flex items-center justify-between p-3 bg-gray-50 rounded">
                        <span className="text-gray-700">{ind.name}</span>
                        <span className="font-semibold">{ind.value} {ind.unit}</span>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
