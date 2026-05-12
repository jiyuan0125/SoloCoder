import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import { format } from 'date-fns';
import { summaryApi, patientApi } from '../api';
import type { PatientSummary, Patient } from '../types';

export default function SummaryPanel() {
  const { id } = useParams<{ id: string }>();
  const patientId = parseInt(id || '0');

  const [patient, setPatient] = useState<Patient | null>(null);
  const [summary, setSummary] = useState<PatientSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [dateRange, setDateRange] = useState({
    startDate: '',
    endDate: ''
  });

  useEffect(() => {
    loadData();
  }, [patientId]);

  const loadData = async () => {
    try {
      const [patientRes, summaryRes] = await Promise.all([
        patientApi.get(patientId),
        summaryApi.getPatientSummary(patientId, dateRange.startDate || undefined, dateRange.endDate || undefined)
      ]);
      setPatient(patientRes.data);
      setSummary(summaryRes.data);
    } catch (error) {
      console.error('加载数据失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleApplyDateFilter = () => {
    loadData();
  };

  const getEffectivenessText = (effectiveness: string) => {
    if (effectiveness === 'poor') return '效果不佳';
    if (effectiveness === 'good') return '效果良好';
    return '评估中';
  };

  const getDeviationClass = (deviation: number) => {
    return deviation >= 0 ? 'text-green-600' : 'text-red-600';
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="text-gray-500">加载中...</div>
      </div>
    );
  }

  if (!summary) {
    return (
      <div className="bg-white rounded-lg shadow p-8 text-center">
        <p className="text-gray-500">暂无汇总数据</p>
      </div>
    );
  }

  const assessmentChartData = summary.assessmentTrends.map((trend) => ({
    date: format(new Date(trend.assessmentDate), 'MM-dd'),
    score: trend.score,
    improvement: trend.improvementPercent
  }));

  return (
    <div>
      <div className="mb-6">
        <div className="flex items-center space-x-4">
          <Link to="/" className="text-blue-600 hover:text-blue-800">
            &larr; 返回列表
          </Link>
        </div>
        <h2 className="text-2xl font-bold text-gray-900 mt-2">
          {patient?.name} - 汇总面板
        </h2>
        <p className="text-gray-500 mt-1">
          查看康复周期内的趋势数据
        </p>
      </div>

      <div className="bg-white rounded-lg shadow p-4 mb-6">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div className="text-center">
            <div className="text-3xl font-bold text-blue-600">
              {summary.planProgress.toFixed(0)}%
            </div>
            <div className="text-sm text-gray-500">计划进度</div>
          </div>
          <div className="text-center">
            <div className="text-3xl font-bold text-green-600">
              {getEffectivenessText(summary.effectiveness)}
            </div>
            <div className="text-sm text-gray-500">康复效果</div>
          </div>
          <div className="text-center">
            <div className="text-3xl font-bold text-purple-600 text-sm">
              {summary.planName}
            </div>
            <div className="text-sm text-gray-500">康复计划</div>
          </div>
          <div className="text-center">
            <div className="text-3xl font-bold text-orange-600">
              {summary.exerciseSummaries.length}
            </div>
            <div className="text-sm text-gray-500">训练项目数</div>
          </div>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow p-4 mb-6">
        <h3 className="text-lg font-semibold mb-4">日期筛选</h3>
        <div className="flex items-end space-x-4">
          <div>
            <label className="block text-sm text-gray-600">开始日期</label>
            <input
              type="date"
              value={dateRange.startDate}
              onChange={(e) => setDateRange({ ...dateRange, startDate: e.target.value })}
              className="border rounded px-3 py-2"
            />
          </div>
          <div>
            <label className="block text-sm text-gray-600">结束日期</label>
            <input
              type="date"
              value={dateRange.endDate}
              onChange={(e) => setDateRange({ ...dateRange, endDate: e.target.value })}
              className="border rounded px-3 py-2"
            />
          </div>
          <button
            onClick={handleApplyDateFilter}
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
          >
            应用筛选
          </button>
        </div>
      </div>

      {assessmentChartData.length > 0 && (
        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <h3 className="text-lg font-semibold mb-4">评估分数趋势</h3>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={assessmentChartData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="date" />
              <YAxis />
              <Tooltip />
              <Legend />
              <Line type="monotone" dataKey="score" stroke="#3b82f6" name="评分" />
              <Line type="monotone" dataKey="improvement" stroke="#10b981" name="改善%" />
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}

      {summary.exerciseSummaries.length > 0 && (
        <div>
          <h3 className="text-lg font-semibold mb-4">训练项目汇总</h3>
          <div className="space-y-4">
            {summary.exerciseSummaries.map((ex, idx) => (
              <div key={idx} className="bg-white rounded-lg shadow p-6">
                <h4 className="font-medium text-lg mb-4">{ex.exerciseName}</h4>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div>
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <div className="text-sm text-gray-500">平均完成质量</div>
                        <div className="text-xl font-bold text-green-600">
                          {ex.avgQualityScore.toFixed(1)} 分
                        </div>
                      </div>
                      <div>
                        <div className="text-sm text-gray-500">完成次数</div>
                        <div className="text-xl font-bold text-blue-600">
                          {ex.totalCompleted} 次
                        </div>
                      </div>
                      <div>
                        <div className="text-sm text-gray-500">计划时长</div>
                        <div className="text-xl font-bold text-purple-600">
                          {ex.plannedDuration} 分钟
                        </div>
                      </div>
                      <div>
                        <div className="text-sm text-gray-500">平均实际时长</div>
                        <div className="text-xl font-bold text-orange-600">
                          {ex.avgActualDuration.toFixed(1)} 分钟
                        </div>
                      </div>
                      <div>
                        <div className="text-sm text-gray-500">时长偏差</div>
                        <div className={`text-xl font-bold ${getDeviationClass(ex.durationDeviation)}`}>
                          {ex.durationDeviation > 0 ? '+' : ''}{ex.durationDeviation.toFixed(1)} 分钟
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
                {ex.difficultyChanges && ex.difficultyChanges.length > 0 && (
                  <div className="mt-4">
                    <h5 className="font-medium text-sm text-gray-600 mb-2">难度调整历史</h5>
                    <div className="space-y-2">
                      {ex.difficultyChanges.map((change) => (
                        <div key={change.id} className="flex items-center space-x-3 p-2 bg-gray-50 rounded">
                          <span className="text-sm text-gray-700">
                            {format(new Date(change.adjustedAt), 'yyyy-MM-dd')}
                          </span>
                          <span className="text-sm">
                            {change.oldDifficulty}级 → {change.newDifficulty}级
                          </span>
                          <span className="text-xs text-gray-500">
                            {change.adjustReason}
                          </span>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
