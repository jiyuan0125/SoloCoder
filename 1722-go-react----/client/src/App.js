import React, { useState } from 'react';
import { Routes, Route, NavLink } from 'react-router-dom';
import KnowledgeManagement from './pages/KnowledgeManagement';
import QuestionManagement from './pages/QuestionManagement';
import ExamPage from './pages/ExamPage';
import ReportPage from './pages/ReportPage';

function App() {
  const [studentId] = useState('student_001');

  return (
    <div className="app">
      <header className="header">
        <h1>智能题库系统</h1>
        <nav className="nav">
          <NavLink to="/" end>知识点管理</NavLink>
          <NavLink to="/questions">题库管理</NavLink>
          <NavLink to="/exam">测试页面</NavLink>
          <NavLink to="/report">学习报告</NavLink>
        </nav>
      </header>
      <main className="main">
        <Routes>
          <Route path="/" element={<KnowledgeManagement />} />
          <Route path="/questions" element={<QuestionManagement />} />
          <Route path="/exam" element={<ExamPage studentId={studentId} />} />
          <Route path="/report" element={<ReportPage studentId={studentId} />} />
        </Routes>
      </main>
    </div>
  );
}

export default App;
