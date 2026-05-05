import * as http from "http";
import { handleRequest } from "./router";
import { startScheduler, stopScheduler } from "./services/scheduler-service";
import { SERVER_PORT } from "./config";

const server = http.createServer(async (req, res) => {
  await handleRequest(req, res);
});

server.listen(SERVER_PORT, () => {
  console.log(`Payment Gateway server is running on http://localhost:${SERVER_PORT}`);
  startScheduler();
});

function gracefulShutdown(): void {
  console.log("Shutting down server...");
  stopScheduler();
  server.close(() => {
    console.log("Server closed");
    process.exit(0);
  });
}

process.on("SIGTERM", gracefulShutdown);
process.on("SIGINT", gracefulShutdown);
