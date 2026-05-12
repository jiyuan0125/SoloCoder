const VARIABLE_PATTERN = /\{\{(\w+)\}\}/g;

export function validateTemplateVariables(text: string): boolean {
  if (!text) return true;
  
  const matches = text.match(/\{\{[^}]*\}\}/g) || [];
  
  for (const match of matches) {
    if (!VARIABLE_PATTERN.test(match)) {
      return false;
    }
  }
  
  return true;
}

export function extractVariables(text: string): string[] {
  if (!text) return [];
  
  const variables = new Set<string>();
  const matches = text.match(VARIABLE_PATTERN) || [];
  
  for (const match of matches) {
    const varName = match.replace(/\{\{|\}\}/g, "");
    variables.add(varName);
  }
  
  return Array.from(variables);
}

export function renderTemplate(
  template: string,
  variables: Record<string, string>
): string {
  if (!template) return "";
  
  return template.replace(VARIABLE_PATTERN, (_, varName) => {
    return variables[varName] !== undefined ? variables[varName] : `{{${varName}}}`;
  });
}
