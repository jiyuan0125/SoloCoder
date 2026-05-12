import React, { useState } from 'react';
import { 
  Card, 
  Form, 
  Input, 
  Button, 
  Table, 
  Tag, 
  Typography, 
  message, 
  Space,
  Alert,
  Divider
} from 'antd';
import { SearchOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { Diploma, EducationLevel, EducationStatus } from '../types';
import { verifyAPI } from '../utils/api';
import { EDUCATION_LEVEL_MAP, EDUCATION_STATUS_MAP } from '../utils/constants';

const { Title } = Typography;

const Verify: React.FC = () => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<{
    diplomas: Diploma[];
    result: string;
  } | null>(null);

  const handleVerify = async (values: { name: string; id_card: string }) => {
    try {
      setLoading(true);
      const data = await verifyAPI.verify(values);
      setResult(data);
    } catch (error: any) {
      message.error(error.response?.data?.message || '查询失败');
      setResult(null);
    } finally {
      setLoading(false);
    }
  };

  const columns: ColumnsType<Diploma> = [
    {
      title: '姓名',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '身份证号',
      dataIndex: 'id_card',
      key: 'id_card',
    },
    {
      title: '毕业院校',
      dataIndex: 'school',
      key: 'school',
    },
    {
      title: '学历层次',
      dataIndex: 'level',
      key: 'level',
      render: (level: EducationLevel) => EDUCATION_LEVEL_MAP[level],
    },
    {
      title: '专业',
      dataIndex: 'major',
      key: 'major',
    },
    {
      title: '学制年限',
      dataIndex: 'study_years',
      key: 'study_years',
      render: (years: number) => `${years}年`,
    },
    {
      title: '入学日期',
      dataIndex: 'enrollment_date',
      key: 'enrollment_date',
    },
    {
      title: '毕业日期',
      dataIndex: 'graduation_date',
      key: 'graduation_date',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: EducationStatus) => (
        <Tag color={status === 'valid' ? 'green' : 'red'}>
          {EDUCATION_STATUS_MAP[status]}
        </Tag>
      ),
    },
  ];

  return (
    <div>
      <Title level={4}>学历认证查询</Title>
      
      <Card>
        <Form
          form={form}
          layout="inline"
          onFinish={handleVerify}
          style={{ justifyContent: 'center' }}
        >
          <Form.Item
            name="name"
            label="姓名"
            rules={[{ required: true, message: '请输入姓名' }]}
          >
            <Input placeholder="请输入姓名" style={{ width: 200 }} />
          </Form.Item>
          
          <Form.Item
            name="id_card"
            label="身份证号"
            rules={[
              { required: true, message: '请输入身份证号' },
              { 
                pattern: /^[0-9]{17}[0-9X]$/, 
                message: '身份证号格式不正确' 
              }
            ]}
          >
            <Input 
              placeholder="18位身份证号" 
              style={{ width: 220 }} 
              maxLength={18}
            />
          </Form.Item>
          
          <Form.Item>
            <Button 
              type="primary" 
              htmlType="submit" 
              loading={loading}
              icon={<SearchOutlined />}
            >
              查询认证
            </Button>
          </Form.Item>
        </Form>
      </Card>

      {result && (
        <div style={{ marginTop: 24 }}>
          <Divider />
          
          {result.diplomas.length > 0 ? (
            <div>
              <Alert
                message="认证结果"
                description="查询成功，学历信息一致"
                type="success"
                showIcon
                style={{ marginBottom: 16 }}
              />
              
              <Table
                columns={columns}
                dataSource={result.diplomas}
                rowKey="id"
                pagination={false}
                bordered
              />
            </div>
          ) : (
            <Alert
              message="认证结果"
              description="未查到相关学历信息"
              type="warning"
              showIcon
            />
          )}
        </div>
      )}
    </div>
  );
};

export default Verify;
