import { useState, useEffect, useCallback } from 'react';
import { vehicleService } from '../services/api';

function VehicleManagement() {
  const [vehicles, setVehicles] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const fetchVehicles = useCallback(async () => {
    setLoading(true);
    try {
      const res = await vehicleService.listVehicles();
      setVehicles(res.data);
    } catch (err) {
      setError('加载车辆数据失败');
      console.error(err);
    }
    setLoading(false);
  }, []);

  useEffect(() => {
    fetchVehicles();
  }, [fetchVehicles]);

  const getStatusClass = (status) => {
    switch (status) {
      case '空闲': return 'status-idle';
      case '出车中': return 'status-intransit';
      case '返回途中': return 'status-returning';
      case '维护中': return 'status-maintenance';
      default: return '';
    }
  };

  const getNextStatus = (currentStatus) => {
    switch (currentStatus) {
      case '空闲': return '出车中';
      case '出车中': return '返回途中';
      case '返回途中': return '空闲';
      default: return null;
    }
  };

  const handleStatusUpdate = async (vehicleId, newStatus) => {
    setLoading(true);
    try {
      await vehicleService.updateStatus(vehicleId, newStatus);
      await fetchVehicles();
    } catch (err) {
      setError('更新状态失败');
      console.error(err);
    }
    setLoading(false);
  };

  const handleSetMaintenance = async (vehicleId) => {
    setLoading(true);
    try {
      await vehicleService.setMaintenance(vehicleId);
      await fetchVehicles();
    } catch (err) {
      setError('设置维护状态失败');
      console.error(err);
    }
    setLoading(false);
  };

  return (
    <div className="page-content">
      {error && (
        <div style={{ color: '#dc3545', marginBottom: '15px', padding: '10px', backgroundColor: '#f8d7da', borderRadius: '4px' }}>
          {error}
          <button onClick={() => setError(null)} style={{ marginLeft: '15px', background: 'none', border: 'none', cursor: 'pointer' }}>×</button>
        </div>
      )}

      <div className="vehicles-management">
        <div className="vehicles-management-header">
          <h2 style={{ fontSize: 20, fontWeight: 600 }}>救护车管理</h2>
          <button 
            className="refresh-btn" 
            onClick={fetchVehicles}
            disabled={loading}
          >
            {loading ? '刷新中...' : '刷新'}
          </button>
        </div>

        <div className="vehicles-management-body">
          <table className="vehicle-table">
            <thead>
              <tr>
                <th>车牌号</th>
                <th>车辆类型</th>
                <th>当前状态</th>
                <th>当前位置</th>
                <th>医生</th>
                <th>护士</th>
                <th>患者</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {vehicles.map(vehicle => {
                const nextStatus = getNextStatus(vehicle.current_status);
                return (
                  <tr key={vehicle.id}>
                    <td style={{ fontWeight: 600 }}>{vehicle.vehicle_number}</td>
                    <td>{vehicle.vehicle_type}</td>
                    <td>
                      <span className={`status-badge ${getStatusClass(vehicle.current_status)}`}>
                        {vehicle.current_status}
                      </span>
                    </td>
                    <td>{vehicle.current_location}</td>
                    <td>{vehicle.doctor_count}</td>
                    <td>{vehicle.nurse_count}</td>
                    <td>{vehicle.patient_count}/3</td>
                    <td>
                      {nextStatus && (
                        <button 
                          className="btn btn-secondary"
                          style={{ marginRight: 8 }}
                          onClick={() => handleStatusUpdate(vehicle.id, nextStatus)}
                          disabled={loading}
                        >
                          转为{nextStatus}
                        </button>
                      )}
                      {vehicle.current_status === '空闲' && (
                        <button 
                          className="btn btn-danger"
                          onClick={() => handleSetMaintenance(vehicle.id)}
                          disabled={loading}
                        >
                          设为维护
                        </button>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>

      <div style={{ marginTop: 30, backgroundColor: 'white', padding: 20, borderRadius: 8, boxShadow: '0 2px 8px rgba(0,0,0,0.08)' }}>
        <h3 style={{ marginBottom: 15, fontSize: 16 }}>状态流转规则</h3>
        <div style={{ display: 'flex', gap: 20, fontSize: 14, color: '#6c757d' }}>
          <div>
            <span className="status-badge status-idle">空闲</span> → 
            <span className="status-badge status-intransit" style={{ marginLeft: 8 }}>出车中</span> → 
            <span className="status-badge status-returning" style={{ marginLeft: 8 }}>返回途中</span> → 
            <span className="status-badge status-idle" style={{ marginLeft: 8 }}>空闲</span>
          </div>
        </div>
        <p style={{ marginTop: 10, fontSize: 13, color: '#6c757d' }}>
          只有空闲状态的车辆可以设置为维护中。维护中的车辆不能派车（除非是一级或二级紧急求救）。
        </p>
      </div>
    </div>
  );
}

export default VehicleManagement;
