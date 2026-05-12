import { TextNormalizeResult } from '../types';

function isChinese(char: string): boolean {
  return /[\u4e00-\u9fa5]/.test(char);
}

function isLetter(char: string): boolean {
  return /[a-zA-Z]/.test(char);
}

function isDigit(char: string): boolean {
  return /[0-9]/.test(char);
}

function isAllowed(char: string): boolean {
  return isChinese(char) || isLetter(char) || isDigit(char);
}

function fullWidthToHalfWidth(char: string): string {
  const code = char.charCodeAt(0);
  if (code >= 0xFF01 && code <= 0xFF5E) {
    return String.fromCharCode(code - 0xFEE0);
  }
  if (code === 0x3000) {
    return ' ';
  }
  return char;
}

export function normalizeText(original: string): TextNormalizeResult {
  const normalizedChars: string[] = [];
  const mappings: { normalizedIndex: number; originalIndex: number }[] = [];
  
  let normalizedIndex = 0;
  let prevSpace = false;

  for (let i = 0; i < original.length; i++) {
    let char = fullWidthToHalfWidth(original[i]);
    
    if (char === ' ') {
      if (!prevSpace) {
        normalizedChars.push(char);
        mappings.push({ normalizedIndex, originalIndex: i });
        normalizedIndex++;
        prevSpace = true;
      }
      continue;
    }

    if (isAllowed(char)) {
      normalizedChars.push(char);
      mappings.push({ normalizedIndex, originalIndex: i });
      normalizedIndex++;
      prevSpace = false;
    }
  }

  return {
    normalized: normalizedChars.join(''),
    mappings
  };
}

export function mapBackToOriginal(
  matches: { word: string; start: number; end: number }[],
  mappings: { normalizedIndex: number; originalIndex: number }[]
): { word: string; start: number; end: number }[] {
  if (mappings.length === 0) {
    return [];
  }

  return matches.map(match => {
    let originalStart = 0;
    let originalEnd = 0;

    for (let i = 0; i < mappings.length; i++) {
      if (mappings[i].normalizedIndex === match.start) {
        originalStart = mappings[i].originalIndex;
      }
      if (mappings[i].normalizedIndex === match.end - 1) {
        originalEnd = mappings[i].originalIndex + 1;
      }
    }

    return {
      word: match.word,
      start: originalStart,
      end: originalEnd
    };
  });
}
