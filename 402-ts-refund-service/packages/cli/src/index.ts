#!/usr/bin/env node

import { parseArgs } from './args.js';
import { runCommand } from './commands.js';

async function main(): Promise<void> {
  try {
    const args = process.argv.slice(2);
    const command = parseArgs(args);
    await runCommand(command);
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Unknown error';
    console.error(`Error: ${message}`);
    process.exit(1);
  }
}

main();
