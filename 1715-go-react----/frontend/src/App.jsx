import React, { useState, useEffect } from 'react'
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom'
import { Layout, Menu, Badge, Alert, Spin } from 'antd'
import {
  FileTextOutlined,
  SearchOutlined,
  MedicineBoxOutlined,
  BarChartOutlined,
  BellOutlined
} from '@ant-design/icons'

import OutbreakPage from './pages/OutbreakPage'
import InvestigationPage from './pages/InvestigationPage'
import VaccinePage from './pages/VaccinePage'
import DashboardPage from './pages/DashboardPage'
import { outbreakApi, todoApi } from './api'

const { Header, Sider, Content } = Layout

function App() {
  const [alerts, setAlerts] = useState([])
  const [pendingTodos, setPendingTodos] = useState(0)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadData()
    const interval = setInterval(loadData, 30000)
    return () => clearInterval(interval)
  }, [])

  const loadData = async () => {
    try {
      const [alertRes, todoRes] = await Promise.all([
        outbreakApi.getAlerts(),
        todoApi.list()
      ])
      setAlerts(alertRes.data.filter(a => a.isActive))
      const pending = todoRes.data.filter(t => t.status === '待办').length
      setPendingTodos(pending)
    } catch (err) {
      console.error('加载数据失败:', err)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <Spin size="large" />
      </div>
    )
  }

  const menuItems = [
    { key: '/', icon: <FileTextOutlined />, label: <Link to="/">疫情报告</Link> },
    { key: '/investigation', icon: <SearchOutlined />, label: <Link to="/investigation">流调管理</Link> },
    { key: '/vaccine', icon: <MedicineBoxOutlined />, label: <Link to="/vaccine">疫苗接种</Link> },
    { key: '/dashboard', icon: <BarChartOutlined />, label: <Link to="/dashboard">统计面板</Link> },
  ]

  return (
    <Router>
      <Layout style={{ minHeight: '100vh' }}>
        <Header style={{ padding: 0, background: '#001529' }}>
          <div style={{ display: 'flex', alignItems: 'center', height: '100%', padding: '0 24px' }}>
            <h1 style={{ color: 'white', margin: 0, fontSize: '20px', flex: 1 }}>
              疫情报告管理系统
            </h1>
            <Badge count={pendingTodos} style={{ marginRight: 24 }}>
              <BellOutlined style={{ color: 'white', fontSize: 20 }} />
            </Badge>
          </div>
        </Header>

        {alerts.length > 0 && (
          <div style={{ padding: '8px 24px', background: '#fff1f0' }}>
            {alerts.map(alert => (
              <Alert
                key={alert.id}
                className="alert-banner"
                message={alert.message}
                type="error"
                showIcon
                style={{ marginBottom: alerts.length > 1 ? 8 : 0 }}
              />
            ))}
          </div>
        )}

        <Layout>
          <Sider width={200} style={{ background: '#fff' }}>
            <Menu
              mode="inline"
              defaultSelectedKeys={['/']}
              style={{ height: '100%', borderRight: 0 }}
              items={menuItems}
            />
          </Sider>
          <Layout style={{ padding: 24 }}>
            <Content
              style={{
                background: '#fff',
                padding: 24,
                margin: 0,
                minHeight: 280,
                borderRadius: 8
              }}
            >
              <Routes>
                <Route path="/" element={<OutbreakPage />} />
                <Route path="/investigation" element={<InvestigationPage />} />
                <Route path="/vaccine" element={<VaccinePage />} />
                <Route path="/dashboard" element={<DashboardPage />} />
              </Routes>
            </Content>
          </Layout>
        </Layout>
      </Layout>
    </Router>
  )
}

export default App
