import express, { Request, Response, NextFunction } from 'express';
import {
  createPatient,
  getPatient,
  getAllPatients,
  updatePatient,
  updatePatientRisk,
  manuallyDowngradeRisk,
  PatientValidationError,
} from './modules/patientModule';
import {
  createFollowUpRecord,
  getFollowUpRecordsByPatient,
  getAllFollowUpRecords,
  CreateFollowUpInput,
  DuplicateFollowUpError,
} from './modules/followUpModule';
import {
  addMedication,
  getPatientMedications,
  removeMedication,
  getPendingReminders,
  processRefillReminders,
  initBuiltinContraindications,
  getAllContraindications,
} from './modules/medicationModule';
import {
  calculateControlRates,
  calculateFollowUpCompletionRate,
  getPatientStatistics,
  getDiseaseStatistics,
} from './modules/statisticsModule';
import {
  createBudget,
  getBudget,
  getAllBudgets,
  adjustBudget,
  settleBudgetItem,
  addBudgetItem,
  removeBudgetItem,
} from './modules/budgetModule';
import {
  initDefaultEntityConfigs,
  getEntityConfigsByType,
  getAllEntityConfigs,
  createEntityConfig,
  getEntityConfig,
  updateEntityConfig,
} from './modules/entityConfigModule';
import { ChronicDiseaseType, RiskLevel, Budget, EntityConfig } from './models/types';

const app = express();
app.use(express.json());

initBuiltinContraindications();
initDefaultEntityConfigs();

interface ApiError {
  status: number;
  message: string;
}

function handleError(
  err: Error,
  req: Request,
  res: Response,
  next: NextFunction
): void {
  if (err instanceof PatientValidationError) {
    res.status(400).json({ error: err.message });
    return;
  }
  if (err instanceof DuplicateFollowUpError) {
    res.status(409).json({ error: err.message });
    return;
  }
  console.error(err);
  res.status(500).json({ error: '服务器内部错误' });
}

app.use(handleError);

