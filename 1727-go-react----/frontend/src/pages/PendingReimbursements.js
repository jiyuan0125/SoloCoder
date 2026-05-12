import React, { useState, useEffect } from 'react';
import { Table, Card, Tag, Modal, Form, Input, message, Button, Space } from 'antd';
import { reimbursementAPI, projectAPI, fenToYuan, CATEGORIES } from '../api';

export default function PendingReimbursements() {
  const [data, setData] = useState([]);
  const [projects, setProjects] = useState({});
  const [loading, setLoading] = useState(false);
  const [rejectModal, setRejectModal] = useState(false);
  const [currentItem, setCurrentItem] = useState(null);
  const [form] = Form.useForm();

  const loadData = async () => {
    setLoading(true);
    try {
      const [reimbs, projList] = await Promise.all([
        reimbursementAPI.pending(),
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

  const handleApprove = async (item) => {
    Modal.confirm({
      title: '确认批准',
      content: '批准后该报销单不可修改，是否继续？',
      onOk: async () => {
        try {
          await reimbursementAPI.approve(item.id);
          message.success('审批通过');
          loadData();
        } catch (e) {
          message.error(e.message);
        }
      },
    });
  };

  const handleReject = (item) => {
    setCurrentItem(item);
    form.resetFields();
    setRejectModal(true);
  };

  const submitReject = async (values) => {
    try {
      await reimbursementAPI.reject(currentItem.id, values.reason);
      message.success('已驳回');
      setRejectModal(false);
      loadData();
    } catch (e) {
      message.error(e.message);
    }
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
      title: '操作',
      key: 'action',
      render: (_, r) => (
        <Space>
          <Button type="primary" size="small" onClick={() => handleApprove(r)}>
            批准
          </Button>
          <Button danger size="small" onClick={() => handleReject(r)}>
            驳回
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Card title="待办审批（待审批报销单）" loading={loading}>
        {data.length === 0 ? (
          <div style={{ textAlign: 'center', padding: 40, color: '#888' }}>
            暂无待审批的报销单
          </div>
        ) : (
          <Table
            columns={columns}
            dataSource={data}
            rowKey="id"
            pagination={false}
          />
        )}
      </Card>

      <Modal
        title="驳回报销单"
        open={rejectModal}
        onCancel={() => setRejectModal(false)}
        footer={null}
      >
        <Form form={form} layout="vertical" onFinish={submitReject}>
          <Form.Item
            name="reason"
            label="驳回理由"
            rules={[{ required: true, message: '请填写驳回理由' }]}
          >
            <Input.TextArea rows={4} placeholder="请输入驳回理由" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" danger block>
              确认驳回
            </Button>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
