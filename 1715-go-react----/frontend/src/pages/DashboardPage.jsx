import React, { useState, useEffect } from 'react'
import {
  Row,
  Col,
  Card,
  Statistic,
  Table,
  Button,
  message,
  Spin,
  Space
} from 'antd'
import {
  UserOutlined,
  WarningOutlined,
  TeamOutlined,
  CheckCircleOutlined,
  FileTextOutlined,
  MedicineBoxOutlined,
  ReloadOutlined
} from '@ant-design/icons'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  ResponsiveContainer
} from 'recharts'
import dayjs from 'dayjs'

import { statsApi, todoApi } from '../api'

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884d8', '#82ca9d']

function DashboardPage() {
  const [stats, setStats] = useState(null)
  const [weeklyReports, setWeeklyReports] = useState([])
  const [todos, setTodos] = useState([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setLoading(true)
    try {
      const [statsRes, weeklyRes, todoRes] = await Promise.all([
        statsApi.getStatistics(),
        statsApi.listWeeklyReports(),
        todoApi.list()
      ])
      setStats(statsRes.data)
      setWeeklyReports(weeklyRes.data)
      setTodos(todoRes.data)
    } catch (err) {
      message.error('加载数据失败')
    } finally {
      setLoading(false)
    }
  }

  const handleGenerateWeekly = async () => {
    try {
      await statsApi.generateWeeklyReport()
      message.success('周报生成成功')
      loadData()
    } catch (err) {
      message.error('生成失败')
    }
  }

  const handleTodoStatus = async (id, status) => {
    try {
      await todoApi.updateStatus(id, status)
      loadData()
      message.success('状态更新成功')
    } catch (err) {
      message.error('更新失败')
    }
  }

  const getPriorityColor = (priority) => {
    switch (priority) {
      case '最高': return 'red'
      case '高': return 'orange'
      case '中': return 'blue'
      default: return 'default'
    }
  }

  const todoColumns = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
    { title: '标题', dataIndex: 'title', key: 'title' },
    { title: '描述', dataIndex: 'description', key: 'description' },
    {
      title: '优先级',
      dataIndex: 'priority',
      key: 'priority',
      render: (p) => <span style={{ color: getPriorityColor(p), fontWeight: 'bold' }}>{p}</span>
    },
    { title: '状态', dataIndex: 'status', key: 'status' },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (t) => dayjs(t).format('MM-DD HH:mm')
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => {
        if (record.status === '待办') {
          return (
            <Button type="link" size="small" onClick={() => handleTodoStatus(record.id, '处理中')}>
              开始处理
            </Button>
          )
        }
        if (record.status === '处理中') {
          return (
            <Button type="link" size="small" onClick={() => handleTodoStatus(record.id, '已完成')}>
              完成
            </Button>
          )
        }
        return null
      }
    },
  ]

  const weeklyColumns = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
    {
      title: '周期',
      key: 'period',
      render: (_, r) => `${dayjs(r.weekStart).format('MM-DD')} ~ ${dayjs(r.weekEnd).format('MM-DD')}`
    },
    { title: '新增病例', dataIndex: 'newCases', key: 'newCases' },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (t) => dayjs(t).format('YYYY-MM-DD HH:mm')
    },
  ]

  if (loading || !stats) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '60vh' }}>
        <Spin size="large" />
      </div>
    )
  }

  const trendData = JSON.parse(stats.trend || '[]')
  const diseaseDist = JSON.parse(stats.diseaseDistribution || '{}')
  const diseaseData = Object.keys(diseaseDist).map(k => ({ name: k, value: diseaseDist[k] }))
  const regionDist = JSON.parse(stats.regionDistribution || '{}')
  const regionData = Object.keys(regionDist).map(k => ({ name: k, value: regionDist[k] }))

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button icon={<ReloadOutlined />} onClick={loadData}>刷新</Button>
        <Button type="primary" onClick={handleGenerateWeekly}>生成周报</Button>
      </Space>

      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={6}>
          <Card>
            <Statistic
              title="总病例数"
              value={stats.totalCases || 0}
              prefix={<UserOutlined />}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="迟报数"
              value={stats.delayedCount || 0}
              prefix={<WarningOutlined />}
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="观察中密接"
              value={stats.activeContacts || 0}
              prefix={<TeamOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="已排除"
              value={stats.excludedContacts || 0}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={12}>
          <Card title="近30天病例趋势">
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={trendData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="date" />
                <YAxis />
                <Tooltip />
                <Legend />
                <Line type="monotone" dataKey="count" stroke="#8884d8" name="病例数" />
              </LineChart>
            </ResponsiveContainer>
          </Card>
        </Col>
        <Col span={12}>
          <Card title="病种分布">
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={diseaseData}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                  outerRadius={80}
                  fill="#8884d8"
                  dataKey="value"
                >
                  {diseaseData.map((_, index) => (
                    <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip />
              </PieChart>
            </ResponsiveContainer>
          </Card>
        </Col>
      </Row>

      <Row gutter={16}>
        <Col span={12}>
          <Card title="待办事项" extra={<FileTextOutlined />}>
            <Table
              columns={todoColumns}
              dataSource={todos}
              rowKey="id"
              size="small"
              pagination={{ pageSize: 5 }}
            />
          </Card>
        </Col>
        <Col span={12}>
          <Card title="地区分布">
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={regionData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="name" />
                <YAxis />
                <Tooltip />
                <Bar dataKey="value" fill="#82ca9d" name="病例数" />
              </BarChart>
            </ResponsiveContainer>
          </Card>
        </Col>
      </Row>

      <Row style={{ marginTop: 16 }}>
        <Col span={24}>
          <Card title="疫情周报">
            <Table
              columns={weeklyColumns}
              dataSource={weeklyReports}
              rowKey="id"
              pagination={{ pageSize: 5 }}
            />
          </Card>
        </Col>
      </Row>
    </div>
  )
}

export default DashboardPage
