import { Router, Request, Response } from 'express';
import { PipelineService } from '../services/pipelineService';
import { CreatePipelineRequest, UpdatePipelineRequest } from '../models/pipeline';
import { CreateStepRequest, UpdateStepRequest } from '../models/step';
import { RetryExecutionRequest } from '../models/execution';

const router = Router();
const pipelineService = new PipelineService();

router.get('/', async (req: Request, res: Response) => {
  try {
    const pipelines = await pipelineService.getAllPipelines();
    res.json(pipelines);
  } catch (error) {
    res.status(500).json({ error: (error as Error).message });
  }
});

router.post('/', async (req: Request, res: Response) => {
  try {
    const request: CreatePipelineRequest = req.body;
    const pipeline = await pipelineService.createPipeline(request);
    res.status(201).json(pipeline);
  } catch (error) {
    res.status(400).json({ error: (error as Error).message });
  }
});

router.get('/:id', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const pipeline = await pipelineService.getPipelineById(id);
    
    if (!pipeline) {
      res.status(404).json({ error: 'Pipeline not found' });
      return;
    }
    
    res.json(pipeline);
  } catch (error) {
    res.status(500).json({ error: (error as Error).message });
  }
});

router.put('/:id', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const request: UpdatePipelineRequest = req.body;
    
    const updatedPipeline = await pipelineService.updatePipeline(id, request);
    
    if (!updatedPipeline) {
      res.status(404).json({ error: 'Pipeline not found' });
      return;
    }
    
    res.json(updatedPipeline);
  } catch (error) {
    res.status(400).json({ error: (error as Error).message });
  }
});

router.delete('/:id', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    
    const pipeline = await pipelineService.getPipelineById(id);
    if (!pipeline) {
      res.status(404).json({ error: 'Pipeline not found' });
      return;
    }
    
    await pipelineService.deletePipeline(id);
    res.status(204).send();
  } catch (error) {
    res.status(500).json({ error: (error as Error).message });
  }
});

router.get('/:id/steps', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    
    const pipeline = await pipelineService.getPipelineById(id);
    if (!pipeline) {
      res.status(404).json({ error: 'Pipeline not found' });
      return;
    }
    
    const steps = await pipelineService.getPipelineSteps(id);
    res.json(steps);
  } catch (error) {
    res.status(500).json({ error: (error as Error).message });
  }
});

router.post('/:id/steps', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const request: CreateStepRequest = req.body;
    
    const step = await pipelineService.addStepToPipeline(id, request);
    res.status(201).json(step);
  } catch (error) {
    const errorMessage = (error as Error).message;
    if (errorMessage.includes('not found')) {
      res.status(404).json({ error: errorMessage });
    } else if (errorMessage.includes('already exists')) {
      res.status(409).json({ error: errorMessage });
    } else {
      res.status(400).json({ error: errorMessage });
    }
  }
});

router.get('/:id/steps/:stepNumber', async (req: Request, res: Response) => {
  try {
    const { id, stepNumber } = req.params;
    
    const pipeline = await pipelineService.getPipelineById(id);
    if (!pipeline) {
      res.status(404).json({ error: 'Pipeline not found' });
      return;
    }
    
    const step = await pipelineService.getPipelineStep(id, parseInt(stepNumber, 10));
    
    if (!step) {
      res.status(404).json({ error: 'Step not found' });
      return;
    }
    
    res.json(step);
  } catch (error) {
    res.status(500).json({ error: (error as Error).message });
  }
});

router.put('/:id/steps/:stepId', async (req: Request, res: Response) => {
  try {
    const { id, stepId } = req.params;
    const request: UpdateStepRequest = req.body;
    
    const updatedStep = await pipelineService.updatePipelineStep(id, stepId, request);
    
    if (!updatedStep) {
      res.status(404).json({ error: 'Step not found' });
      return;
    }
    
    res.json(updatedStep);
  } catch (error) {
    const errorMessage = (error as Error).message;
    if (errorMessage.includes('does not belong')) {
      res.status(400).json({ error: errorMessage });
    } else if (errorMessage.includes('already exists')) {
      res.status(409).json({ error: errorMessage });
    } else {
      res.status(400).json({ error: errorMessage });
    }
  }
});

router.delete('/:id/steps/:stepId', async (req: Request, res: Response) => {
  try {
    const { id, stepId } = req.params;
    
    const deleted = await pipelineService.deletePipelineStep(id, stepId);
    
    if (!deleted) {
      res.status(404).json({ error: 'Step not found' });
      return;
    }
    
    res.status(204).send();
  } catch (error) {
    const errorMessage = (error as Error).message;
    if (errorMessage.includes('does not belong')) {
      res.status(400).json({ error: errorMessage });
    } else {
      res.status(500).json({ error: errorMessage });
    }
  }
});

router.post('/:id/executions', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    
    const execution = await pipelineService.executePipeline(id);
    res.status(201).json(execution);
  } catch (error) {
    const errorMessage = (error as Error).message;
    if (errorMessage.includes('not found')) {
      res.status(404).json({ error: errorMessage });
    } else if (errorMessage.includes('already running')) {
      res.status(409).json({ error: errorMessage });
    } else {
      res.status(400).json({ error: errorMessage });
    }
  }
});

router.get('/:id/executions', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    
    const pipeline = await pipelineService.getPipelineById(id);
    if (!pipeline) {
      res.status(404).json({ error: 'Pipeline not found' });
      return;
    }
    
    const executions = await pipelineService.getPipelineExecutions(id);
    res.json(executions);
  } catch (error) {
    res.status(500).json({ error: (error as Error).message });
  }
});

router.get('/:id/executions/:execId', async (req: Request, res: Response) => {
  try {
    const { id, execId } = req.params;
    
    const execution = await pipelineService.getPipelineExecution(id, execId);
    
    if (!execution) {
      res.status(404).json({ error: 'Execution not found' });
      return;
    }
    
    res.json(execution);
  } catch (error) {
    const errorMessage = (error as Error).message;
    if (errorMessage.includes('does not belong')) {
      res.status(400).json({ error: errorMessage });
    } else {
      res.status(500).json({ error: errorMessage });
    }
  }
});

router.post('/:id/executions/:execId/retry', async (req: Request, res: Response) => {
  try {
    const { id, execId } = req.params;
    const request: RetryExecutionRequest = req.body || {};
    
    const execution = await pipelineService.retryExecution(
      id,
      execId,
      request.stepNumber
    );
    
    res.status(201).json(execution);
  } catch (error) {
    const errorMessage = (error as Error).message;
    if (errorMessage.includes('not found')) {
      res.status(404).json({ error: errorMessage });
    } else if (errorMessage.includes('already running')) {
      res.status(409).json({ error: errorMessage });
    } else if (errorMessage.includes('does not exist in pipeline') || errorMessage.includes('does not belong')) {
      res.status(400).json({ error: errorMessage });
    } else {
      res.status(400).json({ error: errorMessage });
    }
  }
});

export default router;
