import { Response } from 'express';
import { ApiError } from './types';

export const sendError = (res: Response, error: ApiError) => {
  res.status(error.code).json({ error: error.message });
};

export const addDays = (date: Date, days: number): Date => {
  const result = new Date(date);
  result.setDate(result.getDate() + days);
  return result;
};

export const nowISO = (): string => new Date().toISOString();

export const parseISO = (isoString: string): Date => new Date(isoString);

export const isAfter = (date1: Date, date2: Date): boolean => date1 > date2;

export const isBefore = (date1: Date, date2: Date): boolean => date1 < date2;
