import { DEFAULT_DICTIONARY, STOP_WORDS } from './data/dictionary';

let dictionary: Set<string> = new Set(DEFAULT_DICTIONARY);
const stopWords: Set<string> = new Set(STOP_WORDS.map(w => w.toLowerCase()));

export function setDictionary(words: string[]): void {
  dictionary = new Set(words);
}

export function addWord(word: string): void {
  dictionary.add(word);
}

export function removeWord(word: string): void {
  dictionary.delete(word);
}

export function getDictionary(): string[] {
  return Array.from(dictionary);
}

export function isStopWord(word: string): boolean {
  return stopWords.has(word.toLowerCase());
}

function forwardMaxMatch(text: string): string[] {
  const tokens: string[] = [];
  let index = 0;
  
  while (index < text.length) {
    let maxLen = Math.min(10, text.length - index);
    let found = false;
    
    for (let len = maxLen; len >= 1; len--) {
      const candidate = text.substring(index, index + len);
      if (dictionary.has(candidate)) {
        tokens.push(candidate);
        index += len;
        found = true;
        break;
      }
    }
    
    if (!found) {
      tokens.push(text[index]);
      index++;
    }
  }
  
  return tokens;
}

function tokenizeEnglish(text: string): string[] {
  const tokens = text.toLowerCase().split(/[\s.,!?;:()\[\]{}"'\-\_\/\\@#\$%\^&\*\~\+\=\|\<\>\?]+/).filter(t => t.length > 0);
  return tokens;
}

function isChinese(char: string): boolean {
  const code = char.charCodeAt(0);
  return code >= 0x4e00 && code <= 0x9fff;
}

function splitByLanguage(text: string): Array<{ type: 'chinese' | 'english' | 'other'; content: string }> {
  const segments: Array<{ type: 'chinese' | 'english' | 'other'; content: string }> = [];
  let currentType: 'chinese' | 'english' | 'other' | null = null;
  let currentSegment = '';
  
  for (const char of text) {
    let charType: 'chinese' | 'english' | 'other';
    
    if (isChinese(char)) {
      charType = 'chinese';
    } else if (/[a-zA-Z0-9]/.test(char)) {
      charType = 'english';
    } else {
      charType = 'other';
    }
    
    if (currentType !== charType) {
      if (currentType && currentSegment) {
        segments.push({ type: currentType, content: currentSegment });
      }
      currentType = charType;
      currentSegment = char;
    } else {
      currentSegment += char;
    }
  }
  
  if (currentType && currentSegment) {
    segments.push({ type: currentType, content: currentSegment });
  }
  
  return segments;
}

export interface TokenizeResult {
  tokens: string[];
  isAllStopWords: boolean;
}

export function tokenize(text: string): TokenizeResult {
  if (!text || !text.trim()) {
    return { tokens: [], isAllStopWords: true };
  }
  
  const segments = splitByLanguage(text);
  const allTokens: string[] = [];
  
  for (const segment of segments) {
    if (segment.type === 'chinese') {
      const chineseTokens = forwardMaxMatch(segment.content);
      allTokens.push(...chineseTokens);
    } else if (segment.type === 'english') {
      const englishTokens = tokenizeEnglish(segment.content);
      allTokens.push(...englishTokens);
    }
  }
  
  const filteredTokens = allTokens.filter(t => !isStopWord(t));
  
  const isAllStopWords = allTokens.length > 0 && filteredTokens.length === 0;
  
  return {
    tokens: filteredTokens,
    isAllStopWords
  };
}
