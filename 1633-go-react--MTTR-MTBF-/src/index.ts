import express, { Request, Response } from 'express';
import {
  createIncident,
  findIncidentById,
  listIncidents,
  updateIncident,
  updateIncidentStatus,
} from './repository';
import { calculateSeverity, isValidStatusTransition, canBeReopened } from './business';
import { calculateMTTR, calculateMTBF } from './stats';
import {
  CreateIncidentRequest,
  UpdateIncidentRequest,
  UpdateStatusRequest,
  IncidentStatus,
  GroupBy,
} from './types';

const app = express();
app.use(express.json());

const PORT = process.env.PORT || 3000;

app.post('/incidents', async (req: Request, res: Response) => {
  try {
    const body = req.body as CreateIncidentRequest;

    if (
      !body.incidentNumber ||
      !body.service ||
      !body.startTime ||
      !body.discoveredTime ||
      !body.impactScope
    ) {
      return res.status(400).json({ error: 'Missing required fields' });
    }

    const startTime = new Date(body.startTime);
    const discoveredTime = new Date(body.discoveredTime);

    if (isNaN(startTime.getTime()) || isNaN(discoveredTime.getTime())) {
      return res.status(400).json({ error: 'Invalid date format' });
    }

    const severity = body.severity || calculateSeverity(body.impactScope, startTime, null);
    const incident = await createIncident(body, severity, 'discovered');

    return res.status(201).json(incident);
  } catch (err: any) {
    if (err.message?.includes('UNIQUE constraint')) {
      return res.status(409).json({ error: 'Incident number already exists' });
    }
    console.error(err);
    return res.status(500).json({ error: 'Internal server error' });
  }
});

app.get('/incidents', async (req: Request, res: Response) => {
  try {
    const query = {
      service: req.query.service as string | undefined,
      startTimeFrom: req.query.startTimeFrom as string | undefined,
      startTimeTo: req.query.startTimeTo as string | undefined,
    };

    const incidents = await listIncidents(query);
    return res.json(incidents);
  } catch (err) {
    console.error(err);
    return res.status(500).json({ error: 'Internal server error' });
  }
});

app.put('/incidents/:id', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const body = req.body as UpdateIncidentRequest;

    const existing = await findIncidentById(id);
    if (!existing) {
      return res.status(404).json({ error: 'Incident not found' });
    }

    const startTime = body.startTime ? new Date(body.startTime) : existing.startTime;
    const recoveredTime =
      body.recoveredTime !== undefined
        ? body.recoveredTime
          ? new Date(body.recoveredTime)
          : null
        : existing.recoveredTime;
    const impactScope = body.impactScope || existing.impactScope;

    let newSeverity: string | undefined;
    if (body.severity) {
      newSeverity = body.severity;
    } else if (
      body.startTime ||
      body.recoveredTime !== undefined ||
      body.impactScope
    ) {
      newSeverity = calculateSeverity(impactScope, startTime, recoveredTime);
    }

    const updated = await updateIncident(id, body, newSeverity);
    if (!updated) {
      return res.status(404).json({ error: 'Incident not found' });
    }

    return res.json(updated);
  } catch (err) {
    console.error(err);
    return res.status(500).json({ error: 'Internal server error' });
  }
});

app.put('/incidents/:id/status', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const body = req.body as UpdateStatusRequest;

    if (!body.status) {
      return res.status(400).json({ error: 'Status is required' });
    }

    const incident = await findIncidentById(id);
    if (!incident) {
      return res.status(404).json({ error: 'Incident not found' });
    }

    const currentStatus = incident.status;
    const newStatus = body.status;

    if (newStatus === currentStatus) {
      return res.json(incident);
    }

    if (newStatus === 'processing' && currentStatus === 'recovered') {
      if (!canBeReopened(currentStatus)) {
        return res.status(400).json({ error: 'reviewed incident cannot be reopened' });
      }
    } else {
      if (!isValidStatusTransition(currentStatus, newStatus)) {
        return res
          .status(400)
          .json({ error: `invalid status transition from ${currentStatus} to ${newStatus}` });
      }
    }

    if (currentStatus === 'reviewed' && newStatus !== 'reviewed') {
      return res.status(400).json({ error: 'reviewed incident cannot be reopened' });
    }

    let recoveredTime: string | undefined;
    if (newStatus === 'recovered' && !incident.recoveredTime) {
      recoveredTime = new Date().toISOString();
    }

    const updated = await updateIncidentStatus(id, newStatus, recoveredTime);
    return res.json(updated);
  } catch (err) {
    console.error(err);
    return res.status(500).json({ error: 'Internal server error' });
  }
});

app.get('/stats/mttr', async (req: Request, res: Response) => {
  try {
    const service = req.query.service as string | undefined;
    const startTimeFrom = req.query.startTimeFrom as string | undefined;
    const startTimeTo = req.query.startTimeTo as string | undefined;
    const groupBy = req.query.groupBy as GroupBy | undefined;

    const result = await calculateMTTR(service, startTimeFrom, startTimeTo, groupBy);
    return res.json(result);
  } catch (err) {
    console.error(err);
    return res.status(500).json({ error: 'Internal server error' });
  }
});

app.get('/stats/mtbf', async (req: Request, res: Response) => {
  try {
    const service = req.query.service as string | undefined;
    const startTimeFrom = req.query.startTimeFrom as string | undefined;
    const startTimeTo = req.query.startTimeTo as string | undefined;
    const groupBy = req.query.groupBy as GroupBy | undefined;

    const result = await calculateMTBF(service, startTimeFrom, startTimeTo, groupBy);
    return res.json(result);
  } catch (err) {
    console.error(err);
    return res.status(500).json({ error: 'Internal server error' });
  }
});

app.listen(PORT, () => {
  console.log(`Incident management service running on port ${PORT}`);
});
