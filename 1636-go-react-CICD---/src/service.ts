import { v4 as uuidv4 } from 'uuid';
import {
  pipelineRepo,
  executionRepo,
  stageRepo,
  taskRepo,
  notificationRepo,
} from './database';
import {
  PipelineDefinition,
  StageDefinition,
  TaskDefinition,
  TriggerType,
  TaskStatus,
  StageStatus,
  PipelineStatus,
  Task,
  Stage,
  PipelineExecution,
} from './types';

function isValidTaskStatus(status: string): status is TaskStatus {
  return status === 'success' || status === 'failure' || status === 'skipped';
}

function notify(pipelineId: string, executionId: string, content: string) {
  notificationRepo.create({ pipelineId, executionId, content });
}

export class PipelineService {
  create(definition: PipelineDefinition) {
    const pipeline = pipelineRepo.create({
      name: definition.name,
      definition: JSON.stringify(definition),
      allowedWebhookSources: definition.allowedWebhookSources
        ? JSON.stringify(definition.allowedWebhookSources)
        : undefined,
    });
    return pipeline;
  }

  getById(id: string) {
    const pipeline = pipelineRepo.findById(id);
    if (!pipeline) {
      return { error: 'pipeline not found', status: 404 };
    }
    const definition = JSON.parse(pipeline.definition) as PipelineDefinition;
    const execution = executionRepo.findLatest(pipeline.id);
    
    let executionData = null;
    if (execution) {
      const stages = stageRepo.findAllByExecutionId(execution.id);
      const stagesWithTasks = stages.map((stage) => ({
        ...stage,
        tasks: taskRepo.findAllByStageId(stage.id),
      }));
      executionData = { ...execution, stages: stagesWithTasks };
    }

    return {
      data: {
        ...pipeline,
        definition,
        execution: executionData,
      },
      status: 200,
    };
  }

  trigger(pipelineId: string, triggerType: TriggerType, webhookSource?: string) {
    const pipeline = pipelineRepo.findById(pipelineId);
    if (!pipeline) {
      return { error: 'pipeline not found', status: 404 };
    }

    if (triggerType === 'webhook') {
      if (pipeline.allowedWebhookSources) {
        const allowed = JSON.parse(pipeline.allowedWebhookSources) as string[];
        if (!webhookSource || !allowed.includes(webhookSource)) {
          return { error: 'invalid webhook source', status: 401 };
        }
      } else {
        return { error: 'pipeline not found', status: 401 };
      }
    }

    const definition = JSON.parse(pipeline.definition) as PipelineDefinition;

    const deployEnvironments = new Set<string>();
    definition.stages.forEach((stageDef: StageDefinition) => {
      stageDef.tasks.forEach((taskDef: TaskDefinition) => {
        if (taskDef.type === 'deploy' && taskDef.environment) {
          deployEnvironments.add(taskDef.environment);
        }
      });
    });

    for (const env of deployEnvironments) {
      const running = taskRepo.findRunningDeploysByEnvironment(env);
      if (running.length > 0) {
        return { error: `deployment already in progress for environment ${env}`, status: 409 };
      }
    }
    const execution = executionRepo.create({
      pipelineId,
      status: 'pending',
      triggerType,
      triggerTime: Date.now(),
    });

    definition.stages.forEach((stageDef: StageDefinition, stageIndex: number) => {
      const stage = stageRepo.create({
        executionId: execution.id,
        name: stageDef.name,
        status: stageIndex === 0 ? 'running' : 'pending',
        order: stageIndex,
      });
      stageDef.tasks.forEach((taskDef: TaskDefinition, taskIndex: number) => {
        taskRepo.create({
          executionId: execution.id,
          stageId: stage.id,
          name: taskDef.name,
          type: taskDef.type,
          environment: taskDef.environment,
          status: stageIndex === 0 ? 'pending' : 'pending',
          logs: '',
          order: taskIndex,
        });
      });
    });

    executionRepo.updateStatus(execution.id, 'running', Date.now());
    notify(pipelineId, execution.id, `Pipeline started (${triggerType})`);

    this.runExecution(pipelineId, execution.id);

    return { data: { executionId: execution.id }, status: 200 };
  }

  private async runExecution(pipelineId: string, executionId: string) {
    const stages = stageRepo.findAllByExecutionId(executionId);
    
    for (let i = 0; i < stages.length; i++) {
      const stage = stages[i];
      const currentStage = stageRepo.findById(stage.id)!;
      
      if (currentStage.status === 'skipped' || currentStage.status === 'success') {
        continue;
      }

      stageRepo.updateStatus(stage.id, 'running');

      const tasks = taskRepo.findAllByStageId(stage.id);
      const tasksToRun = tasks.filter((t) => t.status !== 'success' && t.status !== 'skipped');
      
      let allSuccess = true;
      const taskPromises = tasksToRun.map((task) => this.runTask(task));
      const results = await Promise.all(taskPromises);
      
      for (const result of results) {
        if (result === 'failure') {
          allSuccess = false;
          break;
        }
      }

      const updatedTasks = taskRepo.findAllByStageId(stage.id);
      const anyFailed = updatedTasks.some((t) => t.status === 'failure');

      if (anyFailed) {
        stageRepo.updateStatus(stage.id, 'failure');
        
        for (let j = i + 1; j < stages.length; j++) {
          stageRepo.updateStatus(stages[j].id, 'skipped');
          const followingTasks = taskRepo.findAllByStageId(stages[j].id);
          followingTasks.forEach((t) => {
            taskRepo.update({ ...t, status: 'skipped' as TaskStatus });
          });
        }

        executionRepo.updateStatus(executionId, 'failure', undefined, Date.now());
        notify(pipelineId, executionId, 'Pipeline failed');
        return;
      } else if (updatedTasks.every((t) => t.status === 'success' || t.status === 'skipped')) {
        stageRepo.updateStatus(stage.id, 'success');
      }
    }

    const finalTasks = [];
    for (const stage of stages) {
      finalTasks.push(...taskRepo.findAllByStageId(stage.id));
    }

    const allDone = finalTasks.every((t) => t.status === 'success' || t.status === 'skipped');
    if (allDone) {
      executionRepo.updateStatus(executionId, 'success', undefined, Date.now());
      notify(pipelineId, executionId, 'Pipeline succeeded');
    }
  }

