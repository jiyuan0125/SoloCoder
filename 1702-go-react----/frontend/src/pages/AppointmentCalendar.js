import { useState, useEffect } from 'react';
import { api } from '../api';

const WEEKDAYS = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'];

function formatDate(date) {
  const month = date.getMonth() + 1;
  const day = date.getDate();
  return `${month}月${day}日`;
}

function formatTime(date) {
  const hours = date.getHours().toString().padStart(2, '0');
  const minutes = date.getMinutes().toString().padStart(2, '0');
  return `${hours}:${minutes}`;
}

function getWeekDates() {
  const today = new Date();
  const day = today.getDay();
  const monday = new Date(today);
  monday.setDate(today.getDate() - (day === 0 ? 6 : day - 1));
  
  const dates = [];
  for (let i = 0; i < 7; i++) {
    const d = new Date(monday);
    d.setDate(monday.getDate() + i);
    dates.push(d);
  }
  return dates;
}

function getTimeSlots() {
  const slots = [];
  for (let hour = 8; hour < 12; hour += 0.5) {
    const startTime = new Date();
    startTime.setHours(Math.floor(hour), hour % 1 === 0 ? 0 : 30, 0, 0);
    slots.push(startTime);
  }
  for (let hour = 14; hour < 17; hour += 0.5) {
    const startTime = new Date();
    startTime.setHours(Math.floor(hour), hour % 1 === 0 ? 0 : 30, 0, 0);
    slots.push(startTime);
  }
  return slots;
}

function isWorking(doctor, time) {
  const hour = time.getHours();
  const minute = time.getMinutes();
  const totalMinutes = hour * 60 + minute;
  
  for (const shift of doctor.work_shifts) {
    const start = shift.start_hour * 60 + shift.start_minute;
    const end = shift.end_hour * 60 + shift.end_minute;
    if (totalMinutes >= start && totalMinutes < end) {
      return true;
    }
  }
  return false;
}

function getAppointment(appointments, doctorId, date, time) {
  const dateStr = date.toISOString().split('T')[0];
  const timeStr = formatTime(time);
  
  return appointments.find(a => {
    if (!a.start_time) return false;
    const apptDate = a.start_time.split('T')[0];
    const apptTime = a.start_time.split('T')[1].substring(0, 5);
    return a.doctor_id === doctorId && apptDate === dateStr && apptTime === timeStr;
  });
}

