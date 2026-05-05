import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

export const config = {
  port: parseInt(process.env.PORT || "3000", 10),
  host: process.env.HOST || "127.0.0.1",
  dataDir: process.env.DATA_DIR || path.resolve(__dirname, "../../../data"),
  dataFile: "invoices.json",
  saveInterval: 5000,
  maxBatchSize: 100,
};

export function getDataFilePath(): string {
  return path.join(config.dataDir, config.dataFile);
}
