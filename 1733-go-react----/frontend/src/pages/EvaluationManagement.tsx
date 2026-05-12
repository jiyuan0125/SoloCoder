import { useEffect, useState } from 'react';
import { evaluationApi, teacherApi } from '../api';
import type { TeachingEvaluation, Teacher, Score } from '../types';

function ScoreInput({ label, value, onChange }: { label: string; value?: Score; onChange: (s: Score) => void }) {
  return (
    <div className="score-input">
      <label>{label}</label>
      <div>
        <span>学生：</span>
        <input
          type="number"
          step="0.1"
          min="1"
          max="5"
          value={value?.student ?? ''}
          onChange={(e) => {
            const v = e.target.value;
            onChange({
              ...value,
              student: v ? parseFloat(v) : undefined,
            });
          }}
        />
        <span>督导：</span>
        <input
          type="number"
          step="0.1"
          min="1"
          max="5"
          value={value?.supervisor ?? ''}
          onChange={(e) => {
            const v = e.target.value;
            onChange({
              ...value,
              supervisor: v ? parseFloat(v) : undefined,
            });
          }}
        />
      </div>
    </div>
  );
}

export default function EvaluationManagement() {
  const [teachers, setTeachers] = useState<Teacher[]>([]);
  const [evaluations, setEvaluations] = useState<TeachingEvaluation[]>([]);
  const [message, setMessage] = useState('');

  const [form, setForm] = useState({
    teacher_id: '',
    semester: '2025-2026-2',
    attitude: undefined as Score | undefined,
    content: undefined as Score | undefined,
    method: undefined as Score | undefined,
    effect: undefined as Score | undefined,
  });

  useEffect(() => {
    loadData();
  }, []);

  async function loadData() {
    const [tRes, eRes] = await Promise.all([
      teacherApi.getAll(),
      evaluationApi.getAll(),
    ]);
    setTeachers(tRes.data);
    setEvaluations(eRes.data);
  }

  async function handleSubmit() {
    if (!form.teacher_id) {
      setMessage('请选择教师');
      return;
    }
    try {
      await evaluationApi.create({
        teacher_id: form.teacher_id,
        semester: form.semester,
        attitude: form.attitude,
        content: form.content,
        method: form.method,
        effect: form.effect,
      });
      setMessage('创建成功');
      loadData();
      setForm({
        teacher_id: '',
        semester: '2025-2026-2',
        attitude: undefined,
        content: undefined,
        method: undefined,
        effect: undefined,
      });
    } catch (e: any) {
      setMessage(e.response?.data?.error || '创建失败');
    }
  }

  return (
    <div className="page">
      <h2>教学考核管理</h2>
      {message && <div className="message">{message}</div>}

      <div className="card">
        <h3>新建教学考核</h3>
        <div className="form-grid">
          <div>
            <label>教师</label>
            <select value={form.teacher_id} onChange={(e) => setForm({ ...form, teacher_id: e.target.value })}>
              <option value="">请选择</option>
              {teachers.map((t) => (
                <option key={t.id} value={t.id}>{t.name}</option>
              ))}
            </select>
          </div>
          <div>
            <label>学期</label>
            <input value={form.semester} onChange={(e) => setForm({ ...form, semester: e.target.value })} />
          </div>
        </div>
        <div className="score-inputs">
          <ScoreInput
            label="教学态度"
            value={form.attitude}
            onChange={(s) => setForm({ ...form, attitude: s })}
          />
          <ScoreInput
            label="教学内容"
            value={form.content}
            onChange={(s) => setForm({ ...form, content: s })}
          />
          <ScoreInput
            label="教学方法"
            value={form.method}
            onChange={(s) => setForm({ ...form, method: s })}
          />
          <ScoreInput
            label="教学效果"
            value={form.effect}
            onChange={(s) => setForm({ ...form, effect: s })}
          />
        </div>
        <p className="hint">提示：分数范围1-5，可为整数或一位小数。学生和督导各占50%权重，若只有一方评分则按已有部分计算。</p>
        <button onClick={handleSubmit}>创建考核</button>
      </div>

      <div className="card">
        <h3>考核记录</h3>
        <table>
          <thead>
            <tr>
              <th>教师ID</th>
              <th>学期</th>
              <th>最终分数</th>
              <th>状态</th>
            </tr>
          </thead>
          <tbody>
            {evaluations.map((e) => (
              <tr key={e.id}>
                <td>{e.teacher_id}</td>
                <td>{e.semester}</td>
                <td>{e.final_score.toFixed(2)}</td>
                <td className={e.qualified ? 'qualified' : 'unqualified'}>
                  {e.qualified ? '合格' : '不合格'}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
