#!/usr/bin/env node

import { runCommand } from "./commands.js";

const args = process.argv.slice(2);

runCommand(args).catch((error) => {
  console.error("CLI 执行错误:", error);
  process.exit(1);
});
