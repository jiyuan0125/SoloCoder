import React, { useState, useEffect } from 'react';
import {
  Table,
  Button,
  Modal,
  Form,
  Input,
  Select,
  InputNumber,
  Tag,
  message,
  Space,
  Card,
  Divider,
  Row,
  Col,
  Popconfirm,
  Alert,
  List,
} from 'antd';
import {
  PlayCircleOutlined,
  PlusCircleOutlined,
  CheckCircleOutlined,
  DeleteOutlined,
  EyeOutlined,
} from '@ant-design/icons';
import {
  getRegistrations,
  getPrescriptionsByRegistration,
  callPatient,
  finishVisit,
  createPrescription,
  updatePrescription,
  confirmPrescription,
  voidPrescription,
  getHerbs,
} from '../api';
import dayjs from 'dayjs';

const { Option } = Select;
const { TextArea } = Input;

const statusMap = {
  waiting: { text: '候诊中', color: 'blue' },
  in_visit: { text: '就诊中', color: 'orange' },
  done: { text: '已完成', color: 'green' },
};

const prescriptionStatusMap = {
  draft: { text: '草稿', color: 'default' },
  confirmed: { text: '已确认', color: 'blue' },
  dispensed: { text: '已发药', color: 'green' },
};

const unitOptions = ['克', '毫升'];
const cookingMethods = ['先煎', '后下', '包煎', '烊化', '冲服', '普通'];

