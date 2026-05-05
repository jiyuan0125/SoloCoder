import { Coupon, DistributionRecord, User } from '@coupon/shared';
import * as fs from 'fs';
import * as path from 'path';

const DATA_DIR = path.join(process.cwd(), 'data');
const COUPONS_FILE = path.join(DATA_DIR, 'coupons.json');
const RECORDS_FILE = path.join(DATA_DIR, 'distribution-records.json');
const USERS_FILE = path.join(DATA_DIR, 'users.json');

export class DataStore {
  private coupons: Map<string, Coupon> = new Map();
  private distributionRecords: Map<string, DistributionRecord> = new Map();
  private users: Map<string, User> = new Map();

  constructor() {
    this.ensureDataDir();
    this.loadData();
    this.initializeMockUsers();
  }

  private ensureDataDir(): void {
    if (!fs.existsSync(DATA_DIR)) {
      fs.mkdirSync(DATA_DIR, { recursive: true });
    }
  }

  private loadData(): void {
    this.loadCoupons();
    this.loadDistributionRecords();
    this.loadUsers();
  }

  private loadCoupons(): void {
    try {
      if (fs.existsSync(COUPONS_FILE)) {
        const data = fs.readFileSync(COUPONS_FILE, 'utf-8');
        const couponsArray: Coupon[] = JSON.parse(data);
        this.coupons = new Map(couponsArray.map(c => [c.id, c]));
      }
    } catch (error) {
      console.error('Error loading coupons:', error);
    }
  }

  private loadDistributionRecords(): void {
    try {
      if (fs.existsSync(RECORDS_FILE)) {
        const data = fs.readFileSync(RECORDS_FILE, 'utf-8');
        const recordsArray: DistributionRecord[] = JSON.parse(data);
        this.distributionRecords = new Map(recordsArray.map(r => [r.id, r]));
      }
    } catch (error) {
      console.error('Error loading distribution records:', error);
    }
  }

  private loadUsers(): void {
    try {
      if (fs.existsSync(USERS_FILE)) {
        const data = fs.readFileSync(USERS_FILE, 'utf-8');
        const usersArray: User[] = JSON.parse(data);
        this.users = new Map(usersArray.map(u => [u.id, u]));
      }
    } catch (error) {
      console.error('Error loading users:', error);
    }
  }

  private saveData(): void {
    this.saveCoupons();
    this.saveDistributionRecords();
    this.saveUsers();
  }

  private saveCoupons(): void {
    try {
      const couponsArray = Array.from(this.coupons.values());
      fs.writeFileSync(COUPONS_FILE, JSON.stringify(couponsArray, null, 2));
    } catch (error) {
      console.error('Error saving coupons:', error);
    }
  }

  private saveDistributionRecords(): void {
    try {
      const recordsArray = Array.from(this.distributionRecords.values());
      fs.writeFileSync(RECORDS_FILE, JSON.stringify(recordsArray, null, 2));
    } catch (error) {
      console.error('Error saving distribution records:', error);
    }
  }

  private saveUsers(): void {
    try {
      const usersArray = Array.from(this.users.values());
      fs.writeFileSync(USERS_FILE, JSON.stringify(usersArray, null, 2));
    } catch (error) {
      console.error('Error saving users:', error);
    }
  }

  private initializeMockUsers(): void {
    if (this.users.size === 0) {
      const now = new Date();
      const mockUsers: User[] = [
        { id: 'user-001', name: '张三', registerTime: new Date(now.getTime() - 1 * 24 * 60 * 60 * 1000).toISOString() },
        { id: 'user-002', name: '李四', registerTime: new Date(now.getTime() - 2 * 24 * 60 * 60 * 1000).toISOString() },
        { id: 'user-003', name: '王五', registerTime: new Date(now.getTime() - 10 * 24 * 60 * 60 * 1000).toISOString() },
        { id: 'user-004', name: '赵六', registerTime: new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000).toISOString() },
        { id: 'user-005', name: '钱七', registerTime: new Date(now.getTime() - 60 * 24 * 60 * 60 * 1000).toISOString() },
      ];
      mockUsers.forEach(user => {
        this.users.set(user.id, user);
      });
      this.saveUsers();
      console.log('Initialized mock users');
    }
  }

  public getCoupon(id: string): Coupon | undefined {
    return this.coupons.get(id);
  }

  public addCoupon(coupon: Coupon): void {
    this.coupons.set(coupon.id, coupon);
    this.saveCoupons();
  }

  public updateCoupon(coupon: Coupon): void {
    this.coupons.set(coupon.id, coupon);
    this.saveCoupons();
  }

  public getUserCoupons(userId: string): Coupon[] {
    return Array.from(this.coupons.values()).filter(c => c.userId === userId);
  }

  public addDistributionRecord(record: DistributionRecord): void {
    this.distributionRecords.set(record.id, record);
    this.saveDistributionRecords();
  }

  public getUser(id: string): User | undefined {
    return this.users.get(id);
  }

  public getAllUsers(): User[] {
    return Array.from(this.users.values());
  }

  public addUser(user: User): void {
    this.users.set(user.id, user);
    this.saveUsers();
  }
}

export const dataStore = new DataStore();
