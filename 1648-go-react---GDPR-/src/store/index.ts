import { DataCategory, UserDataRecord, DataSubjectRequest, Consent } from '../types';

export class InMemoryStore {
  private dataCategories: Map<string, DataCategory> = new Map();
  private userData: Map<string, UserDataRecord> = new Map();
  private requests: Map<string, DataSubjectRequest> = new Map();
  private consents: Map<string, Consent> = new Map();

  private static instance: InMemoryStore;

  static getInstance(): InMemoryStore {
    if (!this.instance) {
      this.instance = new InMemoryStore();
    }
    return this.instance;
  }

  addDataCategory(category: DataCategory): DataCategory {
    this.dataCategories.set(category.id, category);
    return category;
  }

  getDataCategory(id: string): DataCategory | undefined {
    return this.dataCategories.get(id);
  }

  getAllDataCategories(): DataCategory[] {
    return Array.from(this.dataCategories.values());
  }

  addUserData(record: UserDataRecord): UserDataRecord {
    this.userData.set(record.id, record);
    return record;
  }

  getUserDataByUserId(userId: string): UserDataRecord[] {
    return Array.from(this.userData.values()).filter(r => r.userId === userId);
  }

  getAllUserData(): UserDataRecord[] {
    return Array.from(this.userData.values());
  }

  updateUserData(id: string, updates: Partial<UserDataRecord>): UserDataRecord | undefined {
    const record = this.userData.get(id);
    if (!record) return undefined;
    const updated = { ...record, ...updates };
    this.userData.set(id, updated);
    return updated;
  }

  addRequest(request: DataSubjectRequest): DataSubjectRequest {
    this.requests.set(request.id, request);
    return request;
  }

  getRequest(id: string): DataSubjectRequest | undefined {
    return this.requests.get(id);
  }

  updateRequest(id: string, updates: Partial<DataSubjectRequest>): DataSubjectRequest | undefined {
    const request = this.requests.get(id);
    if (!request) return undefined;
    const updated = { ...request, ...updates };
    this.requests.set(id, updated);
    return updated;
  }

  getAllRequests(): DataSubjectRequest[] {
    return Array.from(this.requests.values());
  }

  addConsent(consent: Consent): Consent {
    this.consents.set(consent.id, consent);
    return consent;
  }

  getConsent(id: string): Consent | undefined {
    return this.consents.get(id);
  }

  getConsentsByUserAndCategory(userId: string, categoryId: string): Consent[] {
    return Array.from(this.consents.values()).filter(c => c.userId === userId && c.categoryId === categoryId);
  }
}

export const store = InMemoryStore.getInstance();
