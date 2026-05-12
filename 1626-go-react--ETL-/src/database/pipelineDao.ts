import { BaseDao } from './baseDao';
import { Pipeline, CreatePipelineRequest, UpdatePipelineRequest } from '../models/pipeline';
import { v4 as uuidv4 } from 'uuid';

interface PipelineRow {
  id: string;
  name: string;
  description?: string;
  is_scheduled: number;
  schedule_cron?: string;
  created_at: string;
  updated_at: string;
}

export class PipelineDao extends BaseDao {
  public async create(request: CreatePipelineRequest): Promise<Pipeline> {
    const id = uuidv4();
    const now = new Date().toISOString();
    
    const sql = `
      INSERT INTO pipelines (id, name, description, is_scheduled, schedule_cron, created_at, updated_at)
      VALUES (?, ?, ?, ?, ?, ?, ?)
    `;
    
    await this.run(sql, [
      id,
      request.name,
      request.description || null,
      request.isScheduled ? 1 : 0,
      request.scheduleCron || null,
      now,
      now
    ]);
    
    return this.findById(id) as Promise<Pipeline>;
  }

  public async findById(id: string): Promise<Pipeline | undefined> {
    const sql = `SELECT * FROM pipelines WHERE id = ?`;
    const row = await this.get<PipelineRow>(sql, [id]);
    
    if (!row) {
      return undefined;
    }
    
    return this.mapToPipeline(row);
  }

  public async findAll(): Promise<Pipeline[]> {
    const sql = `SELECT * FROM pipelines ORDER BY created_at DESC`;
    const rows = await this.all<PipelineRow>(sql);
    
    return rows.map(row => this.mapToPipeline(row));
  }

  public async update(id: string, request: UpdatePipelineRequest): Promise<Pipeline | undefined> {
    const existingPipeline = await this.findById(id);
    if (!existingPipeline) {
      return undefined;
    }

    const now = new Date().toISOString();
    const updates: string[] = [];
    const params: any[] = [];

    if (request.name !== undefined) {
      updates.push('name = ?');
      params.push(request.name);
    }
    if (request.description !== undefined) {
      updates.push('description = ?');
      params.push(request.description);
    }
    if (request.isScheduled !== undefined) {
      updates.push('is_scheduled = ?');
      params.push(request.isScheduled ? 1 : 0);
    }
    if (request.scheduleCron !== undefined) {
      updates.push('schedule_cron = ?');
      params.push(request.scheduleCron);
    }
    
    updates.push('updated_at = ?');
    params.push(now);
    params.push(id);

    const sql = `UPDATE pipelines SET ${updates.join(', ')} WHERE id = ?`;
    await this.run(sql, params);
    
    return this.findById(id);
  }

  public async delete(id: string): Promise<boolean> {
    const sql = `DELETE FROM pipelines WHERE id = ?`;
    await this.run(sql, [id]);
    return true;
  }

  private mapToPipeline(row: PipelineRow): Pipeline {
    return {
      id: row.id,
      name: row.name,
      description: row.description,
      isScheduled: row.is_scheduled === 1,
      scheduleCron: row.schedule_cron,
      createdAt: row.created_at,
      updatedAt: row.updated_at
    };
  }
}
