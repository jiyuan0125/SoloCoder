import React, { useState, useEffect } from 'react';
import { Outlet, Link, useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import api from '../api';

const Layout = () => {
  const { user, logout, isAdmin, isTeacher, isParent } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [unreadCount, setUnreadCount] = useState(0);

  useEffect(() => {
    fetchUnreadCount();
    const interval = setInterval(fetchUnreadCount, 10000);
    return () => clearInterval(interval);
  }, []);

  const fetchUnreadCount = async () => {
    try {
      const [msgRes, annRes] = await Promise.all([
        api.get('/messages/unread-count'),
        isParent ? api.get('/announcements/unread-count') : { data: { count: 0 } },
      ]);
      setUnreadCount(msgRes.data.count + (annRes.data?.count || 0));
    } catch (error) {
      console.error('获取未读数量失败:', error);
    }
  };

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const navItems = [
    { path: '/', label: '通知公告', icon: '📢', allowed: ['admin', 'teacher', 'parent'] },
    { path: '/scores', label: '成绩管理', icon: '📊', allowed: ['admin', 'teacher', 'parent'] },
    { path: '/messages', label: '消息中心', icon: '💬', allowed: ['admin', 'teacher', 'parent'] },
  ];

  const isAllowed = (allowedRoles) => {
    return allowedRoles.some(role => {
      if (role === 'admin') return isAdmin;
      if (role === 'teacher') return isTeacher;
      if (role === 'parent') return isParent;
      return false;
    });
  };

  return (
    <div style={styles.app}>
      <header style={styles.header}>
        <div style={styles.headerLeft}>
          <h1 style={styles.logo}>🏫 家校互联平台</h1>
        </div>
        <div style={styles.headerRight}>
          <span style={styles.userInfo}>
            {user?.user?.name} ({getRoleLabel(user?.user?.role)})
          </span>
          <button onClick={handleLogout} style={styles.logoutButton}>
            退出登录
          </button>
        </div>
      </header>

      <div style={styles.main}>
        <nav style={styles.sidebar}>
          {navItems.filter(item => isAllowed(item.allowed)).map(item => (
            <Link
              key={item.path}
              to={item.path}
              style={{
                ...styles.navItem,
                ...(location.pathname === item.path ? styles.navItemActive : {}),
              }}
            >
              <span style={{ fontSize: '20px' }}>{item.icon}</span>
              <span>{item.label}</span>
              {item.path === '/messages' && unreadCount > 0 && (
                <span style={styles.badge}>{unreadCount}</span>
              )}
            </Link>
          ))}
        </nav>

        <main style={styles.content}>
          <Outlet />
        </main>
      </div>
    </div>
  );
};

const getRoleLabel = (role) => {
  const labels = {
    admin: '管理员',
    teacher: '教师',
    parent: '家长',
  };
  return labels[role] || role;
};

const styles = {
  app: {
    minHeight: '100vh',
    background: '#f5f7fa',
  },
  header: {
    background: 'white',
    padding: '0 30px',
    height: '60px',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    boxShadow: '0 2px 8px rgba(0,0,0,0.06)',
    position: 'sticky',
    top: 0,
    zIndex: 100,
  },
  headerLeft: {
    display: 'flex',
    alignItems: 'center',
    gap: '12px',
  },
  logo: {
    margin: 0,
    fontSize: '20px',
    color: '#333',
  },
  headerRight: {
    display: 'flex',
    alignItems: 'center',
    gap: '20px',
  },
  userInfo: {
    color: '#666',
    fontSize: '14px',
  },
  logoutButton: {
    padding: '8px 16px',
    background: '#f5f5f5',
    border: 'none',
    borderRadius: '6px',
    cursor: 'pointer',
    fontSize: '14px',
    color: '#666',
  },
  main: {
    display: 'flex',
    minHeight: 'calc(100vh - 60px)',
  },
  sidebar: {
    width: '220px',
    background: 'white',
    padding: '20px 0',
    borderRight: '1px solid #eee',
  },
  navItem: {
    display: 'flex',
    alignItems: 'center',
    gap: '12px',
    padding: '14px 24px',
    color: '#666',
    textDecoration: 'none',
    fontSize: '15px',
    transition: 'all 0.2s',
  },
  navItemActive: {
    background: '#eef2ff',
    color: '#4f46e5',
    borderRight: '3px solid #4f46e5',
  },
  content: {
    flex: 1,
    padding: '24px',
  },
  badge: {
    background: '#ef4444',
    color: 'white',
    fontSize: '11px',
    padding: '2px 8px',
    borderRadius: '10px',
    marginLeft: 'auto',
  },
};

export default Layout;