app.get('/health', (req: Request, res: Response) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

app.get('/api/r', (req: Request, res: Response) => {
  const patients = getAllPatients();
  res.json(patients);
});

app.post('/api/r', (req: Request, res: Response, next: NextFunction) => {
  try {
    const body = req.body;

    if (!body.chronicDiseases || body.chronicDiseases.length === 0) {
      throw new PatientValidationError('慢病类型不能为空');
    }

    const diagnosisDate = new Date(body.diagnosisDate);

    const patient = createPatient(
      body.basicInfo,
      body.chronicDiseases as ChronicDiseaseType[],
      diagnosisDate,
      body.medicalHistory || [],
      body.initialMetrics
    );

    res.status(201).json(patient);
  } catch (err) {
    next(err);
  }
});

app.get('/api/r/:id', (req: Request, res: Response) => {
  const patient = getPatient(req.params.id);
  if (!patient) {
    res.status(404).json({ error: '患者不存在' });
    return;
  }
  res.json(patient);
});

app.put('/api/r/:id', (req: Request, res: Response, next: NextFunction) => {
  try {
    const patient = updatePatient(req.params.id, req.body);
    if (!patient) {
      res.status(404).json({ error: '患者不存在' });
      return;
    }
    res.json(patient);
  } catch (err) {
    next(err);
  }
});

app.post('/api/r/:id/actions/approve', (req: Request, res: Response, next: NextFunction) => {
  try {
    const { newRiskLevel } = req.body;
    const patient = updatePatientRisk(req.params.id, newRiskLevel as RiskLevel);
    if (!patient) {
      res.status(404).json({ error: '患者不存在' });
      return;
    }
    res.json({
      message: '风险等级升级成功',
      patient,
    });
  } catch (err) {
    next(err);
  }
});

app.post('/api/r/:id/actions/reject', (req: Request, res: Response) => {
  res.json({
    message: '操作取消 - 风险等级保持不变',
  });
});

app.post('/api/r/:id/actions/cancel', (req: Request, res: Response) => {
  res.json({
    message: '操作已取消',
  });
});

app.post('/api/r/:id/actions/downgrade-risk', (req: Request, res: Response, next: NextFunction) => {
  try {
    const { newRiskLevel } = req.body;
    const patient = manuallyDowngradeRisk(req.params.id, newRiskLevel as RiskLevel);
    if (!patient) {
      res.status(404).json({ error: '患者不存在' });
      return;
    }
    res.json({
      message: '风险等级人工降级成功',
      patient,
    });
  } catch (err) {
    next(err);
  }
});

app.get('/api/r/:id/items', (req: Request, res: Response) => {
  const patient = getPatient(req.params.id);
  if (!patient) {
    res.status(404).json({ error: '患者不存在' });
    return;
  }

  const followUps = getFollowUpRecordsByPatient(req.params.id);
  const medications = getPatientMedications(req.params.id);
  const stats = getPatientStatistics(req.params.id);

  res.json({
    patient,
    followUps,
    medications,
    statistics: stats,
  });
});

app.post('/api/r/:id/items', (req: Request, res: Response, next: NextFunction) => {
  try {
    const body = req.body;
    const input: CreateFollowUpInput = {
      patientId: req.params.id,
      date: new Date(body.date),
      symptoms: body.symptoms || [],
      medications: body.medications || [],
      metrics: body.metrics,
      lifestyleAssessment: body.lifestyleAssessment,
      medicationTaken: body.medicationTaken,
      notes: body.notes,
    };

    const result = createFollowUpRecord(input);
    res.status(201).json(result);
  } catch (err) {
    next(err);
  }
});

app.get('/api/r/:id/items/followups', (req: Request, res: Response) => {
  const followUps = getFollowUpRecordsByPatient(req.params.id);
  res.json(followUps);
});

app.get('/api/r/:id/items/medications', (req: Request, res: Response) => {
  const medications = getPatientMedications(req.params.id);
  res.json(medications);
});

app.post('/api/r/:id/items/medications', (req: Request, res: Response, next: NextFunction) => {
  try {
    const body = req.body;
    const medicationData = {
      name: body.name,
      genericName: body.genericName || '',
      dosage: body.dosage,
      frequency: body.frequency,
      startDate: new Date(body.startDate),
      endDate: body.endDate ? new Date(body.endDate) : undefined,
      prescribedBy: body.prescribedBy || 'system',
    };

    const result = addMedication(req.params.id, medicationData);

    if (result.contraindications.length > 0) {
      res.status(400).json({
        error: '存在配伍禁忌',
        contraindications: result.contraindications,
        medication: result.medication,
      });
      return;
    }

    res.status(201).json(result);
  } catch (err) {
    next(err);
  }
});

app.delete('/api/r/:id/items/medications/:medicationId', (req: Request, res: Response) => {
  const success = removeMedication(req.params.id, req.params.medicationId);
  if (!success) {
    res.status(404).json({ error: '用药记录不存在' });
    return;
  }
  res.json({ message: '用药已删除' });
});

app.get('/api/statistics/control-rates', (req: Request, res: Response) => {
  const period = (req.query.period as 'month' | 'quarter' | 'year') || 'month';
  const rates = calculateControlRates(period);
  res.json(rates);
});

app.get('/api/statistics/followup-completion', (req: Request, res: Response) => {
  const period = (req.query.period as 'month' | 'quarter' | 'year') || 'month';
  const completion = calculateFollowUpCompletionRate(period);
  res.json(completion);
});

app.get('/api/statistics/disease/:diseaseType', (req: Request, res: Response) => {
  const stats = getDiseaseStatistics(req.params.diseaseType as ChronicDiseaseType);
  res.json(stats);
});

app.post('/api/reminders/process', (req: Request, res: Response) => {
  const result = processRefillReminders();
  res.json({
    generatedCount: result.generated.length,
    sentCount: result.sentResults.filter(r => r.success).length,
    failedCount: result.sentResults.filter(r => !r.success).length,
    details: result,
  });
});

app.get('/api/reminders/pending', (req: Request, res: Response) => {
  const reminders = getPendingReminders();
  res.json(reminders);
});

app.get('/api/contraindications', (req: Request, res: Response) => {
  const contraindications = getAllContraindications();
  res.json(contraindications);
});

app.get('/api/budgets', (req: Request, res: Response) => {
  const budgets = getAllBudgets();
  res.json(budgets);
});

app.post('/api/budgets', (req: Request, res: Response, next: NextFunction) => {
  try {
    const body = req.body;
    const budget = createBudget(body.period, body.totalAmount || 0, body.items || []);
    res.status(201).json(budget);
  } catch (err) {
    next(err);
  }
});

app.get('/api/budgets/:id', (req: Request, res: Response) => {
  const budget = getBudget(req.params.id);
  if (!budget) {
    res.status(404).json({ error: '预算不存在' });
    return;
  }
  res.json(budget);
});

app.put('/api/budgets/:id', (req: Request, res: Response, next: NextFunction) => {
  try {
    const { totalAmount } = req.body;
    const budget = adjustBudget(req.params.id, totalAmount);
    if (!budget) {
      res.status(404).json({ error: '预算不存在' });
      return;
    }
    res.json(budget);
  } catch (err) {
    next(err);
  }
});

app.post('/api/budgets/:id/items', (req: Request, res: Response, next: NextFunction) => {
  try {
    const budget = addBudgetItem(req.params.id, req.body);
    if (!budget) {
      res.status(404).json({ error: '预算不存在' });
      return;
    }
    res.status(201).json(budget);
  } catch (err) {
    next(err);
  }
});

app.post('/api/budgets/:id/items/:itemId/settle', (req: Request, res: Response, next: NextFunction) => {
  try {
    const budget = settleBudgetItem(req.params.id, req.params.itemId);
    if (!budget) {
      res.status(404).json({ error: '预算或项目不存在' });
      return;
    }
    res.json(budget);
  } catch (err) {
    next(err);
  }
});

app.delete('/api/budgets/:id/items/:itemId', (req: Request, res: Response, next: NextFunction) => {
  try {
    const budget = removeBudgetItem(req.params.id, req.params.itemId);
    if (!budget) {
      res.status(404).json({ error: '预算或项目不存在' });
      return;
    }
    res.json(budget);
  } catch (err) {
    next(err);
  }
});

app.get('/api/configs', (req: Request, res: Response) => {
  const type = req.query.type as EntityConfig['type'] | undefined;
  if (type) {
    const configs = getEntityConfigsByType(type);
    res.json(configs);
  } else {
    const configs = getAllEntityConfigs();
    res.json(configs);
  }
});

app.post('/api/configs', (req: Request, res: Response, next: NextFunction) => {
  try {
    const body = req.body;
    const config = createEntityConfig(
      body.type as EntityConfig['type'],
      body.code,
      body.name,
      body.config || {}
    );
    res.status(201).json(config);
  } catch (err) {
    next(err);
  }
});

app.get('/api/configs/:id', (req: Request, res: Response) => {
  const config = getEntityConfig(req.params.id);
  if (!config) {
    res.status(404).json({ error: '配置不存在' });
    return;
  }
  res.json(config);
});

app.put('/api/configs/:id', (req: Request, res: Response, next: NextFunction) => {
  try {
    const config = updateEntityConfig(req.params.id, req.body);
    if (!config) {
      res.status(404).json({ error: '配置不存在' });
      return;
    }
    res.json(config);
  } catch (err) {
    next(err);
  }
});

app.get('/api/followups', (req: Request, res: Response) => {
  const followUps = getAllFollowUpRecords();
  res.json(followUps);
});

function parsePort(): number {
  const args = process.argv.slice(2);
  for (let i = 0; i < args.length; i++) {
    if ((args[i] === '-p' || args[i] === '--port') && i + 1 < args.length) {
      return parseInt(args[i + 1], 10);
    }
    if (args[i].startsWith('--port=')) {
      return parseInt(args[i].split('=')[1], 10);
    }
  }

  const envPort = process.env.PORT || process.env.SERVER_PORT;
  if (envPort) {
    return parseInt(envPort, 10);
  }

  return 3000;
}

const PORT = parsePort();

app.listen(PORT, () => {
  console.log(`慢病管理系统服务端已启动，监听端口: ${PORT}`);
  console.log(`健康检查: http://localhost:${PORT}/health`);
});

export default app;
