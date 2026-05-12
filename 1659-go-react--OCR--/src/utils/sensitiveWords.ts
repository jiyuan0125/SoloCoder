import { Trie } from './trie';

export class SensitiveWordManager {
  private trie: Trie = new Trie();
  private words: Set<string> = new Set();
  private version: number = 0;

  private rebuild(): void {
    const newTrie = new Trie();
    for (const word of this.words) {
      newTrie.insert(word);
    }
    this.trie = newTrie;
    this.version++;
  }

  add(word: string): void {
    if (!this.words.has(word)) {
      this.words.add(word);
      this.rebuild();
    }
  }

  remove(word: string): void {
    if (this.words.has(word)) {
      this.words.delete(word);
      this.rebuild();
    }
  }

  update(oldWord: string, newWord: string): void {
    if (this.words.has(oldWord)) {
      this.words.delete(oldWord);
    }
    this.words.add(newWord);
    this.rebuild();
  }

  list(): string[] {
    return Array.from(this.words);
  }

  getVersion(): number {
    return this.version;
  }

  scan(text: string): { word: string; start: number; end: number }[] {
    return this.trie.search(text);
  }
}

export const sensitiveWordManager = new SensitiveWordManager();
