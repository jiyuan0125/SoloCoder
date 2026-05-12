import { Router, Request, Response } from 'express';
import { db, Contract, ContractTemplate, SigningTask, ContractStatus } from '../database';
import {
  validateVariables,
  fillTemplateVariables,
  isValidStatusTransition,
  getApproverRole,
} from '../utils';

const router = Router();

router.get('/', (req: Request, res: Response): void => {
  const status = req.query.status as string | undefined;
  const templateType = req.query.templateType as string | undefined;
  const isSupplement = req.query.isSupplement as string | undefined;

  let query = 'SELECT * FROM contracts WHERE 1=1';
  const params: any[] = [];

  if (status) {
    query += ' AND status = ?';
    params.push(status);
  }

  if (templateType) {
    query += ' AND template_type = ?';
    params.push(templateType);
  }

  if (isSupplement !== undefined) {
    query += ' AND is_supplement = ?';
    params.push(isSupplement === 'true' ? 1 : 0);
  }

  query += ' ORDER BY created_at DESC';

  const contracts = db.prepare(query).all(...params) as Contract[];
  res.json(contracts);
});

router.get('/:id', (req: Request, res: Response): void => {
  const contract = db
    .prepare('SELECT * FROM contracts WHERE id = ?')
    .get(req.params.id) as Contract | undefined;

  if (!contract) {
    res.status(404).json({ error: '合同不存在' });
    return;
  }

  const signingTask = db
    .prepare('SELECT * FROM signing_tasks WHERE contract_id = ? ORDER BY created_at DESC LIMIT 1')
    .get(contract.id) as SigningTask | undefined;

  res.json({
    ...contract,
    variables: JSON.parse(contract.variables),
    signingTask,
  });
});

router.post('/', (req: Request, res: Response): void => {
  const { templateId, variables } = req.body as {
    templateId: number;
    variables: Record<string, string | number>;
  };

  if (!templateId) {
    res.status(400).json({ error: '缺少模板ID' });
    return;
  }

  const template = db
    .prepare('SELECT * FROM contract_templates WHERE id = ?')
    .get(templateId) as ContractTemplate | undefined;

  if (!template) {
    res.status(404).json({ error: '模板不存在' });
    return;
  }

  const missingVars = validateVariables(template.content, variables || {});
  if (missingVars.length > 0) {
    res.status(400).json({
      error: '缺少必要变量',
      missingVariables: missingVars,
    });
    return;
  }

  const amount = typeof variables.amount === 'number' ? variables.amount : Number(variables.amount);
  if (isNaN(amount) || amount < 0) {
    res.status(400).json({ error: '合同金额必须是非负数字' });
    return;
  }

  const content = fillTemplateVariables(template.content, variables);

  try {
    const result = db
      .prepare(
        `INSERT INTO contracts 
         (template_id, template_name, template_version, template_type, template_content, variables, content, amount, status)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, '草稿')`,
      )
      .run(
        template.id,
        template.name,
        template.version,
        template.type,
        template.content,
        JSON.stringify(variables),
        content,
        amount,
      );

    const contract = db
      .prepare('SELECT * FROM contracts WHERE id = ?')
      .get(result.lastInsertRowid) as Contract;

    res.status(201).json({
      ...contract,
      variables: JSON.parse(contract.variables),
    });
  } catch (err) {
    res.status(500).json({ error: '创建合同失败' });
  }
});

