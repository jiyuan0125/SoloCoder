import { runAsync, getAsync, allAsync } from '../db';
import { Rating } from '../types';
import { updateProductStats } from './productService';

function isValidScore(score: number): boolean {
  return Number.isInteger(score) && score >= 1 && score <= 5;
}

async function getRatingsByUser(userId: number): Promise<Rating[]> {
  const rows = await allAsync(`
    SELECT user_id as userId, product_id as productId, score, updated_at as updatedAt
    FROM ratings WHERE user_id = ?
    ORDER BY updated_at DESC
  `, userId);
  return rows.map(row => ({
    userId: row.userId,
    productId: row.productId,
    score: row.score,
    updatedAt: row.updatedAt
  }));
}

async function getRating(userId: number, productId: number): Promise<Rating | null> {
  const row = await getAsync(`
    SELECT user_id as userId, product_id as productId, score, updated_at as updatedAt
    FROM ratings WHERE user_id = ? AND product_id = ?
  `, userId, productId);
  if (!row) return null;
  return {
    userId: row.userId,
    productId: row.productId,
    score: row.score,
    updatedAt: row.updatedAt
  };
}

async function upsertRating(userId: number, productId: number, score: number): Promise<Rating | null> {
  if (!isValidScore(score)) {
    return null;
  }
  await runAsync(`
    INSERT INTO ratings (user_id, product_id, score, updated_at)
    VALUES (?, ?, ?, datetime('now'))
    ON CONFLICT(user_id, product_id)
    DO UPDATE SET score = excluded.score, updated_at = datetime('now')
  `, userId, productId, score);
  await updateProductStats(productId);
  return getRating(userId, productId);
}

async function updateRating(userId: number, productId: number, score: number): Promise<Rating | null> {
  if (!isValidScore(score)) {
    return null;
  }
  const existing = await getRating(userId, productId);
  if (!existing) return null;
  await runAsync(
    `UPDATE ratings SET score = ?, updated_at = datetime('now') WHERE user_id = ? AND product_id = ?`,
    score, userId, productId
  );
  await updateProductStats(productId);
  return getRating(userId, productId);
}

async function deleteRating(userId: number, productId: number): Promise<boolean> {
  const result = await runAsync(
    `DELETE FROM ratings WHERE user_id = ? AND product_id = ?`,
    userId, productId
  );
  if (result.changes > 0) {
    await updateProductStats(productId);
    return true;
  }
  return false;
}

async function getUserRatingCount(userId: number): Promise<number> {
  const row = await getAsync(
    `SELECT COUNT(*) as count FROM ratings WHERE user_id = ?`,
    userId
  );
  return row.count;
}

async function userExists(userId: number): Promise<boolean> {
  const row = await getAsync(
    `SELECT 1 FROM ratings WHERE user_id = ? LIMIT 1`,
    userId
  );
  return !!row;
}

async function getUsersWithRatings(): Promise<number[]> {
  const rows = await allAsync(`SELECT DISTINCT user_id as id FROM ratings`);
  return rows.map(r => r.id);
}

async function getProductsWithRatings(): Promise<number[]> {
  const rows = await allAsync(`SELECT DISTINCT product_id as id FROM ratings WHERE total_ratings >= 5`);
  return rows.map(r => r.id);
}

export {
  isValidScore,
  getRatingsByUser,
  getRating,
  upsertRating,
  updateRating,
  deleteRating,
  getUserRatingCount,
  userExists,
  getUsersWithRatings,
  getProductsWithRatings
};
