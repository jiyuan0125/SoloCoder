import React, { useState } from 'react'
import { BrowserRouter, Routes, Route, NavLink } from 'react-router-dom'
import CourseGraph from './pages/CourseGraph'
import LearningPath from './pages/LearningPath'
import CourseLearning from './pages/CourseLearning'
import Profile from './pages/Profile'

const DEMO_STUDENT_ID = 'demo-student-001'

function App() {
  const [studentId] = useState(DEMO_STUDENT_ID)

  return (
    <BrowserRouter>
      <nav>
        <ul>
          <li>
            <NavLink to="/" end>课程图谱</NavLink>
          </li>
          <li>
            <NavLink to="/path">学习路径</NavLink>
          </li>
          <li>
            <NavLink to="/learning">课程学习</NavLink>
          </li>
          <li>
            <NavLink to="/profile">个人中心</NavLink>
          </li>
        </ul>
      </nav>

      <div className="container">
        <Routes>
          <Route path="/" element={<CourseGraph studentId={studentId} />} />
          <Route path="/path" element={<LearningPath studentId={studentId} />} />
          <Route path="/learning" element={<CourseLearning studentId={studentId} />} />
          <Route path="/profile" element={<Profile studentId={studentId} />} />
        </Routes>
      </div>
    </BrowserRouter>
  )
}

export default App