export default function AppointmentCalendar() {
  const [doctors, setDoctors] = useState([]);
  const [appointments, setAppointments] = useState([]);
  const [patients, setPatients] = useState([]);
  const [selectedSlot, setSelectedSlot] = useState(null);
  const [formData, setFormData] = useState({
    patientId: '',
    treatment: ''
  });
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loading, setLoading] = useState(false);
  const [weekDates] = useState(getWeekDates());

  useEffect(() => {
    loadData();
  }, []);

  async function loadData() {
    try {
      const [doctorsData, appointmentsData, patientsData] = await Promise.all([
        api.getDoctors(),
        api.getAppointments(),
        api.getPatients(),
      ]);
      setDoctors(doctorsData);
      setAppointments(appointmentsData);
      setPatients(patientsData);
    } catch (err) {
      console.error(err);
    }
  }

  function handleSlotClick(doctor, date, time, appointment) {
    if (appointment) return;
    if (!isWorking(doctor, time)) return;
    
    setSelectedSlot({ doctor, date, time });
    setError('');
    setSuccess('');
  }

  function closeModal() {
    setSelectedSlot(null);
    setFormData({ patientId: '', treatment: '' });
    setError('');
    setSuccess('');
  }

  async function handleSubmit(e) {
    e.preventDefault();
    if (!formData.patientId || !formData.treatment) {
      setError('请填写完整信息');
      return;
    }

    setLoading(true);
    setError('');

    try {
      const dateStr = selectedSlot.date.toISOString().split('T')[0];
      const timeStr = formatTime(selectedSlot.time);
      
      await api.bookAppointment({
        patient_id: formData.patientId,
        doctor_id: selectedSlot.doctor.id,
        date: dateStr,
        start_time: timeStr,
        treatments: [formData.treatment],
      });

      setSuccess('预约成功！');
      await loadData();
      setTimeout(closeModal, 1500);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  const today = new Date();
  const timeSlots = getTimeSlots();

  return (
    <div className="container">
      <div className="page-header">
        <h1>预约日历</h1>
        <p>查看医生排班和预约情况</p>
      </div>

      <div className="card">
        <div className="card-header">
          <h2 className="card-title">本周排班</h2>
        </div>

        <div className="calendar">
          <div className="calendar-week-header">
            <div className="calendar-day-header">医生</div>
            {weekDates.map((date, idx) => (
              <div
                key={idx}
                className={`calendar-day-header ${date.toDateString() === today.toDateString() ? 'today' : ''}`}
              >
                <div>{WEEKDAYS[date.getDay()]}</div>
                <div className="text-sm text-muted">{formatDate(date)}</div>
              </div>
            ))}
          </div>

          <div className="calendar-body">
            {doctors.map(doctor => (
              <div key={doctor.id} className="calendar-doctor-row">
                <div className="calendar-doctor-cell">
                  <div>
                    <div>{doctor.name}</div>
                    <div className="dept">{doctor.department}</div>
                  </div>
                </div>
                {weekDates.map((date, dateIdx) => {
                  const workingSlots = timeSlots.filter(time => isWorking(doctor, time));
                  if (workingSlots.length === 0) {
                    return (
                      <div key={dateIdx} className="calendar-time-slot">
                        <div className="calendar-empty">休息</div>
                      </div>
                    );
                  }
                  
                  return (
                    <div key={dateIdx} style={{ display: 'flex', flexDirection: 'column', gap: '2px' }}>
                      {workingSlots.map((time, timeIdx) => {
                        const appointment = getAppointment(appointments, doctor.id, date, time);
                        const patientName = patients.find(p => p.id === appointment?.patient_id)?.name;

                        return (
                          <div
                            key={timeIdx}
                            onClick={() => handleSlotClick(doctor, date, time, appointment)}
                            className={`calendar-time-slot ${appointment ? 'occupied' : ''}`}
                          >
                            {appointment ? (
                              <div className="calendar-appointment">
                                <div>{patientName}</div>
                                <div className="time">{formatTime(time)}</div>
                              </div>
                            ) : (
                              <div className="calendar-empty">
                                可预约
                                <div className="text-sm">{formatTime(time)}</div>
                              </div>
                            )}
                          </div>
                        );
                      })}
                    </div>
                  );
                })}
              </div>
            ))}
          </div>
        </div>
      </div>

      {selectedSlot && (
        <div className="modal-overlay" onClick={closeModal}>
          <div className="modal" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <h3 className="modal-title">预约挂号</h3>
              <button className="modal-close" onClick={closeModal}>×</button>
            </div>
            <div className="modal-body">
              <div className="mb-2">
                <strong>医生：</strong>{selectedSlot.doctor.name}
              </div>
              <div className="mb-2">
                <strong>科室：</strong>{selectedSlot.doctor.department}
              </div>
              <div className="mb-2">
                <strong>日期：</strong>{formatDate(selectedSlot.date)}
              </div>
              <div className="mb-3">
                <strong>时间：</strong>{formatTime(selectedSlot.time)}
              </div>

              {error && <div className="alert alert-danger">{error}</div>}
              {success && <div className="alert alert-success">{success}</div>}

              <form onSubmit={handleSubmit}>
                <div className="form-group">
                  <label className="form-label">选择患者</label>
                  <select
                    className="form-select"
                    value={formData.patientId}
                    onChange={e => setFormData({...formData, patientId: e.target.value})}
                  >
                    <option value="">请选择患者</option>
                    {patients.map(p => (
                      <option key={p.id} value={p.id}>{p.name}</option>
                    ))}
                  </select>
                </div>
                <div className="form-group">
                  <label className="form-label">治疗项目</label>
                  <input
                    type="text"
                    className="form-input"
                    value={formData.treatment}
                    onChange={e => setFormData({...formData, treatment: e.target.value})}
                    placeholder="如：洗牙、补牙、检查"
                  />
                </div>
              </form>
            </div>
            <div className="modal-footer">
              <button className="btn btn-outline" onClick={closeModal}>取消</button>
              <button className="btn btn-primary" onClick={handleSubmit} disabled={loading}>
                {loading ? '预约中...' : '确认预约'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
