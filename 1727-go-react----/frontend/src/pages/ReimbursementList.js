import React, { useState, useEffect } from 'react';
import { Table, Card, Tag, Modal, Form, Input, message, Button } from 'antd';
import { reimbursementAPI, projectAPI, fenToYuan, CATEGORIES } from '../api';

const STATUS_MAP = {
  pending: { text: '待审批', color: 'gold' },
  approved: { text: '已批准', color: 'green' },
  rejected: { text: '已驳回', color: 'red' },
};

export default function ReimbursementList() {
  const [data, setData] = useState([]);
  const [projects, setProjects] = useState({});
  const [loading, setLoading] = useState(false);
  const [resubmitModal, setResubmitModal] = useState(false);
  const [currentItem, setCurrentItem] = useState(null);
  const [form] = Form.useForm();

  const loadData = async () => {
    setLoading(true);
    try {
      const [reimbs, projList] = await Promise.all([
        reimbursementAPI.list(),
        projectAPI.list(),
      ]);
      setData(reimbs);
      const projMap = {};
      projList.forEach(p => (projMap[p.id] = p));
      setProjects(projMap);
    } catch (e) {
      message.error(e.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const handleResubmit = async (values) => {
    try {
      const payload = {};
      if (values.amount !== undefined) {
        payload.amount = values.amount.toString();
      }
      if (values.reason) payload.reason = values.reason;
      if (values.date) payload.date = values.date;
      await reimbursementAPI.resubmit(currentItem.id, payload);
      message.success('重新提交成功');
      setResubmitModal(false);
      form.resetFields();
      loadData();
    } catch (e) {
      message.error(e.message);
    }
  };

  const openResubmit = (item) => {
    setCurrentItem(item);
    form.setFieldsValue({
      amount: parseFloat(fenToYuan(item.amount)),
      reason: item.reason,
      date: item.date,
    });
    setResubmitModal(true);
  };

  const columns = [
    { title: '报销单编号', dataIndex: 'id', key: 'id', width: 180 },
    {
      title: '所属项目',
      key: 'project',
      render: (_, r) => projects[r.project_id]?.name || r.project_id,
    },
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
    {
      title: '操作',
      key: 'action',
      render: (_, r) =>
        r.status === 'rejected' ? (
          <Button type="link" onClick={() => openResubmit(r)}>重新提交</Button>
        ) : null,
    },
  ];

  return (
    <Card title="报销单列表" loading={loading}>
      <Table
        columns={columns}
        dataSource={data}
        rowKey="id"
        pagination={{ pageSize: 10 }}
      />

      <Modal
        title="重新提交报销单"
        open={resubmitModal}
        onCancel={() => setResubmitModal(false)}
        footer={null}
      >
        <Form form={form} layout="vertical" onFinish={handleResubmit}>
          <Form.Item
            name="amount"
            label="金额（元，精确到分）"
            rules={[{ type: 'number', min: 0.01, message: '请输入有效金额' }]}
          >
            <InputNumber step="0.01" style={{ width: '100%' }} placeholder="请输入金额" />
          </Form.Item>
          <Form.Item name="reason" label="事由">
            <Input.TextArea rows={3} placeholder="请输入报销事由" />
          </Form.Item>
          <Form.Item name="date" label="日期">
            <Input placeholder="YYYY-MM-DD" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block>
              重新提交
            </Button>
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
}
