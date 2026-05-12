export const SCHEMA_STATEMENTS = [
  `CREATE TABLE IF NOT EXISTS elders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    idCard TEXT NOT NULL UNIQUE,
    gender TEXT,
    birthDate TEXT,
    emergencyContact TEXT NOT NULL,
    emergencyContactPhone TEXT NOT NULL,
    medicalHistory TEXT,
    careLevel TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    checkInDate TEXT NOT NULL,
    checkOutDate TEXT,
    roomNumber TEXT,
    bedNumber TEXT,
    depositAmount INTEGER NOT NULL DEFAULT 0,
    createdAt TEXT NOT NULL
  )`,

  `CREATE TABLE IF NOT EXISTS carePlans (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    elderId INTEGER NOT NULL,
    medicationReminder TEXT,
    dietArrangement TEXT,
    rehabilitationProject TEXT,
    createdAt TEXT NOT NULL,
    updatedAt TEXT NOT NULL,
    FOREIGN KEY (elderId) REFERENCES elders(id)
  )`,

  `CREATE TABLE IF NOT EXISTS carePlanChanges (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    carePlanId INTEGER NOT NULL,
    changeReason TEXT NOT NULL,
    approvedBy TEXT NOT NULL,
    changedAt TEXT NOT NULL,
    snapshot TEXT NOT NULL,
    FOREIGN KEY (carePlanId) REFERENCES carePlans(id)
  )`,

  `CREATE TABLE IF NOT EXISTS careRecords (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    elderId INTEGER NOT NULL,
    recordDate TEXT NOT NULL,
    temperature REAL NOT NULL,
    bloodPressure TEXT NOT NULL,
    diet TEXT,
    specialNotes TEXT,
    createdBy TEXT,
    createdAt TEXT NOT NULL,
    FOREIGN KEY (elderId) REFERENCES elders(id)
  )`,

  `CREATE TABLE IF NOT EXISTS healthWarnings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    elderId INTEGER NOT NULL,
    warningType TEXT NOT NULL,
    description TEXT NOT NULL,
    snapshot TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    createdAt TEXT NOT NULL,
    notifiedAt TEXT,
    notifiedTo TEXT,
    FOREIGN KEY (elderId) REFERENCES elders(id)
  )`,

  `CREATE TABLE IF NOT EXISTS doctorNotifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    warningId INTEGER NOT NULL,
    doctorName TEXT NOT NULL,
    notifiedAt TEXT NOT NULL,
    FOREIGN KEY (warningId) REFERENCES healthWarnings(id)
  )`,

  `CREATE TABLE IF NOT EXISTS familyMembers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    elderId INTEGER NOT NULL,
    name TEXT NOT NULL,
    phone TEXT NOT NULL,
    relation TEXT NOT NULL,
    lastViewDate TEXT,
    viewCount INTEGER NOT NULL DEFAULT 0,
    createdAt TEXT NOT NULL,
    FOREIGN KEY (elderId) REFERENCES elders(id)
  )`,

  `CREATE TABLE IF NOT EXISTS familyMessages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    familyMemberId INTEGER NOT NULL,
    elderId INTEGER NOT NULL,
    content TEXT NOT NULL,
    createdAt TEXT NOT NULL,
    FOREIGN KEY (familyMemberId) REFERENCES familyMembers(id),
    FOREIGN KEY (elderId) REFERENCES elders(id)
  )`,

  `CREATE TABLE IF NOT EXISTS visitAppointments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    familyMemberId INTEGER NOT NULL,
    elderId INTEGER NOT NULL,
    visitDate TEXT NOT NULL,
    timeSlot TEXT NOT NULL,
    visitorName TEXT NOT NULL,
    visitorPhone TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    createdAt TEXT NOT NULL,
    FOREIGN KEY (familyMemberId) REFERENCES familyMembers(id),
    FOREIGN KEY (elderId) REFERENCES elders(id),
    UNIQUE(elderId, visitDate, timeSlot)
  )`,

  `CREATE TABLE IF NOT EXISTS billingRecords (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    elderId INTEGER NOT NULL,
    type TEXT NOT NULL,
    amount INTEGER NOT NULL,
    description TEXT NOT NULL,
    relatedRecordId INTEGER,
    createdAt TEXT NOT NULL,
    FOREIGN KEY (elderId) REFERENCES elders(id)
  )`,

  `CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
  )`
];

export const INITIAL_SETTINGS = [
  { key: 'totalBeds', value: '100' }
];
