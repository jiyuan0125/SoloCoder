import Database from 'better-sqlite3';

let db: Database.Database | null = null;

export interface Document {
  id: string;
  title: string;
  content: string;
  createdAt: number;
  updatedAt: number;
}

export interface InvertedIndexItem {
  token: string;
  docId: string;
  score: number;
}

export interface SearchStat {
  id?: number;
  query: string;
  hasResult: number;
  createdAt: number;
}

export interface HotWord {
  word: string;
  count: number;
}

export function initDatabase(dbPath: string = './search.db'): Database.Database {
  if (db) return db;
  
  db = new Database(dbPath);
  db.pragma('journal_mode = WAL');
  
  db.exec(`
    CREATE TABLE IF NOT EXISTS documents (
      id TEXT PRIMARY KEY,
      title TEXT NOT NULL,
      content TEXT NOT NULL,
      created_at INTEGER NOT NULL,
      updated_at INTEGER NOT NULL
    );
    
    CREATE TABLE IF NOT EXISTS inverted_index (
      token TEXT NOT NULL,
      doc_id TEXT NOT NULL,
      score INTEGER NOT NULL,
      PRIMARY KEY (token, doc_id),
      FOREIGN KEY (doc_id) REFERENCES documents(id) ON DELETE CASCADE
    );
    
    CREATE INDEX IF NOT EXISTS idx_inverted_index_token ON inverted_index(token);
    CREATE INDEX IF NOT EXISTS idx_inverted_index_doc_id ON inverted_index(doc_id);
    
    CREATE TABLE IF NOT EXISTS search_stats (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      query TEXT NOT NULL,
      has_result INTEGER NOT NULL,
      created_at INTEGER NOT NULL
    );
    
    CREATE INDEX IF NOT EXISTS idx_search_stats_query ON search_stats(query);
    CREATE INDEX IF NOT EXISTS idx_search_stats_created_at ON search_stats(created_at);
  `);
  
  return db;
}

export function getDb(): Database.Database {
  if (!db) {
    return initDatabase();
  }
  return db;
}

export function closeDatabase(): void {
  if (db) {
    db.close();
    db = null;
  }
}