router.put('/:id', (req: Request, res: Response): void => {
  const contract = db
    .prepare('SELECT * FROM contracts WHERE id = ?')
    .get(req.params.id) as Contract | undefined;

  if (!contract) {
    res.status(404).json({ error: '合同不存在' });
    return;
  }

  if (contract.status === '审批中') {
    res.status(409).json({ error: '请先撤回审批' });
    return;
  }

  if (['已批准', '签署中', '已签署', '已归档', '已作废'].includes(contract.status)) {
    res.status(400).json({
      error: '当前状态无法修改',
      currentStatus: contract.status,
    });
    return;
  }

  const { variables } = req.body as {
    variables: Record<string, string | number>;
  };

  if (!variables) {
    res.status(400).json({ error: '缺少变量数据' });
    return;
  }

  const missingVars = validateVariables(contract.template_content, variables);
  if (missingVars.length > 0) {
    res.status(400).json({
      error: '缺少必要变量',
      missingVariables: missingVars,
    });
    return;
  }

  const amount = typeof variables.amount === 'number' ? variables.amount : Number(variables.amount);
  if (isNaN(amount) || amount < 0) {
    res.status(400).json({ error: '合同金额必须是非负数字' });
    return;
  }

  const content = fillTemplateVariables(contract.template_content, variables);

  try {
    db
      .prepare(
        'UPDATE contracts SET variables = ?, content = ?, amount = ? WHERE id = ?',
      )
      .run(JSON.stringify(variables), content, amount, req.params.id);

    const updated = db
      .prepare('SELECT * FROM contracts WHERE id = ?')
      .get(req.params.id) as Contract;

    res.json({
      ...updated,
      variables: JSON.parse(updated.variables),
    });
  } catch (err) {
    res.status(500).json({ error: '更新合同失败' });
  }
});

router.post('/:id/submit-approval', (req: Request, res: Response): void => {
  const contract = db
    .prepare('SELECT * FROM contracts WHERE id = ?')
    .get(req.params.id) as Contract | undefined;

  if (!contract) {
    res.status(404).json({ error: '合同不存在' });
    return;
  }

  if (!isValidStatusTransition(contract.status, '审批中')) {
    res.status(400).json({
      error: '非法状态跳转',
      currentStatus: contract.status,
    });
    return;
  }

  const approverRole = getApproverRole(contract.amount);

  try {
    db.prepare('UPDATE contracts SET status = ? WHERE id = ?').run('审批中', contract.id);

    db
      .prepare(
        'INSERT INTO approval_records (contract_id, approver_role, approved) VALUES (?, ?, 0)',
      )
      .run(contract.id, approverRole);

    const updated = db
      .prepare('SELECT * FROM contracts WHERE id = ?')
      .get(contract.id) as Contract;

    res.json({
      ...updated,
      variables: JSON.parse(updated.variables),
      approverRole,
    });
  } catch (err) {
    res.status(500).json({ error: '提交审批失败' });
  }
});

router.post('/:id/withdraw-approval', (req: Request, res: Response): void => {
  const contract = db
    .prepare('SELECT * FROM contracts WHERE id = ?')
    .get(req.params.id) as Contract | undefined;

  if (!contract) {
    res.status(404).json({ error: '合同不存在' });
    return;
  }

  if (contract.status !== '审批中') {
    res.status(400).json({
      error: '只有审批中的合同可以撤回',
      currentStatus: contract.status,
    });
    return;
  }

  try {
    db.prepare('UPDATE contracts SET status = ? WHERE id = ?').run('草稿', contract.id);
    db.prepare('DELETE FROM approval_records WHERE contract_id = ? AND approved = 0').run(contract.id);

    const updated = db
      .prepare('SELECT * FROM contracts WHERE id = ?')
      .get(contract.id) as Contract;

    res.json({
      ...updated,
      variables: JSON.parse(updated.variables),
    });
  } catch (err) {
    res.status(500).json({ error: '撤回审批失败' });
  }
});

