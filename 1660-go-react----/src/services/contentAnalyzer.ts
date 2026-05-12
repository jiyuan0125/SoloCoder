import type { ContentAnalysisResult } from '../types';

export function analyzeContent(content: string): ContentAnalysisResult {
  const reasons: string[] = [];

  if (/https?:\/\//i.test(content)) {
    reasons.push('contains_external_link');
  }

  if (/(.)\1{5,}/.test(content)) {
    reasons.push('consecutive_repeated_chars');
  }

  const letters = content.match(/[a-zA-Z]/g);
  if (letters && letters.length > 0) {
    const uppercaseCount = letters.filter(c => c === c.toUpperCase()).length;
    const uppercaseRatio = uppercaseCount / letters.length;
    if (uppercaseRatio > 0.5) {
      reasons.push('high_uppercase_ratio');
    }
  }

  return {
    isSuspicious: reasons.length > 0,
    reasons,
  };
}