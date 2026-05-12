import React from 'react';
import { Layout, Menu, Avatar, Dropdown, Button } from 'antd';
import { 
  BookOutlined, 
  SearchOutlined, 
  FileTextOutlined, 
  TeamOutlined,
  UserOutlined,
  LogoutOutlined 
} from '@ant-design/icons';
import { useNavigate, useLocation, Outlet } from 'react-router-dom';
import { getStoredUser, clearAuth } from '../utils/auth';
import { ROLE_MAP } from '../utils/constants';

const { Header, Sider, Content } = Layout;

const AppLayout: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const user = getStoredUser();

  const handleLogout = () => {
    clearAuth();
    navigate('/login');
  };

  const getMenuItems = () => {
    const items = [];
    
    if (user?.role === 'admin') {
      items.push({
        key: '/diplomas',
        icon: <BookOutlined />,
        label: '学历信息管理',
      });
    }
    
    if (user?.role === 'admin' || user?.role === 'verifier') {
      items.push({
        key: '/verify',
        icon: <SearchOutlined />,
        label: '学历认证查询',
      });
    }
    
    items.push({
      key: '/logs',
      icon: <FileTextOutlined />,
      label: '认证日志',
    });
    
    if (user?.role === 'admin') {
      items.push({
        key: '/users',
        icon: <TeamOutlined />,
        label: '系统管理',
      });
    }
    
    return items;
  };

  const userMenu = {
    items: [
      {
        key: '1',
        label: (
          <div>
            <div><strong>{user?.username}</strong></div>
            <div style={{ fontSize: 12, color: '#999' }}>{ROLE_MAP[user?.role || 'viewer']}</div>
          </div>
        ),
        disabled: true,
      },
      {
        type: 'divider',
      },
      {
        key: '2',
        icon: <LogoutOutlined />,
        label: '退出登录',
        onClick: handleLogout,
      },
    ],
  };

  const defaultSelectedKey = getMenuItems().find(
    item => location.pathname.startsWith(item.key as string)
  )?.key as string || getMenuItems()[0]?.key as string;

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header style={{ 
        display: 'flex', 
        alignItems: 'center', 
        justifyContent: 'space-between',
        background: '#001529',
        padding: '0 24px'
      }}>
        <div style={{ color: 'white', fontSize: 20, fontWeight: 'bold' }}>
          学历认证管理系统
        </div>
        
        <Dropdown menu={userMenu} placement="bottomRight">
          <Button type="text" style={{ color: 'white' }}>
            <Avatar icon={<UserOutlined />} style={{ marginRight: 8 }} />
            <span>{user?.username}</span>
          </Button>
        </Dropdown>
      </Header>
      
      <Layout>
        <Sider width={200} style={{ background: '#fff' }}>
          <Menu
            mode="inline"
            selectedKeys={[defaultSelectedKey]}
            style={{ height: '100%', borderRight: 0 }}
            items={getMenuItems()}
            onClick={({ key }) => navigate(key)}
          />
        </Sider>
        
        <Layout style={{ padding: '24px' }}>
          <Content 
            style={{ 
              padding: 24, 
              margin: 0, 
              minHeight: 280, 
              background: '#fff',
              borderRadius: 8 
            }}
          >
            <Outlet />
          </Content>
        </Layout>
      </Layout>
    </Layout>
  );
};

export default AppLayout;