router.post('/:id/approve', (req: Request, res: Response): void => {
  const contract = db
    .prepare('SELECT * FROM contracts WHERE id = ?')
    .get(req.params.id) as Contract | undefined;

  if (!contract) {
    res.status(404).json({ error: '合同不存在' });
    return;
  }

  if (contract.status !== '审批中') {
    res.status(400).json({
      error: '只有审批中的合同可以审批',
      currentStatus: contract.status,
    });
    return;
  }

  const { approverName } = req.body as { approverName?: string };

  const approverRole = getApproverRole(contract.amount);

  try {
    const approveTransaction = db.transaction(() => {
      db.prepare('UPDATE contracts SET status = ?, approved_amount = ? WHERE id = ?').run(
        '已批准',
        contract.amount,
        contract.id,
      );

      db
        .prepare(
          'UPDATE approval_records SET approved = 1, approver_name = ? WHERE contract_id = ? AND approved = 0',
        )
        .run(approverName || null, contract.id);

      try {
        db
          .prepare(
            'INSERT INTO signing_tasks (contract_id, contract_amount, status) VALUES (?, ?, ?)',
          )
          .run(contract.id, contract.amount, '待签署');
      } catch (signingErr) {
        db.prepare('UPDATE contracts SET needs_manual_processing = 1 WHERE id = ?').run(contract.id);
      }
    });

    approveTransaction();

    const updated = db
      .prepare('SELECT * FROM contracts WHERE id = ?')
      .get(contract.id) as Contract;

    const signingTask = db
      .prepare('SELECT * FROM signing_tasks WHERE contract_id = ? ORDER BY created_at DESC LIMIT 1')
      .get(contract.id) as SigningTask | undefined;

    res.json({
      ...updated,
      variables: JSON.parse(updated.variables),
      signingTask,
    });
  } catch (err) {
    res.status(500).json({ error: '审批失败' });
  }
});

router.post('/:id/start-signing', (req: Request, res: Response): void => {
  const contract = db
    .prepare('SELECT * FROM contracts WHERE id = ?')
    .get(req.params.id) as Contract | undefined;

  if (!contract) {
    res.status(404).json({ error: '合同不存在' });
    return;
  }

  if (!isValidStatusTransition(contract.status, '签署中')) {
    res.status(400).json({
      error: '非法状态跳转',
      currentStatus: contract.status,
    });
    return;
  }

  const signingTask = db
    .prepare('SELECT * FROM signing_tasks WHERE contract_id = ? ORDER BY created_at DESC LIMIT 1')
    .get(contract.id) as SigningTask | undefined;

  if (!signingTask) {
    res.status(400).json({ error: '签署任务不存在' });
    return;
  }

  try {
    db.prepare('UPDATE contracts SET status = ? WHERE id = ?').run('签署中', contract.id);
    db.prepare('UPDATE signing_tasks SET status = ? WHERE id = ?').run('签署中', signingTask.id);

    const updated = db
      .prepare('SELECT * FROM contracts WHERE id = ?')
      .get(contract.id) as Contract;

    const updatedTask = db
      .prepare('SELECT * FROM signing_tasks WHERE id = ?')
      .get(signingTask.id) as SigningTask;

    res.json({
      ...updated,
      variables: JSON.parse(updated.variables),
      signingTask: updatedTask,
    });
  } catch (err) {
    res.status(500).json({ error: '开始签署失败' });
  }
});

router.post('/:id/complete-signing', (req: Request, res: Response): void => {
  const contract = db
    .prepare('SELECT * FROM contracts WHERE id = ?')
    .get(req.params.id) as Contract | undefined;

  if (!contract) {
    res.status(404).json({ error: '合同不存在' });
    return;
  }

  if (!isValidStatusTransition(contract.status, '已签署')) {
    res.status(400).json({
      error: '非法状态跳转',
      currentStatus: contract.status,
    });
    return;
  }

  const signingTask = db
    .prepare('SELECT * FROM signing_tasks WHERE contract_id = ? ORDER BY created_at DESC LIMIT 1')
    .get(contract.id) as SigningTask | undefined;

  if (!signingTask) {
    res.status(400).json({ error: '签署任务不存在' });
    return;
  }

  try {
    db.prepare('UPDATE contracts SET status = ? WHERE id = ?').run('已签署', contract.id);
    db.prepare('UPDATE signing_tasks SET status = ? WHERE id = ?').run('已签署', signingTask.id);

    const updated = db
      .prepare('SELECT * FROM contracts WHERE id = ?')
      .get(contract.id) as Contract;

    const updatedTask = db
      .prepare('SELECT * FROM signing_tasks WHERE id = ?')
      .get(signingTask.id) as SigningTask;

    res.json({
      ...updated,
      variables: JSON.parse(updated.variables),
      signingTask: updatedTask,
    });
  } catch (err) {
    res.status(500).json({ error: '完成签署失败' });
  }
});

