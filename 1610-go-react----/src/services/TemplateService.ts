import { Template, ChannelType } from '../types';
import { v4 as uuidv4 } from 'uuid';
import { store } from '../storage/store';

const VARIABLE_REGEX = /\{\{(\w+)\}\}/g;

export class TemplateService {
  extractVariables(content: string): string[] {
    const matches = new Set<string>();
    let match;
    const regex = new RegExp(VARIABLE_REGEX);
    while ((match = regex.exec(content)) !== null) {
      matches.add(match[1]);
    }
    return Array.from(matches);
  }

  validateVariables(template: Template, variables: Record<string, any>): string[] {
    const missing: string[] = [];
    for (const v of template.variables) {
      if (variables[v] === undefined) {
        missing.push(v);
      }
    }
    return missing;
  }

  render(template: Template, variables: Record<string, any>): string {
    let content = template.content;
    for (const v of template.variables) {
      const value = variables[v] !== undefined ? String(variables[v]) : '';
      content = content.replace(new RegExp(`\\{\\{${v}\\}\\}`, 'g'), value);
    }
    return content;
  }

  create(name: string, type: ChannelType, content: string): Template {
    const now = Date.now();
    const template: Template = {
      id: uuidv4(),
      name,
      type,
      content,
      variables: this.extractVariables(content),
      createdAt: now,
      updatedAt: now,
    };
    store.addTemplate(template);
    return template;
  }

  list(): Template[] {
    return store.getTemplates();
  }

  get(id: string): Template | undefined {
    return store.getTemplate(id);
  }

  update(id: string, updates: Partial<Pick<Template, 'name' | 'content'>>): Template | undefined {
    const existing = store.getTemplate(id);
    if (!existing) return undefined;

    const updated: Template = {
      ...existing,
      name: updates.name ?? existing.name,
      content: updates.content ?? existing.content,
      variables: updates.content !== undefined
        ? this.extractVariables(updates.content)
        : existing.variables,
      updatedAt: Date.now(),
    };
    store.updateTemplate(id, updated);
    return updated;
  }

  delete(id: string): boolean {
    if (!store.getTemplate(id)) return false;
    store.deleteTemplate(id);
    return true;
  }
}

export const templateService = new TemplateService();
