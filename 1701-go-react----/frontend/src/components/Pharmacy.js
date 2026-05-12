import React, { useState, useEffect } from 'react';
import {
  Table,
  Button,
  Modal,
  Form,
  Input,
  InputNumber,
  Select,
  Tag,
  message,
  Space,
  Card,
  Tabs,
  Divider,
  Popconfirm,
  Row,
  Col,
  Badge,
  Descriptions,
} from 'antd';
import { CheckCircleOutlined, MedicineBoxOutlined, EditOutlined } from '@ant-design/icons';
import {
  getPrescriptions,
  getHerbs,
  dispensePrescription,
  updateHerb,
} from '../api';
import dayjs from 'dayjs';

const { Option } = Select;
const { TabPane } = Tabs;

const prescriptionStatusMap = {
  draft: { text: '草稿', color: 'default' },
  confirmed: { text: '待发药', color: 'blue' },
  dispensed: { text: '已发药', color: 'green' },
};

const herbStatusMap = {
  normal: { text: '正常', color: 'green' },
  warning: { text: '临期', color: 'gold' },
  expired: { text: '过期', color: 'red' },
};

function Pharmacy() {
  const [prescriptions, setPrescriptions] = useState([]);
  const [herbs, setHerbs] = useState([]);
  const [viewPrescription, setViewPrescription] = useState(null);
  const [dispenseModalVisible, setDispenseModalVisible] = useState(false);
  const [currentDispensePrescription, setCurrentDispensePrescription] = useState(null);
  const [editHerbModalVisible, setEditHerbModalVisible] = useState(false);
  const [currentEditHerb, setCurrentEditHerb] = useState(null);
  const [dispenseForm] = Form.useForm();
  const [herbForm] = Form.useForm();

  const fetchData = async () => {
    try {
      const [prescriptionsResp, herbsResp] = await Promise.all([
        getPrescriptions(),
        getHerbs(),
      ]);
      setPrescriptions(prescriptionsResp.data.prescriptions);
      setHerbs(herbsResp.data.herbs);
    } catch (error) {
      message.error('获取数据失败');
    }
  };

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 5000);
    return () => clearInterval(interval);
  }, []);

  const confirmedPrescriptions = prescriptions.filter(
    (p) => p.status === 'confirmed' && p.status_reason !== '已作废'
  );

  const decoctionTasks = prescriptions.filter(
    (p) =>
      p.status === 'dispensed' &&
      p.dispensing_mode === 'decoction' &&
      p.status_reason !== '已作废'
  );

  const handleDispense = (prescription) => {
    setCurrentDispensePrescription(prescription);
    setViewPrescription(prescription);
    setDispenseModalVisible(true);
    dispenseForm.resetFields();
    dispenseForm.setFieldsValue({
      dispensing_mode: 'self',
      minutes_from_now: 60,
    });
  };

  const handleSubmitDispense = async (values) => {
    try {
      await dispensePrescription(currentDispensePrescription.id, values);
      message.success('发药成功');
      setDispenseModalVisible(false);
      setViewPrescription(null);
      fetchData();
    } catch (error) {
      const errData = error.response?.data || {};
      let msg = errData.error || '发药失败';
      if (errData.missing_herbs) {
        const missing = Object.entries(errData.missing_herbs)
          .map(([name, stock]) => `${name}(库存: ${stock}克)`)
          .join('、');
        msg = `药材库存不足: ${missing}`;
      }
      message.error(msg);
    }
  };

  const handleEditHerb = (herb) => {
    setCurrentEditHerb(herb);
    herbForm.setFieldsValue({
      name: herb.name,
      stock: herb.stock,
      price: herb.price,
    });
    setEditHerbModalVisible(true);
  };

  const handleSubmitHerbEdit = async (values) => {
    try {
      await updateHerb(values);
      message.success('药材信息已更新');
      setEditHerbModalVisible(false);
      fetchData();
    } catch (error) {
      message.error(error.response?.data?.error || '更新失败');
    }
  };

  const prescriptionColumns = [
    {
      title: '序号',
      key: 'serial',
      render: (_, __, index) => index + 1,
      width: 60,
    },
    {
      title: '患者',
      dataIndex: 'patient_name',
      key: 'patient_name',
    },
    {
      title: '诊断',
      dataIndex: 'diagnosis',
      key: 'diagnosis',
    },
    {
      title: '药味数',
      key: 'count',
      render: (_, record) => record.items?.length || 0,
    },
    {
      title: '总剂量',
      dataIndex: 'total_dosage',
      key: 'total_dosage',
      render: (v) => `${v}克`,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (t) => dayjs(t).format('MM-DD HH:mm'),
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Space>
          <Button onClick={() => setViewPrescription(record)}>查看</Button>
          <Button
            type="primary"
            icon={<CheckCircleOutlined />}
            onClick={() => handleDispense(record)}
          >
            发药
          </Button>
        </Space>
      ),
    },
  ];

  const decoctionColumns = [
    {
      title: '患者',
      dataIndex: 'patient_name',
      key: 'patient_name',
    },
    {
      title: '诊断',
      dataIndex: 'diagnosis',
      key: 'diagnosis',
    },
    {
      title: '煎药机',
      dataIndex: 'decoction_machine_id',
      key: 'decoction_machine_id',
    },
    {
      title: '锅数',
      dataIndex: 'pot_count',
      key: 'pot_count',
    },
    {
      title: '预计完成时间',
      dataIndex: 'estimated_finish',
      key: 'estimated_finish',
      render: (t) => {
        const dt = dayjs(t);
        const isPast = dt.isBefore(dayjs());
        return (
          <Tag color={isPast ? 'green' : 'orange'}>
            {dt.format('MM-DD HH:mm')}
          </Tag>
        );
      },
    },
    {
      title: '状态',
      key: 'status',
      render: (_, record) => {
        const dt = dayjs(record.estimated_finish);
        if (dt.isBefore(dayjs())) {
          return <Badge status="success" text="可领取" />;
        }
        return <Badge status="processing" text="煎煮中" />;
      },
    },
  ];

  const herbColumns = [
    {
      title: '药材名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '库存',
      dataIndex: 'stock',
      key: 'stock',
      render: (stock, record) => {
        let color = 'inherit';
        if (record.status === 'warning') color = '#faad14';
        if (record.status === 'expired') color = '#ff4d4f';
        return <span style={{ color }}>{stock} 克</span>;
      },
    },
    {
      title: '效期',
      dataIndex: 'expiry_date',
      key: 'expiry_date',
      render: (date, record) => {
        const info = herbStatusMap[record.status];
        return (
          <Space>
            <span>{date}</span>
            <Tag color={info.color}>{info.text}</Tag>
          </Space>
        );
      },
    },
    {
      title: '单价',
      dataIndex: 'price',
      key: 'price',
      render: (price) => `¥${price}`,
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Button
          icon={<EditOutlined />}
          onClick={() => handleEditHerb(record)}
        >
          调整
        </Button>
      ),
    },
  ];

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <h2 style={{ margin: 0 }}>药房管理台</h2>
      </div>

      <Row gutter={16}>
        <Col span={8}>
          <Card
            size="small"
            title={
              <Space>
                <MedicineBoxOutlined />
                待发药处方
              </Space>
            }
            extra={<Badge count={confirmedPrescriptions.length} color="blue" />}
          >
            <Table
              dataSource={confirmedPrescriptions}
              columns={prescriptionColumns}
              rowKey="id"
              pagination={{ pageSize: 5 }}
              size="small"
              scroll={{ y: 400 }}
            />
          </Card>
        </Col>

        <Col span={16}>
          <Card size="small" title="进行中的煎药任务">
            <Table
              dataSource={decoctionTasks}
              columns={decoctionColumns}
              rowKey="id"
              pagination={false}
              size="small"
            />
          </Card>
        </Col>
      </Row>

      <Card title="库存管理" size="small" style={{ marginTop: 16 }}>
        <Table
          dataSource={herbs}
          columns={herbColumns}
          rowKey="name"
          pagination={{ pageSize: 10 }}
          rowClassName={(record) => {
            if (record.status === 'expired') return 'table-row-danger';
            if (record.status === 'warning') return 'table-row-warning';
            return '';
          }}
        />
      </Card>

      <style>{`
        .table-row-danger {
          background-color: #fff1f0 !important;
        }
        .table-row-warning {
          background-color: #fffbe6 !important;
        }
      `}</style>

      <Modal
        title="处方详情"
        open={!!viewPrescription && !dispenseModalVisible}
        onCancel={() => setViewPrescription(null)}
        footer={null}
        width={700}
      >
        {viewPrescription && (
          <div>
            <Descriptions bordered size="small" column={2}>
              <Descriptions.Item label="患者">
                {viewPrescription.patient_name}
              </Descriptions.Item>
              <Descriptions.Item label="诊断">
                {viewPrescription.diagnosis}
              </Descriptions.Item>
              <Descriptions.Item label="证型">
                {viewPrescription.syndrome}
              </Descriptions.Item>
              <Descriptions.Item label="总剂量">
                {viewPrescription.total_dosage}克
              </Descriptions.Item>
            </Descriptions>

            <Divider>处方明细</Divider>

            <Table
              dataSource={viewPrescription.items}
              columns={[
                { title: '药材', dataIndex: 'herb_name', key: 'herb_name' },
                { title: '剂量', dataIndex: 'dosage', key: 'dosage' },
                { title: '单位', dataIndex: 'unit', key: 'unit' },
                { title: '煎煮方式', dataIndex: 'cooking_method', key: 'cooking_method' },
              ]}
              rowKey="herb_name"
              pagination={false}
              size="small"
            />
          </div>
        )}
      </Modal>

      <Modal
        title="发药确认"
        open={dispenseModalVisible}
        onCancel={() => {
          setDispenseModalVisible(false);
          setViewPrescription(null);
        }}
        onOk={() => dispenseForm.submit()}
        width={800}
        okText="确认发药"
        cancelText="取消"
      >
        {currentDispensePrescription && (
          <div>
            <Descriptions bordered size="small" column={2} style={{ marginBottom: 16 }}>
              <Descriptions.Item label="患者">
                {currentDispensePrescription.patient_name}
              </Descriptions.Item>
              <Descriptions.Item label="诊断">
                {currentDispensePrescription.diagnosis}
              </Descriptions.Item>
              <Descriptions.Item label="总剂量" span={2}>
                {currentDispensePrescription.total_dosage}克
              </Descriptions.Item>
            </Descriptions>

            <Table
              title={() => <strong>处方明细</strong>}
              dataSource={currentDispensePrescription.items}
              columns={[
                { title: '药材', dataIndex: 'herb_name', key: 'herb_name' },
                { title: '剂量', dataIndex: 'dosage', key: 'dosage' },
                { title: '单位', dataIndex: 'unit', key: 'unit' },
                { title: '煎煮方式', dataIndex: 'cooking_method', key: 'cooking_method' },
              ]}
              rowKey="herb_name"
              pagination={false}
              size="small"
            />

            <Divider>发放设置</Divider>

            <Form
              form={dispenseForm}
              layout="vertical"
              onFinish={handleSubmitDispense}
            >
              <Form.Item
                name="dispensing_mode"
                label="发放方式"
                rules={[{ required: true, message: '请选择发放方式' }]}
              >
                <Select>
                  <Option value="self">自煎（直接发药给患者）</Option>
                  <Option value="decoction">代煎（进行煎药）</Option>
                </Select>
              </Form.Item>

              <Form.Item
                noStyle
                shouldUpdate={(prevValues, curValues) =>
                  prevValues.dispensing_mode !== curValues.dispensing_mode
                }
              >
                {({ getFieldValue }) =>
                  getFieldValue('dispensing_mode') === 'decoction' ? (
                    <>
                      <Row gutter={16}>
                        <Col span={12}>
                          <Form.Item
                            name="decoction_machine_id"
                            label="煎药机编号"
                            rules={[{ required: true, message: '请输入煎药机编号' }]}
                          >
                            <Input placeholder="如：JY-001" />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item
                            name="pot_count"
                            label="煎煮锅数"
                            rules={[{ required: true, message: '请输入锅数' }]}
                          >
                            <InputNumber min={1} style={{ width: '100%' }} />
                          </Form.Item>
                        </Col>
                      </Row>
                      <Form.Item
                        name="minutes_from_now"
                        label="预计完成时间（分钟后）"
                        initialValue={60}
                      >
                        <InputNumber min={1} style={{ width: '100%' }} />
                      </Form.Item>
                    </>
                  ) : null
                }
              </Form.Item>
            </Form>
          </div>
        )}
      </Modal>

      <Modal
        title="调整药材信息"
        open={editHerbModalVisible}
        onCancel={() => setEditHerbModalVisible(false)}
        onOk={() => herbForm.submit()}
      >
        {currentEditHerb && (
          <Form form={herbForm} layout="vertical" onFinish={handleSubmitHerbEdit}>
            <Form.Item name="name" label="药材名称">
              <Input disabled />
            </Form.Item>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item name="stock" label="库存（克）">
                  <InputNumber min={0} style={{ width: '100%' }} />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item name="price" label="单价（元）">
                  <InputNumber min={0} step={0.01} style={{ width: '100%' }} />
                </Form.Item>
              </Col>
            </Row>
          </Form>
        )}
      </Modal>
    </div>
  );
}

export default Pharmacy;
