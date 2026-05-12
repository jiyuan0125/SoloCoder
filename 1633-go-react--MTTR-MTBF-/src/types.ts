export type IncidentSeverity = 'P1' | 'P2' | 'P3' | 'P4';

export type IncidentStatus = 'discovered' | 'processing' | 'recovered' | 'reviewed';

export type ImpactScope = 'core' | 'non-core';

export type GroupBy = 'week' | 'month' | 'quarter';

export interface Incident {
  id: string;
  incidentNumber: string;
  service: string;
  startTime: Date;
  discoveredTime: Date;
  recoveredTime: Date | null;
  severity: IncidentSeverity;
  status: IncidentStatus;
  impactScope: ImpactScope;
  createdAt: Date;
  updatedAt: Date;
}

export interface CreateIncidentRequest {
  incidentNumber: string;
  service: string;
  startTime: string;
  discoveredTime: string;
  impactScope: ImpactScope;
  severity?: IncidentSeverity;
}

export interface UpdateIncidentRequest {
  service?: string;
  startTime?: string;
  discoveredTime?: string;
  recoveredTime?: string | null;
  impactScope?: ImpactScope;
  severity?: IncidentSeverity;
}

export interface UpdateStatusRequest {
  status: IncidentStatus;
}

export interface ListIncidentsQuery {
  service?: string;
  startTimeFrom?: string;
  startTimeTo?: string;
}

export interface MTTRResult {
  mttr: number | null;
  unit: string;
  count: number;
}

export interface MTBFResult {
  mtbf: number | null;
  unit: string;
  count: number;
  message?: string;
}

export interface GroupedStats {
  group: string;
  mttr: number | null;
  count: number;
}

export interface GroupedMTBF {
  group: string;
  mtbf: number | null;
  count: number;
  message?: string;
}
