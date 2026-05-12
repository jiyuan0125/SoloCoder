import { Command } from 'commander';
import fetch, { RequestInit as FetchRequestInit } from 'node-fetch';

const program = new Command();

const getBaseUrl = (): string => {
  return process.env.SERVER_URL || 'http://localhost:3000';
};

async function apiCall<T>(
  endpoint: string,
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' = 'GET',
  body?: unknown
): Promise<T> {
  const url = `${getBaseUrl()}${endpoint}`;
  const options: FetchRequestInit = {
    method,
    headers: {
      'Content-Type': 'application/json',
    },
  };

  if (body && (method === 'POST' || method === 'PUT')) {
    options.body = JSON.stringify(body);
  }

  const response = await fetch(url, options);
  const data = await response.json();

  if (!response.ok) {
    throw new Error(`API 错误 ${response.status}: ${JSON.stringify(data)}`);
  }

  return data as T;
}

program
  .name('cdm-cli')
  .description('慢病管理系统命令行客户端')
  .version('1.0.0');

program
  .command('health')
  .description('检查服务端健康状态')
  .action(async () => {
    try {
      const result = await apiCall<{ status: string; timestamp: string }>('/health');
      console.log('服务端状态:', result.status);
      console.log('时间戳:', result.timestamp);
    } catch (error) {
      console.error('健康检查失败:', (error as Error).message);
      process.exit(1);
    }
  });

const patient = program.command('patient').description('患者管理');

patient
  .command('list')
  .description('列出所有患者')
  .action(async () => {
    try {
      const patients = await apiCall<unknown[]>('/api/r');
      console.log(`\n共 ${patients.length} 位患者:\n`);
      patients.forEach((p: any) => {
        console.log(`ID: ${p.id}`);
        console.log(`  姓名: ${p.basicInfo.name}`);
        console.log(`  年龄: ${p.basicInfo.age}`);
        console.log(`  慢病类型: ${p.chronicDiseases.join(', ')}`);
        console.log(`  风险等级: ${p.riskLevel}`);
        console.log(`  随访间隔: ${p.followUpIntervalDays} 天`);
        console.log(`  下次随访: ${p.nextFollowUpDate || '待安排'}`);
        console.log('');
      });
    } catch (error) {
      console.error('获取患者列表失败:', (error as Error).message);
    }
  });

patient
  .command('create')
  .description('创建患者档案')
  .option('-n, --name <name>', '姓名')
  .option('-g, --gender <gender>', '性别 (male/female/other)')
  .option('-a, --age <age>', '年龄', parseInt)
  .option('-p, --phone <phone>', '电话')
  .option('-i, --idcard <idcard>', '身份证号')
  .option('-d, --disease <disease>', '慢病类型 (逗号分隔)')
  .option('--diagnosis-date <date>', '确诊日期 (YYYY-MM-DD)')
  .option('--systolic <value>', '收缩压', parseInt)
  .option('--diastolic <value>', '舒张压', parseInt)
  .option('--sugar <value>', '空腹血糖', parseFloat)
  .option('--weight <value>', '体重 (kg)', parseFloat)
  .option('--height <value>', '身高 (cm)', parseFloat)
  .action(async (options) => {
    try {
      const diseases = options.disease ? options.disease.split(',') : ['hypertension'];
      const diagnosisDate = new Date(options.diagnosisDate || '2024-01-01');

      const patientData = {
        basicInfo: {
          name: options.name || '测试患者',
          gender: options.gender || 'male',
          age: options.age || 60,
          phone: options.phone || '13800138000',
          idCard: options.idcard || '110101196001010001',
        },
        chronicDiseases: diseases,
        diagnosisDate: diagnosisDate.toISOString(),
        medicalHistory: [],
        initialMetrics: {
          systolicBP: options.systolic || 140,
          diastolicBP: options.diastolic || 90,
          fastingBloodSugar: options.sugar || 7.5,
          weight: options.weight || 70,
          height: options.height || 170,
        },
      };

      const result = await apiCall<any>('/api/r', 'POST', patientData);
      console.log('\n患者创建成功!');
      console.log('ID:', result.id);
      console.log('姓名:', result.basicInfo.name);
      console.log('风险等级:', result.riskLevel);
      console.log('随访间隔:', result.followUpIntervalDays, '天');
    } catch (error) {
      console.error('创建患者失败:', (error as Error).message);
    }
  });

