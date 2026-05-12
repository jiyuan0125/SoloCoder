import { getDb, Document } from '../database';
import { tokenize } from '../tokenizer';

function generateId(): string {
  return `doc_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
}

function getTokensWithWeight(title: string, content: string): Map<string, number> {
  const tokens = new Map<string, number>();
  
  const titleResult = tokenize(title);
  for (const t of titleResult.tokens) {
    tokens.set(t, 3);
  }
  
  const contentResult = tokenize(content);
  for (const t of contentResult.tokens) {
    if (!tokens.has(t)) {
      tokens.set(t, 1);
    }
  }
  
  return tokens;
}

export interface CreateDocumentInput {
  id?: string;
  title: string;
  content: string;
}

export interface BatchImportResult {
  successCount: number;
  skippedIds: string[];
  skippedReasons: string[];
}

export function createDocument(input: CreateDocumentInput): Document | null {
  const db = getDb();
  const id = input.id || generateId();
  const now = Date.now();
  
  const tokens = getTokensWithWeight(input.title, input.content);
  
  if (tokens.size === 0) {
    return null;
  }
  
  const insertDoc = db.prepare(`
    INSERT INTO documents (id, title, content, created_at, updated_at)
    VALUES (?, ?, ?, ?, ?)
  `);
  
  const insertIndex = db.prepare(`
    INSERT OR REPLACE INTO inverted_index (token, doc_id, score)
    VALUES (?, ?, ?)
  `);
  
  const transaction = db.transaction(() => {
    insertDoc.run(id, input.title, input.content, now, now);
    
    for (const [token, score] of tokens) {
      insertIndex.run(token, id, score);
    }
  });
  
  transaction();
  
  return {
    id,
    title: input.title,
    content: input.content,
    createdAt: now,
    updatedAt: now
  };
}

export function batchImportDocuments(docs: CreateDocumentInput[]): BatchImportResult {
  const successIds: string[] = [];
  const skippedIds: string[] = [];
  const skippedReasons: string[] = [];
  
  for (const doc of docs) {
    const result = createDocument(doc);
    if (result) {
      successIds.push(result.id);
    } else {
      const docId = doc.id || '(未指定ID)';
      skippedIds.push(docId);
      skippedReasons.push('分词结果为空，可能全是特殊字符或停用词');
    }
  }
  
  return {
    successCount: successIds.length,
    skippedIds,
    skippedReasons
  };
}

export function getDocument(id: string): Document | null {
  const db = getDb();
  const row = db.prepare(`
    SELECT id, title, content, created_at as createdAt, updated_at as updatedAt
    FROM documents WHERE id = ?
  `).get(id) as any;
  
  return row || null;
}

export function updateDocument(id: string, updates: { title?: string; content?: string }): Document | null {
  const db = getDb();
  const doc = getDocument(id);
  
  if (!doc) {
    return null;
  }
  
  const newTitle = updates.title !== undefined ? updates.title : doc.title;
  const newContent = updates.content !== undefined ? updates.content : doc.content;
  const now = Date.now();
  
  const tokens = getTokensWithWeight(newTitle, newContent);
  
  if (tokens.size === 0) {
    return null;
  }
  
  const updateDoc = db.prepare(`
    UPDATE documents
    SET title = ?, content = ?, updated_at = ?
    WHERE id = ?
  `);
  
  const deleteOldIndex = db.prepare(`
    DELETE FROM inverted_index WHERE doc_id = ?
  `);
  
  const insertIndex = db.prepare(`
    INSERT OR REPLACE INTO inverted_index (token, doc_id, score)
    VALUES (?, ?, ?)
  `);
  
  const transaction = db.transaction(() => {
    updateDoc.run(newTitle, newContent, now, id);
    deleteOldIndex.run(id);
    
    for (const [token, score] of tokens) {
      insertIndex.run(token, id, score);
    }
  });
  
  transaction();
  
  return {
    ...doc,
    title: newTitle,
    content: newContent,
    updatedAt: now
  };
}

export function deleteDocument(id: string): boolean {
  const db = getDb();
  const doc = getDocument(id);
  
  if (!doc) {
    return false;
  }
  
  const deleteDoc = db.prepare('DELETE FROM documents WHERE id = ?');
  const deleteIndex = db.prepare('DELETE FROM inverted_index WHERE doc_id = ?');
  
  const transaction = db.transaction(() => {
    deleteIndex.run(id);
    deleteDoc.run(id);
  });
  
  transaction();
  
  return true;
}
