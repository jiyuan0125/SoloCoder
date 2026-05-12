import { useEffect, useState } from 'react';
import { trainingApi, registrationApi, teacherApi } from '../api';
import type { Training, Registration, Teacher, AttendanceStatus, AttendanceItem } from '../types';

const ATTENDANCE_OPTIONS: AttendanceStatus[] = ['出勤', '迟到', '请假', '缺席'];

export default function TrainingCenter() {
  const [trainings, setTrainings] = useState<Training[]>([]);
  const [teachers, setTeachers] = useState<Teacher[]>([]);
  const [registrations, setRegistrations] = useState<Registration[]>([]);
  const [selectedTeacher, setSelectedTeacher] = useState('');
  const [message, setMessage] = useState('');

  const [createForm, setCreateForm] = useState({
    name: '',
    type: '教学能力' as Training['type'],
    form: '讲座' as Training['form'],
    date: '',
    hours: 0,
    lecturer: '',
    capacity: 50,
  });

  const [selectedTraining, setSelectedTraining] = useState<Training | null>(null);
  const [editReg, setEditReg] = useState<{ reg: Registration; attendance: AttendanceItem[]; score: string; studyReport: boolean } | null>(null);

  useEffect(() => {
    loadData();
  }, []);

  async function loadData() {
    const [tRes, teRes] = await Promise.all([
      trainingApi.getAll(),
      teacherApi.getAll(),
    ]);
    setTrainings(tRes.data);
    setTeachers(teRes.data);
    if (teRes.data.length > 0 && !selectedTeacher) {
      setSelectedTeacher(teRes.data[0].id);
    }
  }

  useEffect(() => {
    if (selectedTeacher) {
      registrationApi.getAll(selectedTeacher).then((res) => {
        setRegistrations(res.data);
      });
    }
  }, [selectedTeacher]);

  async function handleCreateTraining() {
    try {
      await trainingApi.create(createForm);
      setMessage('培训创建成功');
      loadData();
    } catch (e: any) {
      setMessage(e.response?.data?.error || '创建失败');
    }
  }

  async function handleRegister(trainingId: string) {
    if (!selectedTeacher) {
      setMessage('请先选择教师');
      return;
    }
    try {
      await trainingApi.register(trainingId, selectedTeacher);
      setMessage('报名成功');
      if (selectedTeacher) {
        registrationApi.getAll(selectedTeacher).then((res) => {
          setRegistrations(res.data);
        });
      }
    } catch (e: any) {
      setMessage(e.response?.data?.error || '报名失败');
    }
  }

  async function handleOpenEdit(reg: Registration) {
    const t = trainings.find((tr) => tr.id === reg.training_id);
    const attendance = reg.attendance || Array(t?.hours || 0).fill(null).map(() => ({ status: '出勤' as AttendanceStatus }));
    setEditReg({
      reg,
      attendance,
      score: reg.score?.toString() || '',
      studyReport: reg.study_report || false,
    });
  }

  async function handleSaveEdit() {
    if (!editReg) return;
    try {
      await registrationApi.update(editReg.reg.id, {
        attendance: editReg.attendance,
        score: editReg.score ? parseFloat(editReg.score) : undefined,
        study_report: editReg.studyReport,
      });
      setMessage('保存成功');
      setEditReg(null);
      if (selectedTeacher) {
        registrationApi.getAll(selectedTeacher).then((res) => {
          setRegistrations(res.data);
        });
      }
    } catch (e: any) {
      setMessage(e.response?.data?.error || '保存失败');
    }
  }

  return (
    <div className="page">
      <h2>培训中心</h2>
      {message && <div className="message">{message}</div>}

      <div className="card">
        <h3>选择教师</h3>
        <select value={selectedTeacher} onChange={(e) => setSelectedTeacher(e.target.value)}>
          <option value="">请选择</option>
          {teachers.map((t) => (
            <option key={t.id} value={t.id}>{t.name} - {t.current_title}</option>
          ))}
        </select>
      </div>

      <div className="card">
        <h3>创建培训（管理员）</h3>
        <div className="form-grid">
          <input placeholder="培训名称" value={createForm.name} onChange={(e) => setCreateForm({ ...createForm, name: e.target.value })} />
          <select value={createForm.type} onChange={(e) => setCreateForm({ ...createForm, type: e.target.value as Training['type'] })}>
            <option>教学能力</option>
            <option>科研能力</option>
            <option>管理能力</option>
            <option>师德师风</option>
          </select>
          <select value={createForm.form} onChange={(e) => setCreateForm({ ...createForm, form: e.target.value as Training['form'] })}>
            <option>讲座</option>
            <option>工作坊</option>
            <option>在线课程</option>
            <option>实践研修</option>
          </select>
          <input type="date" value={createForm.date} onChange={(e) => setCreateForm({ ...createForm, date: e.target.value })} />
          <input type="number" placeholder="学时" value={createForm.hours} onChange={(e) => setCreateForm({ ...createForm, hours: parseInt(e.target.value) || 0 })} />
          <input placeholder="讲师" value={createForm.lecturer} onChange={(e) => setCreateForm({ ...createForm, lecturer: e.target.value })} />
          <input type="number" placeholder="容量" value={createForm.capacity} onChange={(e) => setCreateForm({ ...createForm, capacity: parseInt(e.target.value) || 0 })} />
        </div>
        <button onClick={handleCreateTraining}>创建培训</button>
      </div>

      <div className="card">
        <h3>培训列表</h3>
        <table>
          <thead>
            <tr>
              <th>名称</th>
              <th>类型</th>
              <th>形式</th>
              <th>日期</th>
              <th>学时</th>
              <th>讲师</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {trainings.map((t) => (
              <tr key={t.id}>
                <td>{t.name}</td>
                <td>{t.type}</td>
                <td>{t.form}</td>
                <td>{t.date}</td>
                <td>{t.hours}</td>
                <td>{t.lecturer}</td>
                <td>
                  <button onClick={() => handleRegister(t.id)}>报名</button>
                  <button onClick={() => setSelectedTraining(t)}>详情</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {selectedTraining && (
        <div className="card">
          <h3>培训详情: {selectedTraining.name}</h3>
          <p>类型: {selectedTraining.type} | 形式: {selectedTraining.form} | 日期: {selectedTraining.date} | 学时: {selectedTraining.hours} | 讲师: {selectedTraining.lecturer} | 容量: {selectedTraining.capacity}</p>
          <button onClick={() => setSelectedTraining(null)}>关闭</button>
        </div>
      )}

      <div className="card">
        <h3>参训记录</h3>
        <table>
          <thead>
            <tr>
              <th>培训ID</th>
              <th>出勤</th>
              <th>分数</th>
              <th>学习心得</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {registrations.map((r) => (
              <tr key={r.id}>
                <td>{r.training_id}</td>
                <td>{r.attendance ? `${r.attendance.length}次记录` : '未录入'}</td>
                <td>{r.score ?? '-'}</td>
                <td>{r.study_report ? '已提交' : '未提交'}</td>
                <td><button onClick={() => handleOpenEdit(r)}>录入出勤/成绩</button></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {editReg && (
        <div className="card">
          <h3>编辑参训记录</h3>
          <h4>出勤录入（每学时一次）</h4>
          <div className="form-grid">
            {editReg.attendance.map((item, i) => (
              <div key={i}>
                <label>学时{i + 1}</label>
                <select
                  value={item.status}
                  onChange={(e) => {
                    const newAtt = [...editReg.attendance];
                    newAtt[i] = { status: e.target.value as AttendanceStatus };
                    setEditReg({ ...editReg, attendance: newAtt });
                  }}
                >
                  {ATTENDANCE_OPTIONS.map((o) => (
                    <option key={o} value={o}>{o}</option>
                  ))}
                </select>
              </div>
            ))}
          </div>
          <div className="form-grid">
            <div>
              <label>考核分数（百分制）</label>
              <input type="number" value={editReg.score} onChange={(e) => setEditReg({ ...editReg, score: e.target.value })} />
            </div>
            <div>
              <label>补交学习心得</label>
              <input type="checkbox" checked={editReg.studyReport} onChange={(e) => setEditReg({ ...editReg, studyReport: e.target.checked })} />
            </div>
          </div>
          <button onClick={handleSaveEdit}>保存</button>
          <button onClick={() => setEditReg(null)}>取消</button>
        </div>
      )}
    </div>
  );
}
