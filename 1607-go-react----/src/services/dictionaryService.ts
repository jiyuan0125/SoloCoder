import { getDictionary, addWord, removeWord } from '../tokenizer';
import { DEFAULT_DICTIONARY } from '../data/dictionary';

export function listDictionary(): string[] {
  return getDictionary();
}

export function addDictionaryWord(word: string): { success: boolean; message: string } {
  if (!word || !word.trim()) {
    return {
      success: false,
      message: '词条不能为空'
    };
  }
  
  const trimmedWord = word.trim();
  const dict = getDictionary();
  
  if (dict.includes(trimmedWord)) {
    return {
      success: false,
      message: '词条已存在'
    };
  }
  
  addWord(trimmedWord);
  
  return {
    success: true,
    message: '添加成功'
  };
}

export function deleteDictionaryWord(word: string): { success: boolean; message: string } {
  if (!word || !word.trim()) {
    return {
      success: false,
      message: '词条不能为空'
    };
  }
  
  const trimmedWord = word.trim();
  const dict = getDictionary();
  
  if (!dict.includes(trimmedWord)) {
    return {
      success: false,
      message: '词条不存在'
    };
  }
  
  if (DEFAULT_DICTIONARY.includes(trimmedWord)) {
    return {
      success: false,
      message: '不能删除默认词典中的词条'
    };
  }
  
  removeWord(trimmedWord);
  
  return {
    success: true,
    message: '删除成功'
  };
}

export function resetDictionary(): { success: boolean; message: string } {
  const { setDictionary } = require('../tokenizer');
  setDictionary([...DEFAULT_DICTIONARY]);
  
  return {
    success: true,
    message: '词典已重置为默认配置'
  };
}
