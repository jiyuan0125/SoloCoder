import { Routes, Route, Link } from 'react-router-dom'
import PatientList from './pages/PatientList'
import TrainingCalendar from './pages/TrainingCalendar'
import AssessmentRecords from './pages/AssessmentRecords'
import SummaryPanel from './pages/SummaryPanel'

function App() {
  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white shadow-lg">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex items-center">
              <span className="text-xl font-bold text-blue-600">康复管理系统</span>
            </div>
            <div className="flex space-x-4 items-center">
              <Link
                to="/"
                className="px-3 py-2 rounded-md text-sm font-medium text-gray-700 hover:text-blue-600 hover:bg-blue-50"
              >
                患者列表
              </Link>
            </div>
          </div>
        </div>
      </nav>

      <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <Routes>
          <Route path="/" element={<PatientList />} />
          <Route path="/patients/:id/training" element={<TrainingCalendar />} />
          <Route path="/patients/:id/assessments" element={<AssessmentRecords />} />
          <Route path="/patients/:id/summary" element={<SummaryPanel />} />
        </Routes>
      </main>
    </div>
  )
}

export default App
