import * as sqlite3 from 'sqlite3';
import { Database } from 'sqlite3';

export class DatabaseConnection {
  private static instance: DatabaseConnection;
  private db: Database;

  private constructor() {
    this.db = new sqlite3.Database('./pipeline.db');
    this.initialize();
  }

  public static getInstance(): DatabaseConnection {
    if (!DatabaseConnection.instance) {
      DatabaseConnection.instance = new DatabaseConnection();
    }
    return DatabaseConnection.instance;
  }

  private initialize(): void {
    const createTables = [
      `CREATE TABLE IF NOT EXISTS pipelines (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        description TEXT,
        is_scheduled INTEGER NOT NULL DEFAULT 0,
        schedule_cron TEXT,
        created_at TEXT NOT NULL,
        updated_at TEXT NOT NULL
      )`,
      `CREATE TABLE IF NOT EXISTS data_sources (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        description TEXT,
        config TEXT NOT NULL,
        is_active INTEGER NOT NULL DEFAULT 1,
        created_at TEXT NOT NULL,
        updated_at TEXT NOT NULL
      )`,
      `CREATE TABLE IF NOT EXISTS steps (
        id TEXT PRIMARY KEY,
        pipeline_id TEXT NOT NULL,
        step_number INTEGER NOT NULL,
        name TEXT NOT NULL,
        data_source_id TEXT NOT NULL,
        transformations TEXT NOT NULL,
        target_config TEXT NOT NULL,
        created_at TEXT NOT NULL,
        updated_at TEXT NOT NULL,
        FOREIGN KEY (pipeline_id) REFERENCES pipelines(id),
        FOREIGN KEY (data_source_id) REFERENCES data_sources(id)
      )`,
      `CREATE TABLE IF NOT EXISTS pipeline_executions (
        id TEXT PRIMARY KEY,
        pipeline_id TEXT NOT NULL,
        status TEXT NOT NULL,
        start_time TEXT NOT NULL,
        end_time TEXT,
        failed_step_number INTEGER,
        error_message TEXT,
        created_at TEXT NOT NULL,
        FOREIGN KEY (pipeline_id) REFERENCES pipelines(id)
      )`,
      `CREATE TABLE IF NOT EXISTS step_executions (
        id TEXT PRIMARY KEY,
        execution_id TEXT NOT NULL,
        pipeline_id TEXT NOT NULL,
        step_number INTEGER NOT NULL,
        status TEXT NOT NULL,
        start_time TEXT NOT NULL,
        end_time TEXT,
        records_processed INTEGER NOT NULL DEFAULT 0,
        error_message TEXT,
        data_consumed INTEGER NOT NULL DEFAULT 0,
        FOREIGN KEY (execution_id) REFERENCES pipeline_executions(id),
        FOREIGN KEY (pipeline_id) REFERENCES pipelines(id)
      )`
    ];

    createTables.forEach((sql) => {
      this.db.run(sql);
    });
  }

  public getDb(): Database {
    return this.db;
  }

  public close(): void {
    this.db.close();
  }
}
