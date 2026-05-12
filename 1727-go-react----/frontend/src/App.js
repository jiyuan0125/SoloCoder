import React from 'react';
import { Routes, Route, Link, useNavigate, useLocation } from 'react-router-dom';
import { Layout, Menu, Badge } from 'antd';
import { ProjectOutlined, FileTextOutlined, ClockCircleOutlined } from '@ant-design/icons';
import ProjectList from './pages/ProjectList';
import ProjectDetail from './pages/ProjectDetail';
import ReimbursementList from './pages/ReimbursementList';
import PendingReimbursements from './pages/PendingReimbursements';
import FinalReport from './pages/FinalReport';

const { Header, Content, Sider } = Layout;

function App() {
  const navigate = useNavigate();
  const location = useLocation();

  const getSelectedKey = () => {
    if (location.pathname.startsWith('/pending')) return '2';
    if (location.pathname.startsWith('/reimbursements')) return '3';
    return '1';
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header style={{ background: '#001529', padding: '0 24px', display: 'flex', alignItems: 'center' }}>
        <h1 style={{ color: 'white', margin: 0, fontSize: '20px' }}>科研经费管理系统</h1>
      </Header>
      <Layout>
        <Sider width={200} style={{ background: '#fff' }}>
          <Menu
            mode="inline"
            selectedKeys={[getSelectedKey()]}
            style={{ height: '100%', borderRight: 0 }}
          >
            <Menu.Item key="1" icon={<ProjectOutlined />} onClick={() => navigate('/')}>
              项目管理
            </Menu.Item>
            <Menu.Item key="2" icon={<ClockCircleOutlined />} onClick={() => navigate('/pending')}>
              待办审批
            </Menu.Item>
            <Menu.Item key="3" icon={<FileTextOutlined />} onClick={() => navigate('/reimbursements')}>
              报销单管理
            </Menu.Item>
          </Menu>
        </Sider>
        <Layout style={{ padding: '24px' }}>
          <Content
            style={{
              padding: 24,
              margin: 0,
              minHeight: 280,
              background: '#fff',
              borderRadius: '8px',
            }}
          >
            <Routes>
              <Route path="/" element={<ProjectList />} />
              <Route path="/projects/:id" element={<ProjectDetail />} />
              <Route path="/projects/:id/final-report" element={<FinalReport />} />
              <Route path="/pending" element={<PendingReimbursements />} />
              <Route path="/reimbursements" element={<ReimbursementList />} />
            </Routes>
          </Content>
        </Layout>
      </Layout>
    </Layout>
  );
}

export default App;
