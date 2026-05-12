import { getDb, Document, HotWord } from '../database';
import { tokenize } from '../tokenizer';

export interface SearchResultItem {
  document: Document;
  score: number;
  highlightedTitle: string;
  highlightedContent: string;
}

export interface SearchResult {
  success: boolean;
  message?: string;
  results: SearchResultItem[];
  total: number;
}

function highlightText(text: string, tokens: string[]): string {
  let result = text;
  
  const sortedTokens = [...tokens].sort((a, b) => b.length - a.length);
  
  for (const token of sortedTokens) {
    const lowerToken = token.toLowerCase();
    const escapedToken = token.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    const regex = new RegExp(`(${escapedToken})`, 'gi');
    
    result = result.replace(regex, (match) => {
      if (match.toLowerCase() === lowerToken) {
        return `<em>${match}</em>`;
      }
      return match;
    });
  }
  
  return result;
}

function recordSearchStat(query: string, hasResult: boolean): void {
  const db = getDb();
  const stmt = db.prepare(`
    INSERT INTO search_stats (query, has_result, created_at)
    VALUES (?, ?, ?)
  `);
  stmt.run(query, hasResult ? 1 : 0, Date.now());
}

export function search(query: string): SearchResult {
  if (!query || query.trim() === '') {
    return {
      success: false,
      message: '搜索词不能为空',
      results: [],
      total: 0
    };
  }
  
  const tokenizeResult = tokenize(query);
  
  if (tokenizeResult.isAllStopWords) {
    recordSearchStat(query, false);
    return {
      success: false,
      message: '搜索词均为停用词',
      results: [],
      total: 0
    };
  }
  
  const tokens = tokenizeResult.tokens;
  
  if (tokens.length === 0) {
    recordSearchStat(query, false);
    return {
      success: true,
      results: [],
      total: 0
    };
  }
  
  const db = getDb();
  const placeholders = tokens.map(() => '?').join(',');
  
  const rows = db.prepare(`
    SELECT 
      d.id,
      d.title,
      d.content,
      d.created_at as createdAt,
      d.updated_at as updatedAt,
      SUM(ii.score) as totalScore
    FROM inverted_index ii
    JOIN documents d ON ii.doc_id = d.id
    WHERE ii.token IN (${placeholders})
    GROUP BY d.id
    ORDER BY totalScore DESC
  `).all(...tokens) as any[];
  
  const hasResult = rows.length > 0;
  recordSearchStat(query, hasResult);
  
  const results: SearchResultItem[] = rows.map(row => ({
    document: {
      id: row.id,
      title: row.title,
      content: row.content,
      createdAt: row.createdAt,
      updatedAt: row.updatedAt
    },
    score: row.totalScore,
    highlightedTitle: highlightText(row.title, tokens),
    highlightedContent: highlightText(row.content, tokens)
  }));
  
  return {
    success: true,
    results,
    total: results.length
  };
}

export function getSearchSuggestions(prefix: string): string[] {
  if (!prefix || prefix.trim() === '') {
    return [];
  }
  
  const db = getDb();
  const searchPattern = `${prefix.trim()}%`;
  
  const rows = db.prepare(`
    SELECT DISTINCT title
    FROM documents
    WHERE title LIKE ?
    LIMIT 5
  `).all(searchPattern) as any[];
  
  return rows.map(row => row.title);
}

export function getHotWords(limit: number = 20): HotWord[] {
  const db = getDb();
  
  const rows = db.prepare(`
    SELECT query as word, COUNT(*) as count
    FROM search_stats
    GROUP BY query
    ORDER BY count DESC
    LIMIT ?
  `).all(limit) as any[];
  
  return rows;
}

export function getNoResultRate(): { totalSearches: number; noResultCount: number; rate: number } {
  const db = getDb();
  
  const stats = db.prepare(`
    SELECT 
      COUNT(*) as total,
      SUM(CASE WHEN has_result = 0 THEN 1 ELSE 0 END) as noResult
    FROM search_stats
  `).get() as any;
  
  const total = stats.total || 0;
  const noResult = stats.noResult || 0;
  const rate = total > 0 ? noResult / total : 0;
  
  return {
    totalSearches: total,
    noResultCount: noResult,
    rate
  };
}
