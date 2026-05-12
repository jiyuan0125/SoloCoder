import React, { useState, useEffect } from 'react'
import {
  Table,
  Button,
  Modal,
  Form,
  Input,
  Select,
  DatePicker,
  InputNumber,
  Space,
  message,
  Card,
  Tabs,
  Tag
} from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'

import { vaccineApi } from '../api'

const { Option } = Select
const { TabPane } = Tabs

function VaccinePage() {
  const [vaccines, setVaccines] = useState([])
  const [records, setRecords] = useState([])
  const [loading, setLoading] = useState(false)
  const [vaccineModalVisible, setVaccineModalVisible] = useState(false)
  const [recordModalVisible, setRecordModalVisible] = useState(false)
  const [vaccineForm] = Form.useForm()
  const [recordForm] = Form.useForm()

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setLoading(true)
    try {
      const [vaccineRes, recordRes] = await Promise.all([
        vaccineApi.listVaccines(),
        vaccineApi.listAllRecords()
      ])
      setVaccines(vaccineRes.data)
      setRecords(recordRes.data)
    } catch (err) {
      message.error('加载数据失败')
    } finally {
      setLoading(false)
    }
  }

  const handleCreateVaccine = async (values) => {
    try {
      const data = {
        ...values,
        expiryDate: values.expiryDate?.format('YYYY-MM-DD')
      }
      await vaccineApi.createVaccine(data)
      message.success('疫苗创建成功')
      setVaccineModalVisible(false)
      vaccineForm.resetFields()
      loadData()
    } catch (err) {
      message.error(err.response?.data?.error || '创建失败')
    }
  }

  const handleRecordVaccination = async (values) => {
    try {
      const data = {
        ...values,
        vaccinationDate: values.vaccinationDate.format('YYYY-MM-DD')
      }
      await vaccineApi.recordVaccination(data)
      message.success('接种记录成功')
      setRecordModalVisible(false)
      recordForm.resetFields()
      loadData()
    } catch (err) {
      if (err.response?.status === 409) {
        message.error('该剂次已接种')
      } else if (err.response?.status === 400) {
        message.error(err.response?.data?.error || '超出允许时间范围')
      } else {
        message.error(err.response?.data?.error || '记录失败')
      }
    }
  }

  const vaccineColumns = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
    { title: '疫苗名称', dataIndex: 'name', key: 'name' },
    { title: '生产厂家', dataIndex: 'manufacturer', key: 'manufacturer' },
    { title: '批号', dataIndex: 'batchNumber', key: 'batchNumber' },
    {
      title: '有效期',
      dataIndex: 'expiryDate',
      key: 'expiryDate',
      render: (t) => t ? dayjs(t).format('YYYY-MM-DD') : '-'
    },
    { title: '接种程序', dataIndex: 'schedule', key: 'schedule' },
    { title: '剂次', dataIndex: 'doses', key: 'doses' },
  ]

  const recordColumns = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
    { title: '接种者', dataIndex: 'recipientName', key: 'recipientName' },
    { title: '接种者ID', dataIndex: 'recipientId', key: 'recipientId' },
    { title: '疫苗名称', dataIndex: 'vaccineName', key: 'vaccineName' },
    { title: '剂次', dataIndex: 'doseNumber', key: 'doseNumber' },
    {
      title: '接种日期',
      dataIndex: 'vaccinationDate',
      key: 'vaccinationDate',
      render: (t) => dayjs(t).format('YYYY-MM-DD')
    },
    { title: '接种单位', dataIndex: 'unit', key: 'unit' },
    { title: '接种医生', dataIndex: 'doctor', key: 'doctor' },
  ]

  return (
    <div>
      <Tabs defaultActiveKey="1">
        <TabPane tab="疫苗信息" key="1">
          <Space style={{ marginBottom: 16 }}>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => setVaccineModalVisible(true)}
            >
              添加疫苗
            </Button>
            <Button onClick={loadData}>刷新</Button>
          </Space>

          <Card>
            <Table
              columns={vaccineColumns}
              dataSource={vaccines}
              rowKey="id"
              loading={loading}
            />
          </Card>
        </TabPane>

        <TabPane tab="接种记录" key="2">
          <Space style={{ marginBottom: 16 }}>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => setRecordModalVisible(true)}
            >
              记录接种
            </Button>
            <Button onClick={loadData}>刷新</Button>
          </Space>

          <Card>
            <Table
              columns={recordColumns}
              dataSource={records}
              rowKey="id"
              loading={loading}
            />
          </Card>
        </TabPane>
      </Tabs>

      <Modal
        title="添加疫苗"
        open={vaccineModalVisible}
        onCancel={() => setVaccineModalVisible(false)}
        onOk={() => vaccineForm.submit()}
        okText="提交"
        cancelText="取消"
      >
        <Form form={vaccineForm} layout="vertical" onFinish={handleCreateVaccine}>
          <Form.Item name="name" label="疫苗名称" rules={[{ required: true }]}>
            <Input placeholder="请输入疫苗名称" />
          </Form.Item>
          <Form.Item name="manufacturer" label="生产厂家">
            <Input placeholder="请输入生产厂家" />
          </Form.Item>
          <Form.Item name="batchNumber" label="批号">
            <Input placeholder="请输入批号" />
          </Form.Item>
          <Form.Item name="expiryDate" label="有效期">
            <DatePicker style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="schedule" label="接种程序">
            <Input placeholder="如：0-1-6" />
          </Form.Item>
          <Form.Item name="doses" label="剂次">
            <InputNumber min={1} style={{ width: '100%' }} placeholder="请输入剂次" />
          </Form.Item>
          <Form.Item name="intervalDays" label="间隔天数">
            <Input placeholder="如：0,30,180" />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="记录接种"
        open={recordModalVisible}
        onCancel={() => setRecordModalVisible(false)}
        onOk={() => recordForm.submit()}
        okText="提交"
        cancelText="取消"
      >
        <Form form={recordForm} layout="vertical" onFinish={handleRecordVaccination}>
          <Form.Item name="recipientName" label="接种者姓名" rules={[{ required: true }]}>
            <Input placeholder="请输入接种者姓名" />
          </Form.Item>
          <Form.Item name="recipientId" label="接种者ID" rules={[{ required: true }]}>
            <Input placeholder="请输入接种者ID" />
          </Form.Item>
          <Form.Item name="vaccineName" label="疫苗名称" rules={[{ required: true }]}>
            <Select placeholder="请选择疫苗">
              {vaccines.map(v => (
                <Option key={v.id} value={v.name}>
                  {v.name} ({v.schedule})
                </Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item name="doseNumber" label="剂次" rules={[{ required: true }]}>
            <InputNumber min={1} style={{ width: '100%' }} placeholder="请输入剂次" />
          </Form.Item>
          <Form.Item
            name="vaccinationDate"
            label="接种日期"
            rules={[{ required: true }]}
          >
            <DatePicker style={{ width: '100%' }} defaultValue={dayjs()} />
          </Form.Item>
          <Form.Item name="unit" label="接种单位">
            <Input placeholder="请输入接种单位" />
          </Form.Item>
          <Form.Item name="doctor" label="接种医生">
            <Input placeholder="请输入接种医生" />
          </Form.Item>
          <Form.Item>
            <Tag color="orange">注意：系统会自动校验时间窗口和重复接种</Tag>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default VaccinePage
