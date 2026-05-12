import express from 'express';
import { PipelineService } from './service';
import { executionRepo, stageRepo, taskRepo } from './database';
import { TriggerType } from './types';

const app = express();
const port = process.env.PORT || 3000;

app.use(express.json());

const pipelineService = new PipelineService();

app.post('/pipelines', (req, res) => {
  const pipeline = pipelineService.create(req.body);
  res.status(201).json(pipeline);
});

app.post('/pipelines/:id/trigger', (req, res) => {
  const { id } = req.params;
  const { triggerType, webhookSource } = req.body as { triggerType: TriggerType; webhookSource?: string };
  
  const result = pipelineService.trigger(id, triggerType, webhookSource);
  
  if (result.error) {
    res.status(result.status).json({ error: result.error });
  } else {
    res.status(200).json(result.data);
  }
});

app.get('/pipelines/:id', (req, res) => {
  const { id } = req.params;
  const result = pipelineService.getById(id);
  
  if (result.error) {
    res.status(result.status).json({ error: result.error });
  } else {
    res.status(200).json(result.data);
  }
});

app.post('/pipelines/:id/stages/:stageId/retry', (req, res) => {
  const { id, stageId } = req.params;
  
  const latestExecution = executionRepo.findLatest(id);
  if (!latestExecution) {
    return res.status(404).json({ error: 'pipeline not found' });
  }

  const stage = stageRepo.findById(stageId);
  if (!stage || stage.executionId !== latestExecution.id) {
    return res.status(404).json({ error: 'pipeline not found' });
  }

  const result = pipelineService.retryStage(id, latestExecution.id, stageId);
  
  if (result.error) {
    res.status(result.status).json({ error: result.error });
  } else {
    res.status(200).json(result.data);
  }
});

app.post('/pipelines/:id/stages/:stageId/tasks/:taskId/retry', (req, res) => {
  const { id, stageId, taskId } = req.params;
  
  const latestExecution = executionRepo.findLatest(id);
  if (!latestExecution) {
    return res.status(404).json({ error: 'pipeline not found' });
  }

  const task = taskRepo.findById(taskId);
  if (!task || task.stageId !== stageId || task.executionId !== latestExecution.id) {
    return res.status(404).json({ error: 'pipeline not found' });
  }

  const result = pipelineService.retryTask(id, latestExecution.id, stageId, taskId);
  
  if (result.error) {
    res.status(result.status).json({ error: result.error });
  } else {
    res.status(200).json(result.data);
  }
});

app.get('/pipelines/:id/logs', (req, res) => {
  const { id } = req.params;
  const result = pipelineService.getLogs(id);
  
  if (result.error) {
    res.status(result.status).json({ error: result.error });
  } else {
    res.status(200).json(result.data);
  }
});

app.listen(port, () => {
  console.log(`CI/CD Pipeline Manager running on port ${port}`);
});
