import { useEffect, useState } from 'react';
import { statsApi, teacherApi } from '../api';
import type { StatsResponse, Teacher, Title } from '../types';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  PieChart,
  Pie,
  Cell,
  ResponsiveContainer,
} from 'recharts';

const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444'];

export default function Dashboard() {
  const [stats, setStats] = useState<StatsResponse | null>(null);
  const [teachers, setTeachers] = useState<Teacher[]>([]);
  const [message, setMessage] = useState('');

  const [newTeacher, setNewTeacher] = useState({
    name: '',
    current_title: '讲师' as Title,
  });

  useEffect(() => {
    loadData();
  }, []);

  async function loadData() {
    try {
      const [sRes, tRes] = await Promise.all([
        statsApi.getAll(),
        teacherApi.getAll(),
      ]);
      setStats(sRes.data);
      setTeachers(tRes.data);
    } catch (e) {
      console.error(e);
    }
  }

  async function handleCreateTeacher() {
    if (!newTeacher.name) {
      setMessage('请输入教师姓名');
      return;
    }
    try {
      await teacherApi.create(newTeacher);
      setMessage('教师创建成功');
      setNewTeacher({ name: '', current_title: '讲师' });
      loadData();
    } catch (e: any) {
      setMessage(e.response?.data?.error || '创建失败');
    }
  }

  if (!stats) {
    return <div>加载中...</div>;
  }

  const hoursData = Object.entries(stats.training_hours).map(([title, data]) => ({
    title,
    已完成: data.completed,
    要求: data.required,
  }));

  const evalData = [
    { name: '合格', value: stats.evaluations.qualified },
    { name: '不合格', value: stats.evaluations.total - stats.evaluations.qualified },
  ];

  const reviewData = [
    { name: '通过', value: stats.review_rate.passed },
    { name: '未通过', value: stats.review_rate.total - stats.review_rate.passed },
  ];

  return (
    <div className="page">
      <h2>统计看板</h2>
      {message && <div className="message">{message}</div>}

      <div className="card">
        <h3>快速初始化：创建教师</h3>
        <div className="form-grid">
          <input placeholder="教师姓名" value={newTeacher.name} onChange={(e) => setNewTeacher({ ...newTeacher, name: e.target.value })} />
          <select value={newTeacher.current_title} onChange={(e) => setNewTeacher({ ...newTeacher, current_title: e.target.value as Title })}>
            <option>助教</option>
            <option>讲师</option>
            <option>副教授</option>
            <option>教授</option>
          </select>
        </div>
        <button onClick={handleCreateTeacher}>创建教师</button>
      </div>

      <div className="card">
        <h3>年度培训学时分布</h3>
        <p className="hint">要求：初级60学时，中级40学时，副高及以上20学时</p>
        <ResponsiveContainer width="100%" height={300}>
          <BarChart data={hoursData}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="title" />
            <YAxis />
            <Tooltip />
            <Legend />
            <Bar dataKey="已完成" fill="#3b82f6" />
            <Bar dataKey="要求" fill="#e5e7eb" />
          </BarChart>
        </ResponsiveContainer>
      </div>

      <div className="charts-row">
        <div className="card">
          <h3>考核分布</h3>
          <p>总计：{stats.evaluations.total} 人次</p>
          <p>合格：{stats.evaluations.qualified} | 不合格：{stats.evaluations.total - stats.evaluations.qualified}</p>
          <p>平均分：{stats.evaluations.avg_score.toFixed(2)}</p>
          <ResponsiveContainer width="100%" height={250}>
            <PieChart>
              <Pie
                data={evalData}
                cx="50%"
                cy="50%"
                outerRadius={80}
                dataKey="value"
                label={({ name, percent }) => `${name} ${((percent || 0) * 100).toFixed(0)}%`}
              >
                {evalData.map((_, index) => (
                  <Cell key={`cell-${index}`} fill={COLORS[index % 2]} />
                ))}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
        </div>

        <div className="card">
          <h3>职称评审通过率</h3>
          <p>总计：{stats.review_rate.total} 人次</p>
          <p>通过：{stats.review_rate.passed}</p>
          <p>通过率：{(stats.review_rate.rate * 100).toFixed(1)}%</p>
          <ResponsiveContainer width="100%" height={250}>
            <PieChart>
              <Pie
                data={reviewData}
                cx="50%"
                cy="50%"
                outerRadius={80}
                dataKey="value"
                label={({ name, percent }) => `${name} ${((percent || 0) * 100).toFixed(0)}%`}
              >
                {reviewData.map((_, index) => (
                  <Cell key={`cell-${index}`} fill={COLORS[index % 2]} />
                ))}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
        </div>
      </div>

      <div className="card">
        <h3>教师列表</h3>
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>姓名</th>
              <th>当前职称</th>
            </tr>
          </thead>
          <tbody>
            {teachers.map((t) => (
              <tr key={t.id}>
                <td>{t.id}</td>
                <td>{t.name}</td>
                <td>{t.current_title}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
