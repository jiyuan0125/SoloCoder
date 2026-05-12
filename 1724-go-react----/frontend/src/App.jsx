import React from 'react'
import { BrowserRouter as Router, Routes, Route, Link, useNavigate } from 'react-router-dom'
import TasksPage from './pages/TasksPage.jsx'
import QuestionnairePage from './pages/QuestionnairePage.jsx'
import ResultsPage from './pages/ResultsPage.jsx'
import FeedbackPage from './pages/FeedbackPage.jsx'

function App() {
  return (
    <Router>
      <div className="app">
        <header className="header">
          <h1>教学评估系统</h1>
          <nav className="nav">
            <Link to="/">评估任务</Link>
            <Link to="/results">结果统计</Link>
            <Link to="/feedback">反馈跟踪</Link>
          </nav>
        </header>
        <main className="main">
          <Routes>
            <Route path="/" element={<TasksPage />} />
            <Route path="/questionnaire/:taskId/:courseId" element={<QuestionnairePage />} />
            <Route path="/results" element={<ResultsPage />} />
            <Route path="/feedback" element={<FeedbackPage />} />
          </Routes>
        </main>
      </div>
    </Router>
  )
}

export default App
