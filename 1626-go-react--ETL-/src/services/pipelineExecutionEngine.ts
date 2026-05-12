import { PipelineExecution, StepExecution, StepExecutionStatus } from '../models/execution';
import { Step } from '../models/step';
import { StepDao } from '../database/stepDao';
import { ExecutionDao } from '../database/executionDao';
import { PipelineDao } from '../database/pipelineDao';
import { TransformationEngine } from './transformationEngine';
import { DataSourceService } from './dataSourceService';

const BATCH_SIZE = 1000;

export class PipelineExecutionEngine {
  private stepDao: StepDao;
  private executionDao: ExecutionDao;
  private pipelineDao: PipelineDao;
  private transformationEngine: TransformationEngine;
  private dataSourceService: DataSourceService;

  constructor() {
    this.stepDao = new StepDao();
    this.executionDao = new ExecutionDao();
    this.pipelineDao = new PipelineDao();
    this.transformationEngine = new TransformationEngine();
    this.dataSourceService = new DataSourceService();
  }

  public async executePipeline(pipelineId: string): Promise<PipelineExecution> {
    const pipeline = await this.pipelineDao.findById(pipelineId);
    if (!pipeline) {
      throw new Error(`Pipeline '${pipelineId}' not found`);
    }

    const runningExecution = await this.executionDao.findRunningExecutionByPipelineId(pipelineId);
    if (runningExecution) {
      throw new Error(`Pipeline '${pipelineId}' is already running`);
    }

    const steps = await this.stepDao.findByPipelineId(pipelineId);
    if (steps.length === 0) {
      throw new Error(`Pipeline '${pipelineId}' has no steps`);
    }

    const execution = await this.executionDao.createPipelineExecution(pipelineId);
    
    try {
      await this.executeSteps(execution.id, pipelineId, steps);
      await this.executionDao.updatePipelineExecutionStatus(execution.id, 'completed');
      
      const updatedExecution = await this.executionDao.findPipelineExecutionById(execution.id);
      return updatedExecution!;
    } catch (error) {
      const errorMessage = (error as Error).message;
      const failedStepMatch = errorMessage.match(/Step (\d+)/);
      const failedStepNumber = failedStepMatch ? parseInt(failedStepMatch[1], 10) : undefined;
      
      await this.executionDao.updatePipelineExecutionStatus(
        execution.id,
        'failed',
        failedStepNumber,
        errorMessage
      );
      
      const updatedExecution = await this.executionDao.findPipelineExecutionById(execution.id);
      return updatedExecution!;
    }
  }

  public async retryExecution(
    pipelineId: string,
    execId: string,
    startFromStepNumber?: number
  ): Promise<PipelineExecution> {
    const originalExecution = await this.executionDao.findPipelineExecutionById(execId);
    if (!originalExecution) {
      throw new Error(`Execution '${execId}' not found`);
    }

    if (originalExecution.pipelineId !== pipelineId) {
      throw new Error(`Execution '${execId}' does not belong to pipeline '${pipelineId}'`);
    }

    const runningExecution = await this.executionDao.findRunningExecutionByPipelineId(pipelineId);
    if (runningExecution) {
      throw new Error(`Pipeline '${pipelineId}' is already running`);
    }

    const steps = await this.stepDao.findByPipelineId(pipelineId);
    if (steps.length === 0) {
      throw new Error(`Pipeline '${pipelineId}' has no steps`);
    }

    let stepToStartFrom: number;
    if (startFromStepNumber !== undefined) {
      const stepExists = steps.some(step => step.stepNumber === startFromStepNumber);
      if (!stepExists) {
        throw new Error(`Step ${startFromStepNumber} does not exist in pipeline`);
      }
      stepToStartFrom = startFromStepNumber;
    } else if (originalExecution.failedStepNumber !== undefined) {
      stepToStartFrom = originalExecution.failedStepNumber;
    } else {
      throw new Error('No step specified to retry from and no failed step found in execution');
    }

    const newExecution = await this.executionDao.createPipelineExecution(pipelineId);
    
    try {
      await this.executeSteps(newExecution.id, pipelineId, steps, stepToStartFrom);
      await this.executionDao.updatePipelineExecutionStatus(newExecution.id, 'completed');
      
      const updatedExecution = await this.executionDao.findPipelineExecutionById(newExecution.id);
      return updatedExecution!;
    } catch (error) {
      const errorMessage = (error as Error).message;
      const failedStepMatch = errorMessage.match(/Step (\d+)/);
      const failedStepNumber = failedStepMatch ? parseInt(failedStepMatch[1], 10) : undefined;
      
      await this.executionDao.updatePipelineExecutionStatus(
        newExecution.id,
        'failed',
        failedStepNumber,
        errorMessage
      );
      
      const updatedExecution = await this.executionDao.findPipelineExecutionById(newExecution.id);
      return updatedExecution!;
    }
  }

