import React, { useState, useEffect, useCallback } from 'react';
import { api } from '../utils/api';

const EXAM_DURATION = 30 * 60;

function ExamPage({ studentId }) {
  const [examState, setExamState] = useState('idle');
  const [exam, setExam] = useState(null);
  const [currentQuestion, setCurrentQuestion] = useState(null);
  const [questionNum, setQuestionNum] = useState(0);
  const [totalQuestions, setTotalQuestions] = useState(20);
  const [selectedAnswer, setSelectedAnswer] = useState('');
  const [answerResult, setAnswerResult] = useState(null);
  const [timeLeft, setTimeLeft] = useState(EXAM_DURATION);
  const [questionStartTime, setQuestionStartTime] = useState(null);
  const [examResult, setExamResult] = useState(null);

  const parseOptions = useCallback((options) => {
    if (!options) return [];
    try {
      return JSON.parse(options);
    } catch {
      return [options];
    }
  }, []);

  const startExam = async () => {
    try {
      const data = await api.startExam(studentId);
      setExam(data.exam);
      setCurrentQuestion(data.question);
      setQuestionNum(data.question_number);
      setTotalQuestions(data.total_questions);
      setExamState('ongoing');
      setTimeLeft(EXAM_DURATION);
      setQuestionStartTime(Date.now());
      setSelectedAnswer('');
      setAnswerResult(null);
    } catch (error) {
      alert(error.message);
    }
  };

  const submitAnswer = async () => {
    if (!selectedAnswer) {
      alert('请先选择或输入答案');
      return;
    }

    const timeSpent = questionStartTime ? Math.floor((Date.now() - questionStartTime) / 1000) : 0;

    try {
      const data = await api.submitAnswer(exam.id, selectedAnswer, timeSpent);
      setExam(data.exam);
      setAnswerResult(data.current_answer_result || null);

      if (data.completed) {
        setExamState('completed');
        setExamResult(data);
        return;
      }

      setTimeout(() => {
        setCurrentQuestion(data.next_question);
        setQuestionNum(data.question_number);
        setSelectedAnswer('');
        setAnswerResult(null);
        setQuestionStartTime(Date.now());
      }, 1500);
    } catch (error) {
      alert(error.message);
    }
  };

  const markIncomplete = useCallback(async () => {
    if (!exam) return;
    try {
      await api.markExamIncomplete(exam.id);
      setExamState('idle');
      setExam(null);
      setCurrentQuestion(null);
    } catch (error) {
      console.error('Failed to mark incomplete:', error);
    }
  }, [exam]);

  useEffect(() => {
    if (examState !== 'ongoing') return;

    const timer = setInterval(() => {
      setTimeLeft((prev) => {
        if (prev <= 1) {
          clearInterval(timer);
          markIncomplete();
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    const handleBeforeUnload = () => {
      markIncomplete();
    };
    window.addEventListener('beforeunload', handleBeforeUnload);

    return () => {
      clearInterval(timer);
      window.removeEventListener('beforeunload', handleBeforeUnload);
    };
  }, [examState, markIncomplete]);

  const formatTime = (seconds) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  };

  const progress = (questionNum / totalQuestions) * 100;

  if (examState === 'idle') {
    return (
      <div className="exam-container">
        <div className="question-card">
          <div style={{ textAlign: 'center', padding: '2rem' }}>
            <h2 style={{ marginBottom: '1rem' }}>自适应测试</h2>
            <p style={{ marginBottom: '1rem', color: '#666' }}>
              每次测试共20题，限时30分钟。系统会根据你的答题情况自动调整题目难度。
            </p>
            <ul style={{ textAlign: 'left', margin: '1.5rem auto', maxWidth: '400px', color: '#555' }}>
              <li>初始难度为3级</li>
              <li>连续答对2题，难度提升1级</li>
              <li>连续答错2题，难度降低1级</li>
              <li>答对答错交替，难度不变</li>
              <li>提交答案后无法回看</li>
            </ul>
            <button
              className="btn btn-primary"
              style={{ fontSize: '1.1rem', padding: '0.75rem 2rem', marginTop: '1rem' }}
              onClick={startExam}
            >
              开始测试
            </button>
          </div>
        </div>
      </div>
    );
  }

  if (examState === 'completed') {
    return (
      <div className="exam-container">
        <div className="question-card">
          <div style={{ textAlign: 'center', padding: '1rem' }}>
            <h2 style={{ marginBottom: '1.5rem', color: '#28a745' }}>测试完成！</h2>
            <div style={{ fontSize: '3rem', fontWeight: 'bold', color: '#667eea', marginBottom: '1rem' }}>
              {exam?.TotalScore || examResult?.exam?.total_score || 0} 分
            </div>
            <p style={{ color: '#666', marginBottom: '0.5rem' }}>
              答对 {exam?.CorrectCount || examResult?.exam?.correct_count || 0} / {totalQuestions} 题
            </p>
            <p style={{ color: '#666', marginBottom: '2rem' }}>
              正确率: {totalQuestions > 0 ? Math.round((exam?.CorrectCount || examResult?.exam?.correct_count || 0) / totalQuestions * 100) : 0}%
            </p>
            <button
              className="btn btn-primary"
              onClick={() => {
                setExamState('idle');
                setExamResult(null);
              }}
            >
              返回
            </button>
          </div>
        </div>
      </div>
    );
  }

  const options = parseOptions(currentQuestion?.options);

  return (
    <div className="exam-container">
      <div className="card" style={{ marginBottom: '1rem' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.5rem' }}>
          <span>
            第 {questionNum} 题 / 共 {totalQuestions} 题
          </span>
          <span className={`timer ${timeLeft < 300 ? 'warning' : ''}`}>
            ⏱ {formatTime(timeLeft)}
          </span>
        </div>
        <div className="progress-bar">
          <div className="progress-fill" style={{ width: `${progress}%` }} />
        </div>
        <div style={{ display: 'flex', gap: '1rem', fontSize: '0.9rem', color: '#666' }}>
          <span>当前难度: L{exam?.current_difficulty || 3}</span>
        </div>
      </div>

      <div className="question-card">
        <div className="question-info">
          <span>
            类型: {{
              single_choice: '单选题',
              multiple_choice: '多选题',
              true_false: '判断题',
              fill_blank: '填空题',
              essay: '简答题',
            }[currentQuestion?.type] || currentQuestion?.type}
          </span>
          <span>难度: L{currentQuestion?.difficulty}</span>
          <span>分值: {currentQuestion?.score}分</span>
        </div>

        <div className="question-content">
          {currentQuestion?.content}
        </div>

        {options.length > 0 ? (
          <ul className="options-list">
            {options.map((option, index) => (
              <li
                key={index}
                className={`option-item ${selectedAnswer === option ? 'selected' : ''}`}
                onClick={() => !answerResult && setSelectedAnswer(option)}
              >
                <span className="option-letter">
                  {String.fromCharCode(65 + index)}
                </span>
                {option}
              </li>
            ))}
          </ul>
        ) : (
          <div className="form-group">
            <label>请输入答案：</label>
            <textarea
              value={selectedAnswer}
              onChange={(e) => !answerResult && setSelectedAnswer(e.target.value)}
              placeholder="请输入你的答案..."
              rows={4}
              disabled={!!answerResult}
            />
          </div>
        )}

        {answerResult && (
          <div className={`answer-result ${answerResult.is_correct ? 'answer-correct' : 'answer-wrong'}`}>
            <p style={{ fontWeight: 'bold', marginBottom: '0.5rem' }}>
              {answerResult.is_correct ? '✓ 回答正确！' : '✗ 回答错误'}
            </p>
            {answerResult.needs_grading && (
              <p style={{ color: '#f0ad4e' }}>主观题需要人工判分</p>
            )}
            {answerResult.explanation && (
              <p style={{ marginTop: '0.5rem' }}>
                <strong>解析：</strong>{answerResult.explanation}
              </p>
            )}
          </div>
        )}

        <div className="submit-area">
          {!answerResult && (
            <button
              className="btn btn-primary"
              style={{ fontSize: '1rem', padding: '0.75rem 2rem' }}
              onClick={submitAnswer}
            >
              提交答案
            </button>
          )}
          {answerResult && (
            <p style={{ color: '#666' }}>正在加载下一题...</p>
          )}
        </div>
      </div>
    </div>
  );
}

export default ExamPage;
