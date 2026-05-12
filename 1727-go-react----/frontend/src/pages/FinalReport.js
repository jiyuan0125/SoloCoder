import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Descriptions, Table, Statistic, Row, Col, Button, Alert, Tag } from 'antd';
import { ArrowLeftOutlined, WarningOutlined } from '@ant-design/icons';
import { projectAPI, fenToYuan } from '../api';

export default function FinalReport() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [report, setReport] = useState(null);

  useEffect(() => {
    const load = async () => {
      setLoading(true);
      try {
        const data = await projectAPI.finalReport(id);
        setReport(data);
      } catch (e) {
        alert(e.message);
      } finally {
        setLoading(false);
      }
    };
    load();
  }, [id]);

  if (!report) {
    return <div>加载中...</div>;
  }

  const columns = [
    { title: '费用类别', dataIndex: 'category_name', key: 'category_name' },
    { title: '预算金额', dataIndex: 'budget', key: 'budget', render: v => `¥${fenToYuan(v)}` },
    { title: '实际报销', dataIndex: 'spent', key: 'spent', render: v => `¥${fenToYuan(v)}` },
    { title: '执行率', dataIndex: 'execution_rate', key: 'execution_rate', render: v => `${v.toFixed(2)}%` },
    {
      title: '情况',
      key: 'status',
      render: (_, r) => {
        if (r.budget === 0 && r.spent === 0) return <Tag color="gray">无预算</Tag>;
        if (r.execution_rate < 30) return <Tag color="red">执行率偏低</Tag>;
        return <Tag color="green">正常</Tag>;
      },
    },
  ];

  return (
    <div>
      <Button
        icon={<ArrowLeftOutlined />}
        onClick={() => navigate(-1)}
        style={{ marginBottom: 16 }}
      >
        返回
      </Button>

      <Card title="决算报告" loading={loading}>
        {report.is_abnormal && (
          <Alert
            message="预算执行异常"
            description={`整体执行率为 ${report.execution_rate.toFixed(2)}%，低于30%，请关注预算执行情况。`}
            type="warning"
            showIcon
            icon={<WarningOutlined />}
            style={{ marginBottom: 24 }}
          />
        )}

        <Descriptions column={2} bordered style={{ marginBottom: 24 }}>
          <Descriptions.Item label="项目编号">{report.project_id}</Descriptions.Item>
          <Descriptions.Item label="项目名称">{report.project_name}</Descriptions.Item>
          <Descriptions.Item label="报告生成时间">
            {new Date(report.generated_at).toLocaleString()}
          </Descriptions.Item>
          <Descriptions.Item label="执行状态">
            {report.is_abnormal ? (
              <Tag color="red" icon={<WarningOutlined />}>预算执行异常</Tag>
            ) : (
              <Tag color="green">正常</Tag>
            )}
          </Descriptions.Item>
        </Descriptions>

        <Row gutter={16} style={{ marginBottom: 24 }}>
          <Col span={8}>
            <Statistic title="预算总额" value={fenToYuan(report.total_budget)} prefix="¥" />
          </Col>
          <Col span={8}>
            <Statistic
              title="实际报销总额"
              value={fenToYuan(report.total_spent)}
              prefix="¥"
              valueStyle={{ color: '#3f8600' }}
            />
          </Col>
          <Col span={8}>
            <Statistic
              title="整体执行率"
              value={report.execution_rate.toFixed(2)}
              suffix="%"
              valueStyle={{ color: report.is_abnormal ? '#cf1322' : '#3f8600' }}
            />
          </Col>
        </Row>

        <Card title="各类别执行明细" size="small">
          <Table
            columns={columns}
            dataSource={report.details}
            rowKey="category"
            pagination={false}
          />
        </Card>
      </Card>
    </div>
  );
}
