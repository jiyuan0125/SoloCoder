import { Router, Request, Response } from 'express';
import { db, ContractTemplate } from '../database';
import { isValidTemplateType, extractVariables } from '../utils';

const router = Router();

router.get('/', (req: Request, res: Response): void => {
  const type = req.query.type as string | undefined;
  
  let query = 'SELECT * FROM contract_templates WHERE 1=1';
  const params: any[] = [];
  
  if (type) {
    query += ' AND type = ?';
    params.push(type);
  }
  
  query += ' ORDER BY created_at DESC';
  
  const templates = db.prepare(query).all(...params) as ContractTemplate[];
  res.json(templates);
});

router.get('/:id', (req: Request, res: Response): void => {
  const template = db
    .prepare('SELECT * FROM contract_templates WHERE id = ?')
    .get(req.params.id) as ContractTemplate | undefined;

  if (!template) {
    res.status(404).json({ error: '模板不存在' });
    return;
  }

  const variables = extractVariables(template.content);
  res.json({ ...template, variables });
});

router.post('/', (req: Request, res: Response): void => {
  const { name, type, version, content } = req.body as Partial<ContractTemplate>;

  if (!name || !type || !version || !content) {
    res.status(400).json({ error: '缺少必要字段' });
    return;
  }

  if (!isValidTemplateType(type)) {
    res.status(400).json({ error: '无效的模板类型，必须是：采购、销售、服务协议、保密协议' });
    return;
  }

  const existing = db
    .prepare('SELECT * FROM contract_templates WHERE name = ? AND version = ?')
    .get(name, version) as ContractTemplate | undefined;

  if (existing) {
    res.status(409).json({ error: '模板名称加版本号已存在' });
    return;
  }

  try {
    const result = db
      .prepare(
        'INSERT INTO contract_templates (name, type, version, content) VALUES (?, ?, ?, ?)',
      )
      .run(name, type, version, content);

    const template = db
      .prepare('SELECT * FROM contract_templates WHERE id = ?')
      .get(result.lastInsertRowid) as ContractTemplate;

    const variables = extractVariables(template.content);
    res.status(201).json({ ...template, variables });
  } catch (err) {
    res.status(500).json({ error: '创建模板失败' });
  }
});

router.put('/:id', (req: Request, res: Response): void => {
  const template = db
    .prepare('SELECT * FROM contract_templates WHERE id = ?')
    .get(req.params.id) as ContractTemplate | undefined;

  if (!template) {
    res.status(404).json({ error: '模板不存在' });
    return;
  }

  const { name, type, version, content } = req.body as Partial<ContractTemplate>;

  if (type && !isValidTemplateType(type)) {
    res.status(400).json({ error: '无效的模板类型，必须是：采购、销售、服务协议、保密协议' });
    return;
  }

  if (name && version && (name !== template.name || version !== template.version)) {
    const existing = db
      .prepare('SELECT * FROM contract_templates WHERE name = ? AND version = ? AND id != ?')
      .get(name, version, req.params.id) as ContractTemplate | undefined;

    if (existing) {
      res.status(409).json({ error: '模板名称加版本号已存在' });
      return;
    }
  }

  const newName = name ?? template.name;
  const newType = type ?? template.type;
  const newVersion = version ?? template.version;
  const newContent = content ?? template.content;

  try {
    db
      .prepare(
        'UPDATE contract_templates SET name = ?, type = ?, version = ?, content = ? WHERE id = ?',
      )
      .run(newName, newType, newVersion, newContent, req.params.id);

    const updated = db
      .prepare('SELECT * FROM contract_templates WHERE id = ?')
      .get(req.params.id) as ContractTemplate;

    const variables = extractVariables(updated.content);
    res.json({ ...updated, variables });
  } catch (err) {
    res.status(500).json({ error: '更新模板失败' });
  }
});

router.delete('/:id', (req: Request, res: Response): void => {
  const template = db
    .prepare('SELECT * FROM contract_templates WHERE id = ?')
    .get(req.params.id) as ContractTemplate | undefined;

  if (!template) {
    res.status(404).json({ error: '模板不存在' });
    return;
  }

  const relatedContracts = db
    .prepare('SELECT COUNT(*) as count FROM contracts WHERE template_id = ?')
    .get(req.params.id) as { count: number };

  if (relatedContracts.count > 0) {
    res.status(400).json({ error: '该模板已有合同使用，无法删除' });
    return;
  }

  db.prepare('DELETE FROM contract_templates WHERE id = ?').run(req.params.id);
  res.status(204).send();
});

export default router;
