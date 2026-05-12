import { ConditionParseError, Transaction } from '../types';

export interface ParsedCondition {
  evaluate(transaction: Transaction): boolean;
}

const COMPARISON_OPERATORS = ['==', '!=', '>', '>=', '<', '<='];
const LOGICAL_OPERATORS = ['AND', 'OR'];

function createParseError(message: string): ConditionParseError {
  const error = new Error(message) as ConditionParseError;
  error.code = 'INVALID_CONDITION';
  return error;
}

function parseValue(token: string): string | number | boolean | null {
  const trimmed = token.trim();

  if (trimmed.startsWith('"') && trimmed.endsWith('"')) {
    return trimmed.slice(1, -1);
  }
  if (trimmed.startsWith("'") && trimmed.endsWith("'")) {
    return trimmed.slice(1, -1);
  }
  if (trimmed === 'true') return true;
  if (trimmed === 'false') return false;
  if (trimmed === 'null') return null;

  const num = Number(trimmed);
  if (!Number.isNaN(num)) return num;

  return trimmed;
}

function isFieldReference(token: string): boolean {
  const trimmed = token.trim();
  return /^[a-zA-Z_][a-zA-Z0-9_]*$/.test(trimmed);
}

function getFieldValue(transaction: Transaction, field: string): any {
  return transaction[field];
}

class ComparisonCondition implements ParsedCondition {
  private field: string;
  private operator: string;
  private value: any;

  constructor(field: string, operator: string, value: any) {
    this.field = field;
    this.operator = operator;
    this.value = value;
  }

  evaluate(transaction: Transaction): boolean {
    const fieldValue = getFieldValue(transaction, this.field);
    const value = this.value;

    switch (this.operator) {
      case '==':
        return fieldValue === value;
      case '!=':
        return fieldValue !== value;
      case '>':
        return (fieldValue as number) > value;
      case '>=':
        return (fieldValue as number) >= value;
      case '<':
        return (fieldValue as number) < value;
      case '<=':
        return (fieldValue as number) <= value;
      default:
        return false;
    }
  }
}

class InCondition implements ParsedCondition {
  private field: string;
  private values: any[];
  private negated: boolean;

  constructor(field: string, values: any[], negated: boolean = false) {
    this.field = field;
    this.values = values;
    this.negated = negated;
  }

  evaluate(transaction: Transaction): boolean {
    const fieldValue = getFieldValue(transaction, this.field);
    const result = this.values.some(v => v === fieldValue);
    return this.negated ? !result : result;
  }
}

class BetweenCondition implements ParsedCondition {
  private field: string;
  private lower: number;
  private upper: number;

  constructor(field: string, lower: number, upper: number) {
    this.field = field;
    this.lower = lower;
    this.upper = upper;
  }

  evaluate(transaction: Transaction): boolean {
    const fieldValue = getFieldValue(transaction, this.field) as number;
    return fieldValue >= this.lower && fieldValue <= this.upper;
  }
}

class LogicalCondition implements ParsedCondition {
  private operator: 'AND' | 'OR';
  private left: ParsedCondition;
  private right: ParsedCondition;

  constructor(operator: 'AND' | 'OR', left: ParsedCondition, right: ParsedCondition) {
    this.operator = operator;
    this.left = left;
    this.right = right;
  }

  evaluate(transaction: Transaction): boolean {
    if (this.operator === 'AND') {
      return this.left.evaluate(transaction) && this.right.evaluate(transaction);
    }
    return this.left.evaluate(transaction) || this.right.evaluate(transaction);
  }
}

class NotCondition implements ParsedCondition {
  private child: ParsedCondition;

  constructor(child: ParsedCondition) {
    this.child = child;
  }

  evaluate(transaction: Transaction): boolean {
    return !this.child.evaluate(transaction);
  }
}

function tokenize(expression: string): string[] {
  const tokens: string[] = [];
  let current = '';
  let inString = false;
  let stringChar = '';
  let inBrackets = false;
  let bracketDepth = 0;

  for (let i = 0; i < expression.length; i++) {
    const char = expression[i];

    if ((char === '"' || char === "'") && expression[i - 1] !== '\\') {
      if (!inString) {
        if (current.trim()) tokens.push(current.trim());
        current = char;
        inString = true;
        stringChar = char;
      } else if (char === stringChar) {
        current += char;
        tokens.push(current);
        current = '';
        inString = false;
        stringChar = '';
      } else {
        current += char;
      }
      continue;
    }

    if (inString) {
      current += char;
      continue;
    }

    if (char === '[') {
      if (current.trim()) tokens.push(current.trim());
      current = char;
      inBrackets = true;
      bracketDepth = 1;
      continue;
    }

    if (inBrackets) {
      if (char === '[') bracketDepth++;
      if (char === ']') {
        bracketDepth--;
        current += char;
        if (bracketDepth === 0) {
          tokens.push(current);
          current = '';
          inBrackets = false;
        }
        continue;
      }
      current += char;
      continue;
    }

    if (char === '(' || char === ')') {
      if (current.trim()) tokens.push(current.trim());
      tokens.push(char);
      current = '';
      continue;
    }

    if (char === ' ' || char === '\t' || char === '\n') {
      if (current.trim()) {
        tokens.push(current.trim());
        current = '';
      }
      continue;
    }

    current += char;
  }

  if (current.trim()) {
    tokens.push(current.trim());
  }

  return tokens;
}

