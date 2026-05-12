import Database from 'better-sqlite3';
import path from 'path';

const dbPath = path.join(__dirname, '..', 'volunteer.db');
const db = new Database(dbPath);

db.exec(`
  CREATE TABLE IF NOT EXISTS volunteers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    phone TEXT UNIQUE NOT NULL,
    id_card TEXT,
    skills TEXT NOT NULL,
    availability TEXT NOT NULL,
    total_hours REAL DEFAULT 0,
    is_backbone INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
  );

  CREATE TABLE IF NOT EXISTS activities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    date TEXT NOT NULL,
    start_time TEXT NOT NULL,
    end_time TEXT NOT NULL,
    location TEXT NOT NULL,
    required_people INTEGER NOT NULL,
    skills_required TEXT NOT NULL,
    created_by INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES volunteers(id)
  );

  CREATE TABLE IF NOT EXISTS activity_registrations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    volunteer_id INTEGER NOT NULL,
    activity_id INTEGER NOT NULL,
    status TEXT DEFAULT 'registered',
    actual_start_time TEXT,
    actual_end_time TEXT,
    calculated_hours REAL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (volunteer_id) REFERENCES volunteers(id),
    FOREIGN KEY (activity_id) REFERENCES activities(id),
    UNIQUE(volunteer_id, activity_id)
  );

  CREATE TABLE IF NOT EXISTS meeting_rooms (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    capacity INTEGER NOT NULL
  );

  CREATE TABLE IF NOT EXISTS room_bookings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    room_id INTEGER NOT NULL,
    date TEXT NOT NULL,
    start_time TEXT NOT NULL,
    end_time TEXT NOT NULL,
    team_name TEXT NOT NULL,
    purpose TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (room_id) REFERENCES meeting_rooms(id)
  );

  CREATE TABLE IF NOT EXISTS service_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    volunteer_id INTEGER NOT NULL,
    activity_id INTEGER NOT NULL,
    hours REAL NOT NULL,
    year INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (volunteer_id) REFERENCES volunteers(id),
    FOREIGN KEY (activity_id) REFERENCES activities(id)
  );
`);

const insertRoom = db.prepare(`INSERT OR IGNORE INTO meeting_rooms (name, capacity) VALUES (?, ?)`);
insertRoom.run('会议室A', 20);
insertRoom.run('会议室B', 30);
insertRoom.run('会议室C', 10);

export default db;
