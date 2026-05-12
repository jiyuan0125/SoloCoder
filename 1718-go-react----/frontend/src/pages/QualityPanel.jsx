import { useState, useEffect } from 'react';
import { qualityApi } from '../api';

function QualityPanel() {
  const [metrics, setMetrics] = useState([]);
  const [selectedYear, setSelectedYear] = useState(new Date().getFullYear());
  const [selectedMonth, setSelectedMonth] = useState(new Date().getMonth() + 1);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadMetrics();
  }, [selectedYear, selectedMonth]);

  const loadMetrics = async () => {
    setLoading(true);
    try {
      const res = await qualityApi.getMetrics(selectedYear, selectedMonth);
      setMetrics(res.data.data || []);
    } catch (err) {
      console.error('加载质控指标失败', err);
    } finally {
      setLoading(false);
    }
  };

  const years = Array.from({ length: 5 }, (_, i) => new Date().getFullYear() - i);
  const months = Array.from({ length: 12 }, (_, i) => i + 1);

  const getRateBarWidth = (rate, max = 100) => {
    return Math.min(100, (rate / max) * 100);
  };

  const getRateColor = (rate, type) => {
    if (type === 'complete') {
      if (rate >= 70) return 'bg-green-500';
      if (rate >= 50) return 'bg-yellow-500';
      return 'bg-red-500';
    }
    if (type === 'variation') {
      if (rate <= 40) return 'bg-green-500';
      if (rate <= 60) return 'bg-yellow-500';
      return 'bg-red-500';
    }
    return 'bg-blue-500';
  };

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold text-gray-800">质控面板</h1>
        <div className="flex gap-2">
          <select
            value={selectedYear}
            onChange={(e) => setSelectedYear(parseInt(e.target.value))}
            className="border rounded px-3 py-2"
          >
            {years.map((y) => (
              <option key={y} value={y}>{y}年</option>
            ))}
          </select>
          <select
            value={selectedMonth}
            onChange={(e) => setSelectedMonth(parseInt(e.target.value))}
            className="border rounded px-3 py-2"
          >
            {months.map((m) => (
              <option key={m} value={m}>{m}月</option>
            ))}
          </select>
          <button
            onClick={loadMetrics}
            className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
          >
            刷新
          </button>
        </div>
      </div>

      {loading ? (
        <div className="text-center py-12 text-gray-400">加载中...</div>
      ) : (
        <div className="space-y-4">
          {metrics.filter(m => m.has_warning).length > 0 && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-4">
              <h3 className="font-semibold text-red-700 mb-2">⚠️ 质控预警</h3>
              <div className="space-y-1">
                {metrics.filter(m => m.has_warning).map((m) => (
                  <div key={m.path_id} className="text-sm text-red-600">
                    <span className="font-medium">{m.path_name}</span>: {m.warning_reason}
                  </div>
                ))}
              </div>
            </div>
          )}

          {metrics.map((m) => (
            <div key={m.path_id} className={`bg-white rounded-lg shadow ${m.has_warning ? 'border-2 border-red-300' : ''}`}>
              <div className="p-4 border-b bg-gray-50">
                <div className="flex justify-between items-center">
                  <div>
                    <h3 className="font-semibold text-lg">{m.path_name}</h3>
                    <div className="text-sm text-gray-500">
                      {m.path_code} | 标准住院: {m.standard_min_days}-{m.standard_max_days}天 | 
                      标准费用: {m.standard_min_cost.toLocaleString()}-{m.standard_max_cost.toLocaleString()}元
                    </div>
                  </div>
                  {m.has_warning && (
                    <span className="bg-red-100 text-red-700 px-3 py-1 rounded-full text-sm font-medium">
                      ⚠️ 预警
                    </span>
                  )}
                </div>
              </div>
              
              <div className="p-4">
                <div className="grid grid-cols-5 gap-6">
                  <div className="text-center">
                    <div className="text-2xl font-bold text-blue-600">{m.total_eligible}</div>
                    <div className="text-sm text-gray-500">符合条件患者</div>
                  </div>
                  <div className="text-center">
                    <div className="text-2xl font-bold text-blue-600">{m.enrolled_count}</div>
                    <div className="text-sm text-gray-500">入径患者</div>
                  </div>
                  <div className="text-center">
                    <div className="text-2xl font-bold text-green-600">{m.completed_count}</div>
                    <div className="text-sm text-gray-500">已完成</div>
                  </div>
                  <div className="text-center">
                    <div className="text-2xl font-bold text-gray-600">{m.exited_count}</div>
                    <div className="text-sm text-gray-500">已退径</div>
                  </div>
                  <div className="text-center">
                    <div className="text-2xl font-bold text-orange-600">{m.variation_count}</div>
                    <div className="text-sm text-gray-500">变异患者</div>
                  </div>
                </div>

                <div className="mt-6 space-y-4">
                  <div>
                    <div className="flex justify-between text-sm mb-1">
                      <span className="text-gray-600">入径率</span>
                      <span className="font-medium">{m.enroll_rate}%</span>
                    </div>
                    <div className="h-3 bg-gray-200 rounded-full overflow-hidden">
                      <div 
                        className="h-full bg-blue-500 transition-all" style={{ width: `${getRateBarWidth(m.enroll_rate)}%` }}></div>
                    </div>
                  </div>

                  <div>
                    <div className="flex justify-between text-sm mb-1">
                      <span className={`${m.complete_rate < 70 ? 'text-red-600 font-medium' : 'text-gray-600'}`}>
                        完成率 {m.complete_rate < 70 && <span className="text-red-600">(⚠️ 低于70%)</span>}
                      </span>
                      <span className="font-medium">{m.complete_rate}%</span>
                    </div>
                    <div className="h-3 bg-gray-200 rounded-full overflow-hidden">
                      <div 
                        className={`h-full transition-all ${getRateColor(m.complete_rate, 'complete')}`} style={{ width: `${getRateBarWidth(m.complete_rate)}%` }}></div>
                    </div>
                    <div className="text-xs text-gray-400 mt-0.5">阈值: ≥70%</div>
                  </div>

                  <div>
                    <div className="flex justify-between text-sm mb-1">
                      <span className={`${m.variation_rate > 40 ? 'text-red-600 font-medium' : 'text-gray-600'}`}>
                        变异率 {m.variation_rate > 40 && <span className="text-red-600">(⚠️ 高于40%)</span>}
                      </span>
                      <span className="font-medium">{m.variation_rate}%</span>
                    </div>
                    <div className="h-3 bg-gray-200 rounded-full overflow-hidden">
                      <div 
                        className={`h-full transition-all ${getRateColor(m.variation_rate, 'variation')}`} style={{ width: `${getRateBarWidth(m.variation_rate)}%` }}></div>
                    </div>
                    <div className="text-xs text-gray-400 mt-0.5">阈值: ≤40%</div>
                  </div>

                  <div className="grid grid-cols-2 gap-4 pt-2">
                    <div className="bg-blue-50 rounded-lg p-3">
                      <div className="text-sm text-gray-500">平均住院日</div>
                      <div className="text-xl font-bold">
                        {m.avg_stay_days} 天
                        <span className={`ml-2 text-sm font-normal ${
                          m.avg_stay_days >= m.standard_min_days && m.avg_stay_days <= m.standard_max_days
                            ? 'text-green-600' : 'text-orange-600'
                        }`}>
                          (标准: {m.standard_min_days}-{m.standard_max_days}天)
                        </span>
                      </div>
                    </div>
                    <div className="bg-green-50 rounded-lg p-3">
                      <div className="text-sm text-gray-500">平均费用</div>
                      <div className="text-xl font-bold">
                        ¥{m.avg_cost.toLocaleString()}
                        <span className={`ml-2 text-sm font-normal ${
                          m.avg_cost >= m.standard_min_cost && m.avg_cost <= m.standard_max_cost
                            ? 'text-green-600' : 'text-orange-600'
                        }`}>
                          (标准: ¥{m.standard_min_cost.toLocaleString()}-¥{m.standard_max_cost.toLocaleString()})
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          ))}

          {metrics.length === 0 && (
            <div className="bg-white rounded-lg shadow p-12 text-center text-gray-400">
              暂无质控数据
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export default QualityPanel;
