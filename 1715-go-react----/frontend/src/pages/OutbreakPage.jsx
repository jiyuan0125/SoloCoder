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
  Card
} from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'

import { outbreakApi, statsApi } from '../api'

const { Option } = Select
const { RangePicker } = DatePicker

function OutbreakPage() {
  const [reports, setReports] = useState([])
  const [diseases, setDiseases] = useState([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [form] = Form.useForm()

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setLoading(true)
    try {
      const [reportRes, diseaseRes] = await Promise.all([
        outbreakApi.listReports(),
        statsApi.listDiseases()
      ])
      setReports(reportRes.data)
      setDiseases(diseaseRes.data)
    } catch (err) {
      message.error('加载数据失败')
    } finally {
      setLoading(false)
    }
  }

  const handleSubmit = async (values) => {
    try {
      const data = {
        ...values,
        onsetTime: values.onsetTime.format('YYYY-MM-DDTHH:mm:ss'),
        reportTime: values.reportTime.format('YYYY-MM-DDTHH:mm:ss')
      }
      await outbreakApi.createReport(data)
      message.success('报告提交成功')
      setModalVisible(false)
      form.resetFields()
      loadData()
    } catch (err) {
      message.error(err.response?.data?.error || '提交失败')
    }
  }

  const columns = [
    { title: '患者姓名', dataIndex: 'patientName', key: 'patientName' },
    { title: '患者ID', dataIndex: 'patientId', key: 'patientId' },
    { title: '传染病', dataIndex: 'diseaseName', key: 'diseaseName' },
    { title: '地区', dataIndex: 'region', key: 'region' },
    { title: '区县', dataIndex: 'district', key: 'district' },
    {
      title: '发病时间',
      dataIndex: 'onsetTime',
      key: 'onsetTime',
      render: (t) => dayjs(t).format('YYYY-MM-DD HH:mm')
    },
    {
      title: '报告时间',
      dataIndex: 'reportTime',
      key: 'reportTime',
      render: (t) => dayjs(t).format('YYYY-MM-DD HH:mm')
    },
    {
      title: '状态',
      dataIndex: 'isDelayed',
      key: 'isDelayed',
      render: (delayed) =>
        delayed ? <Tag color="red">迟报</Tag> : <Tag color="green">及时</Tag>
    },
    {
      title: '聚集性',
      dataIndex: 'hasClustering',
      key: 'hasClustering',
      render: (clustering) =>
        clustering ? <Tag color="orange">是</Tag> : <Tag>否</Tag>
    },
  ]

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalVisible(true)}>
          新建疫情报告
        </Button>
        <Button onClick={loadData}>刷新</Button>
      </Space>

      <Card title="疫情报告列表">
        <Table
          columns={columns}
          dataSource={reports}
          rowKey="id"
          loading={loading}
        />
      </Card>

      <Modal
        title="新建疫情报告"
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => form.submit()}
        okText="提交"
        cancelText="取消"
        width={600}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item name="patientName" label="患者姓名" rules={[{ required: true }]}>
            <Input placeholder="请输入患者姓名" />
          </Form.Item>
          <Form.Item name="patientId" label="患者ID" rules={[{ required: true }]}>
            <Input placeholder="请输入患者ID" />
          </Form.Item>
          <Form.Item name="diseaseName" label="传染病名称" rules={[{ required: true }]}>
            <Select placeholder="请选择传染病">
              {diseases.map(d => (
                <Option key={d.id} value={d.name}>
                  {d.name} ({d.class})
                </Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item name="region" label="地区" rules={[{ required: true }]}>
            <Input placeholder="请输入地区（如：北京市）" />
          </Form.Item>
          <Form.Item name="district" label="区县" rules={[{ required: true }]}>
            <Input placeholder="请输入区县（如：朝阳区）" />
          </Form.Item>
          <Form.Item name="onsetTime" label="发病时间" rules={[{ required: true }]}>
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="reportTime" label="报告时间" rules={[{ required: true }]}>
            <DatePicker showTime style={{ width: '100%' }} defaultValue={dayjs()} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default OutbreakPage
