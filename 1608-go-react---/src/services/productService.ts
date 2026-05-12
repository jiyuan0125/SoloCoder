import { runAsync, getAsync, allAsync } from '../db';
import { Product } from '../types';

async function getProducts(): Promise<Product[]> {
  const rows = await allAsync(`
    SELECT id, name, category, avg_score as avgScore, total_ratings as totalRatings, created_at as createdAt
    FROM products
    ORDER BY id
  `);
  return rows.map(row => ({
    id: row.id,
    name: row.name,
    category: row.category,
    avgScore: row.avgScore,
    totalRatings: row.totalRatings,
    createdAt: row.createdAt
  }));
}

async function getProductById(id: number): Promise<Product | null> {
  const row = await getAsync(`
    SELECT id, name, category, avg_score as avgScore, total_ratings as totalRatings, created_at as createdAt
    FROM products WHERE id = ?
  `, id);
  if (!row) return null;
  return {
    id: row.id,
    name: row.name,
    category: row.category,
    avgScore: row.avgScore,
    totalRatings: row.totalRatings,
    createdAt: row.createdAt
  };
}

async function createProduct(name: string, category: string): Promise<Product> {
  const result = await runAsync(
    `INSERT INTO products (name, category) VALUES (?, ?)`,
    name, category
  );
  return getProductById(result.lastID) as Promise<Product>;
}

async function updateProduct(id: number, name: string, category: string): Promise<Product | null> {
  const existing = await getProductById(id);
  if (!existing) return null;
  await runAsync(
    `UPDATE products SET name = ?, category = ? WHERE id = ?`,
    name, category, id
  );
  return getProductById(id);
}

async function deleteProduct(id: number): Promise<{ success: boolean; hasRatings: boolean }> {
  const ratingCount = await getAsync(
    `SELECT COUNT(*) as count FROM ratings WHERE product_id = ?`,
    id
  );
  if (ratingCount.count > 0) {
    return { success: false, hasRatings: true };
  }
  const result = await runAsync(`DELETE FROM products WHERE id = ?`, id);
  return { success: result.changes > 0, hasRatings: false };
}

async function updateProductStats(productId: number): Promise<void> {
  const stats = await getAsync(`
    SELECT COUNT(*) as total, AVG(score) as avg
    FROM ratings WHERE product_id = ?
  `, productId);
  await runAsync(
    `UPDATE products SET total_ratings = ?, avg_score = ? WHERE id = ?`,
    stats.total, stats.avg || 0, productId
  );
}

async function getTopRatedProducts(limit: number = 20): Promise<Product[]> {
  const rows = await allAsync(`
    SELECT id, name, category, avg_score as avgScore, total_ratings as totalRatings, created_at as createdAt
    FROM products
    WHERE total_ratings >= 5
    ORDER BY avg_score DESC, total_ratings DESC
    LIMIT ?
  `, limit);
  return rows.map(row => ({
    id: row.id,
    name: row.name,
    category: row.category,
    avgScore: row.avgScore,
    totalRatings: row.totalRatings,
    createdAt: row.createdAt
  }));
}

async function getEligibleProducts(): Promise<number[]> {
  const rows = await allAsync(`
    SELECT id FROM products WHERE total_ratings >= 5
  `);
  return rows.map(r => r.id);
}

export {
  getProducts,
  getProductById,
  createProduct,
  updateProduct,
  deleteProduct,
  updateProductStats,
  getTopRatedProducts,
  getEligibleProducts
};
