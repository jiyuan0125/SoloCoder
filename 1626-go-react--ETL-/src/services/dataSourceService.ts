import { DataSourceDao } from '../database/dataSourceDao';
import { DataSource, CreateDataSourceRequest, UpdateDataSourceRequest } from '../models/dataSource';
import { DataSourceConfig } from '../models/step';

export class DataSourceService {
  private dataSourceDao: DataSourceDao;

  constructor() {
    this.dataSourceDao = new DataSourceDao();
  }

  public async createDataSource(request: CreateDataSourceRequest): Promise<DataSource> {
    this.validateDataSourceConfig(request.config);
    return this.dataSourceDao.create(request);
  }

  public async getDataSourceById(id: string): Promise<DataSource | undefined> {
    return this.dataSourceDao.findById(id);
  }

  public async getAllDataSources(): Promise<DataSource[]> {
    return this.dataSourceDao.findAll();
  }

  public async updateDataSource(id: string, request: UpdateDataSourceRequest): Promise<DataSource | undefined> {
    if (request.config) {
      this.validateDataSourceConfig(request.config);
    }
    return this.dataSourceDao.update(id, request);
  }

  public async deleteDataSource(id: string): Promise<boolean> {
    return this.dataSourceDao.delete(id);
  }

  public async fetchDataFromDataSource(dataSourceId: string): Promise<any[]> {
    const dataSource = await this.dataSourceDao.findById(dataSourceId);
    
    if (!dataSource) {
      throw new Error(`Data source '${dataSourceId}' not found`);
    }
    
    if (!dataSource.isActive) {
      throw new Error(`Data source '${dataSourceId}' is not active`);
    }
    
    return this.fetchData(dataSource.config);
  }

  private async fetchData(config: DataSourceConfig): Promise<any[]> {
    switch (config.type) {
      case 'http':
        return this.fetchFromHttp(config);
      case 'memory':
        return this.fetchFromMemory(config);
      default:
        throw new Error(`Unsupported data source type: ${config.type}`);
    }
  }

  private async fetchFromHttp(config: DataSourceConfig): Promise<any[]> {
    if (!config.httpConfig) {
      throw new Error('HTTP config is required for HTTP data source');
    }
    
    const { url, method = 'GET', headers = {}, body } = config.httpConfig;
    
    try {
      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
          ...headers
        },
        body: body ? JSON.stringify(body) : undefined
      });
      
      if (!response.ok) {
        throw new Error(`HTTP request failed with status ${response.status}`);
      }
      
      const data = await response.json();
      
      if (Array.isArray(data)) {
        return data;
      } else if (typeof data === 'object' && data !== null) {
        return [data];
      } else {
        return [];
      }
    } catch (error) {
      throw new Error(`Failed to fetch data from HTTP: ${(error as Error).message}`);
    }
  }

  private async fetchFromMemory(config: DataSourceConfig): Promise<any[]> {
    if (!config.memoryConfig || !config.memoryConfig.data) {
      throw new Error('Memory data is required for memory data source');
    }
    
    return config.memoryConfig.data;
  }

  private validateDataSourceConfig(config: DataSourceConfig): void {
    if (!config.type) {
      throw new Error('Data source type is required');
    }
    
    if (config.type === 'http' && !config.httpConfig) {
      throw new Error('HTTP config is required for HTTP data source');
    }
    
    if (config.type === 'memory' && (!config.memoryConfig || !config.memoryConfig.data)) {
      throw new Error('Memory data is required for memory data source');
    }
  }
}
