import React, { useState, useEffect } from 'react'
import {
  Table,
  Button,
  Modal,
  Form,
  Input,
  Select,
  DatePicker,
  Tag,
  Space,
  message,
  Card,
  Timeline,
  Row,
  Col,
  Divider,
  Popconfirm
} from 'antd'
import { PlusOutlined, UserOutlined, MedicineBoxOutlined, CheckCircleOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'

import { investigationApi, contactApi, outbreakApi, statsApi } from '../api'

const { Option } = Select

function InvestigationPage() {
  const [investigations, setInvestigations] = useState([])
  const [reports, setReports] = useState([])
  const [diseases, setDiseases] = useState([])
  const [contacts, setContacts] = useState([])
  const [selectedInv, setSelectedInv] = useState(null)
  const [loading, setLoading] = useState(false)
  const [invModalVisible, setInvModalVisible] = useState(false)
  const [contactModalVisible, setContactModalVisible] = useState(false)
  const [invForm] = Form.useForm()
  const [contactForm] = Form.useForm()

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setLoading(true)
    try {
      const [invRes, reportRes, diseaseRes] = await Promise.all([
        investigationApi.list(),
        outbreakApi.listReports(),
        statsApi.listDiseases()
      ])
      setInvestigations(invRes.data)
      setReports(reportRes.data)
      setDiseases(diseaseRes.data)
    } catch (err) {
      message.error('加载数据失败')
    } finally {
      setLoading(false)
    }
  }

  const loadContacts = async (invId) => {
    try {
      const res = await contactApi.list(invId)
      setContacts(res.data)
    } catch (err) {
      message.error('加载密接者失败')
    }
  }

  const handleCreateInvestigation = async (values) => {
    try {
      const data = {
        ...values,
        onsetTime: values.onsetTime?.format('YYYY-MM-DDTHH:mm:ss'),
        visitTime: values.visitTime?.format('YYYY-MM-DDTHH:mm:ss'),
        confirmTime: values.confirmTime?.format('YYYY-MM-DDTHH:mm:ss')
      }
      await investigationApi.create(data)
      message.success('流调创建成功')
      setInvModalVisible(false)
      invForm.resetFields()
      loadData()
    } catch (err) {
      message.error(err.response?.data?.error || '创建失败')
    }
  }

  const handleCreateContact = async (values) => {
    try {
      const data = {
        ...values,
        investigationId: selectedInv.id,
        firstContactDate: values.firstContactDate.format('YYYY-MM-DD'),
        lastContactDate: values.lastContactDate.format('YYYY-MM-DD')
      }
      await contactApi.create(data)
      message.success('密接者添加成功')
      setContactModalVisible(false)
      contactForm.resetFields()
      loadContacts(selectedInv.id)
    } catch (err) {
      message.error(err.response?.data?.error || '添加失败')
    }
  }

  const handleUpdateContactStatus = async (contactId, newStatus) => {
    try {
      await contactApi.updateStatus(contactId, newStatus)
      message.success('状态更新成功')
      loadContacts(selectedInv.id)
    } catch (err) {
      message.error(err.response?.data?.error || '更新失败')
    }
  }

  const getStatusColor = (status) => {
    switch (status) {
      case '正常': return 'green'
      case '出现症状': return 'orange'
      case '确诊': return 'red'
      case '排除': return 'default'
      default: return 'blue'
    }
  }

  const statusOptions = (currentStatus) => {
    const transitions = {
      '正常': ['出现症状', '排除'],
      '出现症状': ['确诊', '排除'],
      '确诊': [],
      '排除': []
    }
    return transitions[currentStatus] || []
  }

  const invColumns = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
    { title: '关联报告ID', dataIndex: 'reportId', key: 'reportId' },
    { title: '患者姓名', dataIndex: 'patientName', key: 'patientName' },
    {
      title: '发病时间',
      dataIndex: 'onsetTime',
      key: 'onsetTime',
      render: (t) => t ? dayjs(t).format('YYYY-MM-DD') : '-'
    },
    {
      title: '就诊时间',
      dataIndex: 'visitTime',
      key: 'visitTime',
      render: (t) => t ? dayjs(t).format('YYYY-MM-DD') : '-'
    },
    {
      title: '确诊时间',
      dataIndex: 'confirmTime',
      key: 'confirmTime',
      render: (t) => t ? dayjs(t).format('YYYY-MM-DD') : '-'
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Button type="link" onClick={() => {
          setSelectedInv(record)
          loadContacts(record.id)
        }}>
          查看密接者
        </Button>
      )
    },
  ]

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setInvModalVisible(true)}>
          新建流调
        </Button>
        <Button onClick={loadData}>刷新</Button>
      </Space>

      <Row gutter={16}>
        <Col span={selectedInv ? 12 : 24}>
          <Card title="流调列表" loading={loading}>
            <Table
              columns={invColumns}
              dataSource={investigations}
              rowKey="id"
              pagination={{ pageSize: 5 }}
              onRow={(record) => ({
                onClick: () => {
                  setSelectedInv(record)
                  loadContacts(record.id)
                }
              })}
              rowClassName={(record) =>
                selectedInv?.id === record.id ? 'ant-table-row-selected' : ''
              }
            />
          </Card>
        </Col>

        {selectedInv && (
          <Col span={12}>
            <Card
              title={`密接者管理 - 流调 #${selectedInv.id}`}
              extra={
                <Button
                  type="primary"
                  size="small"
                  icon={<PlusOutlined />}
                  onClick={() => setContactModalVisible(true)}
                >
                  添加密接者
                </Button>
              }
            >
              <Divider>传播链时间线</Divider>
              <Timeline mode="left">
                <Timeline.Item color="blue">
                  <p><strong>病例: {selectedInv.patientName}</strong></p>
                  <p style={{ color: '#666' }}>
                    发病: {selectedInv.onsetTime ? dayjs(selectedInv.onsetTime).format('YYYY-MM-DD') : '未知'}
                  </p>
                </Timeline.Item>
                {contacts.map((contact, idx) => (
                  <Timeline.Item
                    key={contact.id}
                    color={getStatusColor(contact.status) === 'red' ? 'red' : 'green'}
                    className="timeline-node"
                  >
                    <Card size="small" style={{ marginBottom: 8 }}>
                      <Space direction="vertical" size="small" style={{ width: '100%' }}>
                        <Space>
                          <UserOutlined />
                          <strong>{contact.name}</strong>
                          <Tag color={getStatusColor(contact.status)}>{contact.status}</Tag>
                        </Space>
                        <p style={{ margin: 0, fontSize: 12, color: '#666' }}>
                          接触方式: {contact.contactType}
                        </p>
                        <p style={{ margin: 0, fontSize: 12, color: '#666' }}>
                          观察期: {dayjs(contact.observationStart).format('MM-DD')} - {dayjs(contact.observationEnd).format('MM-DD')}
                        </p>
                        <Space>
                          {statusOptions(contact.status).map(s => (
                            <Popconfirm
                              key={s}
                              title={`确定将状态更新为"${s}"？`}
                              onConfirm={() => handleUpdateContactStatus(contact.id, s)}
                            >
                              <Button size="small" type={s === '确诊' ? 'primary' : 'default'} danger={s === '确诊'}>
                                {s}
                              </Button>
                            </Popconfirm>
                          ))}
                        </Space>
                      </Space>
                    </Card>
                  </Timeline.Item>
                ))}
              </Timeline>
            </Card>
          </Col>
        )}
      </Row>

      <Modal
        title="新建流调"
        open={invModalVisible}
        onCancel={() => setInvModalVisible(false)}
        onOk={() => invForm.submit()}
        okText="提交"
        cancelText="取消"
      >
        <Form form={invForm} layout="vertical" onFinish={handleCreateInvestigation}>
          <Form.Item name="reportId" label="关联报告" rules={[{ required: true }]}>
            <Select placeholder="请选择疫情报告">
              {reports.filter(r => !r.investigationId).map(r => (
                <Option key={r.id} value={r.id}>
                  #{r.id} - {r.patientName} ({r.diseaseName})
                </Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item name="patientName" label="患者姓名" rules={[{ required: true }]}>
            <Input placeholder="请输入患者姓名" />
          </Form.Item>
          <Form.Item name="onsetTime" label="发病时间">
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="visitTime" label="就诊时间">
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="confirmTime" label="确诊时间">
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="clinical" label="临床表现">
            <Input.TextArea rows={3} placeholder="请输入临床表现" />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="添加密切接触者"
        open={contactModalVisible}
        onCancel={() => setContactModalVisible(false)}
        onOk={() => contactForm.submit()}
        okText="提交"
        cancelText="取消"
      >
        <Form form={contactForm} layout="vertical" onFinish={handleCreateContact}>
          <Form.Item name="name" label="姓名" rules={[{ required: true }]}>
            <Input placeholder="请输入密接者姓名" />
          </Form.Item>
          <Form.Item name="phone" label="联系方式">
            <Input placeholder="请输入联系方式" />
          </Form.Item>
          <Form.Item name="contactType" label="接触方式" rules={[{ required: true }]}>
            <Select placeholder="请选择接触方式">
              <Option value="同住">同住</Option>
              <Option value="同餐">同餐</Option>
              <Option value="同工作">同工作</Option>
              <Option value="同乘车">同乘车</Option>
              <Option value="其他">其他</Option>
            </Select>
          </Form.Item>
          <Form.Item name="firstContactDate" label="首次接触日期" rules={[{ required: true }]}>
            <DatePicker style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="lastContactDate" label="最后接触日期" rules={[{ required: true }]}>
            <DatePicker style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="diseaseName" label="传染病名称" rules={[{ required: true }]}>
            <Select placeholder="请选择传染病">
              {diseases.map(d => (
                <Option key={d.id} value={d.name}>{d.name}</Option>
              ))}
            </Select>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default InvestigationPage
