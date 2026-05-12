import { v4 as uuidv4 } from 'uuid';
import { SensitiveWord, SensitiveWordLevel } from '../types';

class SensitiveWordStore {
  private words: Map<string, SensitiveWord> = new Map();

  create(word: string, level: SensitiveWordLevel): SensitiveWord {
    const id = uuidv4();
    const now = new Date();
    const sensitiveWord: SensitiveWord = {
      id,
      word,
      level,
      createdAt: now,
      updatedAt: now,
    };
    this.words.set(id, sensitiveWord);
    return sensitiveWord;
  }

  findById(id: string): SensitiveWord | undefined {
    return this.words.get(id);
  }

  findAll(): SensitiveWord[] {
    return Array.from(this.words.values());
  }

  update(id: string, word?: string, level?: SensitiveWordLevel): SensitiveWord | undefined {
    const existing = this.words.get(id);
    if (!existing) return undefined;
    
    const updated: SensitiveWord = {
      ...existing,
      word: word ?? existing.word,
      level: level ?? existing.level,
      updatedAt: new Date(),
    };
    this.words.set(id, updated);
    return updated;
  }

  delete(id: string): boolean {
    return this.words.delete(id);
  }

  matchText(text: string): { word: SensitiveWord; matched: string }[] {
    const results: { word: SensitiveWord; matched: string }[] = [];
    const normalizedText = text.replace(/\s+/g, '').toLowerCase();
    
    for (const sensitiveWord of this.words.values()) {
      const normalizedWord = sensitiveWord.word.replace(/\s+/g, '').toLowerCase();
      if (normalizedText.includes(normalizedWord)) {
        results.push({ word: sensitiveWord, matched: sensitiveWord.word });
      }
    }
    
    return results;
  }
}

export const sensitiveWordStore = new SensitiveWordStore();
