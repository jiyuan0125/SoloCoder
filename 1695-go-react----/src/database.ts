import sqlite3 from 'sqlite3';
import path from 'path';

const DB_PATH = path.join(__dirname, '../emergency.db');

export const db = new sqlite3.Database(DB_PATH, (err) => {
  if (err) {
    console.error('数据库连接失败:', err);
  } else {
    console.log('已连接到 SQLite 数据库');
  }
});

export const initDatabase = (): Promise<void> => {
  return new Promise((resolve, reject) => {
    db.serialize(() => {
      db.run(`
        CREATE TABLE IF NOT EXISTS shelters (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          name TEXT NOT NULL UNIQUE,
          location TEXT NOT NULL,
          type TEXT NOT NULL,
          area REAL NOT NULL,
          capacity INTEGER NOT NULL,
          occupied INTEGER DEFAULT 0,
          facilities TEXT NOT NULL,
          manager TEXT NOT NULL,
          status TEXT NOT NULL DEFAULT 'AVAILABLE',
          latitude REAL,
          longitude REAL,
          created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
          updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
      `, (err) => {
        if (err) reject(err);
      });

      db.run(`
        CREATE TABLE IF NOT EXISTS materials (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          name TEXT NOT NULL,
          shelter_id INTEGER NOT NULL,
          quantity INTEGER NOT NULL,
          expiry_date TEXT NOT NULL,
          min_stock INTEGER NOT NULL,
          created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
          updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
          FOREIGN KEY (shelter_id) REFERENCES shelters(id) ON DELETE CASCADE
        )
      `, (err) => {
        if (err) reject(err);
      });

      db.run(`
        CREATE TABLE IF NOT EXISTS replenishment_todos (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          material_id INTEGER NOT NULL,
          shelter_id INTEGER NOT NULL,
          material_name TEXT NOT NULL,
          current_stock INTEGER NOT NULL,
          min_stock INTEGER NOT NULL,
          required INTEGER NOT NULL,
          status TEXT NOT NULL DEFAULT 'PENDING',
          created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
          FOREIGN KEY (material_id) REFERENCES materials(id) ON DELETE CASCADE,
          FOREIGN KEY (shelter_id) REFERENCES shelters(id) ON DELETE CASCADE
        )
      `, (err) => {
        if (err) reject(err);
      });

      db.run(`
        CREATE TABLE IF NOT EXISTS transfers (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          from_shelter_id INTEGER NOT NULL,
          to_shelter_id INTEGER NOT NULL,
          material_id INTEGER NOT NULL,
          material_name TEXT NOT NULL,
          quantity INTEGER NOT NULL,
          status TEXT NOT NULL DEFAULT 'PENDING',
          created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
          updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
          FOREIGN KEY (from_shelter_id) REFERENCES shelters(id),
          FOREIGN KEY (to_shelter_id) REFERENCES shelters(id),
          FOREIGN KEY (material_id) REFERENCES materials(id)
        )
      `, (err) => {
        if (err) reject(err);
      });

      db.run(`
        CREATE TABLE IF NOT EXISTS assignments (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          shelter_id INTEGER NOT NULL,
          people_count INTEGER NOT NULL,
          reason TEXT,
          created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
          FOREIGN KEY (shelter_id) REFERENCES shelters(id)
        )
      `, (err) => {
        if (err) reject(err);
      });

      console.log('数据库表初始化完成');
      resolve();
    });
  });
};

export const runQuery = <T>(sql: string, params: any[] = []): Promise<T> => {
  return new Promise((resolve, reject) => {
    db.all(sql, params, (err, rows) => {
      if (err) reject(err);
      else resolve(rows as T);
    });
  });
};

export const runInsert = (sql: string, params: any[] = []): Promise<number> => {
  return new Promise((resolve, reject) => {
    db.run(sql, params, function (err) {
      if (err) reject(err);
      else resolve(this.lastID);
    });
  });
};

export const runUpdate = (sql: string, params: any[] = []): Promise<number> => {
  return new Promise((resolve, reject) => {
    db.run(sql, params, function (err) {
      if (err) reject(err);
      else resolve(this.changes);
    });
  });
};
