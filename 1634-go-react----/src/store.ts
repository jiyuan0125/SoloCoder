import { v4 as uuidv4 } from 'uuid';
import { Server, Metric, Alert, ResourceType, AlertConfig } from './types';

class DataStore {
  private servers: Map<string, Server> = new Map();
  private metrics: Metric[] = [];
  private alerts: Map<string, Alert[]> = new Map();
  public alertConfig: AlertConfig = {
    threshold: 80,
    durationMinutes: 5
  };

  addServer(name: string): Server {
    const server: Server = {
      id: uuidv4(),
      name,
      createdAt: new Date()
    };
    this.servers.set(server.id, server);
    return server;
  }

  getServer(id: string): Server | undefined {
    return this.servers.get(id);
  }

  addMetric(metric: Metric): void {
    this.metrics.push(metric);
    this.cleanupOldMetrics();
  }

  getMetrics(serverId: string, startTime: Date, endTime: Date): Metric[] {
    return this.metrics.filter(
      m => m.serverId === serverId && m.timestamp >= startTime && m.timestamp <= endTime
    );
  }

  getActiveAlert(serverId: string, resourceType: ResourceType): Alert | undefined {
    const serverAlerts = this.alerts.get(serverId) || [];
    return serverAlerts.find(a => a.resourceType === resourceType && a.active);
  }

  createAlert(serverId: string, resourceType: ResourceType): Alert {
    const alert: Alert = {
      id: uuidv4(),
      serverId,
      resourceType,
      startTime: new Date(),
      threshold: this.alertConfig.threshold,
      active: true
    };

    if (!this.alerts.has(serverId)) {
      this.alerts.set(serverId, []);
    }
    this.alerts.get(serverId)!.push(alert);
    return alert;
  }

  getAlerts(serverId: string): Alert[] {
    return this.alerts.get(serverId) || [];
  }

  private cleanupOldMetrics(): void {
    const thirtyDaysAgo = new Date(Date.now() - 30 * 24 * 60 * 60 * 1000);
    this.metrics = this.metrics.filter(m => m.timestamp >= thirtyDaysAgo);
  }
}

export const dataStore = new DataStore();
