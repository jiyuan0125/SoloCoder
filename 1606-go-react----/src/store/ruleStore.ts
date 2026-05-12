import { v4 as uuidv4 } from 'uuid';
import { Rule, Condition } from '../types';

class RuleStore {
  private rules: Map<string, Rule> = new Map();

  create(
    name: string,
    priority: number,
    conditionLogic: 'AND' | 'OR',
    conditions: Condition[],
    action: 'REJECT' | 'PENDING' | 'PASS' | 'LOG',
    sensitiveWordIds: string[] = []
  ): Rule {
    const id = uuidv4();
    const now = new Date();
    const rule: Rule = {
      id,
      name,
      priority,
      conditionLogic,
      conditions,
      action,
      sensitiveWordIds,
      createdAt: now,
      updatedAt: now,
    };
    this.rules.set(id, rule);
    return rule;
  }

  findById(id: string): Rule | undefined {
    return this.rules.get(id);
  }

  findAll(): Rule[] {
    return Array.from(this.rules.values()).sort((a, b) => b.priority - a.priority);
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
    const existing = this.rules.get(id);
    if (!existing) return undefined;
    
    const updated: Rule = {
      ...existing,
      name: name ?? existing.name,
      priority: priority ?? existing.priority,
      conditionLogic: conditionLogic ?? existing.conditionLogic,
      conditions: conditions ?? existing.conditions,
      action: action ?? existing.action,
      sensitiveWordIds: sensitiveWordIds ?? existing.sensitiveWordIds,
      updatedAt: new Date(),
    };
    this.rules.set(id, updated);
    return updated;
  }

  delete(id: string): boolean {
    return this.rules.delete(id);
  }

  findBySensitiveWordId(wordId: string): Rule[] {
    return this.findAll().filter(rule => 
      rule.sensitiveWordIds.includes(wordId)
    );
  }
}

export const ruleStore = new RuleStore();
