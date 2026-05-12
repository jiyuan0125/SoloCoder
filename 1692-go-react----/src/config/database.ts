import { DataSource } from 'typeorm';
import 'reflect-metadata';

export const AppDataSource = new DataSource({
  type: 'sqlite',
  database: './database.sqlite',
  synchronize: true,
  logging: false,
  entities: [
    __dirname + '/../models/**/*.ts',
    __dirname + '/../models/**/*.js'
  ],
  migrations: [],
  subscribers: [],
});

export const initializeDatabase = async (): Promise<void> => {
  if (!AppDataSource.isInitialized) {
    await AppDataSource.initialize();
  }
};
