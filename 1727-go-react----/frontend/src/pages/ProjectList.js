import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Table, Button, Modal, Form, Input, DatePicker, InputNumber, message, Tag, Space, Card } from 'antd';
import { PlusOutlined, EyeOutlined, StopOutlined, FileTextOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { projectAPI, fenToYuan, CATEGORIES } from '../api';

export default function ProjectList() {
  const [projects, setProjects] = useState([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [form] = Form.useForm();
  const navigate = useNavigate();

  const loadProjects = async () => {
    setLoading(true);
    try {
      const data = await projectAPI.list();
      setProjects(data);
    } catch (e) {
      message.error(e.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadProjects();
  }, []);

  const handleSubmit = async (values) => {
    try {
      const budget = {};
      let totalFen = 0;
      CATEGORIES.forEach(cat => {
        const val = values[cat.key] || 0;
        budget[cat.key] = val.toString();
        const parts = String(val).split('.');
        let intPart = parseInt(parts[0] || '0', 10);
        let decPart = parts[1] ? parseInt(parts[1].padEnd(2, '0').slice(0, 2), 10) : 0;
        totalFen += intPart * 100 + decPart;
      });

      const totalVal = values.total_amount;
      const totalParts = String(totalVal).split('.');
      let totalInt = parseInt(totalParts[0] || '0', 10);
      let totalDec = totalParts[1] ? parseInt(totalParts[1].padEnd(2, '0').slice(0, 2), 10) : 0;
      const expectedTotal = totalInt * 100 + totalDec;

      if (totalFen !== expectedTotal) {
        message.error('预算各类别之和必须精确等于总额，差一分钱都不允许');
        return;
      }

      await projectAPI.create({
        name: values.name,
        principal: values.principal,
        start_date: values.date_range[0].format('YYYY-MM-DD'),
        end_date: values.date_range[1].format('YYYY-MM-DD'),
        total_amount: values.total_amount.toString(),
        budget,
      });
      message.success('项目创建成功');
      setModalVisible(false);
      form.resetFields();
      loadProjects();
    } catch (e) {
      message.error(e.message);
    }
  };

  const handleClose = async (id) => {
    Modal.confirm({
      title: '确认结题',
      content: '结题后无法再提交报销单，是否继续？',
      onOk: async () => {
        try {
          await projectAPI.close(id);
          message.success('项目已结题');
          loadProjects();
        } catch (e) {
          message.error(e.message);
        }
      },
    });
  };

  const columns = [
    { title: '项目编号', dataIndex: 'id', key: 'id', width: 180 },
    { title: '项目名称', dataIndex: 'name', key: 'name' },
    { title: '负责人', dataIndex: 'principal', key: 'principal' },
    { title: '起止日期', key: 'date', render: (_, r) => `${r.start_date} 至 ${r.end_date}` },
    { title: '经费总额', key: 'total', render: (_, r) => `¥${fenToYuan(r.total_amount)}` },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (s) => (
        <Tag color={s === 'active' ? 'green' : 'gray'}>
          {s === 'active' ? '进行中' : '已结题'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      render: (_, r) => (
        <Space>
          <Button type="link" icon={<EyeOutlined />} onClick={() => navigate(`/projects/${r.id}`)}>
            详情
          </Button>
          {r.status === 'active' && (
            <Button type="link" danger icon={<StopOutlined />} onClick={() => handleClose(r.id)}>
              结题
            </Button>
          )}
          {r.status === 'closed' && (
            <Button type="link" icon={<FileTextOutlined />} onClick={() => navigate(`/projects/${r.id}/final-report`)}>
              决算报告
            </Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Card>
        <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
          <h2 style={{ margin: 0 }}>项目列表</h2>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalVisible(true)}>
            新建项目
          </Button>
        </div>
        <Table
          columns={columns}
          dataSource={projects}
          rowKey="id"
          loading={loading}
        />
      </Card>

      <Modal
        title="新建项目"
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
        width={600}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item name="name" label="项目名称" rules={[{ required: true }]}>
            <Input placeholder="请输入项目名称" />
          </Form.Item>
          <Form.Item name="principal" label="负责人" rules={[{ required: true }]}>
            <Input placeholder="请输入负责人姓名" />
          </Form.Item>
          <Form.Item name="date_range" label="起止日期" rules={[{ required: true }]}>
            <DatePicker.RangePicker style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item
            name="total_amount"
            label="经费总额（元，精确到分）"
            rules={[{ required: true, type: 'number', min: 0.01, message: '请输入有效金额' }]}
          >
            <InputNumber step="0.01" style={{ width: '100%' }} placeholder="例如：500000.50" />
          </Form.Item>
          <div style={{ marginBottom: 8, fontWeight: 'bold' }}>预算分配（各类别之和必须精确等于总额）</div>
          {CATEGORIES.map(cat => (
            <Form.Item
              key={cat.key}
              name={cat.key}
              label={cat.name}
              initialValue={0}
              rules={[{ required: true, type: 'number', min: 0, message: '请输入有效金额' }]}
            >
              <InputNumber step="0.01" style={{ width: '100%' }} placeholder="请输入金额" />
            </Form.Item>
          ))}
          <Form.Item>
            <Button type="primary" htmlType="submit" block>
              创建项目
            </Button>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
