import Database from 'better-sqlite3';
import path from 'path';

const dbPath = path.join(__dirname, '..', 'community.db');
const db = new Database(dbPath);

db.pragma('journal_mode = WAL');

db.exec(`
  CREATE TABLE IF NOT EXISTS announcements (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('通知', '紧急', '活动预告', '维修公告')),
    scope_type TEXT NOT NULL CHECK(scope_type IN ('全小区', '指定楼栋')),
    scope_buildings TEXT,
    is_withdrawn INTEGER DEFAULT 0,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    edit_count INTEGER DEFAULT 0,
    edit_summary TEXT
  );

  CREATE TABLE IF NOT EXISTS repairs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category TEXT NOT NULL CHECK(category IN ('水电', '门窗', '电梯', '公共设施', '其他')),
    location TEXT NOT NULL,
    description TEXT NOT NULL,
    phone TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT '待接单' CHECK(status IN ('待接单', '已分配', '处理中', '已完成')),
    priority TEXT NOT NULL DEFAULT '普通' CHECK(priority IN ('普通', '紧急')),
    assigned_staff_id INTEGER,
    assigned_at INTEGER,
    completed_at INTEGER,
    evaluation_score INTEGER,
    evaluation_comment TEXT,
    created_at INTEGER NOT NULL
  );

  CREATE TABLE IF NOT EXISTS staff (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    phone TEXT NOT NULL,
    created_at INTEGER NOT NULL
  );

  CREATE TABLE IF NOT EXISTS residents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    building TEXT NOT NULL,
    unit TEXT NOT NULL,
    room TEXT NOT NULL,
    created_at INTEGER NOT NULL
  );

  CREATE TABLE IF NOT EXISTS fee_bills (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    resident_id INTEGER NOT NULL,
    year INTEGER NOT NULL,
    quarter INTEGER NOT NULL CHECK(quarter IN (1, 2, 3, 4)),
    property_fee INTEGER NOT NULL DEFAULT 0,
    water_fee INTEGER NOT NULL DEFAULT 0,
    electricity_fee INTEGER NOT NULL DEFAULT 0,
    gas_fee INTEGER NOT NULL DEFAULT 0,
    other_fee INTEGER NOT NULL DEFAULT 0,
    late_fee INTEGER NOT NULL DEFAULT 0,
    is_paid INTEGER DEFAULT 0,
    paid_at INTEGER,
    due_date INTEGER NOT NULL,
    created_at INTEGER NOT NULL,
    FOREIGN KEY (resident_id) REFERENCES residents(id),
    UNIQUE(resident_id, year, quarter)
  );

  CREATE TABLE IF NOT EXISTS receipts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    bill_id INTEGER NOT NULL,
    amount INTEGER NOT NULL,
    payment_method TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    FOREIGN KEY (bill_id) REFERENCES fee_bills(id)
  );

  CREATE TABLE IF NOT EXISTS announcement_edits (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    announcement_id INTEGER NOT NULL,
    old_title TEXT,
    old_content TEXT,
    new_title TEXT,
    new_content TEXT,
    summary TEXT,
    edited_at INTEGER NOT NULL,
    FOREIGN KEY (announcement_id) REFERENCES announcements(id)
  );
`);

const staffCount = db.prepare('SELECT COUNT(*) as count FROM staff').get() as { count: number };
if (staffCount.count === 0) {
  const now = Date.now();
  const insertStaff = db.prepare('INSERT INTO staff (name, phone, created_at) VALUES (?, ?, ?)');
  insertStaff.run('张工', '13800138001', now);
  insertStaff.run('李工', '13800138002', now);
  insertStaff.run('王工', '13800138003', now);
}

const residentCount = db.prepare('SELECT COUNT(*) as count FROM residents').get() as { count: number };
if (residentCount.count === 0) {
  const now = Date.now();
  const insertResident = db.prepare('INSERT INTO residents (name, building, unit, room, created_at) VALUES (?, ?, ?, ?, ?)');
  insertResident.run('张三', '1号楼', '1单元', '101', now);
  insertResident.run('李四', '1号楼', '2单元', '202', now);
  insertResident.run('王五', '2号楼', '1单元', '301', now);
}

export default db;
