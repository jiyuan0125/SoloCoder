import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { format, addDays, subDays } from 'date-fns';
import { trainingApi, patientApi } from '../api';
import type { Task, Patient } from '../types';

export default function TrainingCalendar() {
  const { id } = useParams<{ id: string }>();
  const patientId = parseInt(id || '0');

  const [patient, setPatient] = useState<Patient | null>(null);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [currentDate, setCurrentDate] = useState(new Date());
  const [loading, setLoading] = useState(true);
  const [selectedTask, setSelectedTask] = useState<Task | null>(null);
  const [showRecordModal, setShowRecordModal] = useState(false);
  const [recordForm, setRecordForm] = useState({
    actualDuration: 15,
    qualityScore: 8,
    subjectiveFeeling: 'normal' as const,
    notes: ''
  });
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    loadData();
  }, [patientId, currentDate]);

  const loadData = async () => {
    try {
      const [patientRes, tasksRes] = await Promise.all([
        patientApi.get(patientId),
        trainingApi.getPatientTasks(patientId, format(currentDate, 'yyyy-MM-dd'))
      ]);
      setPatient(patientRes.data);
      setTasks(tasksRes.data);
    } catch (error) {
      console.error('加载数据失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleOpenRecord = (task: Task) => {
    setSelectedTask(task);
    setRecordForm({
      actualDuration: task.exercise?.durationMinutes || 15,
      qualityScore: 8,
      subjectiveFeeling: 'normal',
      notes: ''
    });
    setShowRecordModal(true);
  };

  const handleSubmitRecord = async () => {
    if (!selectedTask) return;
    setSubmitting(true);
    try {
      await trainingApi.createRecord({
        taskId: selectedTask.id,
        exerciseId: selectedTask.exerciseId,
        planId: selectedTask.planId,
        patientId: patientId,
        actualDuration: recordForm.actualDuration,
        qualityScore: recordForm.qualityScore,
        subjectiveFeeling: recordForm.subjectiveFeeling,
        notes: recordForm.notes,
      });
      setShowRecordModal(false);
      setSelectedTask(null);
      loadData();
    } catch (error: any) {
      alert('提交失败: ' + (error.response?.data?.error || error.message));
    } finally {
      setSubmitting(false);
    }
  };

  const getTaskStatusStyle = (status: string) => {
    switch (status) {
      case 'completed':
        return 'bg-green-100 border-l-green-400 text-green-800';
      case 'overdue':
        return 'bg-orange-100 border-l-orange-400 text-orange-800';
      case 'abandoned':
        return 'bg-gray-100 border-l-gray-400 text-gray-600';
      default:
        return 'bg-gray-50 border-l-gray-300 text-gray-800';
    }
  };

  const getTaskStatusText = (status: string) => {
    switch (status) {
      case 'completed': return '已完成';
      case 'overdue': return '逾期';
      case 'abandoned': return '已放弃';
      default: return '待完成';
    }
  };

  const getDifficultyText = (level: number) => {
    const levels = ['', '最简单', '简单', '中等', '困难', '最困难'];
    return levels[level] || '未知';
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="text-gray-500">加载中...</div>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <div>
          <div className="flex items-center space-x-4">
            <Link to="/" className="text-blue-600 hover:text-blue-800">
              &larr; 返回列表
            </Link>
          </div>
          <h2 className="text-2xl font-bold text-gray-900 mt-2">
            {patient?.name} - 训练日历
          </h2>
          <p className="text-gray-500 mt-1">按天查看训练任务，完成后录入训练记录</p>
        </div>

        <div className="flex items-center space-x-4">
          <button
            onClick={() => setCurrentDate(subDays(currentDate, 1))}
            className="px-4 py-2 border rounded hover:bg-gray-50"
          >
            &larr;
          </button>
          <span className="text-lg font-medium">
            {format(currentDate, 'yyyy年MM月dd日')}
          </span>
          <button
            onClick={() => setCurrentDate(addDays(currentDate, 1))}
            className="px-4 py-2 border rounded hover:bg-gray-50"
          >
            &rarr;
          </button>
        </div>
      </div>

      {tasks.length === 0 ? (
        <div className="bg-white rounded-lg shadow p-8 text-center">
          <p className="text-gray-500">当天暂无训练任务</p>
        </div>
      ) : (
        <div className="space-y-4">
          {tasks.map((task) => (
            <div
              key={task.id}
              className={`p-4 rounded-lg border-l-4 ${getTaskStatusStyle(task.status)}`}
            >
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center space-x-3">
                    <h3 className="text-lg font-semibold">
                      {task.exercise?.name}
                    </h3>
                    <span className="px-2 py-1 bg-white bg-opacity-50 rounded text-xs">
                      {getTaskStatusText(task.status)}
                    </span>
                  </div>
                  <div className="mt-2 text-sm space-y-1">
                    <p>计划时长: {task.exercise?.durationMinutes} 分钟</p>
                    <p>难度等级: {task.exercise?.difficultyLevel}级 ({getDifficultyText(task.exercise?.difficultyLevel || 1)})</p>
                    <p>频率: {task.exercise?.frequencyType === 'daily' ? '每天' : '每周'} {task.exercise?.frequencyCount} 次</p>
                    <p>计划时间: {task.taskTime}</p>
                    {task.exercise?.goalDescription && (
                      <p className="text-gray-600">
                        目标: {task.exercise.goalDescription}
                      </p>
                    )}
                  </div>
                </div>
                {task.status !== 'completed' && task.status !== 'abandoned' && (
                  <button
                    onClick={() => handleOpenRecord(task)}
                    className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition"
                  >
                    录入记录
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {showRecordModal && selectedTask && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
            <div className="p-6">
              <h3 className="text-xl font-bold mb-4">录入训练记录</h3>
              <p className="text-gray-600 mb-4">
                {selectedTask.exercise?.name}
              </p>

              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700">
                    实际时长（分钟）
                  </label>
                  <input
                    type="number"
                    value={recordForm.actualDuration}
                    onChange={(e) => setRecordForm({ ...recordForm, actualDuration: parseInt(e.target.value) || 0 })}
                    className="mt-1 block w-full border rounded px-3 py-2"
                    min="1"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700">
                    完成质量评分（1-10）
                  </label>
                  <input
                    type="number"
                    value={recordForm.qualityScore}
                    onChange={(e) => setRecordForm({ ...recordForm, qualityScore: parseInt(e.target.value) || 0 })}
                    className="mt-1 block w-full border rounded px-3 py-2"
                    min="1"
                    max="10"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700">
                    患者主观感受
                  </label>
                  <select
                    value={recordForm.subjectiveFeeling}
                    onChange={(e) => setRecordForm({ ...recordForm, subjectiveFeeling: e.target.value as any })}
                    className="mt-1 block w-full border rounded px-3 py-2"
                  >
                    <option value="easy">轻松</option>
                    <option value="normal">正常</option>
                    <option value="hard">吃力</option>
                    <option value="very_hard">很吃力</option>
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700">
                    备注
                  </label>
                  <textarea
                    value={recordForm.notes}
                    onChange={(e) => setRecordForm({ ...recordForm, notes: e.target.value })}
                    className="mt-1 block w-full border rounded px-3 py-2"
                    rows={3}
                  />
                </div>
              </div>

              <div className="mt-6 flex justify-end space-x-3">
                <button
                  onClick={() => setShowRecordModal(false)}
                  className="px-4 py-2 border rounded hover:bg-gray-50"
                  disabled={submitting}
                >
                  取消
                </button>
                <button
                  onClick={handleSubmitRecord}
                  className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
                  disabled={submitting}
                >
                  {submitting ? '提交中...' : '提交'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
