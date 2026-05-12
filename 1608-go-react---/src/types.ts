export interface Rating {
  userId: number;
  productId: number;
  score: number;
  updatedAt: string;
}

export interface Product {
  id: number;
  name: string;
  category: string;
  avgScore: number;
  totalRatings: number;
  createdAt: string;
}

export interface DailyStats {
  date: string;
  coverage: number;
  accuracy: number;
  diversity: number;
  totalExposures: number;
  totalInteractions: number;
}

export interface RecommendHistory {
  id: number;
  userId: number;
  productId: number;
  exposedAt: string;
  interactedAt: string | null;
}
