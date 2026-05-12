import { runAsync, allAsync } from '../db';
import { getUserRatingCount, userExists } from './ratingService';
import { getTopRatedProducts, getProductById } from './productService';
import { combinedRecommend } from './collaborativeFiltering';
import { Product } from '../types';

async function recordExposure(userId: number, productIds: number[]): Promise<void> {
  const stmt = `INSERT INTO recommend_history (user_id, product_id) VALUES (?, ?)`;
  for (const pid of productIds) {
    await runAsync(stmt, userId, pid);
  }
}

async function recordInteraction(userId: number, productId: number): Promise<boolean> {
  const result = await runAsync(`
    UPDATE recommend_history
    SET interacted_at = datetime('now')
    WHERE user_id = ? AND product_id = ? AND interacted_at IS NULL
  `, userId, productId);
  return result.changes > 0;
}

async function getRecommendations(userId: number, limit: number = 20): Promise<{
  fromColdStart: boolean;
  products: Product[];
}> {
  const hasRatings = await userExists(userId);
  
  const ratingCount = await getUserRatingCount(userId);
  
  let productIds: number[] = [];
  let fromColdStart = false;

  if (ratingCount < 3) {
    fromColdStart = true;
    const coldStartProducts = await getTopRatedProducts(limit);
    await recordExposure(userId, coldStartProducts.map(p => p.id));
    return { fromColdStart: true, products: coldStartProducts };
  }

  productIds = await combinedRecommend(userId, limit);
  
  const products: Product[] = [];
  for (const pid of productIds) {
    const p = await getProductById(pid);
    if (p) products.push(p);
  }

  await recordExposure(userId, products.map(p => p.id));
  return { fromColdStart: false, products };
}

export {
  getRecommendations,
  recordExposure,
  recordInteraction
};
