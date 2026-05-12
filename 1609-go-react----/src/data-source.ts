import "reflect-metadata";
import { DataSource } from "typeorm";

export const AppDataSource = new DataSource({
  type: "better-sqlite3",
  database: "./message-center.db",
  synchronize: true,
  logging: false,
  entities: [__dirname + "/entities/*.{ts,js}"],
  migrations: [],
  subscribers: [],
});
