import { useEffect, useState } from 'react';
import { applicationApi, teacherApi } from '../api';
import type { TitleApplication, Teacher, Title, ReviewStage, ReviewStatus } from '../types';

const TITLES: Title[] = ['助教', '讲师', '副教授', '教授'];
const STATUSES: ReviewStatus[] = ['通过', '不通过'];

export default function TitleReview() {
  const [teachers, setTeachers] = useState<Teacher[]>([]);
  const [applications, setApplications] = useState<TitleApplication[]>([]);
  const [message, setMessage] = useState('');
  const [selectedTeacher, setSelectedTeacher] = useState('');

  const [applyForm, setApplyForm] = useState({
    teacher_id: '',
    apply_title: '讲师' as Title,
    materials: '',
    achievement_summary: '',
  });

  const [reviewForm, setReviewForm] = useState<{
    appId: string;
    stage: ReviewStage;
    status: ReviewStatus;
    comment: string;
  } | null>(null);

  useEffect(() => {
    loadData();
  }, []);

  async function loadData() {
    const [tRes, aRes] = await Promise.all([
      teacherApi.getAll(),
      applicationApi.getAll(),
    ]);
    setTeachers(tRes.data);
    setApplications(aRes.data);
    if (tRes.data.length > 0 && !selectedTeacher) {
      setSelectedTeacher(tRes.data[0].id);
    }
  }

  async function handleApply() {
    if (!applyForm.teacher_id) {
      setMessage('请选择教师');
      return;
    }
    try {
      await applicationApi.create(applyForm);
      setMessage('申请成功');
      loadData();
    } catch (e: any) {
      setMessage(e.response?.data?.error || '申请失败');
    }
  }

  async function handleReviewSubmit() {
    if (!reviewForm) return;
    try {
      await applicationApi.submitReview(reviewForm.appId, {
        stage: reviewForm.stage,
        status: reviewForm.status,
        comment: reviewForm.comment,
      });
      setMessage('评审提交成功');
      setReviewForm(null);
      loadData();
    } catch (e: any) {
      setMessage(e.response?.data?.error || '评审失败');
    }
  }

  const filteredApps = selectedTeacher
    ? applications.filter((a) => a.teacher_id === selectedTeacher)
    : applications;

  return (
    <div className="page">
      <h2>职称评审</h2>
      {message && <div className="message">{message}</div>}

      <div className="card">
        <h3>职称申报</h3>
        <div className="form-grid">
          <div>
            <label>教师</label>
            <select value={applyForm.teacher_id} onChange={(e) => setApplyForm({ ...applyForm, teacher_id: e.target.value })}>
              <option value="">请选择</option>
              {teachers.map((t) => (
                <option key={t.id} value={t.id}>{t.name} - {t.current_title}</option>
              ))}
            </select>
          </div>
          <div>
            <label>申报职称</label>
            <select value={applyForm.apply_title} onChange={(e) => setApplyForm({ ...applyForm, apply_title: e.target.value as Title })}>
              {TITLES.map((t) => (
                <option key={t} value={t}>{t}</option>
              ))}
            </select>
          </div>
        </div>
        <div>
          <label>材料说明</label>
          <textarea value={applyForm.materials} onChange={(e) => setApplyForm({ ...applyForm, materials: e.target.value })} rows={3} />
        </div>
        <div>
          <label>成果摘要</label>
          <textarea value={applyForm.achievement_summary} onChange={(e) => setApplyForm({ ...applyForm, achievement_summary: e.target.value })} rows={3} />
        </div>
        <p className="hint">提示：不能跨级申报（助教→讲师→副教授→教授）</p>
        <button onClick={handleApply}>提交申请</button>
      </div>

      <div className="card">
        <h3>申请列表 / 进度跟踪</h3>
        <div>
          <label>筛选教师：</label>
          <select value={selectedTeacher} onChange={(e) => setSelectedTeacher(e.target.value)}>
            <option value="">全部</option>
            {teachers.map((t) => (
              <option key={t.id} value={t.id}>{t.name}</option>
            ))}
          </select>
        </div>
        <table>
          <thead>
            <tr>
              <th>申请ID</th>
              <th>教师ID</th>
              <th>申报职称</th>
              <th>年度</th>
              <th>当前阶段</th>
              <th>最终状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {filteredApps.map((a) => (
              <tr key={a.id}>
                <td>{a.id}</td>
                <td>{a.teacher_id}</td>
                <td>{a.apply_title}</td>
                <td>{a.year}</td>
                <td>{a.current_stage}</td>
                <td className={a.final_status === true ? 'qualified' : a.final_status === false ? 'unqualified' : ''}>
                  {a.final_status === true ? '通过' : a.final_status === false ? '不通过' : '评审中'}
                </td>
                <td>
                  {a.final_status === undefined && (
                    <button onClick={() =>
                      setReviewForm({
                        appId: a.id,
                        stage: a.current_stage,
                        status: '通过',
                        comment: '',
                      })
                    }>评审</button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {reviewForm && (
        <div className="card">
          <h3>评审 - {reviewForm.stage}</h3>
          <div className="form-grid">
            <div>
              <label>结果</label>
              <select value={reviewForm.status} onChange={(e) => setReviewForm({ ...reviewForm, status: e.target.value as ReviewStatus })}>
                {STATUSES.map((s) => (
                  <option key={s} value={s}>{s}</option>
                ))}
              </select>
            </div>
          </div>
          <div>
            <label>意见</label>
            <textarea value={reviewForm.comment} onChange={(e) => setReviewForm({ ...reviewForm, comment: e.target.value })} rows={3} />
          </div>
          <p className="hint">提示：评审需按顺序推进（初审→盲审→终审），不能跳步。终审不通过需等待下一年评审窗口。</p>
          <button onClick={handleReviewSubmit}>提交评审</button>
          <button onClick={() => setReviewForm(null)}>取消</button>
        </div>
      )}
    </div>
  );
}