patient
  .command('get <id>')
  .description('获取患者详情')
  .action(async (id) => {
    try {
      const patient = await apiCall<any>(`/api/r/${id}`);
      console.log('\n患者详情:');
      console.log(JSON.stringify(patient, null, 2));
    } catch (error) {
      console.error('获取患者详情失败:', (error as Error).message);
    }
  });

patient
  .command('risk <id>')
  .description('手动升级风险等级')
  .option('-l, --level <level>', '目标风险等级 (low/medium/high)')
  .option('-d, --downgrade', '人工降级模式')
  .action(async (id, options) => {
    try {
      if (options.downgrade) {
        const result = await apiCall<any>(`/api/r/${id}/actions/downgrade-risk`, 'POST', {
          newRiskLevel: options.level || 'low',
        });
        console.log('\n人工降级成功!');
        console.log('新风险等级:', result.patient.riskLevel);
      } else {
        const result = await apiCall<any>(`/api/r/${id}/actions/approve`, 'POST', {
          newRiskLevel: options.level || 'medium',
        });
        console.log('\n风险等级升级成功!');
        console.log('新风险等级:', result.patient.riskLevel);
      }
    } catch (error) {
      console.error('操作失败:', (error as Error).message);
    }
  });

const followup = program.command('followup').description('随访管理');

followup
  .command('list')
  .description('列出所有随访记录')
  .action(async () => {
    try {
      const followups = await apiCall<unknown[]>('/api/followups');
      console.log(`\n共 ${followups.length} 条随访记录:\n`);
      followups.forEach((f: any) => {
        console.log(`ID: ${f.id}`);
        console.log(`  患者ID: ${f.patientId}`);
        console.log(`  日期: ${f.date}`);
        console.log(`  血压: ${f.metrics.systolicBP}/${f.metrics.diastolicBP}`);
        console.log(`  血糖: ${f.metrics.fastingBloodSugar}`);
        console.log('');
      });
    } catch (error) {
      console.error('获取随访记录失败:', (error as Error).message);
    }
  });

followup
  .command('create')
  .description('创建随访记录')
  .option('-p, --patient <id>', '患者ID')
  .option('-d, --date <date>', '随访日期 (YYYY-MM-DD)')
  .option('--systolic <value>', '收缩压', parseInt)
  .option('--diastolic <value>', '舒张压', parseInt)
  .option('--sugar <value>', '空腹血糖', parseFloat)
  .option('--taken <bool>', '是否按时服药', 'true')
  .action(async (options) => {
    try {
      if (!options.patient) {
        console.error('请指定患者ID (--patient)');
        return;
      }

      const followupData = {
        date: options.date || new Date().toISOString().split('T')[0],
        symptoms: [],
        medications: [],
        metrics: {
          systolicBP: options.systolic || 135,
          diastolicBP: options.diastolic || 85,
          fastingBloodSugar: options.sugar || 6.5,
        },
        lifestyleAssessment: {
          smoking: 'none',
          alcohol: 'none',
          exercise: 'occasional',
          diet: 'moderate',
          sleepHours: 7,
        },
        medicationTaken: options.taken === 'true',
      };

      const result = await apiCall<any>(`/api/r/${options.patient}/items`, 'POST', followupData);
      console.log('\n随访记录创建成功!');
      console.log('记录ID:', result.record.id);
      console.log('血压控制:', result.controlStatus.bloodPressureControlled ? '达标' : '未达标');
      console.log('血糖控制:', result.controlStatus.bloodSugarControlled ? '达标' : '未达标');
      if (result.riskLevelChanged) {
        console.log(`\n⚠️  风险等级已变更: ${result.oldRiskLevel} → ${result.newRiskLevel}`);
      }
      if (result.notificationRequired) {
        console.log('⚠️  依从性差，已通知家庭医生');
      }
    } catch (error) {
      console.error('创建随访记录失败:', (error as Error).message);
    }
  });

followup
  .command('patient <id>')
  .description('获取患者的随访记录')
  .action(async (id) => {
    try {
      const data = await apiCall<any>(`/api/r/${id}/items`);
      console.log(`\n患者 ${data.patient.basicInfo.name} 的详情:\n`);
      console.log(`随访记录 (${data.followUps.length} 条):`);
      data.followUps.forEach((f: any, i: number) => {
        console.log(`  ${i + 1}. 日期: ${f.date}`);
        console.log(`     血压: ${f.metrics.systolicBP}/${f.metrics.diastolicBP}`);
        console.log(`     血糖: ${f.metrics.fastingBloodSugar}`);
      });

      if (data.statistics) {
        console.log(`\n统计数据:`);
        console.log(`  总随访次数: ${data.statistics.totalFollowUps}`);
        console.log(`  达标次数: ${data.statistics.controlledCount}`);
        console.log(`  依从率: ${data.statistics.adherenceRate}%`);
        console.log(`  趋势: ${data.statistics.recentTrend}`);
      }
    } catch (error) {
      console.error('获取患者详情失败:', (error as Error).message);
    }
  });

