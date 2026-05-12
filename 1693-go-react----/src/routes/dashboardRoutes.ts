import { Router, Request, Response } from 'express';
import * as dashboardService from '../services/dashboardService';

const router = Router();

router.get('/', (req: Request, res: Response) => {
  try {
    const summary = dashboardService.getSummary();
    res.json(summary);
  } catch (err: any) {
    res.status(500).json({ error: err.message });
  }
});

router.get('/county/:county', (req: Request, res: Response) => {
  try {
    const stats = dashboardService.getCountyDashboard(req.params.county);
    res.json(stats);
  } catch (err: any) {
    res.status(500).json({ error: err.message });
  }
});

router.get('/county/:county/township/:township', (req: Request, res: Response) => {
  try {
    const stats = dashboardService.getTownshipDashboard(
      req.params.county,
      req.params.township
    );
    res.json(stats);
  } catch (err: any) {
    res.status(500).json({ error: err.message });
  }
});

router.get('/county/:county/township/:township/village/:village', (req: Request, res: Response) => {
  try {
    const stats = dashboardService.getVillageDashboard(
      req.params.county,
      req.params.township,
      req.params.village
    );
    res.json(stats);
  } catch (err: any) {
    res.status(500).json({ error: err.message });
  }
});

export default router;