  private async executeSteps(
    executionId: string,
    pipelineId: string,
    steps: Step[],
    startFromStepNumber: number = 1
  ): Promise<void> {
    const stepsToExecute = steps.filter(step => step.stepNumber >= startFromStepNumber)
      .sort((a, b) => a.stepNumber - b.stepNumber);

    for (const step of stepsToExecute) {
      await this.executeStep(executionId, pipelineId, step);
    }
  }

  private async executeStep(
    executionId: string,
    pipelineId: string,
    step: Step
  ): Promise<void> {
    const stepExecution = await this.executionDao.createStepExecution(
      executionId,
      pipelineId,
      step.stepNumber
    );

    let totalRecordsProcessed = 0;

    try {
      let data: any[];
      
      try {
        data = await this.dataSourceService.fetchDataFromDataSource(step.dataSourceId);
      } catch (error) {
        throw new Error(`Step ${step.stepNumber}: Data source not available - ${(error as Error).message}`);
      }

      if (data.length === 0) {
        await this.completeStepExecution(stepExecution.id, 0, true);
        return;
      }

      const batches = this.splitIntoBatches(data, BATCH_SIZE);
      
      for (let i = 0; i < batches.length; i++) {
        const batch = batches[i];
        
        try {
          const transformedData = await this.transformationEngine.applyTransformations(
            batch,
            step.transformations
          );
          
          await this.writeToTarget(transformedData, step.targetConfig);
          totalRecordsProcessed += transformedData.length;
        } catch (error) {
          throw new Error(`Step ${step.stepNumber}: ${(error as Error).message}`);
        }
      }

      await this.completeStepExecution(stepExecution.id, totalRecordsProcessed, true);
    } catch (error) {
      await this.failStepExecution(stepExecution.id, totalRecordsProcessed, (error as Error).message);
      throw error;
    }
  }

  private splitIntoBatches(data: any[], batchSize: number): any[][] {
    const batches: any[][] = [];
    for (let i = 0; i < data.length; i += batchSize) {
      batches.push(data.slice(i, i + batchSize));
    }
    return batches;
  }

  private async writeToTarget(data: any[], targetConfig: any): Promise<void> {
    if (!targetConfig) {
      return;
    }

    if (targetConfig.type === 'memory') {
      console.log(`Writing ${data.length} records to memory target:`, JSON.stringify(targetConfig));
      return;
    }

    if (targetConfig.type === 'http' && targetConfig.url) {
      try {
        const response = await fetch(targetConfig.url, {
          method: targetConfig.method || 'POST',
          headers: {
            'Content-Type': 'application/json',
            ...(targetConfig.headers || {})
          },
          body: JSON.stringify(data)
        });
        
        if (!response.ok) {
          throw new Error(`Failed to write to target: HTTP ${response.status}`);
        }
      } catch (error) {
        throw new Error(`Failed to write to target: ${(error as Error).message}`);
      }
    }
  }

  private async completeStepExecution(
    stepExecutionId: string,
    recordsProcessed: number,
    dataConsumed: boolean
  ): Promise<void> {
    await this.executionDao.updateStepExecution(
      stepExecutionId,
      'completed' as StepExecutionStatus,
      recordsProcessed,
      undefined,
      dataConsumed
    );
  }

  private async failStepExecution(
    stepExecutionId: string,
    recordsProcessed: number,
    errorMessage: string
  ): Promise<void> {
    await this.executionDao.updateStepExecution(
      stepExecutionId,
      'failed' as StepExecutionStatus,
      recordsProcessed,
      errorMessage,
      false
    );
  }
}