  private async runTask(task: Task): Promise<TaskStatus> {
    if (task.type === 'deploy' && task.environment) {
      const running = taskRepo.findRunningDeploysByEnvironment(task.environment);
      if (running.length > 0) {
        taskRepo.update({
          ...task,
          status: 'failure',
          endTime: Date.now(),
          logs: `deployment already in progress for environment ${task.environment}`,
        });
        return 'failure';
      }
    }

    const startTime = Date.now();
    taskRepo.update({
      ...task,
      status: 'running',
      startTime,
      logs: `${task.name} started\n`,
    });

    await new Promise((resolve) => setTimeout(resolve, 100));

    const success = Math.random() > 0.1;
    const endTime = Date.now();
    const status: TaskStatus = success ? 'success' : 'failure';
    const logs = `${task.name} started\n${task.name} ${success ? 'succeeded' : 'failed'}\n`;

    taskRepo.update({
      ...task,
      status,
      endTime,
      logs,
    });

    return status;
  }

  retryStage(pipelineId: string, executionId: string, stageId: string) {
    const pipeline = pipelineRepo.findById(pipelineId);
    if (!pipeline) {
      return { error: 'pipeline not found', status: 404 };
    }

    const stage = stageRepo.findById(stageId);
    if (!stage || stage.executionId !== executionId) {
      return { error: 'pipeline not found', status: 404 };
    }

    const execution = executionRepo.findById(executionId);
    if (!execution) {
      return { error: 'pipeline not found', status: 404 };
    }

    const tasks = taskRepo.findAllByStageId(stageId);
    for (const task of tasks) {
      if (task.status !== 'success') {
        taskRepo.update({
          ...task,
          status: 'pending',
          startTime: undefined,
          endTime: undefined,
          logs: '',
        });
      }
    }

    stageRepo.updateStatus(stageId, 'running');
    executionRepo.updateStatus(executionId, 'running');
    notify(pipelineId, executionId, `Retrying stage: ${stage.name}`);

    this.runExecution(pipelineId, executionId);

    return { data: { message: 'stage retry triggered' }, status: 200 };
  }

  retryTask(pipelineId: string, executionId: string, stageId: string, taskId: string) {
    const pipeline = pipelineRepo.findById(pipelineId);
    if (!pipeline) {
      return { error: 'pipeline not found', status: 404 };
    }

    const task = taskRepo.findById(taskId);
    if (!task || task.stageId !== stageId || task.executionId !== executionId) {
      return { error: 'pipeline not found', status: 404 };
    }

    const execution = executionRepo.findById(executionId);
    if (!execution) {
      return { error: 'pipeline not found', status: 404 };
    }

    if (task.status !== 'success') {
      taskRepo.update({
        ...task,
        status: 'pending',
        startTime: undefined,
        endTime: undefined,
        logs: '',
      });
    }

    const stage = stageRepo.findById(stageId)!;
    stageRepo.updateStatus(stageId, 'running');
    executionRepo.updateStatus(executionId, 'running');
    notify(pipelineId, executionId, `Retrying task: ${task.name}`);

    this.runExecution(pipelineId, executionId);

    return { data: { message: 'task retry triggered' }, status: 200 };
  }

  getLogs(pipelineId: string) {
    const pipeline = pipelineRepo.findById(pipelineId);
    if (!pipeline) {
      return { error: 'pipeline not found', status: 404 };
    }

    const execution = executionRepo.findLatest(pipelineId);
    if (!execution) {
      return { data: { logs: '' }, status: 200 };
    }

    const stages = stageRepo.findAllByExecutionId(execution.id);
    let logs = `Pipeline Execution: ${execution.id}\n`;
    logs += `Trigger: ${execution.triggerType} at ${new Date(execution.triggerTime).toISOString()}\n`;
    logs += `Status: ${execution.status}\n\n`;

    for (const stage of stages) {
      logs += `--- Stage: ${stage.name} (${stage.status}) ---\n`;
      const tasks = taskRepo.findAllByStageId(stage.id);
      for (const task of tasks) {
        logs += `  Task: ${task.name} [${task.type}] (${task.status})\n`;
        if (task.logs) {
          logs += task.logs
            .split('\n')
            .map((line) => `    ${line}`)
            .join('\n');
        }
        logs += '\n';
      }
    }

    return { data: { logs }, status: 200 };
  }

  static validateTaskStatus(status: string): boolean {
    return isValidTaskStatus(status);
  }
}