router.post('/:id/archive', (req: Request, res: Response): void => {
  const contract = db
    .prepare('SELECT * FROM contracts WHERE id = ?')
    .get(req.params.id) as Contract | undefined;

  if (!contract) {
    res.status(404).json({ error: '合同不存在' });
    return;
  }

  if (!isValidStatusTransition(contract.status, '已归档')) {
    res.status(400).json({
      error: '非法状态跳转',
      currentStatus: contract.status,
    });
    return;
  }

  try {
    db.prepare('UPDATE contracts SET status = ? WHERE id = ?').run('已归档', contract.id);

    const updated = db
      .prepare('SELECT * FROM contracts WHERE id = ?')
      .get(contract.id) as Contract;

    res.json({
      ...updated,
      variables: JSON.parse(updated.variables),
    });
  } catch (err) {
    res.status(500).json({ error: '归档失败' });
  }
});

router.post('/:id/void', (req: Request, res: Response): void => {
  const contract = db
    .prepare('SELECT * FROM contracts WHERE id = ?')
    .get(req.params.id) as Contract | undefined;

  if (!contract) {
    res.status(404).json({ error: '合同不存在' });
    return;
  }

  if (contract.status === '已作废') {
    res.status(400).json({
      error: '已作废的合同不能恢复',
      currentStatus: contract.status,
    });
    return;
  }

  try {
    db.prepare('UPDATE contracts SET status = ? WHERE id = ?').run('已作废', contract.id);

    const updated = db
      .prepare('SELECT * FROM contracts WHERE id = ?')
      .get(contract.id) as Contract;

    res.json({
      ...updated,
      variables: JSON.parse(updated.variables),
    });
  } catch (err) {
    res.status(500).json({ error: '作废失败' });
  }
});

router.post('/:id/create-supplement', (req: Request, res: Response): void => {
  const originalContract = db
    .prepare('SELECT * FROM contracts WHERE id = ?')
    .get(req.params.id) as Contract | undefined;

  if (!originalContract) {
    res.status(404).json({ error: '原合同不存在' });
    return;
  }

  if (originalContract.status !== '已签署' && originalContract.status !== '已归档') {
    res.status(400).json({
      error: '只有已签署或已归档的合同可以创建补充协议',
      currentStatus: originalContract.status,
    });
    return;
  }

  const { variables } = req.body as {
    variables: Record<string, string | number>;
  };

  if (!variables) {
    res.status(400).json({ error: '缺少变量数据' });
    return;
  }

  const missingVars = validateVariables(originalContract.template_content, variables);
  if (missingVars.length > 0) {
    res.status(400).json({
      error: '缺少必要变量',
      missingVariables: missingVars,
    });
    return;
  }

  const amount = typeof variables.amount === 'number' ? variables.amount : Number(variables.amount);
  if (isNaN(amount) || amount < 0) {
    res.status(400).json({ error: '合同金额必须是非负数字' });
    return;
  }

  const content = fillTemplateVariables(originalContract.template_content, variables);

  try {
    const result = db
      .prepare(
        `INSERT INTO contracts 
         (template_id, template_name, template_version, template_type, template_content, variables, content, amount, status, is_supplement, parent_contract_id)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, '草稿', 1, ?)`,
      )
      .run(
        originalContract.template_id,
        originalContract.template_name,
        originalContract.template_version,
        originalContract.template_type,
        originalContract.template_content,
        JSON.stringify(variables),
        content,
        amount,
        originalContract.id,
      );

    const contract = db
      .prepare('SELECT * FROM contracts WHERE id = ?')
      .get(result.lastInsertRowid) as Contract;

    res.status(201).json({
      ...contract,
      variables: JSON.parse(contract.variables),
    });
  } catch (err) {
    res.status(500).json({ error: '创建补充协议失败' });
  }
});

export default router;
