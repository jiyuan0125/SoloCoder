import { Rule, Condition, Content, SensitiveWord, ReviewResult, SensitiveWordLevel } from '../types';
import { ruleStore } from '../store/ruleStore';
import { sensitiveWordService } from './sensitiveWordService';

interface EvaluationContext {
  content: Content;
  matchedSensitiveWords: { word: SensitiveWord; matched: string; level: SensitiveWordLevel }[];
}

export class RuleEngine {
  create(
    name: string,
    priority: number,
    conditionLogic: 'AND' | 'OR',
    conditions: Condition[],
    action: 'REJECT' | 'PENDING' | 'PASS' | 'LOG',
    sensitiveWordIds: string[] = []
  ): Rule {
    return ruleStore.create(name, priority, conditionLogic, conditions, action, sensitiveWordIds);
  }

  findById(id: string): Rule | undefined {
    return ruleStore.findById(id);
  }

  findAll(): Rule[] {
    return ruleStore.findAll();
  }

  update(
    id: string,
    name?: string,
    priority?: number,
    conditionLogic?: 'AND' | 'OR',
    conditions?: Condition[],
    action?: 'REJECT' | 'PENDING' | 'PASS' | 'LOG',
    sensitiveWordIds?: string[]
  ): Rule | undefined {
    return ruleStore.update(id, name, priority, conditionLogic, conditions, action, sensitiveWordIds);
  }

  delete(id: string): boolean {
    return ruleStore.delete(id);
  }

  private evaluateCondition(condition: Condition, context: EvaluationContext): boolean {
    const { field, operator, value } = condition;
    const { content, matchedSensitiveWords } = context;

    if (field === 'text') {
      if (!content.text) return false;
      if (operator === 'contains') {
        if (Array.isArray(value)) {
          return value.some(v => (content.text as string).includes(String(v)));
        }
        return (content.text as string).includes(String(value));
      }
      if (operator === 'contains_sensitive_word') {
        return matchedSensitiveWords.length > 0;
      }
      if (operator === 'contains_sensitive_word_level') {
        const levelValue = Number(value);
        return matchedSensitiveWords.some(m => m.level === levelValue);
      }
    }

    if (field === 'has_image') {
      if (operator === 'equals') {
        return !!content.imageUrl === Boolean(value);
      }
    }

    if (field === 'sensitive_word_count') {
      const count = matchedSensitiveWords.length;
      if (operator === 'greater_than') return count > Number(value);
      if (operator === 'less_than') return count < Number(value);
      if (operator === 'equals') return count === Number(value);
    }

    return false;
  }

  private evaluateRule(rule: Rule, context: EvaluationContext): boolean {
    if (rule.conditions.length === 0) return false;
    if (!rule.action) return false;

    const results = rule.conditions.map(cond => this.evaluateCondition(cond, context));

    if (rule.conditionLogic === 'AND') {
      return results.every(r => r);
    } else {
      return results.some(r => r);
    }
  }

  evaluate(
    content: Content,
    matchedWords: { word: SensitiveWord; matched: string }[]
  ): {
    action: 'REJECT' | 'PENDING' | 'PASS' | 'LOG' | null;
    matchedRule?: Rule;
    level1Hits: SensitiveWord[];
    level2Hits: SensitiveWord[];
    level3Hits: SensitiveWord[];
  } {
    const matchedSensitiveWords = matchedWords.map(m => ({
      ...m,
      level: m.word.level,
    }));

    const context: EvaluationContext = {
      content,
      matchedSensitiveWords,
    };

    const level1Hits = matchedSensitiveWords.filter(m => m.level === SensitiveWordLevel.LEVEL_1).map(m => m.word);
    const level2Hits = matchedSensitiveWords.filter(m => m.level === SensitiveWordLevel.LEVEL_2).map(m => m.word);
    const level3Hits = matchedSensitiveWords.filter(m => m.level === SensitiveWordLevel.LEVEL_3).map(m => m.word);

    const rules = this.findAll();

    for (const rule of rules) {
      if (rule.conditions.length === 0 || !rule.action) {
        continue;
      }

      if (this.evaluateRule(rule, context)) {
        return {
          action: rule.action,
          matchedRule: rule,
          level1Hits,
          level2Hits,
          level3Hits,
        };
      }
    }

    if (level1Hits.length > 0) {
      return { action: 'REJECT', level1Hits, level2Hits, level3Hits };
    }
    if (level2Hits.length > 0) {
      return { action: 'PENDING', level1Hits, level2Hits, level3Hits };
    }
    if (level3Hits.length > 0) {
      return { action: 'LOG', level1Hits, level2Hits, level3Hits };
    }

    return { action: null, level1Hits, level2Hits, level3Hits };
  }
}

export const ruleEngine = new RuleEngine();
