import { v4 as uuidv4 } from 'uuid';
import { Content, ContentStatus } from '../types';

class ContentStore {
  private contents: Map<string, Content> = new Map();

  create(text?: string, imageUrl?: string): Content {
    const id = uuidv4();
    const content: Content = {
      id,
      text,
      imageUrl,
      status: ContentStatus.SUBMITTED,
      submittedAt: new Date(),
    };
    this.contents.set(id, content);
    return content;
  }

  findById(id: string): Content | undefined {
    return this.contents.get(id);
  }

  findAll(): Content[] {
    return Array.from(this.contents.values());
  }

  update(id: string, updates: Partial<Content>): Content | undefined {
    const existing = this.contents.get(id);
    if (!existing) return undefined;
    
    const updated: Content = { ...existing, ...updates };
    this.contents.set(id, updated);
    return updated;
  }
}

export const contentStore = new ContentStore();
