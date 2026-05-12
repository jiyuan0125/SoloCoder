import { PipelineDao } from '../database/pipelineDao';
import { StepDao } from '../database/stepDao';
import { ExecutionDao } from '../database/executionDao';
import { Pipeline, CreatePipelineRequest, UpdatePipelineRequest } from '../models/pipeline';
import { Step, CreateStepRequest, UpdateStepRequest } from '../models/step';
import { PipelineExecution } from '../models/execution';
import { PipelineExecutionEngine } from './pipelineExecutionEngine';
import { DataSourceService } from './dataSourceService';

export class PipelineService {
  private pipelineDao: PipelineDao;
  private stepDao: StepDao;
  private executionDao: ExecutionDao;
  private executionEngine: PipelineExecutionEngine;
  private dataSourceService: DataSourceService;

  constructor() {
    this.pipelineDao = new PipelineDao();
    this.stepDao = new StepDao();
    this.executionDao = new ExecutionDao();
    this.executionEngine = new PipelineExecutionEngine();
    this.dataSourceService = new DataSourceService();
  }

  public async createPipeline(request: CreatePipelineRequest): Promise<Pipeline> {
    if (!request.name || request.name.trim() === '') {
      throw new Error('Pipeline name is required');
    }
    return this.pipelineDao.create(request);
  }

  public async getPipelineById(id: string): Promise<Pipeline | undefined> {
    return this.pipelineDao.findById(id);
  }

  public async getAllPipelines(): Promise<Pipeline[]> {
    return this.pipelineDao.findAll();
  }

  public async updatePipeline(id: string, request: UpdatePipelineRequest): Promise<Pipeline | undefined> {
    return this.pipelineDao.update(id, request);
  }

  public async deletePipeline(id: string): Promise<boolean> {
    await this.stepDao.deleteByPipelineId(id);
    return this.pipelineDao.delete(id);
  }

  public async addStepToPipeline(pipelineId: string, request: CreateStepRequest): Promise<Step> {
    const pipeline = await this.pipelineDao.findById(pipelineId);
    if (!pipeline) {
      throw new Error(`Pipeline '${pipelineId}' not found`);
    }

    const existingSteps = await this.stepDao.findByPipelineId(pipelineId);
    const stepNumberExists = existingSteps.some(step => step.stepNumber === request.stepNumber);
    if (stepNumberExists) {
      throw new Error(`Step number ${request.stepNumber} already exists in pipeline`);
    }

    const dataSource = await this.dataSourceService.getDataSourceById(request.dataSourceId);
    if (!dataSource) {
      throw new Error(`Data source '${request.dataSourceId}' not found`);
    }

    return this.stepDao.create(pipelineId, request);
  }

  public async getPipelineSteps(pipelineId: string): Promise<Step[]> {
    return this.stepDao.findByPipelineId(pipelineId);
  }

  public async getPipelineStep(pipelineId: string, stepNumber: number): Promise<Step | undefined> {
    return this.stepDao.findByPipelineIdAndStepNumber(pipelineId, stepNumber);
  }

  public async updatePipelineStep(
    pipelineId: string,
    stepId: string,
    request: UpdateStepRequest
  ): Promise<Step | undefined> {
    const existingStep = await this.stepDao.findById(stepId);
    if (!existingStep) {
      return undefined;
    }

    if (existingStep.pipelineId !== pipelineId) {
      throw new Error(`Step '${stepId}' does not belong to pipeline '${pipelineId}'`);
    }

    if (request.stepNumber !== undefined) {
      const otherSteps = await this.stepDao.findByPipelineId(pipelineId);
      const stepNumberExists = otherSteps.some(
        step => step.stepNumber === request.stepNumber && step.id !== stepId
      );
      if (stepNumberExists) {
        throw new Error(`Step number ${request.stepNumber} already exists in pipeline`);
      }
    }

    if (request.dataSourceId !== undefined) {
      const dataSource = await this.dataSourceService.getDataSourceById(request.dataSourceId);
      if (!dataSource) {
        throw new Error(`Data source '${request.dataSourceId}' not found`);
      }
    }

    return this.stepDao.update(stepId, request);
  }

  public async deletePipelineStep(pipelineId: string, stepId: string): Promise<boolean> {
    const existingStep = await this.stepDao.findById(stepId);
    if (!existingStep) {
      return false;
    }

    if (existingStep.pipelineId !== pipelineId) {
      throw new Error(`Step '${stepId}' does not belong to pipeline '${pipelineId}'`);
    }

    return this.stepDao.delete(stepId);
  }

  public async executePipeline(pipelineId: string): Promise<PipelineExecution> {
    return this.executionEngine.executePipeline(pipelineId);
  }

  public async getPipelineExecutions(pipelineId: string): Promise<PipelineExecution[]> {
    return this.executionDao.findPipelineExecutionsByPipelineId(pipelineId);
  }

  public async getPipelineExecution(pipelineId: string, executionId: string): Promise<PipelineExecution | undefined> {
    const execution = await this.executionDao.findPipelineExecutionById(executionId);
    if (execution && execution.pipelineId !== pipelineId) {
      throw new Error(`Execution '${executionId}' does not belong to pipeline '${pipelineId}'`);
    }
    return execution;
  }

  public async retryExecution(
    pipelineId: string,
    executionId: string,
    startFromStepNumber?: number
  ): Promise<PipelineExecution> {
    return this.executionEngine.retryExecution(pipelineId, executionId, startFromStepNumber);
  }
}