function DoctorWorkstation() {
  const [registrations, setRegistrations] = useState([]);
  const [selectedPatient, setSelectedPatient] = useState(null);
  const [prescriptions, setPrescriptions] = useState([]);
  const [herbs, setHerbs] = useState([]);
  const [modalVisible, setModalVisible] = useState(false);
  const [viewPrescription, setViewPrescription] = useState(null);
  const [currentPrescription, setCurrentPrescription] = useState(null);
  const [form] = Form.useForm();
  const [doctorName, setDoctorName] = useState('张医生');
  const [warnings, setWarnings] = useState([]);

  const fetchData = async () => {
    try {
      const [regResp, herbsResp] = await Promise.all([
        getRegistrations(),
        getHerbs(),
      ]);
      setRegistrations(regResp.data.registrations);
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

  const handleCallPatient = async (registration) => {
    try {
      await callPatient({
        registration_id: registration.id,
        doctor_name: doctorName,
      });
      message.success(`已叫号: ${registration.serial_num}`);
      fetchData();
    } catch (error) {
      message.error(error.response?.data?.error || '叫号失败');
    }
  };

  const handleSelectPatient = async (registration) => {
    setSelectedPatient(registration);
    try {
      const response = await getPrescriptionsByRegistration(registration.id);
      setPrescriptions(response.data.prescriptions);
    } catch (error) {
      message.error('获取处方失败');
    }
  };

  const handleFinishVisit = async () => {
    if (!selectedPatient) return;
    try {
      await finishVisit(selectedPatient.id);
      message.success('就诊完成');
      setSelectedPatient(null);
      setPrescriptions([]);
      fetchData();
    } catch (error) {
      message.error(error.response?.data?.error || '操作失败');
    }
  };

  const handleCreatePrescription = () => {
    setCurrentPrescription(null);
    setWarnings([]);
    form.resetFields();
    form.setFieldsValue({ items: [{ unit: '克', cooking_method: '普通' }] });
    setModalVisible(true);
  };

  const handleEditPrescription = (prescription) => {
    setCurrentPrescription(prescription);
    setWarnings([]);
    form.setFieldsValue({
      diagnosis: prescription.diagnosis,
      syndrome: prescription.syndrome,
      items: prescription.items,
    });
    setModalVisible(true);
  };

  const handleSubmitPrescription = async (values) => {
    try {
      let response;
      if (currentPrescription) {
        response = await updatePrescription(currentPrescription.id, values);
      } else {
        response = await createPrescription({
          registration_id: selectedPatient.id,
          ...values,
        });
      }
      message.success(currentPrescription ? '处方更新成功' : '处方创建成功');
      
      if (response.data.warnings && response.data.warnings.length > 0) {
        setWarnings(response.data.warnings);
      }
      
      setModalVisible(false);
      const regResp = await getPrescriptionsByRegistration(selectedPatient.id);
      setPrescriptions(regResp.data.prescriptions);
    } catch (error) {
      const errMsg = error.response?.data?.error || '操作失败';
      if (error.response?.data?.warnings) {
        setWarnings(error.response.data.warnings);
      }
      message.error(errMsg);
    }
  };

  const handleConfirmPrescription = async (prescription) => {
    try {
      await confirmPrescription(prescription.id);
      message.success('处方已确认');
      const regResp = await getPrescriptionsByRegistration(selectedPatient.id);
      setPrescriptions(regResp.data.prescriptions);
    } catch (error) {
      message.error(error.response?.data?.error || '操作失败');
    }
  };

  const handleVoidPrescription = async (prescription) => {
    try {
      await voidPrescription(prescription.id);
      message.success('处方已作废');
      const regResp = await getPrescriptionsByRegistration(selectedPatient.id);
      setPrescriptions(regResp.data.prescriptions);
    } catch (error) {
      message.error(error.response?.data?.error || '操作失败');
    }
  };

  const availableHerbs = herbs.filter(h => h.status !== 'expired');

  const waitingColumns = [
    {
      title: '序号',
      dataIndex: 'serial_num',
      key: 'serial_num',
      width: 120,
    },
    {
      title: '姓名',
      dataIndex: 'patient_name',
      key: 'patient_name',
    },
    {
      title: '主诉',
      dataIndex: 'chief_complaint',
      key: 'chief_complaint',
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Space>
          {record.status === 'waiting' && (
            <Button
              type="primary"
              icon={<PlayCircleOutlined />}
              onClick={() => handleCallPatient(record)}
            >
              叫号
            </Button>
          )}
          {record.status === 'in_visit' && (
            <Button onClick={() => handleSelectPatient(record)}>
              接诊
            </Button>
          )}
        </Space>
      ),
    },
  ];

  const prescriptionColumns = [
    {
      title: '诊断',
      dataIndex: 'diagnosis',
      key: 'diagnosis',
    },
    {
      title: '证型',
      dataIndex: 'syndrome',
      key: 'syndrome',
    },
    {
      title: '药味数',
      key: 'items_count',
      render: (_, record) => record.items?.length || 0,
    },
    {
      title: '总剂量',
      dataIndex: 'total_dosage',
      key: 'total_dosage',
      render: (v) => `${v}克`,
    },
    {
      title: '状态',
      key: 'status',
      render: (_, record) => {
        if (record.status_reason === '已作废') {
          return <Tag color="red">已作废</Tag>;
        }
        const info = prescriptionStatusMap[record.status] || { text: record.status, color: 'default' };
        return <Tag color={info.color}>{info.text}</Tag>;
      },
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => {
        const isVoided = record.status_reason === '已作废';
        return (
          <Space>
            <Button icon={<EyeOutlined />} onClick={() => setViewPrescription(record)}>
              查看
            </Button>
            {record.status === 'draft' && !isVoided && (
              <Button onClick={() => handleEditPrescription(record)}>
                编辑
              </Button>
            )}
            {record.status === 'draft' && !isVoided && (
              <Popconfirm
                title="确认确认该处方？"
                onConfirm={() => handleConfirmPrescription(record)}
              >
                <Button type="primary" icon={<CheckCircleOutlined />}>
                  确认
                </Button>
              </Popconfirm>
            )}
            {record.status === 'confirmed' && !isVoided && (
              <Popconfirm
                title="确认作废该处方？作废后需要重新开处方"
                onConfirm={() => handleVoidPrescription(record)}
              >
                <Button danger icon={<DeleteOutlined />}>
                  作废
                </Button>
              </Popconfirm>
            )}
          </Space>
        );
      },
    },
  ];

  const waitingList = registrations.filter((r) => r.status === 'waiting');
  const inVisitList = registrations.filter((r) => r.status === 'in_visit');

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h2 style={{ margin: 0 }}>诊室工作台</h2>
        <Space>
          <span>当前医生:</span>
          <Input
            value={doctorName}
            onChange={(e) => setDoctorName(e.target.value)}
            style={{ width: 150 }}
          />
        </Space>
      </div>

      <Row gutter={16}>
        <Col span={8}>
          <Card title={`候诊队列 (${waitingList.length}人)`} size="small">
            <Table
              dataSource={waitingList}
              columns={waitingColumns}
              rowKey="id"
              pagination={false}
              size="small"
              scroll={{ y: 300 }}
            />
          </Card>
        </Col>

        <Col span={16}>
          <Card
            title="当前患者"
            size="small"
            extra={
              selectedPatient && (
                <Space>
                  <Button type="primary" onClick={handleCreatePrescription}>
                    开处方
                  </Button>
                  <Popconfirm title="确认完成就诊？" onConfirm={handleFinishVisit}>
                    <Button>完成就诊</Button>
                  </Popconfirm>
                </Space>
              )
            }
          >
            {selectedPatient ? (
              <div>
                <Space direction="vertical" style={{ width: '100%', marginBottom: 16 }}>
                  <div>
                    <strong>序号: </strong>{selectedPatient.serial_num}
                    <span style={{ marginLeft: 20 }}>
                      <strong>姓名: </strong>{selectedPatient.patient_name}
                    </span>
                    <span style={{ marginLeft: 20 }}>
                      <strong>电话: </strong>{selectedPatient.patient_phone}
                    </span>
                  </div>
                  <div>
                    <strong>主诉: </strong>{selectedPatient.chief_complaint}
                  </div>
                </Space>

                <Divider>处方记录</Divider>

                <Table
                  dataSource={prescriptions}
                  columns={prescriptionColumns}
                  rowKey="id"
                  pagination={false}
                  size="small"
                />
              </div>
            ) : (
              <div style={{ textAlign: 'center', color: '#999', padding: 40 }}>
                请先从左侧候诊队列中选择一位患者接诊
              </div>
            )}
          </Card>
        </Col>
      </Row>

      <Card title="就诊中患者" size="small" style={{ marginTop: 16 }}>
        <Table
          dataSource={inVisitList}
          columns={[
            { title: '序号', dataIndex: 'serial_num', key: 'serial_num' },
            { title: '姓名', dataIndex: 'patient_name', key: 'patient_name' },
            { title: '主诉', dataIndex: 'chief_complaint', key: 'chief_complaint' },
            {
              title: '操作',
              key: 'action',
              render: (_, record) => (
                <Button onClick={() => handleSelectPatient(record)}>处理</Button>
              ),
            },
          ]}
          rowKey="id"
          pagination={false}
          size="small"
        />
      </Card>

      <Modal
        title={currentPrescription ? '编辑处方' : '新开处方'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => form.submit()}
        width={900}
        okText="保存为草稿"
        cancelText="取消"
      >
        {warnings.length > 0 && (
          <Alert
            message="提示信息"
            description={
              <List
                dataSource={warnings}
                renderItem={(item) => <List.Item>{item}</List.Item>}
              />
            }
            type="warning"
            showIcon
            style={{ marginBottom: 16 }}
          />
        )}
        <Form form={form} layout="vertical" onFinish={handleSubmitPrescription}>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="diagnosis"
                label="中医病名"
                rules={[{ required: true, message: '请输入中医病名' }]}
              >
                <Input placeholder="如：感冒、咳嗽等" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="syndrome"
                label="证型"
                rules={[{ required: true, message: '请输入证型' }]}
              >
                <Input placeholder="如：风寒感冒、风热感冒等" />
              </Form.Item>
            </Col>
          </Row>

          <Form.List name="items">
            {(fields, { add, remove }) => (
              <>
                <div style={{ marginBottom: 8, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <label style={{ fontWeight: 'bold' }}>处方明细（最多20味药）</label>
                  {fields.length < 20 && (
                    <Button
                      type="dashed"
                      onClick={() => add({ unit: '克', cooking_method: '普通' })}
                      icon={<PlusCircleOutlined />}
                    >
                      添加药材
                    </Button>
                  )}
                </div>
                {fields.map(({ key, name, ...restField }) => (
                  <Space key={key} style={{ display: 'flex', marginBottom: 8 }} align="baseline">
                    <Form.Item
                      {...restField}
                      name={[name, 'herb_name']}
                      rules={[{ required: true, message: '请选择药材' }]}
                      style={{ width: 180, marginBottom: 0 }}
                    >
                      <Select placeholder="选择药材" showSearch optionFilterProp="children">
                        {availableHerbs.map((herb) => (
                          <Option key={herb.name} value={herb.name}>
                            {herb.name}
                            {herb.status === 'warning' && ' (临期)'}
                          </Option>
                        ))}
                      </Select>
                    </Form.Item>
                    <Form.Item
                      {...restField}
                      name={[name, 'dosage']}
                      rules={[{ required: true, message: '请输入剂量' }]}
                      style={{ width: 100, marginBottom: 0 }}
                    >
                      <InputNumber placeholder="剂量" min={1} step={5} />
                    </Form.Item>
                    <Form.Item
                      {...restField}
                      name={[name, 'unit']}
                      initialValue="克"
                      style={{ width: 80, marginBottom: 0 }}
                    >
                      <Select>
                        {unitOptions.map((u) => (
                          <Option key={u} value={u}>
                            {u}
                          </Option>
                        ))}
                      </Select>
                    </Form.Item>
                    <Form.Item
                      {...restField}
                      name={[name, 'cooking_method']}
                      initialValue="普通"
                      style={{ width: 100, marginBottom: 0 }}
                    >
                      <Select>
                        {cookingMethods.map((m) => (
                          <Option key={m} value={m}>
                            {m}
                          </Option>
                        ))}
                      </Select>
                    </Form.Item>
                    <Button
                      type="text"
                      danger
                      icon={<DeleteOutlined />}
                      onClick={() => remove(name)}
                    />
                  </Space>
                ))}
              </>
            )}
          </Form.List>
        </Form>
      </Modal>

      <Modal
        title="处方详情"
        open={!!viewPrescription}
        onCancel={() => setViewPrescription(null)}
        footer={null}
        width={700}
      >
        {viewPrescription && (
          <div>
            <Space direction="vertical" style={{ width: '100%' }}>
              <div>
                <strong>患者: </strong>{viewPrescription.patient_name}
              </div>
              <div>
                <strong>诊断: </strong>{viewPrescription.diagnosis}
              </div>
              <div>
                <strong>证型: </strong>{viewPrescription.syndrome}
              </div>
              {viewPrescription.status_reason === '已作废' && (
                <Tag color="red">已作废</Tag>
              )}
              {viewPrescription.status === 'dispensed' && (
                <div>
                  <strong>发药方式: </strong>
                  {viewPrescription.dispensing_mode === 'self' ? '自煎' : '代煎'}
                  {viewPrescription.dispensing_mode === 'decoction' && (
                    <span>
                      <br />
                      <strong>煎药机: </strong>{viewPrescription.decoction_machine_id}
                      <br />
                      <strong>锅数: </strong>{viewPrescription.pot_count}
                      <br />
                      <strong>预计完成: </strong>
                      {dayjs(viewPrescription.estimated_finish).format('YYYY-MM-DD HH:mm')}
                    </span>
                  )}
                </div>
              )}
            </Space>

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

            <div style={{ marginTop: 16, textAlign: 'right' }}>
              <strong>总剂量: </strong>{viewPrescription.total_dosage}克
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
}

export default DoctorWorkstation;
