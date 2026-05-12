import { BaseDao } from './baseDao';
import { Step, CreateStepRequest, UpdateStepRequest, TransformationRule } from '../models/step';
import { v4 as uuidv4 } from 'uuid';

interface StepRow {
  id: string;
  pipeline_id: string;
  step_number: number;
  name: string;
  data_source_id: string;
  transformations: string;
  target_config: string;
  created_at: string;
  updated_at: string;
}

export class StepDao extends BaseDao {
  public async create(pipelineId: string, request: CreateStepRequest): Promise<Step> {
    const id = uuidv4();
    const now = new Date().toISOString();
    
    const sql = `
      INSERT INTO steps (id, pipeline_id, step_number, name, data_source_id, transformations, target_config, created_at, updated_at)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `;
    
    await this.run(sql, [
      id,
      pipelineId,
      request.stepNumber,
      request.name,
      request.dataSourceId,
      JSON.stringify(request.transformations),
      JSON.stringify(request.targetConfig),
      now,
      now
    ]);
    
    return this.findById(id) as Promise<Step>;
  }

  public async findById(id: string): Promise<Step | undefined> {
    const sql = `SELECT * FROM steps WHERE id = ?`;
    const row = await this.get<StepRow>(sql, [id]);
    
    if (!row) {
      return undefined;
    }
    
    return this.mapToStep(row);
  }

  public async findByPipelineId(pipelineId: string): Promise<Step[]> {
    const sql = `SELECT * FROM steps WHERE pipeline_id = ? ORDER BY step_number ASC`;
    const rows = await this.all<StepRow>(sql, [pipelineId]);
    
    return rows.map(row => this.mapToStep(row));
  }

  public async findByPipelineIdAndStepNumber(pipelineId: string, stepNumber: number): Promise<Step | undefined> {
    const sql = `SELECT * FROM steps WHERE pipeline_id = ? AND step_number = ?`;
    const row = await this.get<StepRow>(sql, [pipelineId, stepNumber]);
    
    if (!row) {
      return undefined;
    }
    
    return this.mapToStep(row);
  }

  public async update(id: string, request: UpdateStepRequest): Promise<Step | undefined> {
    const existingStep = await this.findById(id);
    if (!existingStep) {
      return undefined;
    }

    const now = new Date().toISOString();
    const updates: string[] = [];
    const params: any[] = [];

    if (request.name !== undefined) {
      updates.push('name = ?');
      params.push(request.name);
    }
    if (request.stepNumber !== undefined) {
      updates.push('step_number = ?');
      params.push(request.stepNumber);
    }
    if (request.dataSourceId !== undefined) {
      updates.push('data_source_id = ?');
      params.push(request.dataSourceId);
    }
    if (request.transformations !== undefined) {
      updates.push('transformations = ?');
      params.push(JSON.stringify(request.transformations));
    }
    if (request.targetConfig !== undefined) {
      updates.push('target_config = ?');
      params.push(JSON.stringify(request.targetConfig));
    }
    
    updates.push('updated_at = ?');
    params.push(now);
    params.push(id);

    const sql = `UPDATE steps SET ${updates.join(', ')} WHERE id = ?`;
    await this.run(sql, params);
    
    return this.findById(id);
  }

  public async delete(id: string): Promise<boolean> {
    const sql = `DELETE FROM steps WHERE id = ?`;
    await this.run(sql, [id]);
    return true;
  }

  public async deleteByPipelineId(pipelineId: string): Promise<boolean> {
    const sql = `DELETE FROM steps WHERE pipeline_id = ?`;
    await this.run(sql, [pipelineId]);
    return true;
  }

  private mapToStep(row: StepRow): Step {
    return {
      id: row.id,
      pipelineId: row.pipeline_id,
      stepNumber: row.step_number,
      name: row.name,
      dataSourceId: row.data_source_id,
      transformations: JSON.parse(row.transformations) as TransformationRule[],
      targetConfig: JSON.parse(row.target_config),
      createdAt: row.created_at,
      updatedAt: row.updated_at
    };
  }
}