const medication = program.command('medication').description('用药管理');

medication
  .command('add')
  .description('添加用药')
  .option('-p, --patient <id>', '患者ID')
  .option('-n, --name <name>', '药品名称')
  .option('-g, --generic <name>', '通用名')
  .option('-d, --dosage <dosage>', '剂量')
  .option('-f, --frequency <freq>', '频次')
  .option('--start <date>', '开始日期')
  .option('--end <date>', '结束日期')
  .action(async (options) => {
    try {
      if (!options.patient) {
        console.error('请指定患者ID (--patient)');
        return;
      }

      const medData = {
        name: options.name || '阿司匹林',
        genericName: options.generic || '',
        dosage: options.dosage || '100mg',
        frequency: options.frequency || '每日一次',
        startDate: options.start || new Date().toISOString().split('T')[0],
        endDate: options.end,
        prescribedBy: '医生',
      };

      const result = await apiCall<any>(`/api/r/${options.patient}/items/medications`, 'POST', medData);

      if (result.contraindications && result.contraindications.length > 0) {
        console.log('\n⚠️  存在配伍禁忌!');
        result.contraindications.forEach((c: any) => {
          console.log(`  - ${c.drugA} ↔ ${c.drugB}: ${c.description} (${c.severity})`);
        });
        console.log('\n药品未添加，原因: 存在配伍禁忌');
      } else {
        console.log('\n用药添加成功!');
        console.log('药品:', result.medication.name);
        console.log('剂量:', result.medication.dosage);
      }
    } catch (error) {
      console.error('添加用药失败:', (error as Error).message);
    }
  });

medication
  .command('list <patientId>')
  .description('列出患者用药')
  .action(async (patientId) => {
    try {
      const meds = await apiCall<unknown[]>(`/api/r/${patientId}/items/medications`);
      console.log(`\n患者用药 (${meds.length} 项):\n`);
      meds.forEach((m: any, i: number) => {
        console.log(`${i + 1}. ${m.name}`);
        console.log(`   剂量: ${m.dosage} | 频次: ${m.frequency}`);
        if (m.remainingDays !== undefined) {
          console.log(`   剩余天数: ${m.remainingDays} 天`);
        }
        console.log('');
      });
    } catch (error) {
      console.error('获取用药列表失败:', (error as Error).message);
    }
  });

const stats = program.command('stats').description('统计分析');

stats
  .command('control')
  .description('查看控制率')
  .option('-p, --period <period>', '统计周期 (month/quarter/year)', 'month')
  .action(async (options) => {
    try {
      const rates = await apiCall<any>(`/api/statistics/control-rates?period=${options.period}`);
      console.log('\n控制率统计:');
      console.log(`\n统计周期: ${rates.period}`);
      console.log(`总患者数: ${rates.totalPatients}`);
      console.log(`达标患者数: ${rates.controlledPatients}`);
      console.log(`\n高血压控制率: ${rates.hypertension}%`);
      console.log(`糖尿病控制率: ${rates.diabetes}%`);
      console.log(`总体控制率: ${rates.overall}%`);
      console.log(`\n计算时间: ${rates.calculatedAt}`);
    } catch (error) {
      console.error('获取控制率失败:', (error as Error).message);
    }
  });

stats
  .command('completion')
  .description('查看随访完成率')
  .option('-p, --period <period>', '统计周期 (month/quarter/year)', 'month')
  .action(async (options) => {
    try {
      const completion = await apiCall<any>(`/api/statistics/followup-completion?period=${options.period}`);
      console.log('\n随访完成率统计:');
      console.log(`统计周期: ${completion.period}`);
      console.log(`应随访次数: ${completion.totalExpected}`);
      console.log(`已完成次数: ${completion.completed}`);
      console.log(`完成率: ${completion.rate}%`);
    } catch (error) {
      console.error('获取完成率失败:', (error as Error).message);
    }
  });

const budget = program.command('budget').description('预算管理');

