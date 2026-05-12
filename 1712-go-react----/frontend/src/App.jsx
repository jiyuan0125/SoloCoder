import React, { useState } from 'react';
import ResidentsPage from './pages/ResidentsPage';
import CheckupsPage from './pages/CheckupsPage';
import ChronicDiseasePage from './pages/ChronicDiseasePage';
import FamilyPage from './pages/FamilyPage';
import './App.css';

function App() {
  const [currentPage, setCurrentPage] = useState('residents');
  const [selectedResident, setSelectedResident] = useState(null);
  const pages = [
    { id: 'residents', name: '居民档案' },
    { id: 'checkups', name: '体检记录' },
    { id: 'chronic', name: '慢性病管理' },
    { id: 'family', name: '家庭档案' },
  ];

  const renderPage = () => {
    switch (currentPage) {
      case 'residents':
        return (
          <ResidentsPage
            onSelectResident={(resident) => {
              setSelectedResident(resident);
              setCurrentPage('checkups');
            }}
          />
        );
      case 'checkups':
        return <CheckupsPage selectedResident={selectedResident} />;
      case 'chronic':
        return <ChronicDiseasePage />;
      case 'family':
        return <FamilyPage />;
      default:
        return <ResidentsPage />;
    }
  };

  return (
    <div className="app">
      <header className="app-header">
        <h1>居民健康档案管理系统</h1>
        <nav className="app-nav">
          {pages.map((page) => (
            <button
              key={page.id}
              className={currentPage === page.id ? 'active' : ''}
              onClick={() => setCurrentPage(page.id)}
            >
              {page.name}
            </button>
          ))}
        </nav>
      </header>
      <main className="app-main">
        {renderPage()}
      </main>
    </div>
  );
}

export default App;
