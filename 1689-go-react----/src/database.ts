import Database from 'better-sqlite3';
import path from 'path';

const db = new Database(path.join(__dirname, '../data.db'));

export function initDatabase(): void {
  db.exec(`
    CREATE TABLE IF NOT EXISTS users (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      username TEXT UNIQUE NOT NULL,
      password TEXT NOT NULL,
      name TEXT NOT NULL,
      role TEXT NOT NULL CHECK (role IN ('social_worker', 'supervisor'))
    );

    CREATE TABLE IF NOT EXISTS cases (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      social_worker_id INTEGER NOT NULL,
      client_name TEXT NOT NULL,
      client_info TEXT NOT NULL,
      problem_description TEXT NOT NULL,
      needs_assessment TEXT NOT NULL,
      intervention_plan TEXT NOT NULL,
      status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'in_progress', 'extended', 'closed', 'archived')),
      start_date TEXT NOT NULL,
      estimated_end_date TEXT NOT NULL,
      actual_end_date TEXT,
      supervisor_approved BOOLEAN DEFAULT 0,
      created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      FOREIGN KEY (social_worker_id) REFERENCES users(id)
    );

    CREATE TABLE IF NOT EXISTS case_sessions (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      case_id INTEGER NOT NULL,
      session_date TEXT NOT NULL,
      duration INTEGER NOT NULL,
      summary TEXT NOT NULL,
      next_plan TEXT NOT NULL,
      created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      FOREIGN KEY (case_id) REFERENCES cases(id)
    );

    CREATE TABLE IF NOT EXISTS case_extensions (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      case_id INTEGER NOT NULL,
      reason TEXT NOT NULL,
      new_end_date TEXT NOT NULL,
      approved BOOLEAN DEFAULT 0,
      approved_by INTEGER,
      created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      FOREIGN KEY (case_id) REFERENCES cases(id),
      FOREIGN KEY (approved_by) REFERENCES users(id)
    );

    CREATE TABLE IF NOT EXISTS case_archives (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      case_id INTEGER NOT NULL,
      archive_data TEXT NOT NULL,
      archived_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      FOREIGN KEY (case_id) REFERENCES cases(id)
    );

    CREATE TABLE IF NOT EXISTS group_activities (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL,
      goal TEXT NOT NULL,
      activity_count INTEGER NOT NULL,
      completed_count INTEGER NOT NULL DEFAULT 0,
      time_details TEXT NOT NULL,
      location TEXT NOT NULL,
      max_participants INTEGER NOT NULL CHECK (max_participants BETWEEN 6 AND 12),
      status TEXT NOT NULL DEFAULT 'planning' CHECK (status IN ('planning', 'active', 'completed')),
      created_by INTEGER NOT NULL,
      created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      summary_report TEXT,
      FOREIGN KEY (created_by) REFERENCES users(id)
    );

    CREATE TABLE IF NOT EXISTS group_members (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      group_activity_id INTEGER NOT NULL,
      name TEXT NOT NULL,
      contact_info TEXT,
      status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'removed')),
      approved_by INTEGER,
      approved_at TEXT,
      consecutive_absences INTEGER NOT NULL DEFAULT 0,
      created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      FOREIGN KEY (group_activity_id) REFERENCES group_activities(id),
      FOREIGN KEY (approved_by) REFERENCES users(id)
    );

    CREATE TABLE IF NOT EXISTS group_attendance (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      group_activity_id INTEGER NOT NULL,
      member_id INTEGER NOT NULL,
      session_number INTEGER NOT NULL,
      attended BOOLEAN NOT NULL DEFAULT 0,
      leave_approved BOOLEAN NOT NULL DEFAULT 0,
      FOREIGN KEY (group_activity_id) REFERENCES group_activities(id),
      FOREIGN KEY (member_id) REFERENCES group_members(id)
    );

    CREATE TABLE IF NOT EXISTS community_services (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      service_date TEXT NOT NULL,
      content TEXT NOT NULL,
      participants INTEGER NOT NULL,
      effect TEXT NOT NULL,
      created_by INTEGER NOT NULL,
      created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      FOREIGN KEY (created_by) REFERENCES users(id)
    );

    CREATE TABLE IF NOT EXISTS case_reopen_requests (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      case_id INTEGER NOT NULL,
      requested_by INTEGER NOT NULL,
      reason TEXT NOT NULL,
      status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
      approved_by INTEGER,
      approved_at TEXT,
      created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
      FOREIGN KEY (case_id) REFERENCES cases(id),
      FOREIGN KEY (requested_by) REFERENCES users(id),
      FOREIGN KEY (approved_by) REFERENCES users(id)
    );
  `);

  const userCount = db.prepare('SELECT COUNT(*) as count FROM users').get() as { count: number };
  if (userCount.count === 0) {
    const bcrypt = require('bcryptjs');
    const supervisorPassword = bcrypt.hashSync('supervisor123', 10);
    const workerPassword = bcrypt.hashSync('worker123', 10);

    db.prepare(`
      INSERT INTO users (username, password, name, role) VALUES (?, ?, ?, ?)
    `).run('supervisor', supervisorPassword, '张主管', 'supervisor');

    db.prepare(`
      INSERT INTO users (username, password, name, role) VALUES (?, ?, ?, ?)
    `).run('worker1', workerPassword, '李社工', 'social_worker');

    db.prepare(`
      INSERT INTO users (username, password, name, role) VALUES (?, ?, ?, ?)
    `).run('worker2', workerPassword, '王社工', 'social_worker');
  }
}

export default db;
