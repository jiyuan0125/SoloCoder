import { useState, useEffect, useCallback } from 'react';
import { callService, vehicleService, dispatchService } from '../services/api';

function Dashboard() {
  const [pendingCalls, setPendingCalls] = useState([]);
  const [vehicles, setVehicles] = useState([]);
  const [dispatchRecords, setDispatchRecords] = useState([]);
  const [selectedCall, setSelectedCall] = useState(null);
  const [recommendedVehicles, setRecommendedVehicles] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const fetchData = useCallback(async () => {
    try {
      const [callsRes, vehiclesRes, recordsRes] = await Promise.all([
        callService.getPendingCalls(),
        vehicleService.listVehicles(),
        dispatchService.listRecords()
      ]);
      setPendingCalls(callsRes.data);
      setVehicles(vehiclesRes.data);
      setDispatchRecords(recordsRes.data);
    } catch (err) {
      setError('加载数据失败');
      console.error(err);
    }
  }, []);

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 5000);
    return () => clearInterval(interval);
  }, [fetchData]);

  const getSeverityClass = (severity) => {
    switch (severity) {
      case '一级濒危': return 'severity-level1';
      case '二级危重': return 'severity-level2';
      case '三级急症': return 'severity-level3';
      case '四级非急症': return 'severity-level4';
      default: return '';
    }
  };

  const isHighPriority = (severity) => {
    return severity === '一级濒危' || severity === '二级危重';
  };

  const getStatusClass = (status) => {
    switch (status) {
      case '空闲': return 'status-idle';
      case '出车中': return 'status-intransit';
      case '返回途中': return 'status-returning';
      case '维护中': return 'status-maintenance';
      default: return '';
    }
  };

  const handleDispatch = async (call) => {
    setSelectedCall(call);
    setLoading(true);
    try {
      const res = await dispatchService.recommendVehicle(call.id);
      setRecommendedVehicles(res.data);
    } catch (err) {
      setError('获取推荐车辆失败');
      console.error(err);
    }
    setLoading(false);
  };

  const confirmDispatch = async (vehicleId) => {
    setLoading(true);
    try {
      await dispatchService.dispatchVehicle({
        call_id: selectedCall.id,
        vehicle_id: vehicleId
      });
      setSelectedCall(null);
      setRecommendedVehicles([]);
      await fetchData();
    } catch (err) {
      setError('派车失败');
      console.error(err);
    }
    setLoading(false);
  };

  const formatTime = (dateStr) => {
    if (!dateStr) return '-';
    const date = new Date(dateStr);
    return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  };

  const closeModal = () => {
    setSelectedCall(null);
    setRecommendedVehicles([]);
  };

  return (
    <div className="page-content">
      {error && (
        <div style={{ color: '#dc3545', marginBottom: '15px', padding: '10px', backgroundColor: '#f8d7da', borderRadius: '4px' }}>
          {error}
          <button onClick={() => setError(null)} style={{ marginLeft: '15px', background: 'none', border: 'none', cursor: 'pointer' }}>×</button>
        </div>
      )}

      <div className="dashboard">
        <div className="panel">
          <div className="panel-header">待派车求救 ({pendingCalls.length})</div>
          <div className="panel-body">
            {pendingCalls.length === 0 ? (
              <div className="empty-state">暂无待派车求救</div>
            ) : (
              pendingCalls.map(call => (
                <div 
                  key={call.id} 
                  className={`call-card ${isHighPriority(call.severity_level) ? 'high-priority' : ''}`}
                >
                  <div className="call-card-header">
                    <span className="call-id">ID: {call.id.substring(0, 8)}...</span>
                    <span className={`severity-badge ${getSeverityClass(call.severity_level)}`}>
                      {call.severity_level}
                    </span>
                  </div>
                  <div className="call-card-info">
                    <div><strong>求救人:</strong> {call.caller_name} ({call.caller_phone})</div>
                    <div><strong>地点:</strong> {call.location}</div>
                    <div><strong>患者:</strong> {call.patient_count}人 ({call.patient_age_group} {call.patient_gender})</div>
                    <div><strong>症状:</strong> {call.chief_complaint}</div>
                    <div><strong>求救时间:</strong> {formatTime(call.call_time)}</div>
                  </div>
                  <div className="call-card-actions">
                    <button 
                      className="btn btn-primary"
                      onClick={() => handleDispatch(call)}
                      disabled={loading}
                    >
                      派车
                    </button>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>

        <div className="panel">
          <div className="panel-header">车辆状态 - 地图区域</div>
          <div className="panel-body" style={{ padding: 0 }}>
            <table className="vehicle-table">
              <thead>
                <tr>
                  <th>车牌号</th>
                  <th>类型</th>
                  <th>状态</th>
                  <th>当前位置</th>
                  <th>医护人员</th>
                  <th>患者</th>
                </tr>
              </thead>
              <tbody>
                {vehicles.map(vehicle => (
                  <tr key={vehicle.id}>
                    <td style={{ fontWeight: 600 }}>{vehicle.vehicle_number}</td>
                    <td>{vehicle.vehicle_type}</td>
                    <td>
                      <span className={`status-badge ${getStatusClass(vehicle.current_status)}`}>
                        {vehicle.current_status}
                      </span>
                    </td>
                    <td>{vehicle.current_location}</td>
                    <td>{vehicle.doctor_count}医 {vehicle.nurse_count}护</td>
                    <td>{vehicle.patient_count}/3</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        <div className="panel">
          <div className="panel-header">出车记录 ({dispatchRecords.length})</div>
          <div className="panel-body">
            {dispatchRecords.length === 0 ? (
              <div className="empty-state">暂无出车记录</div>
            ) : (
              [...dispatchRecords].reverse().map(record => (
                <div key={record.id} className="dispatch-record">
                  <div className="dispatch-record-header">
                    <span>派车ID: {record.id.substring(0, 8)}...</span>
                    <span style={{ fontSize: 12, color: '#6c757d' }}>
                      {formatTime(record.dispatch_time)}
                    </span>
                  </div>
                  <div className="dispatch-record-details">
                    <div>目标地点: {record.target_location}</div>
                    <div>预计距离: {record.estimated_distance_km}公里</div>
                    <div>预计到达: {record.estimated_arrival_time_minutes}分钟</div>
                    {record.arrival_time && (
                      <div>实际到达: {formatTime(record.arrival_time)}</div>
                    )}
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      </div>

      {selectedCall && (
        <div className="dispatch-modal" onClick={closeModal}>
          <div className="dispatch-modal-content" onClick={e => e.stopPropagation()}>
            <div className="dispatch-modal-header">
              <div className="dispatch-modal-title">选择救护车派车</div>
              <button className="dispatch-modal-close" onClick={closeModal}>×</button>
            </div>
            
            <div style={{ marginBottom: 20, padding: 15, backgroundColor: '#f8f9fa', borderRadius: 6 }}>
              <div><strong>求救人:</strong> {selectedCall.caller_name}</div>
              <div><strong>地点:</strong> {selectedCall.location}</div>
              <div><strong>症状:</strong> {selectedCall.chief_complaint}</div>
              <div><strong>严重程度:</strong> {selectedCall.severity_level}</div>
            </div>

            {loading ? (
              <div className="empty-state">加载中...</div>
            ) : recommendedVehicles.length === 0 ? (
              <div className="empty-state">无可用车辆推荐</div>
            ) : (
              recommendedVehicles.map((item, index) => (
                <div 
                  key={item.vehicle.id}
                  className="vehicle-option"
                  onClick={() => confirmDispatch(item.vehicle.id)}
                >
                  <div className="vehicle-option-header">
                    <span style={{ fontWeight: 600 }}>
                      {index === 0 ? '⭐ ' : ''}{item.vehicle.vehicle_number}
                    </span>
                    <span className={`status-badge ${getStatusClass(item.vehicle.current_status)}`}>
                      {item.vehicle.current_status}
                    </span>
                  </div>
                  <div className="vehicle-option-info">
                    <div>类型: {item.vehicle.vehicle_type}</div>
                    <div>距离: {item.distance}公里</div>
                    <div>位置: {item.vehicle.current_location}</div>
                    <div>医护: {item.vehicle.doctor_count}医 {item.vehicle.nurse_count}护</div>
                    <div>当前患者: {item.vehicle.patient_count}/3</div>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  );
}

export default Dashboard;
