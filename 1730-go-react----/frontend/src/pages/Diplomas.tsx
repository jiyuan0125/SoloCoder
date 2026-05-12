import React, { useEffect, useState } from 'react';
import { 
  Table, 
  Button, 
  Modal, 
  Form, 
  Input, 
  Select, 
  DatePicker, 
  InputNumber,
  message,
  Popconfirm,
  Tag,
  Space,
  Typography
} from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import { Diploma, EducationLevel, EducationStatus } from '../types';
import { diplomaAPI } from '../utils/api';
import { EDUCATION_LEVEL_MAP, EDUCATION_STATUS_MAP } from '../utils/constants';

const { Title } = Typography;

const DiplomaForm: React.FC<{
  visible: boolean;
  onCancel: () => void;
  onSubmit: (values: any) => Promise<void>;
  initialValues?: Diploma;
  loading: boolean;
}> = ({ visible, onCancel, onSubmit, initialValues, loading }) => {
  const [form] = Form.useForm();

  useEffect(() => {
    if (visible) {
      if (initialValues) {
        form.setFieldsValue({
          ...initialValues,
          enrollment_date: initialValues.enrollment_date ? dayjs(initialValues.enrollment_date) : undefined,
          graduation_date: initialValues.graduation_date ? dayjs(initialValues.graduation_date) : undefined,
        });
      } else {
        form.resetFields();
      }
    }
  }, [visible, initialValues, form]);

  const handleFinish = async (values: any) => {
    const formattedValues = {
      ...values,
      enrollment_date: values.enrollment_date?.format('YYYY-MM-DD'),
      graduation_date: values.graduation_date?.format('YYYY-MM-DD'),
    };
    await onSubmit(formattedValues);
  };

  return (
    <Modal
      title={initialValues ? '编辑学历信息' : '新增学历信息'}
      open={visible}
      onCancel={onCancel}
      destroyOnClose
      footer={null}
      width={600}
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={handleFinish}
        initialValues={{ status: 'valid' }}
      >
        <Form.Item
          label="姓名"
          name="name"
          rules={[{ required: true, message: '请输入姓名' }]}
        >
          <Input placeholder="请输入姓名" />
        </Form.Item>

        <Form.Item
          label="身份证号"
          name="id_card"
          rules={[
            { required: true, message: '请输入身份证号' },
            { 
              pattern: /^[0-9]{17}[0-9X]$/, 
              message: '身份证号格式不正确' 
            }
          ]}
          extra="18位，前17位数字，最后一位可以是数字或X"
        >
          <Input placeholder="请输入身份证号" maxLength={18} />
        </Form.Item>

        <Form.Item
          label="毕业院校"
          name="school"
          rules={[{ required: true, message: '请输入毕业院校' }]}
        >
          <Input placeholder="请输入毕业院校" />
        </Form.Item>

        <Form.Item
          label="学历层次"
          name="level"
          rules={[{ required: true, message: '请选择学历层次' }]}
        >
          <Select placeholder="请选择学历层次">
            {Object.entries(EDUCATION_LEVEL_MAP).map(([key, label]) => (
              <Select.Option key={key} value={key}>{label}</Select.Option>
            ))}
          </Select>
        </Form.Item>

        <Form.Item
          label="专业"
          name="major"
          rules={[{ required: true, message: '请输入专业' }]}
        >
          <Input placeholder="请输入专业" />
        </Form.Item>

        <Form.Item
          label="学制年限"
          name="study_years"
          rules={[
            { required: true, message: '请输入学制年限' },
            { type: 'number', min: 1, message: '学制年限必须大于0' }
          ]}
        >
          <InputNumber min={1} placeholder="请输入学制年限" style={{ width: '100%' }} />
        </Form.Item>

        <Form.Item
          label="入学日期"
          name="enrollment_date"
          rules={[{ required: true, message: '请选择入学日期' }]}
        >
          <DatePicker style={{ width: '100%' }} />
        </Form.Item>

        <Form.Item
          label="毕业日期"
          name="graduation_date"
          rules={[{ required: true, message: '请选择毕业日期' }]}
        >
          <DatePicker style={{ width: '100%' }} />
        </Form.Item>

        {initialValues && (
          <Form.Item
            label="学历状态"
            name="status"
            rules={[{ required: true, message: '请选择学历状态' }]}
          >
            <Select placeholder="请选择学历状态">
              {Object.entries(EDUCATION_STATUS_MAP).map(([key, label]) => (
                <Select.Option key={key} value={key}>{label}</Select.Option>
              ))}
            </Select>
          </Form.Item>
        )}

        <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
          <Space>
            <Button onClick={onCancel}>取消</Button>
            <Button type="primary" htmlType="submit" loading={loading}>
              提交
            </Button>
          </Space>
        </Form.Item>
      </Form>
    </Modal>
  );
};

