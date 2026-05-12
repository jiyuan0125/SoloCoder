import { TransformationRule } from '../models/step';

export class TransformationEngine {
  public async applyTransformations(data: any[], rules: TransformationRule[]): Promise<any[]> {
    let result = data;
    
    for (const rule of rules) {
      this.validateRuleFields(rule, result);
      result = await this.applyTransformation(result, rule);
    }
    
    return result;
  }

  private validateRuleFields(rule: TransformationRule, data: any[]): void {
    if (data.length === 0) {
      return;
    }

    const sampleRecord = data[0];
    const availableFields = Object.keys(sampleRecord);

    switch (rule.type) {
      case 'mapping':
        if (rule.mapping) {
          for (const sourceField of Object.keys(rule.mapping)) {
            if (!availableFields.includes(sourceField)) {
              throw new Error(`Field '${sourceField}' does not exist in the data`);
            }
          }
        }
        break;
      
      case 'typeConversion':
        if (rule.typeConversion) {
          for (const field of Object.keys(rule.typeConversion)) {
            if (!availableFields.includes(field)) {
              throw new Error(`Field '${field}' does not exist in the data`);
            }
          }
        }
        break;
      
      case 'filter':
        if (rule.filter) {
          if (!availableFields.includes(rule.filter.field)) {
            throw new Error(`Field '${rule.filter.field}' does not exist in the data`);
          }
        }
        break;
      
      case 'aggregation':
        if (rule.aggregation) {
          for (const groupByField of rule.aggregation.groupBy) {
            if (!availableFields.includes(groupByField)) {
              throw new Error(`Field '${groupByField}' does not exist in the data`);
            }
          }
          for (const agg of rule.aggregation.aggregations) {
            if (!availableFields.includes(agg.field)) {
              throw new Error(`Field '${agg.field}' does not exist in the data`);
            }
          }
        }
        break;
    }
  }

  private async applyTransformation(data: any[], rule: TransformationRule): Promise<any[]> {
    switch (rule.type) {
      case 'mapping':
        return this.applyMapping(data, rule);
      
      case 'typeConversion':
        return this.applyTypeConversion(data, rule);
      
      case 'filter':
        return this.applyFilter(data, rule);
      
      case 'aggregation':
        return this.applyAggregation(data, rule);
      
      default:
        return data;
    }
  }

  private applyMapping(data: any[], rule: TransformationRule): any[] {
    if (!rule.mapping) return data;
    
    return data.map(record => {
      const newRecord: any = {};
      for (const [sourceField, targetField] of Object.entries(rule.mapping!)) {
        if (record.hasOwnProperty(sourceField)) {
          newRecord[targetField] = record[sourceField];
        }
      }
      return newRecord;
    });
  }

  private applyTypeConversion(data: any[], rule: TransformationRule): any[] {
    if (!rule.typeConversion) return data;
    
    return data.map(record => {
      const newRecord = { ...record };
      for (const [field, type] of Object.entries(rule.typeConversion!)) {
        if (newRecord.hasOwnProperty(field)) {
          newRecord[field] = this.convertType(newRecord[field], type);
        }
      }
      return newRecord;
    });
  }

  private convertType(value: any, type: string): any {
    switch (type) {
      case 'number':
        return Number(value);
      case 'string':
        return String(value);
      case 'boolean':
        return Boolean(value);
      case 'integer':
        return parseInt(value, 10);
      case 'float':
        return parseFloat(value);
      default:
        return value;
    }
  }

  private applyFilter(data: any[], rule: TransformationRule): any[] {
    if (!rule.filter) return data;
    
    const { field, operator, value } = rule.filter;
    
    return data.filter(record => {
      if (!record.hasOwnProperty(field)) return false;
      
      const fieldValue = record[field];
      
      switch (operator) {
        case 'equals':
          return fieldValue === value;
        case 'notEquals':
          return fieldValue !== value;
        case 'greaterThan':
          return fieldValue > value;
        case 'lessThan':
          return fieldValue < value;
        case 'contains':
          return String(fieldValue).includes(String(value));
        default:
          return true;
      }
    });
  }

  private applyAggregation(data: any[], rule: TransformationRule): any[] {
    if (!rule.aggregation) return data;
    
    const { groupBy, aggregations } = rule.aggregation;
    
    const groups = new Map<string, any[]>();
    
    for (const record of data) {
      const key = groupBy.map(field => record[field]).join('|');
      
      if (!groups.has(key)) {
        groups.set(key, []);
      }
      
      groups.get(key)!.push(record);
    }
    
    const results: any[] = [];
    
    for (const [, groupRecords] of groups) {
      const result: any = {};
      
      for (const field of groupBy) {
        result[field] = groupRecords[0][field];
      }
      
      for (const agg of aggregations) {
        const values = groupRecords.map(r => r[agg.field]);
        
        switch (agg.function) {
          case 'sum':
            result[agg.alias] = values.reduce((a, b) => a + Number(b), 0);
            break;
          case 'avg':
            result[agg.alias] = values.reduce((a, b) => a + Number(b), 0) / values.length;
            break;
          case 'count':
            result[agg.alias] = values.length;
            break;
          case 'min':
            result[agg.alias] = Math.min(...values.map(v => Number(v)));
            break;
          case 'max':
            result[agg.alias] = Math.max(...values.map(v => Number(v)));
            break;
        }
      }
      
      results.push(result);
    }
    
    return results;
  }
}
