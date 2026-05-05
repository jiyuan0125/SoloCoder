#!/usr/bin/env node

import { parseCommand } from './parser';
import { executeCommand } from './executor';

async function main(): Promise<void> {
  const args = process.argv.slice(2);
  const command = parseCommand(args);
  
  if (!command) {
    console.log(`
Budget Planner CLI

Usage: budget-cli <command> [options]

Commands:
  department list                    List all departments
  department create <id> <name> <managerId>  Create a new department
  
  budget get <departmentId> <year>  Get budget for a department and year
  budget create <departmentId> <year> <totalAmount> --allocations <json>  Create budget
  budget adjust <departmentId> <year> --adjustments <json>  Adjust budget
  budget draft <newYear> <sourceYear>  Create draft budgets for new year
  budget confirm <departmentId> <year> <confirmedBy>  Confirm a draft budget
  
  reimbursement submit <departmentId> <year> <month> <description> <requestedBy> --items <json>  Submit reimbursement
  
  carryover request <departmentId> <year> <fromQuarter> <toQuarter> <amount> <requestedBy>  Request carryover
  
  stats usage [options]     Get budget usage statistics
  stats trend [options]     Get budget usage trend data

Options:
  --help                     Show this help message
  --server <url>              Server URL (default: http://localhost:3000)
`);
    process.exit(0);
  }

  try {
    const result = await executeCommand(command);
    console.log(JSON.stringify(result, null, 2));
  } catch (error) {
    console.error('Error:', error instanceof Error ? error.message : String(error));
    process.exit(1);
  }
}

main().catch((error) => {
  console.error('Fatal error:', error);
  process.exit(1);
});
