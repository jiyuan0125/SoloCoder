import express, { Request, Response } from 'express';
import db from '../db';
import { generateUUID } from '../utils/uuid';

const router = express.Router();

router.post('/', (req: Request, res: Response) => {
  const { name, description } = req.body;
  
  if (!name) {
    return res.status(400).json({ error: 'service name is required' });
  }

  const now = new Date().toISOString();
  const id = generateUUID();

  const stmt = db.prepare(`
    INSERT INTO services (id, name, description, created_at, updated_at)
    VALUES (?, ?, ?, ?, ?)
  `);

  stmt.run(id, name, description || null, now, now);

  const result = db.prepare('SELECT * FROM services WHERE id = ?').get(id) as any;
  
  return res.status(201).json({
    id: result.id,
    name: result.name,
    description: result.description,
    createdAt: result.created_at,
    updatedAt: result.updated_at
  });
});

router.post('/:id/sla', (req: Request, res: Response) => {
  const { id } = req.params;
  const { availabilityTarget, responseTimeTarget, errorRateTarget } = req.body;

  const service = db.prepare('SELECT * FROM services WHERE id = ?').get(id);
  if (!service) {
    return res.status(404).json({ error: 'service not found' });
  }

  if (availabilityTarget === undefined || responseTimeTarget === undefined || errorRateTarget === undefined) {
    return res.status(400).json({ error: 'availabilityTarget, responseTimeTarget, and errorRateTarget are required' });
  }

  const now = new Date().toISOString();
  const slaId = generateUUID();

  const existingSla = db.prepare('SELECT * FROM sla_definitions WHERE service_id = ?').get(id);
  if (existingSla) {
    const updateStmt = db.prepare(`
      UPDATE sla_definitions 
      SET availability_target = ?, response_time_target = ?, error_rate_target = ?, updated_at = ?
      WHERE service_id = ?
    `);
    updateStmt.run(availabilityTarget, responseTimeTarget, errorRateTarget, now, id);
    
    const updated = db.prepare('SELECT * FROM sla_definitions WHERE service_id = ?').get(id) as any;
    return res.json({
      id: updated.id,
      serviceId: updated.service_id,
      availabilityTarget: updated.availability_target,
      responseTimeTarget: updated.response_time_target,
      errorRateTarget: updated.error_rate_target,
      createdAt: updated.created_at,
      updatedAt: updated.updated_at
    });
  }

  const insertStmt = db.prepare(`
    INSERT INTO sla_definitions (id, service_id, availability_target, response_time_target, error_rate_target, created_at, updated_at)
    VALUES (?, ?, ?, ?, ?, ?, ?)
  `);
  insertStmt.run(slaId, id, availabilityTarget, responseTimeTarget, errorRateTarget, now, now);

  const result = db.prepare('SELECT * FROM sla_definitions WHERE id = ?').get(slaId) as any;
  return res.status(201).json({
    id: result.id,
    serviceId: result.service_id,
    availabilityTarget: result.availability_target,
    responseTimeTarget: result.response_time_target,
    errorRateTarget: result.error_rate_target,
    createdAt: result.created_at,
    updatedAt: result.updated_at
  });
});

router.get('/:id/sla', (req: Request, res: Response) => {
  const { id } = req.params;

  const service = db.prepare('SELECT * FROM services WHERE id = ?').get(id);
  if (!service) {
    return res.status(404).json({ error: 'service not found' });
  }

  const sla = db.prepare('SELECT * FROM sla_definitions WHERE service_id = ?').get(id) as any;
  if (!sla) {
    return res.status(404).json({ error: 'SLA not found for this service' });
  }

  return res.json({
    id: sla.id,
    serviceId: sla.service_id,
    availabilityTarget: sla.availability_target,
    responseTimeTarget: sla.response_time_target,
    errorRateTarget: sla.error_rate_target,
    createdAt: sla.created_at,
    updatedAt: sla.updated_at
  });
});

export default router;
