export interface ParsedArgs {
  server: string;
  [key: string]: string | number | boolean;
}

export function parseArgs(args: string[]): ParsedArgs {
  const result: ParsedArgs = {
    server: 'http://127.0.0.1:3000',
  };
  
  for (let i = 0; i < args.length; i++) {
    const arg = args[i];
    
    if (arg.startsWith('--server=')) {
      result.server = arg.slice('--server='.length);
    } else if (arg === '--server' && i + 1 < args.length) {
      result.server = args[i + 1];
      i++;
    } else if (arg.startsWith('--')) {
      const eqIndex = arg.indexOf('=');
      if (eqIndex !== -1) {
        const key = arg.slice(2, eqIndex);
        const value = arg.slice(eqIndex + 1);
        result[key] = parseValue(value);
      } else if (i + 1 < args.length && !args[i + 1].startsWith('--')) {
        const key = arg.slice(2);
        const value = args[i + 1];
        result[key] = parseValue(value);
        i++;
      } else {
        const key = arg.slice(2);
        result[key] = true;
      }
    }
  }
  
  return result;
}

function parseValue(value: string): string | number {
  const numValue = Number(value);
  if (!isNaN(numValue) && value.trim() !== '') {
    return numValue;
  }
  return value;
}
