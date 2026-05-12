import { BaseDao } from './baseDao';
import { DataSource, CreateDataSourceRequest, UpdateDataSourceRequest } from '../models/dataSource';
import { DataSourceConfig } from '../models/step';
import { v4 as uuidv4 } from 'uuid';

interface DataSourceRow {
  id: string;
  name: string;
  description?: string;
  config: string;
  is_active: number;
  created_at: string;
  updated_at: string;
}

export class DataSourceDao extends BaseDao {
  public async create(request: CreateDataSourceRequest): Promise<DataSource> {
    const id = uuidv4();
    const now = new Date().toISOString();
    
    const sql = `
      INSERT INTO data_sources (id, name, description, config, is_active, created_at, updated_at)
      VALUES (?, ?, ?, ?, ?, ?, ?)
    `;
    
    await this.run(sql, [
      id,
      request.name,
      request.description || null,
      JSON.stringify(request.config),
      1,
      now,
      now
    ]);
    
    return this.findById(id) as Promise<DataSource>;
  }

  public async findById(id: string): Promise<DataSource | undefined> {
    const sql = `SELECT * FROM data_sources WHERE id = ?`;
    const row = await this.get<DataSourceRow>(sql, [id]);
    
    if (!row) {
      return undefined;
    }
    
    return this.mapToDataSource(row);
  }

  public async findAll(): Promise<DataSource[]> {
    const sql = `SELECT * FROM data_sources ORDER BY created_at DESC`;
    const rows = await this.all<DataSourceRow>(sql);
    
    return rows.map(row => this.mapToDataSource(row));
  }

  public async update(id: string, request: UpdateDataSourceRequest): Promise<DataSource | undefined> {
    const existingDataSource = await this.findById(id);
    if (!existingDataSource) {
      return undefined;
    }

    const now = new Date().toISOString();
    const updates: string[] = [];
    const params: any[] = [];

    if (request.name !== undefined) {
      updates.push('name = ?');
      params.push(request.name);
    }
    if (request.description !== undefined) {
      updates.push('description = ?');
      params.push(request.description);
    }
    if (request.config !== undefined) {
      updates.push('config = ?');
      params.push(JSON.stringify(request.config));
    }
    if (request.isActive !== undefined) {
      updates.push('is_active = ?');
      params.push(request.isActive ? 1 : 0);
    }
    
    updates.push('updated_at = ?');
    params.push(now);
    params.push(id);

    const sql = `UPDATE data_sources SET ${updates.join(', ')} WHERE id = ?`;
    await this.run(sql, params);
    
    return this.findById(id);
  }

  public async delete(id: string): Promise<boolean> {
    const sql = `DELETE FROM data_sources WHERE id = ?`;
    await this.run(sql, [id]);
    return true;
  }

  private mapToDataSource(row: DataSourceRow): DataSource {
    return {
      id: row.id,
      name: row.name,
      description: row.description,
      config: JSON.parse(row.config) as DataSourceConfig,
      isActive: row.is_active === 1,
      createdAt: row.created_at,
      updatedAt: row.updated_at
    };
  }
}
