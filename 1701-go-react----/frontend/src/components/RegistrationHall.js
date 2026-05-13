import React, { useState, useEffect, useRef } from 'react';
import {
  Table,
  Button,
  Modal,
  Form,
  Input,
  Alert,
  Tag,
  Card,
  message,
  Badge,
  Space,
} from 'antd';
import { PlusOutlined, AudioOutlined } from '@ant-design/icons';
import { getRegistrations, registerPatient } from '../api';
import dayjs from 'dayjs';

const statusMap = {
  waiting: { text: '候诊中', color: 'blue' },
  in_visit: { text: '就诊中', color: 'orange' },
  done: { text: '已完成', color: 'green' },
};

function RegistrationHall() {
  const [registrations, setRegistrations] = useState([]);
  const [modalVisible, setModalVisible] = useState(false);
  const [form] = Form.useForm();
  const [callInfo, setCallInfo] = useState(null);
  const wsRef = useRef(null);

  const fetchData = async () => {
    try {
      const response = await getRegistrations();
      setRegistrations(response.data.registrations);
    } catch (error) {
      message.error('获取挂号列表失败');
    }
  };

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 5000);

    const ws = new WebSocket('ws://localhost:8300/ws');
    wsRef.current = ws;

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === 'call_number') {
        setCallInfo(data);
        message.info(`正在叫号: ${data.serial_num} - ${data.doctor_name}`);
        fetchData();
      }
    };

    return () => {
      clearInterval(interval);
      ws.close();
    };
  }, []);

  const waitingCount = registrations.filter((r) => r.status === 'waiting').length;
  const estimatedWaitTime = waitingCount * 15;

  const handleSubmit = async (values) => {
    try {
      await registerPatient(values);
      message.success('挂号成功');
      setModalVisible(false);
      form.resetFields();
      fetchData();
    } catch (error) {
      message.error(error.response?.data?.error || '挂号失败');
    }
  };

  const columns = [
    {
      title: '挂号序号',
      dataIndex: 'serial_num',
      key: 'serial_num',
      render: (text) => <strong>{text}</strong>,
    },
    {
      title: '患者姓名',
      dataIndex: 'patient_name',
      key: 'patient_name',
    },
    {
      title: '手机号',
      dataIndex: 'patient_phone',
      key: 'patient_phone',
    },
    {
      title: '主诉症状',
      dataIndex: 'chief_complaint',
      key: 'chief_complaint',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status) => {
        const info = statusMap[status] || { text: status, color: 'default' };
        return <Tag color={info.color}>{info.text}</Tag>;
      },
    },
    {
      title: '挂号时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (time) => dayjs(time).format('HH:mm:ss'),
    },
  ];

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <h2 style={{ margin: 0 }}>挂号大厅</h2>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => setModalVisible(true)}
        >
          新增挂号
        </Button>
      </div>

      {waitingCount > 10 && (
        <Alert
          message={
            <span>
              <strong style={{ color: 'red' }}>
                当前候诊人数较多，预计等待时间约{estimatedWaitTime}分钟
              </strong>
            </span>
          }
          type="warning"
          showIcon
          style={{ marginBottom: 16, backgroundColor: '#fff1f0', borderColor: '#ffa39e' }}
        />
      )}

      {callInfo && (
        <Card
          style={{ marginBottom: 16, backgroundColor: '#e6f7ff', borderColor: '#91d5ff' }}
          extra={
            <Badge status="processing" text="正在叫号" />
          }
        >
          <Space size="large">
            <AudioOutlined style={{ fontSize: 24, color: '#1890ff' }} />
            <div>
              <div style={{ fontSize: 24, fontWeight: 'bold', color: '#1890ff' }}>
                {callInfo.serial_num}
              </div>
              <div style={{ color: '#666' }}>{callInfo.doctor_name}</div>
            </div>
          </Space>
        </Card>
      )}

      <Table
        dataSource={registrations}
        columns={columns}
        rowKey="id"
        pagination={{ pageSize: 10 }}
      />

      <Modal
        title="新增挂号"
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => form.submit()}
        okText="确认挂号"
        cancelText="取消"
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item
            name="name"
            label="患者姓名"
            rules={[{ required: true, message: '请输入患者姓名' }]}
          >
            <Input placeholder="请输入患者姓名" />
          </Form.Item>
          <Form.Item
            name="phone"
            label="手机号"
            rules={[
              { required: true, message: '请输入手机号' },
              { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号' },
            ]}
          >
            <Input placeholder="请输入手机号" />
          </Form.Item>
          <Form.Item
            name="chief_complaint"
            label="主诉症状"
            rules={[{ required: true, message: '请输入主诉症状' }]}
          >
            <Input.TextArea rows={3} placeholder="请输入主诉症状" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}

export default RegistrationHall;
