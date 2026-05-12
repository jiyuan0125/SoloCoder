export enum Category {
  搬运 = '搬运',
  代购 = '代购',
  照看宠物 = '照看宠物',
  维修 = '维修',
  接送孩子 = '接送孩子',
  其他 = '其他',
}

export const VALID_CATEGORIES = Object.values(Category);

export enum RequestStatus {
  待接单 = '待接单',
  进行中 = '进行中',
  已完成 = '已完成',
  已过期 = '已过期',
  已归档 = '已归档',
}

export interface User {
  id: string;
  name: string;
  building: string;
  avgRating: number;
  ratingCount: number;
  coolDownUntil?: number;
  createdAt: number;
}

export interface HelpRequest {
  id: string;
  publisherId: string;
  title: string;
  category: Category;
  description: string;
  expectedTime: number;
  willingToPay: boolean;
  status: RequestStatus;
  assigneeId?: string;
  completedAt?: number;
  createdAt: number;
  expiredAt?: number;
  archivedAt?: number;
}

export interface Rating {
  id: string;
  requestId: string;
  fromUserId: string;
  toUserId: string;
  score: number;
  createdAt: number;
}

export interface MonthStat {
  id: string;
  year: number;
  month: number;
  publishCount: number;
  matchSuccessCount: number;
  totalCompleteTime: number;
  completeCount: number;
}

export interface BuildingActivity {
  id: string;
  building: string;
  year: number;
  month: number;
  publishCount: number;
  acceptCount: number;
  score: number;
}