const Diplomas: React.FC = () => {
  const [diplomas, setDiplomas] = useState<Diploma[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingDiploma, setEditingDiploma] = useState<Diploma | undefined>();
  const [submitLoading, setSubmitLoading] = useState(false);

  const fetchDiplomas = async () => {
    try {
      setLoading(true);
      const data = await diplomaAPI.list();
      setDiplomas(data);
    } catch (error: any) {
      message.error(error.response?.data?.message || '获取学历信息失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDiplomas();
  }, []);

  const handleCreate = async (values: any) => {
    try {
      setSubmitLoading(true);
      await diplomaAPI.create(values);
      message.success('创建成功');
      setModalVisible(false);
      fetchDiplomas();
    } catch (error: any) {
      message.error(error.response?.data?.message || '创建失败');
    } finally {
      setSubmitLoading(false);
    }
  };

  const handleUpdate = async (values: any) => {
    if (!editingDiploma) return;
    
    try {
      setSubmitLoading(true);
      await diplomaAPI.update(editingDiploma.id, values);
      message.success('更新成功');
      setModalVisible(false);
      setEditingDiploma(undefined);
      fetchDiplomas();
    } catch (error: any) {
      message.error(error.response?.data?.message || '更新失败');
    } finally {
      setSubmitLoading(false);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await diplomaAPI.delete(id);
      message.success('删除成功');
      fetchDiplomas();
    } catch (error: any) {
      message.error(error.response?.data?.message || '删除失败');
    }
  };

  const columns: ColumnsType<Diploma> = [
    {
      title: '姓名',
      dataIndex: 'name',
      key: 'name',
      width: 100,
    },
    {
      title: '身份证号',
      dataIndex: 'id_card',
      key: 'id_card',
      width: 200,
    },
    {
      title: '毕业院校',
      dataIndex: 'school',
      key: 'school',
      width: 150,
    },
    {
      title: '学历层次',
      dataIndex: 'level',
      key: 'level',
      width: 100,
      render: (level: EducationLevel) => EDUCATION_LEVEL_MAP[level],
    },
    {
      title: '专业',
      dataIndex: 'major',
      key: 'major',
      width: 120,
    },
    {
      title: '学制年限',
      dataIndex: 'study_years',
      key: 'study_years',
      width: 80,
      render: (years: number) => `${years}年`,
    },
    {
      title: '入学日期',
      dataIndex: 'enrollment_date',
      key: 'enrollment_date',
      width: 120,
    },
    {
      title: '毕业日期',
      dataIndex: 'graduation_date',
      key: 'graduation_date',
      width: 120,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: EducationStatus) => (
        <Tag color={status === 'valid' ? 'green' : 'red'}>
          {EDUCATION_STATUS_MAP[status]}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 160,
      fixed: 'right',
      render: (_, record) => (
        <Space size="middle">
          <Button
            type="link"
            icon={<EditOutlined />}
            onClick={() => {
              setEditingDiploma(record);
              setModalVisible(true);
            }}
          >
            编辑
          </Button>
          <Popconfirm
            title="确认删除该学历信息吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确认"
            cancelText="取消"
          >
            <Button type="link" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>学历信息管理</Title>
        <Button 
          type="primary" 
          icon={<PlusOutlined />}
          onClick={() => {
            setEditingDiploma(undefined);
            setModalVisible(true);
          }}
        >
          新增学历
        </Button>
      </div>

      <Table
        columns={columns}
        dataSource={diplomas}
        rowKey="id"
        loading={loading}
        scroll={{ x: 1300 }}
        pagination={{
          pageSize: 10,
          showSizeChanger: true,
          showTotal: (total) => `共 ${total} 条记录`,
        }}
      />

      <DiplomaForm
        visible={modalVisible}
        onCancel={() => {
          setModalVisible(false);
          setEditingDiploma(undefined);
        }}
        onSubmit={editingDiploma ? handleUpdate : handleCreate}
        initialValues={editingDiploma}
        loading={submitLoading}
      />
    </div>
  );
};

export default Diplomas;
