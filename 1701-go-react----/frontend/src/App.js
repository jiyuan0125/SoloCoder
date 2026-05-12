import React from 'react';
import { BrowserRouter as Router, Routes, Route, Link, useLocation } from 'react-router-dom';
import { Layout, Menu, ConfigProvider, theme } from 'antd';
import { HomeOutlined, UserOutlined, MedicineBoxOutlined } from '@ant-design/icons';
import zhCN from 'antd/locale/zh_CN';
import RegistrationHall from './components/RegistrationHall';
import DoctorWorkstation from './components/DoctorWorkstation';
import Pharmacy from './components/Pharmacy';

const { Header, Content, Sider } = Layout;

const menuItems = [
  {
    key: '/',
    icon: <HomeOutlined />,
    label: <Link to="/">挂号大厅</Link>,
  },
  {
    key: '/doctor',
    icon: <UserOutlined />,
    label: <Link to="/doctor">诊室工作台</Link>,
  },
  {
    key: '/pharmacy',
    icon: <MedicineBoxOutlined />,
    label: <Link to="/pharmacy">药房管理台</Link>,
  },
];

const AppLayout = () => {
  const location = useLocation();

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header style={{ display: 'flex', alignItems: 'center' }}>
        <div style={{ color: 'white', fontSize: '20px', fontWeight: 'bold' }}>
          中医馆管理系统
        </div>
      </Header>
      <Layout>
        <Sider width={200} theme="light">
          <Menu
            mode="inline"
            selectedKeys={[location.pathname]}
            items={menuItems}
            style={{ height: '100%', borderRight: 0 }}
          />
        </Sider>
        <Layout style={{ padding: '24px' }}>
          <Content
            style={{
              background: '#fff',
              padding: 24,
              margin: 0,
              minHeight: 280,
              borderRadius: 8,
            }}
          >
            <Routes>
              <Route path="/" element={<RegistrationHall />} />
              <Route path="/doctor" element={<DoctorWorkstation />} />
              <Route path="/pharmacy" element={<Pharmacy />} />
            </Routes>
          </Content>
        </Layout>
      </Layout>
    </Layout>
  );
};

function App() {
  return (
    <ConfigProvider locale={zhCN} theme={{ algorithm: theme.defaultAlgorithm }}>
      <Router>
        <AppLayout />
      </Router>
    </ConfigProvider>
  );
}

export default App;
