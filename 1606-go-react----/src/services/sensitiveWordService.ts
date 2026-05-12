import { SensitiveWord, SensitiveWordLevel } from '../types';
import { sensitiveWordStore } from '../store/sensitiveWordStore';
import { ruleStore } from '../store/ruleStore';

export class SensitiveWordService {
  create(word: string, level: SensitiveWordLevel): SensitiveWord {
    return sensitiveWordStore.create(word, level);
  }

  findById(id: string): SensitiveWord | undefined {
    return sensitiveWordStore.findById(id);
  }

  findAll(): SensitiveWord[] {
    return sensitiveWordStore.findAll();
  }

  update(
    id: string,
    word?: string,
    level?: SensitiveWordLevel
  ): SensitiveWord | undefined {
    return sensitiveWordStore.update(id, word, level);
  }

  delete(id: string): { success: boolean; error?: string } {
    const word = sensitiveWordStore.findById(id);
    if (!word) {
      return { success: false, error: '敏感词不存在' };
    }

    const referencingRules = ruleStore.findBySensitiveWordId(id);
    if (referencingRules.length > 0) {
      return { 
        success: false, 
        error: `敏感词被 ${referencingRules.length} 条审核规则引用，无法删除` 
      };
    }

    const deleted = sensitiveWordStore.delete(id);
    return { success: deleted };
  }

  matchText(text: string): { word: SensitiveWord; matched: string }[] {
    return sensitiveWordStore.matchText(text);
  }
}

export const sensitiveWordService = new SensitiveWordService();
