import { Router, Request, Response } from 'express';
import {
  getAllApplications,
  getAllTraces,
  getAlertsByApplication,
} from '../storage';
import { Topology, TopologyNode, TopologyEdge, TraceData } from '../types';

const router = Router();

function getHealthStatus(applicationId: string): TopologyNode['healthStatus'] {
  const alerts = getAlertsByApplication(applicationId);
  const activeAlerts = alerts.filter(a => a.status === 'active');
  
  const hasCritical = activeAlerts.some(a => a.severity === 'critical');
  if (hasCritical) {
    return 'critical';
  }
  
  const hasWarning = activeAlerts.some(a => a.severity === 'warning');
  if (hasWarning) {
    return 'degraded';
  }
  
  return 'healthy';
}

router.get('/', (req: Request, res: Response) => {
  const applications = getAllApplications();
  const traces = getAllTraces();
  
  const nodes: TopologyNode[] = applications.map(app => ({
    id: app.id,
    name: app.name,
    healthStatus: getHealthStatus(app.id),
  }));
  
  const serviceToAppId = new Map<string, string>();
  applications.forEach(app => {
    serviceToAppId.set(app.name, app.id);
  });
  
  const edgeMap = new Map<string, number>();
  
  const parentToChild = new Map<string, Set<string>>();
  
  for (const trace of traces) {
    if (!parentToChild.has(trace.traceId)) {
      parentToChild.set(trace.traceId, new Set());
    }
  }
  
  const spanMap = new Map<string, TraceData>();
  for (const trace of traces) {
    spanMap.set(trace.spanId, trace);
  }
  
  for (const trace of traces) {
    if (trace.parentSpanId) {
      const parent = spanMap.get(trace.parentSpanId);
      if (parent && parent.serviceName !== trace.serviceName) {
        const sourceId = serviceToAppId.get(parent.serviceName);
        const targetId = serviceToAppId.get(trace.serviceName);
        
        if (sourceId && targetId) {
          const key = `${sourceId}->${targetId}`;
          edgeMap.set(key, (edgeMap.get(key) || 0) + 1);
        }
      }
    }
  }
  
  const edges: TopologyEdge[] = Array.from(edgeMap.entries()).map(([key, count]) => {
    const [source, target] = key.split('->');
    return { source, target, count };
  });
  
  const topology: Topology = { nodes, edges };
  res.status(200).json(topology);
});

export default router;