budget
  .command('list')
  .description('列出所有预算')
  .action(async () => {
    try {
      const budgets = await apiCall<unknown[]>('/api/budgets');
      console.log(`\n共 ${budgets.length} 个预算:\n`);
      budgets.forEach((b: any) => {
        console.log(`ID: ${b.id}`);
        console.log(`  周期: ${b.period}`);
        console.log(`  总额: ${b.totalAmount} 元`);
        console.log(`  项目数: ${b.items.length}`);
        console.log('');
      });
    } catch (error) {
      console.error('获取预算列表失败:', (error as Error).message);
    }
  });

budget
  .command('create')
  .description('创建预算')
  .option('--period <period>', '周期', '2024-Q1')
  .option('--followup <amount>', '随访预算', parseInt)
  .option('--medication <amount>', '用药预算', parseInt)
  .option('--lab <amount>', '检验预算', parseInt)
  .action(async (options) => {
    try {
      const items = [];
      if (options.followup) {
        items.push({ name: '随访', amount: options.followup, settledAmount: 0, category: 'followup', status: 'pending' });
      }
      if (options.medication) {
        items.push({ name: '用药', amount: options.medication, settledAmount: 0, category: 'medication', status: 'pending' });
      }
      if (options.lab) {
        items.push({ name: '检验', amount: options.lab, settledAmount: 0, category: 'lab', status: 'pending' });
      }

      const budget = await apiCall<any>('/api/budgets', 'POST', {
        period: options.period,
        items,
      });
      console.log('\n预算创建成功!');
      console.log('ID:', budget.id);
      console.log('总额:', budget.totalAmount);
    } catch (error) {
      console.error('创建预算失败:', (error as Error).message);
    }
  });

budget
  .command('adjust <id>')
  .description('调整预算总额')
  .option('--amount <amount>', '新总额', parseInt)
  .action(async (id, options) => {
    try {
      const budget = await apiCall<any>(`/api/budgets/${id}`, 'PUT', {
        totalAmount: options.amount,
      });
      console.log('\n预算调整成功!');
      console.log('新总额:', budget.totalAmount);
      console.log('项目明细:');
      budget.items.forEach((item: any) => {
        console.log(`  ${item.name}: ${item.amount} 元 (${item.status})`);
      });
    } catch (error) {
      console.error('调整预算失败:', (error as Error).message);
    }
  });

const reminders = program.command('reminders').description('续药提醒');

reminders
  .command('list')
  .description('列出待发送的提醒')
  .action(async () => {
    try {
      const pending = await apiCall<unknown[]>('/api/reminders/pending');
      console.log(`\n待发送的提醒 (${pending.length} 条):\n`);
      pending.forEach((r: any) => {
        console.log(`ID: ${r.id}`);
        console.log(`  患者ID: ${r.patientId}`);
        console.log(`  药品: ${r.medicationName}`);
        console.log(`  到期日期: ${r.expirationDate}`);
        console.log('');
      });
    } catch (error) {
      console.error('获取提醒失败:', (error as Error).message);
    }
  });

reminders
  .command('process')
  .description('处理续药提醒 (生成+发送)')
  .action(async () => {
    try {
      const result = await apiCall<any>('/api/reminders/process', 'POST');
      console.log('\n续药提醒处理完成:');
      console.log(`新生成: ${result.generatedCount} 条`);
      console.log(`发送成功: ${result.sentCount} 条`);
      console.log(`发送失败: ${result.failedCount} 条`);
      if (result.failedCount > 0) {
        console.log('\n注意: 发送失败的提醒将在下次调度周期重试');
      }
    } catch (error) {
      console.error('处理提醒失败:', (error as Error).message);
    }
  });

const config = program.command('config').description('实体配置');

config
  .command('list')
  .description('列出所有配置')
  .option('-t, --type <type>', '配置类型 (chronic_disease/medication/followup_template)')
  .action(async (options) => {
    try {
      const url = options.type ? `/api/configs?type=${options.type}` : '/api/configs';
      const configs = await apiCall<unknown[]>(url);
      console.log(`\n配置列表 (${configs.length} 项):\n`);
      configs.forEach((c: any) => {
        console.log(`ID: ${c.id}`);
        console.log(`  类型: ${c.type}`);
        console.log(`  代码: ${c.code}`);
        console.log(`  名称: ${c.name}`);
        console.log('');
      });
    } catch (error) {
      console.error('获取配置失败:', (error as Error).message);
    }
  });

program.parseAsync(process.argv);
