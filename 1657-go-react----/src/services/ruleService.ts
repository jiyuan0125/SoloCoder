import { Rule } from '../types';
import { RuleRepository } from '../repositories/ruleRepository';
import { RuleEngine } from '../engine/ruleEngine';
import { validateCondition } from '../engine/conditionParser';

export class RuleValidationError extends Error {
  code: string;
  field: string;

  constructor(message: string, field: string, code: string = 'VALIDATION_ERROR') {
    super(message);
    this.code = code;
    this.field = field;
  }
}

export class RuleService {
  private static instance: RuleService;
  private ruleRepo: RuleRepository;
  private ruleEngine: RuleEngine;

  static getInstance(): RuleService {
    if (!RuleService.instance) {
      RuleService.instance = new RuleService();
    }
    return RuleService.instance;
  }

  constructor() {
    this.ruleRepo = RuleRepository.getInstance();
    this.ruleEngine = RuleEngine.getInstance();
  }

  private validateRule(rule: Partial<Omit<Rule, 'id' | 'createdAt' | 'updatedAt'>>): void {
    if (rule.name !== undefined) {
      if (!rule.name || rule.name.trim() === '') {
        throw new RuleValidationError('Rule name cannot be empty', 'name');
      }
    }

    if (rule.weight !== undefined) {
      if (typeof rule.weight !== 'number' || rule.weight < 1 || rule.weight > 100) {
        throw new RuleValidationError('Weight must be between 1 and 100', 'weight', 'INVALID_WEIGHT');
      }
    }

    if (rule.condition !== undefined) {
      if (!rule.condition || rule.condition.trim() === '') {
        throw new RuleValidationError('Condition cannot be empty', 'condition');
      }

      if (!validateCondition(rule.condition)) {
        throw new RuleValidationError('Invalid condition expression syntax', 'condition', 'INVALID_CONDITION');
      }
    }
  }

  create(rule: Omit<Rule, 'id' | 'createdAt' | 'updatedAt'>): Rule {
    this.validateRule(rule);

    const created = this.ruleRepo.create(rule);
    this.ruleEngine.refresh();
    return created;
  }

  update(id: string, updates: Partial<Omit<Rule, 'id' | 'createdAt'>>): Rule | null {
    const existing = this.ruleRepo.findById(id);
    if (!existing) return null;

    this.validateRule(updates);

    const updated = this.ruleRepo.update(id, updates);
    if (updated) {
      this.ruleEngine.refresh();
    }
    return updated;
  }

  delete(id: string): boolean {
    const deleted = this.ruleRepo.delete(id);
    if (deleted) {
      this.ruleEngine.refresh();
    }
    return deleted;
  }

  findById(id: string): Rule | null {
    return this.ruleRepo.findById(id);
  }

  findAll(): Rule[] {
    return this.ruleRepo.findAll();
  }

  findAllEnabled(): Rule[] {
    return this.ruleRepo.findAllEnabled();
  }
}
