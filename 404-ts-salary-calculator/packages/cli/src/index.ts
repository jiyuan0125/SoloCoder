#!/usr/bin/env node

import { parseArgs } from './args.js';
import { showHelp } from './help.js';
import { handleCalculate, handleSave, handleQuery, handleComparison } from './commands.js';

async function main(): Promise<void> {
  const args = process.argv.slice(2);
  
  if (args.length === 0 || args.includes('--help') || args.includes('-h')) {
    showHelp();
    return;
  }
  
  const command = args[0];
  const parsedArgs = parseArgs(args.slice(1));
  
  switch (command) {
    case 'calculate':
      await handleCalculate(parsedArgs);
      break;
    case 'save':
      await handleSave(parsedArgs);
      break;
    case 'query':
      await handleQuery(parsedArgs);
      break;
    case 'comparison':
      await handleComparison(parsedArgs);
      break;
    default:
      console.error(`未知命令: ${command}`);
      showHelp();
      process.exit(1);
  }
}

main().catch((err: unknown) => {
  console.error('执行失败:', err instanceof Error ? err.message : String(err));
  process.exit(1);
});
