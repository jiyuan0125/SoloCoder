import { Router, Request, Response } from 'express';
import { DataSourceService } from '../services/dataSourceService';
import { CreateDataSourceRequest, UpdateDataSourceRequest } from '../models/dataSource';

const router = Router();
const dataSourceService = new DataSourceService();

router.get('/', async (req: Request, res: Response) => {
  try {
    const dataSources = await dataSourceService.getAllDataSources();
    res.json(dataSources);
  } catch (error) {
    res.status(500).json({ error: (error as Error).message });
  }
});

router.post('/', async (req: Request, res: Response) => {
  try {
    const request: CreateDataSourceRequest = req.body;
    const dataSource = await dataSourceService.createDataSource(request);
    res.status(201).json(dataSource);
  } catch (error) {
    res.status(400).json({ error: (error as Error).message });
  }
});

router.get('/:id', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const dataSource = await dataSourceService.getDataSourceById(id);
    
    if (!dataSource) {
      res.status(404).json({ error: 'Data source not found' });
      return;
    }
    
    res.json(dataSource);
  } catch (error) {
    res.status(500).json({ error: (error as Error).message });
  }
});

router.put('/:id', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const request: UpdateDataSourceRequest = req.body;
    
    const updatedDataSource = await dataSourceService.updateDataSource(id, request);
    
    if (!updatedDataSource) {
      res.status(404).json({ error: 'Data source not found' });
      return;
    }
    
    res.json(updatedDataSource);
  } catch (error) {
    res.status(400).json({ error: (error as Error).message });
  }
});

router.delete('/:id', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    
    const dataSource = await dataSourceService.getDataSourceById(id);
    if (!dataSource) {
      res.status(404).json({ error: 'Data source not found' });
      return;
    }
    
    await dataSourceService.deleteDataSource(id);
    res.status(204).send();
  } catch (error) {
    res.status(500).json({ error: (error as Error).message });
  }
});

export default router;
