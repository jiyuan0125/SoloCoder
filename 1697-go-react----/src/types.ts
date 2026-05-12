export enum DisasterType {
  EARTHQUAKE = '地震',
  FLOOD = '洪水',
  TYPHOON = '台风',
  DROUGHT = '干旱',
  LANDSLIDE = '山体滑坡',
  FIRE = '火灾',
  OTHER = '其他'
}

export enum ReportStatus {
  PENDING = '待核查',
  UNDER_REVIEW = '核查中',
  VERIFIED = '已核查',
  NEEDS_SUPPLEMENT = '需补充',
  PUBLISHED = '已发布'
}

export enum VerificationResult {
  PASSED = '合格',
  NEEDS_SUPPLEMENT = '需补充'
}

export interface DisasterReport {
  id: string;
  disasterType: string;
  occurrenceTime: string;
  location: string;
  affectedPopulation: string | number;
  evacuatedPopulation: string | number;
  deathMissingCount: string | number;
  cropAreaAffected: string | number;
  housesDamaged: string | number;
  directEconomicLoss: string | number;
  status: ReportStatus;
  createdAt: string;
  updatedAt: string;
  mergedFrom: string | null;
  reportCount: number;
}

export interface VerificationRecord {
  id: string;
  reportId: string;
  verifierId: string;
  verifierName: string;
  result: VerificationResult;
  comments: string;
  verifiedAt: string;
  deadline: string;
}

export interface PublishedRecord {
  id: string;
  reportId: string;
  publisherId: string;
  publisherName: string;
  content: string;
  version: number;
  previousContent: string | null;
  publishedAt: string;
}

export interface StatisticsReport {
  type: 'daily' | 'weekly' | 'monthly';
  period: string;
  totalReports: number;
  verifiedReports: number;
  totalAffectedPopulation: number;
  totalEvacuated: number;
  totalDeaths: number;
  totalEconomicLoss: number;
  breakdown: Array<{
    disasterType: string;
    count: number;
    affectedPopulation: number;
    economicLoss: number;
  }>;
}
