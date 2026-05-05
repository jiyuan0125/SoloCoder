import http from "node:http";
import { config } from "./config.js";
import { store } from "./storage.js";
import { handleRequest } from "./router.js";
import { asyncWrapper } from "./utils.js";

async function main(): Promise<void> {
  console.log("Starting tax invoice server...");

  await asyncWrapper(
    async () => store.loadFromFile(),
    (error) => {
      console.error("Failed to load data:", error);
      return undefined;
    }
  );

  store.startAutoSave();

  const server = http.createServer(async (req, res) => {
    await handleRequest(req, res);
  });

  process.on("SIGTERM", async () => {
    console.log("Received SIGTERM, shutting down...");
    store.stopAutoSave();
    try {
      await store.saveToFile();
      console.log("Data saved successfully");
    } catch (error) {
      console.error("Failed to save data during shutdown:", error);
    }
    process.exit(0);
  });

  process.on("SIGINT", async () => {
    console.log("Received SIGINT, shutting down...");
    store.stopAutoSave();
    try {
      await store.saveToFile();
      console.log("Data saved successfully");
    } catch (error) {
      console.error("Failed to save data during shutdown:", error);
    }
    process.exit(0);
  });

  server.listen(config.port, config.host, () => {
    console.log(`Tax invoice server running at http://${config.host}:${config.port}`);
    console.log(`Data directory: ${config.dataDir}`);
  });
}

main().catch((error) => {
  console.error("Server startup failed:", error);
  process.exit(1);
});
