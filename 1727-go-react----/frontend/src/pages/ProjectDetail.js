import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Descriptions, Table, Button, Modal, Form, Input, DatePicker, InputNumber, message, Tag, Space, Statistic, Row, Col } from 'antd';
import { ArrowLeftOutlined, PlusOutlined } from '@ant-design/icons';
import { PieChart, Pie, Cell, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import dayjs from 'dayjs';
import { projectAPI, reimbursementAPI, fenToYuan, CATEGORIES, CATEGORY_COLORS } from '../api';

const STATUS_MAP = {
  pending: { text: '待审批', color: 'gold' },
  approved: { text: '已批准', color: 'green' },
  rejected: { text: '已驳回', color: 'red' },
};

export default function ProjectDetail() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState(null);
  const [reimbModalVisible, setReimbModalVisible] = useState(false);
  const [form] = Form.useForm();

  const loadData = async () => {
    setLoading(true);
    try {
      const res = await projectAPI.get(id);
      setData(res);
    } catch (e) {
      message.error(e.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [id]);

  const handleReimbSubmit = async (values) => {
    try {
      await reimbursementAPI.create({
        project_id: id,
        category: values.category,
        amount: values.amount.toString(),
        reason: values.reason,
        date: values.date.format('YYYY-MM-DD'),
      });
      message.success('报销单提交成功');
      setReimbModalVisible(false);
      form.resetFields();
      loadData();
    } catch (e) {
      message.error(e.message);
    }
  };

  if (!data) {
    return <div>加载中...</div>;
  }

  const { project, budget_details, reimbursements } = data;

  const pieData = CATEGORIES.map(cat => ({
    name: cat.name,
    value: project.budget[cat.key] || 0,
    key: cat.key,
  })).filter(d => d.value > 0);

  const reimbColumns = [
    { title: '报销单编号', dataIndex: 'id', key: 'id', width: 180 },
    {
      title: '费用类别',
      dataIndex: 'category',
      key: 'category',
      render: (c) => CATEGORIES.find(x => x.key === c)?.name || c,
    },
    { title: '金额', key: 'amount', render: (_, r) => `¥${fenToYuan(r.amount)}` },
    { title: '事由', dataIndex: 'reason', key: 'reason' },
    { title: '日期', dataIndex: 'date', key: 'date' },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (s) => <Tag color={STATUS_MAP[s]?.color}>{STATUS_MAP[s]?.text}</Tag>,
    },
    {
      title: '驳回理由',
      dataIndex: 'reject_reason',
      key: 'reject_reason',
      render: (r) => r || '-',
    },
  ];

  let totalSpent = 0;
  for (const key in budget_details.spent) {
    totalSpent += budget_details.spent[key];
  }

  return (
    <div>
      <Button
        icon={<ArrowLeftOutlined />}
        onClick={() => navigate(-1)}
        style={{ marginBottom: 16 }}
      >
        返回
      </Button>

      <Card title="项目基本信息" loading={loading}>
        <Descriptions column={2}>
          <Descriptions.Item label="项目编号">{project.id}</Descriptions.Item>
          <Descriptions.Item label="项目名称">{project.name}</Descriptions.Item>
          <Descriptions.Item label="负责人">{project.principal}</Descriptions.Item>
          <Descriptions.Item label="状态">
            <Tag color={project.status === 'active' ? 'green' : 'gray'}>
              {project.status === 'active' ? '进行中' : '已结题'}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="开始日期">{project.start_date}</Descriptions.Item>
          <Descriptions.Item label="结束日期">{project.end_date}</Descriptions.Item>
        </Descriptions>
      </Card>

      <Card title="预算执行情况" style={{ marginTop: 16 }}>
        <Row gutter={16}>
          <Col span={8}>
            <Statistic title="经费总额" value={fenToYuan(project.total_amount)} prefix="¥" />
          </Col>
          <Col span={8}>
            <Statistic title="已报销金额" value={fenToYuan(totalSpent)} prefix="¥" valueStyle={{ color: '#3f8600' }} />
          </Col>
          <Col span={8}>
            <Statistic title="剩余预算" value={fenToYuan(project.total_amount - totalSpent)} prefix="¥" valueStyle={{ color: '#1890ff' }} />
          </Col>
        </Row>
      </Card>

      <Row gutter={16} style={{ marginTop: 16 }}>
        <Col span={12}>
          <Card title="预算分布饼图">
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={pieData}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                  outerRadius={100}
                  fill="#8884d8"
                  dataKey="value"
                >
                  {pieData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={CATEGORY_COLORS[entry.key]} />
                  ))}
                </Pie>
                <Tooltip formatter={(value) => `¥${fenToYuan(value)}`} />
                <Legend />
              </PieChart>
            </ResponsiveContainer>
          </Card>
        </Col>
        <Col span={12}>
          <Card title="预算明细">
            <Table
              dataSource={CATEGORIES.map(cat => ({
                key: cat.key,
                name: cat.name,
                budget: project.budget[cat.key] || 0,
                spent: budget_details.spent[cat.key] || 0,
                remaining: (project.budget[cat.key] || 0) - (budget_details.spent[cat.key] || 0),
              }))}
              pagination={false}
              size="small"
            >
              <Table.Column title="类别" dataIndex="name" key="name" />
              <Table.Column title="预算" dataIndex="budget" key="budget" render={v => `¥${fenToYuan(v)}`} />
              <Table.Column title="已报销" dataIndex="spent" key="spent" render={v => `¥${fenToYuan(v)}`} />
              <Table.Column title="剩余" dataIndex="remaining" key="remaining" render={v => `¥${fenToYuan(v)}`} />
            </Table>
          </Card>
        </Col>
      </Row>

      <Card
        title="报销单列表"
        style={{ marginTop: 16 }}
        extra={
          project.status === 'active' && (
            <Button type="primary" icon={<PlusOutlined />} onClick={() => setReimbModalVisible(true)}>
              提交报销
            </Button>
          )
        }
      >
        <Table
          columns={reimbColumns}
          dataSource={reimbursements}
          rowKey="id"
          pagination={{ pageSize: 10 }}
        />
      </Card>

      <Modal
        title="提交报销单"
        open={reimbModalVisible}
        onCancel={() => setReimbModalVisible(false)}
        footer={null}
      >
        <Form form={form} layout="vertical" onFinish={handleReimbSubmit}>
          <Form.Item name="category" label="费用类别" rules={[{ required: true }]}>
            <Input.Group>
              <select
                style={{ width: '100%', height: 32, padding: '4px 11px', border: '1px solid #d9d9d9', borderRadius: 6 }}
                onChange={(e) => form.setFieldValue('category', e.target.value)}
                value={form.getFieldValue('category')}
              >
                <option value="">请选择类别</option>
                {CATEGORIES.map(cat => (
                  <option key={cat.key} value={cat.key}>{cat.name}</option>
                ))}
              </select>
            </Input.Group>
          </Form.Item>
          <Form.Item
            name="amount"
            label="金额（元，精确到分）"
            rules={[{ required: true, type: 'number', min: 0.01, message: '请输入有效金额' }]}
          >
            <InputNumber step="0.01" style={{ width: '100%' }} placeholder="请输入金额" />
          </Form.Item>
          <Form.Item name="reason" label="事由" rules={[{ required: true }]}>
            <Input.TextArea rows={3} placeholder="请输入报销事由" />
          </Form.Item>
          <Form.Item name="date" label="日期" rules={[{ required: true }]}>
            <DatePicker style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block>
              提交报销
            </Button>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
