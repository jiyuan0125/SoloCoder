import React, { useState } from 'react'
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom'
import CourseHall from './pages/CourseHall'
import LiveClassroom from './pages/LiveClassroom'
import LearningCenter from './pages/LearningCenter'
import AdminPanel from './pages/AdminPanel'

const App = () => {
  const [currentStudent, setCurrentStudent] = useState({ id: 1, nickname: '小明', balance: 10000 })

  return (
    <Router>
      <div className="min-h-screen bg-gray-50">
        <nav className="bg-blue-600 text-white shadow-lg">
          <div className="max-w-7xl mx-auto px-4">
            <div className="flex items-center justify-between h-16">
              <div className="flex items-center space-x-8">
                <Link to="/" className="text-xl font-bold">在线教育平台</Link>
                <div className="flex space-x-4">
                  <Link to="/" className="hover:bg-blue-700 px-3 py-2 rounded">课程大厅</Link>
                  <Link to="/learning" className="hover:bg-blue-700 px-3 py-2 rounded">学习中心</Link>
                  <Link to="/admin" className="hover:bg-blue-700 px-3 py-2 rounded">管理面板</Link>
                </div>
              </div>
              <div className="flex items-center space-x-4">
                <span>{currentStudent.nickname}</span>
                <span className="bg-blue-500 px-3 py-1 rounded">
                  余额: ¥{(currentStudent.balance / 100).toFixed(2)}
                </span>
              </div>
            </div>
          </div>
        </nav>
        
        <main className="max-w-7xl mx-auto py-6 px-4">
          <Routes>
            <Route path="/" element={<CourseHall currentStudent={currentStudent} setCurrentStudent={setCurrentStudent} />} />
            <Route path="/live/:id" element={<LiveClassroom currentStudent={currentStudent} />} />
            <Route path="/learning" element={<LearningCenter currentStudent={currentStudent} />} />
            <Route path="/admin" element={<AdminPanel />} />
          </Routes>
        </main>
      </div>
    </Router>
  )
}

export default App