function parseArrayLiteral(token: string): any[] {
  if (!token.startsWith('[') || !token.endsWith(']')) {
    throw createParseError(`Invalid array literal: ${token}`);
  }

  const content = token.slice(1, -1).trim();
  if (!content) return [];

  const elements: string[] = [];
  let current = '';
  let depth = 0;
  let inString = false;
  let stringChar = '';

  for (let i = 0; i < content.length; i++) {
    const char = content[i];

    if ((char === '"' || char === "'") && content[i - 1] !== '\\') {
      if (!inString) {
        inString = true;
        stringChar = char;
      } else if (char === stringChar) {
        inString = false;
        stringChar = '';
      }
      current += char;
      continue;
    }

    if (inString) {
      current += char;
      continue;
    }

    if (char === '[') depth++;
    if (char === ']') depth--;

    if (char === ',' && depth === 0) {
      elements.push(current.trim());
      current = '';
      continue;
    }

    current += char;
  }

  if (current.trim()) elements.push(current.trim());

  return elements.map(el => parseValue(el));
}

class Parser {
  private tokens: string[];
  private pos: number;

  constructor(expression: string) {
    this.tokens = tokenize(expression);
    this.pos = 0;
  }

  private current(): string | undefined {
    return this.tokens[this.pos];
  }

  private peek(offset: number = 1): string | undefined {
    return this.tokens[this.pos + offset];
  }

  private consume(): string {
    return this.tokens[this.pos++];
  }

  private expect(token: string): void {
    if (this.current() !== token) {
      throw createParseError(`Expected '${token}', got '${this.current()}'`);
    }
    this.consume();
  }

  parse(): ParsedCondition {
    const result = this.parseOr();
    if (this.pos < this.tokens.length) {
      throw createParseError(`Unexpected token: ${this.current()}`);
    }
    return result;
  }

  private parseOr(): ParsedCondition {
    let left = this.parseAnd();

    while (this.current() === 'OR') {
      this.consume();
      const right = this.parseAnd();
      left = new LogicalCondition('OR', left, right);
    }

    return left;
  }

  private parseAnd(): ParsedCondition {
    let left = this.parseNot();

    while (this.current() === 'AND') {
      this.consume();
      const right = this.parseNot();
      left = new LogicalCondition('AND', left, right);
    }

    return left;
  }

  private parseNot(): ParsedCondition {
    if (this.current() === 'NOT') {
      this.consume();
      const child = this.parsePrimary();
      return new NotCondition(child);
    }
    return this.parsePrimary();
  }

  private parsePrimary(): ParsedCondition {
    if (this.current() === '(') {
      this.consume();
      const expr = this.parseOr();
      this.expect(')');
      return expr;
    }

    const next = this.peek();

    if (next === 'IN' || (next && next.toUpperCase() === 'IN')) {
      return this.parseIn();
    }

    if (next === 'BETWEEN' || (next && next.toUpperCase() === 'BETWEEN')) {
      return this.parseBetween();
    }

    return this.parseComparison();
  }

  private parseIn(): ParsedCondition {
    const field = this.consume();
    const operator = this.consume().toUpperCase();

    if (operator !== 'IN') {
      throw createParseError(`Expected 'IN', got '${operator}'`);
    }

    if (!isFieldReference(field)) {
      throw createParseError(`Invalid field reference: ${field}`);
    }

    const negated = false;
    const arrayToken = this.consume();
    const values = parseArrayLiteral(arrayToken);

    return new InCondition(field, values, negated);
  }

  private parseBetween(): ParsedCondition {
    const field = this.consume();
    const operator = this.consume().toUpperCase();

    if (operator !== 'BETWEEN') {
      throw createParseError(`Expected 'BETWEEN', got '${operator}'`);
    }

    if (!isFieldReference(field)) {
      throw createParseError(`Invalid field reference: ${field}`);
    }

    const lowerToken = this.consume();
    const lower = parseValue(lowerToken);
    if (typeof lower !== 'number') {
      throw createParseError(`BETWEEN lower bound must be a number, got ${lowerToken}`);
    }

    if (this.current()?.toUpperCase() !== 'AND') {
      throw createParseError(`Expected 'AND' in BETWEEN clause, got '${this.current()}'`);
    }
    this.consume();

    const upperToken = this.consume();
    const upper = parseValue(upperToken);
    if (typeof upper !== 'number') {
      throw createParseError(`BETWEEN upper bound must be a number, got ${upperToken}`);
    }

    return new BetweenCondition(field, lower, upper);
  }

  private parseComparison(): ParsedCondition {
    const field = this.consume();

    if (!isFieldReference(field)) {
      throw createParseError(`Invalid field reference: ${field}`);
    }

    const operator = this.current();
    if (!operator) {
      throw createParseError(`Missing operator after field: ${field}`);
    }

    if (!COMPARISON_OPERATORS.includes(operator)) {
      throw createParseError(`Invalid operator: ${operator}`);
    }

    this.consume();
    const valueToken = this.consume();
    const value = parseValue(valueToken);

    return new ComparisonCondition(field, operator, value);
  }
}

export function parseCondition(expression: string): ParsedCondition {
  if (!expression || !expression.trim()) {
    throw createParseError('Condition expression cannot be empty');
  }

  const parser = new Parser(expression);
  return parser.parse();
}

export function validateCondition(expression: string): boolean {
  try {
    parseCondition(expression);
    return true;
  } catch {
    return false;
  }
}
