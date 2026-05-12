export interface Service {
  id: string;
  name: string;
  description?: string;
  createdAt: string;
  updatedAt: string;
}

export interface SLADefinition {
  id: string;
  serviceId: string;
  availabilityTarget: number; // 99.9 = 99.9%
  responseTimeTarget: number; // P99 目标，毫秒
  errorRateTarget: number; // 0.1 = 0.1%
  createdAt: string;
  updatedAt: string;
}

export interface Incident {
  id: string;
  serviceId: string;
  startTime: string;
  endTime?: string;
  durationMinutes: number;
  description?: string;
  isAutomatic: boolean;
  createdAt: string;
}

export interface Metric {
  id: string;
  serviceId: string;
  timestamp: string;
  responseTime: number; // 毫秒
  isError: boolean;
}

export interface Report {
  id: string;
  year: number;
  month: number;
  serviceId: string;
  slaDefinitionId: string;
  availabilityAchieved: number;
  responseTimeP99: number;
  errorRate: number;
  isConfirmed: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface ReportDetail {
  id: string;
  reportId: string;
  metricType: 'availability' | 'responseTime' | 'errorRate';
  target: number;
  actual: number;
  isAchieved: boolean;
  isConfirmed: boolean;
}

export interface Breach {
  id: string;
  reportId: string;
  incidentId: string;
  metricType: 'availability' | 'responseTime' | 'errorRate';
  createdAt: string;
}

export interface MonthlyReport {
  serviceId: string;
  serviceName: string;
  year: number;
  month: number;
  slaDefinition: {
    availabilityTarget: number;
    responseTimeTarget: number;
    errorRateTarget: number;
  };
  achieved: {
    availability: number;
    responseTimeP99: number;
    errorRate: number;
  };
  metrics: {
    availability: {
      target: number;
      actual: number;
      achieved: boolean;
    };
    responseTime: {
      target: number;
      actual: number;
      achieved: boolean;
    };
    errorRate: {
      target: number;
      actual: number;
      achieved: boolean;
    };
  };
  incidents: Incident[];
  complianceRate: number; // 达标率
  breaches: Breach[];
  isConfirmed: boolean;
}
