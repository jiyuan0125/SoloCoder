import { Rule, RuleTreeNode, Transaction } from '../types';
import { parseCondition, ParsedCondition } from './conditionParser';
import { RuleRepository } from '../repositories/ruleRepository';

export interface RuleEvaluationResult {
  matched: boolean;
  matchedRules: Rule[];
  totalScore: number;
}

export class RuleEngine {
  private static instance: RuleEngine;
  private ruleRepo: RuleRepository;
  private parsedConditions: Map<string, ParsedCondition>;
  private lastUpdate: number;

  static getInstance(): RuleEngine {
    if (!RuleEngine.instance) {
      RuleEngine.instance = new RuleEngine();
    }
    return RuleEngine.instance;
  }

  constructor() {
    this.ruleRepo = RuleRepository.getInstance();
    this.parsedConditions = new Map();
    this.lastUpdate = 0;
    this.refreshCache();
  }

  private refreshCache(): void {
    const rules = this.ruleRepo.findAll();
    const newCache = new Map<string, ParsedCondition>();

    for (const rule of rules) {
      if (this.parsedConditions.has(rule.id)) {
        const existing = this.parsedConditions.get(rule.id)!;
        newCache.set(rule.id, existing);
      } else {
        try {
          const parsed = parseCondition(rule.condition);
          newCache.set(rule.id, parsed);
        } catch (e) {
          console.warn(`Failed to parse condition for rule ${rule.id}: ${rule.condition}`);
        }
      }
    }

    this.parsedConditions = newCache;
    this.lastUpdate = Date.now();
  }

  private getParsedCondition(rule: Rule): ParsedCondition | null {
    const cached = this.parsedConditions.get(rule.id);
    if (cached) return cached;

    try {
      const parsed = parseCondition(rule.condition);
      this.parsedConditions.set(rule.id, parsed);
      return parsed;
    } catch (e) {
      return null;
    }
  }

  evaluateRule(rule: Rule, transaction: Transaction): boolean {
    if (!rule.enabled) return false;

    const parsed = this.getParsedCondition(rule);
    if (!parsed) return false;

    try {
      return parsed.evaluate(transaction);
    } catch (e) {
      console.warn(`Error evaluating rule ${rule.id}: ${e}`);
      return false;
    }
  }

  evaluateAll(transaction: Transaction): RuleEvaluationResult {
    this.refreshCache();

    const enabledRules = this.ruleRepo.findAllEnabled();
    const matchedRules: Rule[] = [];
    let totalScore = 0;

    for (const rule of enabledRules) {
      if (this.evaluateRule(rule, transaction)) {
        matchedRules.push(rule);
        totalScore += rule.weight;
      }
    }

    return {
      matched: matchedRules.length > 0,
      matchedRules,
      totalScore
    };
  }

  evaluateWithShortCircuit(transaction: Transaction, tree: RuleTreeNode): boolean {
    this.refreshCache();
    return this.evaluateNode(tree, transaction);
  }

  private evaluateNode(node: RuleTreeNode, transaction: Transaction): boolean {
    if (node.type === 'RULE') {
      if (!node.ruleId) return false;

      const rule = this.ruleRepo.findById(node.ruleId);
      if (!rule) return false;

      return this.evaluateRule(rule, transaction);
    }

    if (!node.children || node.children.length === 0) return false;

    if (node.type === 'AND') {
      for (const child of node.children) {
        if (!this.evaluateNode(child, transaction)) {
          return false;
        }
      }
      return true;
    }

    if (node.type === 'OR') {
      for (const child of node.children) {
        if (this.evaluateNode(child, transaction)) {
          return true;
        }
      }
      return false;
    }

    return false;
  }

  buildTreeFromRules(treeDef: RuleTreeNode, rules: Rule[]): RuleTreeNode {
    if (treeDef.type === 'RULE') {
      return treeDef;
    }

    return {
      ...treeDef,
      children: treeDef.children?.map(child =>
        this.buildTreeFromRules(child, rules)
      ) || []
    };
  }

  invalidateCache(): void {
    this.refreshCache();
  }

  refresh(): void {
    this.refreshCache();
  }
}
