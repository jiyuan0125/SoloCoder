import React from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import Labs from './components/Labs';
import Chemicals from './components/Chemicals';
import Trainings from './components/Trainings';
import Checks from './components/Checks';
import Todos from './components/Todos';
import './App.css';

const App = () => {
  return (
    <Router>
      <div className="app">
        <header className="app-header">
          <h1>实验室安全管理系统</h1>
          <nav className="nav">
            <Link to="/">实验室管理</Link>
            <Link to="/chemicals">危化品台账</Link>
            <Link to="/trainings">培训记录</Link>
            <Link to="/checks">安全检查</Link>
            <Link to="/todos">待办事项</Link>
          </nav>
        </header>
        <main className="app-main">
          <Routes>
            <Route path="/" element={<Labs />} />
            <Route path="/chemicals" element={<Chemicals />} />
            <Route path="/trainings" element={<Trainings />} />
            <Route path="/checks" element={<Checks />} />
            <Route path="/todos" element={<Todos />} />
          </Routes>
        </main>
      </div>
    </Router>
  );
};

export default App;
