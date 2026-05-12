import { DataRecord } from '../types';

export interface DataSource {
  systemId: string;
  getAllRecords(): Promise<DataRecord[]>;
  getRecordsAfterWatermark(watermark: number): Promise<DataRecord[]>;
  getRecordById(id: string): Promise<DataRecord | undefined>;
  upsertRecord(record: DataRecord): Promise<void>;
  deleteRecord(id: string): Promise<void>;
}

const inMemoryStores: Record<string, Record<string, DataRecord>> = {};

export class InMemoryDataSource implements DataSource {
  constructor(public systemId: string) {
    if (!inMemoryStores[systemId]) {
      inMemoryStores[systemId] = {};
    }
  }

  private get store(): Record<string, DataRecord> {
    return inMemoryStores[this.systemId];
  }

  async getAllRecords(): Promise<DataRecord[]> {
    return Object.values(this.store);
  }

  async getRecordsAfterWatermark(watermark: number): Promise<DataRecord[]> {
    return Object.values(this.store).filter((r) => r.updatedAt > watermark);
  }

  async getRecordById(id: string): Promise<DataRecord | undefined> {
    return this.store[id];
  }

  async upsertRecord(record: DataRecord): Promise<void> {
    this.store[record.id] = { ...record };
  }

  async deleteRecord(id: string): Promise<void> {
    delete this.store[id];
  }

  seedData(records: DataRecord[]): void {
    for (const r of records) {
      this.store[r.id] = { ...r };
    }
  }
}

const dataSources: Record<string, DataSource> = {};

export function getDataSource(systemId: string): DataSource {
  if (!dataSources[systemId]) {
    dataSources[systemId] = new InMemoryDataSource(systemId);
  }
  return dataSources[systemId];
}

export function registerDataSource(source: DataSource): void {
  dataSources[source.systemId] = source;
}
