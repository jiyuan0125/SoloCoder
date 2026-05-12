import { BaseDao } from './baseDao';
import { PipelineExecution, StepExecution, ExecutionStatus, StepExecutionStatus } from '../models/execution';
import { v4 as uuidv4 } from 'uuid';

interface PipelineExecutionRow {
  id: string;
  pipeline_id: string;
  status: string;
  start_time: string;
  end_time?: string;
  failed_step_number?: number;
  error_message?: string;
  created_at: string;
}

interface StepExecutionRow {
  id: string;
  execution_id: string;
  pipeline_id: string;
  step_number: number;
  status: string;
  start_time: string;
  end_time?: string;
  records_processed: number;
  error_message?: string;
  data_consumed: number;
}

export class ExecutionDao extends BaseDao {
  public async createPipelineExecution(pipelineId: string): Promise<PipelineExecution> {
    const id = uuidv4();
    const now = new Date().toISOString();
    
    const sql = `
      INSERT INTO pipeline_executions (id, pipeline_id, status, start_time, created_at)
      VALUES (?, ?, ?, ?, ?)
    `;
    
    await this.run(sql, [
      id,
      pipelineId,
      'running' as ExecutionStatus,
      now,
      now
    ]);
    
    return this.findPipelineExecutionById(id) as Promise<PipelineExecution>;
  }

  public async findPipelineExecutionById(id: string): Promise<PipelineExecution | undefined> {
    const sql = `SELECT * FROM pipeline_executions WHERE id = ?`;
    const row = await this.get<PipelineExecutionRow>(sql, [id]);
    
    if (!row) {
      return undefined;
    }
    
    return this.mapToPipelineExecution(row);
  }

  public async findPipelineExecutionsByPipelineId(pipelineId: string): Promise<PipelineExecution[]> {
    const sql = `SELECT * FROM pipeline_executions WHERE pipeline_id = ? ORDER BY created_at DESC`;
    const rows = await this.all<PipelineExecutionRow>(sql, [pipelineId]);
    
    return rows.map(row => this.mapToPipelineExecution(row));
  }

  public async findRunningExecutionByPipelineId(pipelineId: string): Promise<PipelineExecution | undefined> {
    const sql = `SELECT * FROM pipeline_executions WHERE pipeline_id = ? AND status = 'running'`;
    const row = await this.get<PipelineExecutionRow>(sql, [pipelineId]);
    
    if (!row) {
      return undefined;
    }
    
    return this.mapToPipelineExecution(row);
  }

  public async updatePipelineExecutionStatus(
    id: string,
    status: ExecutionStatus,
    failedStepNumber?: number,
    errorMessage?: string
  ): Promise<PipelineExecution | undefined> {
    const now = new Date().toISOString();
    const updates: string[] = ['status = ?'];
    const params: any[] = [status];

    if (status === 'completed' || status === 'failed') {
      updates.push('end_time = ?');
      params.push(now);
    }

    if (failedStepNumber !== undefined) {
      updates.push('failed_step_number = ?');
      params.push(failedStepNumber);
    }

    if (errorMessage !== undefined) {
      updates.push('error_message = ?');
      params.push(errorMessage);
    }

    params.push(id);

    const sql = `UPDATE pipeline_executions SET ${updates.join(', ')} WHERE id = ?`;
    await this.run(sql, params);
    
    return this.findPipelineExecutionById(id);
  }

  public async createStepExecution(
    executionId: string,
    pipelineId: string,
    stepNumber: number
  ): Promise<StepExecution> {
    const id = uuidv4();
    const now = new Date().toISOString();
    
    const sql = `
      INSERT INTO step_executions (id, execution_id, pipeline_id, step_number, status, start_time, records_processed, data_consumed)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `;
    
    await this.run(sql, [
      id,
      executionId,
      pipelineId,
      stepNumber,
      'running' as StepExecutionStatus,
      now,
      0,
      0
    ]);
    
    return this.findStepExecutionById(id) as Promise<StepExecution>;
  }

  public async findStepExecutionById(id: string): Promise<StepExecution | undefined> {
    const sql = `SELECT * FROM step_executions WHERE id = ?`;
    const row = await this.get<StepExecutionRow>(sql, [id]);
    
    if (!row) {
      return undefined;
    }
    
    return this.mapToStepExecution(row);
  }

  public async findStepExecutionsByExecutionId(executionId: string): Promise<StepExecution[]> {
    const sql = `SELECT * FROM step_executions WHERE execution_id = ? ORDER BY step_number ASC`;
    const rows = await this.all<StepExecutionRow>(sql, [executionId]);
    
    return rows.map(row => this.mapToStepExecution(row));
  }

  public async findStepExecutionByExecutionIdAndStepNumber(
    executionId: string,
    stepNumber: number
  ): Promise<StepExecution | undefined> {
    const sql = `SELECT * FROM step_executions WHERE execution_id = ? AND step_number = ?`;
    const row = await this.get<StepExecutionRow>(sql, [executionId, stepNumber]);
    
    if (!row) {
      return undefined;
    }
    
    return this.mapToStepExecution(row);
  }

  public async updateStepExecution(
    id: string,
    status: StepExecutionStatus,
    recordsProcessed: number,
    errorMessage?: string,
    dataConsumed: boolean = false
  ): Promise<StepExecution | undefined> {
    const now = new Date().toISOString();
    const updates: string[] = ['status = ?', 'records_processed = ?', 'data_consumed = ?'];
    const params: any[] = [status, recordsProcessed, dataConsumed ? 1 : 0];

    if (status === 'completed' || status === 'failed' || status === 'skipped') {
      updates.push('end_time = ?');
      params.push(now);
    }

    if (errorMessage !== undefined) {
      updates.push('error_message = ?');
      params.push(errorMessage);
    }

    params.push(id);

    const sql = `UPDATE step_executions SET ${updates.join(', ')} WHERE id = ?`;
    await this.run(sql, params);
    
    return this.findStepExecutionById(id);
  }

  private mapToPipelineExecution(row: PipelineExecutionRow): PipelineExecution {
    return {
      id: row.id,
      pipelineId: row.pipeline_id,
      status: row.status as ExecutionStatus,
      startTime: row.start_time,
      endTime: row.end_time,
      failedStepNumber: row.failed_step_number,
      errorMessage: row.error_message,
      createdAt: row.created_at
    };
  }

  private mapToStepExecution(row: StepExecutionRow): StepExecution {
    return {
      id: row.id,
      executionId: row.execution_id,
      pipelineId: row.pipeline_id,
      stepNumber: row.step_number,
      status: row.status as StepExecutionStatus,
      startTime: row.start_time,
      endTime: row.end_time,
      recordsProcessed: row.records_processed,
      errorMessage: row.error_message,
      dataConsumed: row.data_consumed === 1
    };
  }
}
