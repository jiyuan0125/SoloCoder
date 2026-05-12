import Database from 'better-sqlite3';
import path from 'path';

const dbPath = path.join(process.cwd(), 'observatory.db');

const db = new Database(dbPath);

db.pragma('journal_mode = WAL');
db.pragma('foreign_keys = ON');

db.exec(`
  CREATE TABLE IF NOT EXISTS equipment (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
  );

  CREATE TABLE IF NOT EXISTS observation_tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    target_object TEXT NOT NULL,
    start_time TEXT,
    end_time TEXT,
    equipment_id INTEGER,
    equipment_name TEXT,
    responsible_person TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT '待排期',
    scheduled_month TEXT NOT NULL,
    is_canceled INTEGER NOT NULL DEFAULT 0,
    cancel_reason TEXT,
    fault_description TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
  );

  CREATE TABLE IF NOT EXISTS equipment_bookings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    equipment_id INTEGER NOT NULL,
    task_id INTEGER NOT NULL,
    start_time TEXT NOT NULL,
    end_time TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (equipment_id) REFERENCES equipment(id),
    FOREIGN KEY (task_id) REFERENCES observation_tasks(id)
  );

  CREATE INDEX IF NOT EXISTS idx_bookings_equipment_time 
    ON equipment_bookings(equipment_id, start_time, end_time);

  CREATE TABLE IF NOT EXISTS public_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    start_time TEXT NOT NULL,
    end_time TEXT NOT NULL,
    max_capacity INTEGER NOT NULL DEFAULT 50,
    current_bookings INTEGER NOT NULL DEFAULT 0,
    is_special INTEGER NOT NULL DEFAULT 0,
    is_canceled INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
  );

  CREATE TABLE IF NOT EXISTS bookings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id INTEGER NOT NULL,
    phone TEXT NOT NULL,
    seats INTEGER NOT NULL,
    is_waitlist INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (event_id) REFERENCES public_events(id)
  );

  CREATE INDEX IF NOT EXISTS idx_bookings_event_phone 
    ON bookings(event_id, phone);

  CREATE INDEX IF NOT EXISTS idx_bookings_waitlist 
    ON bookings(event_id, is_waitlist, created_at);
`);

const equipmentCount = db.prepare('SELECT COUNT(*) as count FROM equipment').get() as { count: number };
if (equipmentCount.count === 0) {
  const stmt = db.prepare('INSERT INTO equipment (name, description) VALUES (?, ?)');
  stmt.run('天文望远镜A型', '口径200mm的反射式望远镜，适合深空观测');
  stmt.run('天文望远镜B型', '口径150mm的折射式望远镜，适合行星观测');
  stmt.run('光谱仪', '用于天体光谱分析');
  stmt.run('射电天线', '用于射电天文观测');
}

export { db };
