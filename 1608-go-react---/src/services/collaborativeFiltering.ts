import { allAsync } from '../db';
import { getEligibleProducts } from './productService';

type RatingMap = Map<number, number>;

function cosineSimilarity(a: RatingMap, b: RatingMap): number {
  let dot = 0;
  let normA = 0;
  let normB = 0;
  let commonCount = 0;

  for (const [key, val] of a) {
    normA += val * val;
    if (b.has(key)) {
      dot += val * (b.get(key) as number);
      commonCount++;
    }
  }

  for (const val of b.values()) {
    normB += val * val;
  }

  if (commonCount < 2) return 0;
  const denom = Math.sqrt(normA) * Math.sqrt(normB);
  return denom === 0 ? 0 : dot / denom;
}

async function buildUserItemMatrix(): Promise<Map<number, RatingMap>> {
  const rows = await allAsync(`
    SELECT user_id as userId, product_id as productId, score
    FROM ratings
  `);
  const matrix = new Map<number, RatingMap>();
  for (const row of rows) {
    if (!matrix.has(row.userId)) {
      matrix.set(row.userId, new Map());
    }
    matrix.get(row.userId)?.set(row.productId, row.score);
  }
  return matrix;
}

async function buildItemUserMatrix(): Promise<Map<number, RatingMap>> {
  const rows = await allAsync(`
    SELECT user_id as userId, product_id as productId, score
    FROM ratings
  `);
  const matrix = new Map<number, RatingMap>();
  for (const row of rows) {
    if (!matrix.has(row.productId)) {
      matrix.set(row.productId, new Map());
    }
    matrix.get(row.productId)?.set(row.userId, row.score);
  }
  return matrix;
}

async function userBasedRecommend(userId: number, limit: number = 20): Promise<number[]> {
  const userMatrix = await buildUserItemMatrix();
  const targetRatings = userMatrix.get(userId);
  if (!targetRatings) return [];

  const eligibleProducts = await getEligibleProducts();
  const eligibleSet = new Set(eligibleProducts);

  const ratedProducts = new Set(targetRatings.keys());
  const candidates = new Set<number>();
  const similarities: { userId: number; sim: number }[] = [];

  for (const [otherUserId, otherRatings] of userMatrix) {
    if (otherUserId === userId) continue;
    const sim = cosineSimilarity(targetRatings, otherRatings);
    if (sim > 0) {
      similarities.push({ userId: otherUserId, sim });
      for (const pid of otherRatings.keys()) {
        if (eligibleSet.has(pid) && !ratedProducts.has(pid)) {
          candidates.add(pid);
        }
      }
    }
  }

  similarities.sort((a, b) => b.sim - a.sim);
  const topSimilarUsers = similarities.slice(0, 50);
  const simMap = new Map(topSimilarUsers.map(u => [u.userId, u.sim]));

  const scores: Map<number, number> = new Map();
  const simSum: Map<number, number> = new Map();

  for (const pid of candidates) {
    for (const otherUserId of simMap.keys()) {
      const otherRatings = userMatrix.get(otherUserId);
      if (otherRatings && otherRatings.has(pid)) {
        const sim = simMap.get(otherUserId) as number;
        const score = otherRatings.get(pid) as number;
        scores.set(pid, (scores.get(pid) || 0) + sim * score);
        simSum.set(pid, (simSum.get(pid) || 0) + sim);
      }
    }
  }

  const result: { productId: number; predicted: number }[] = [];
  for (const [pid, total] of scores) {
    const sum = simSum.get(pid) || 1;
    result.push({ productId: pid, predicted: total / sum });
  }

  result.sort((a, b) => b.predicted - a.predicted);
  return result.slice(0, limit).map(r => r.productId);
}

async function itemBasedRecommend(userId: number, limit: number = 20): Promise<number[]> {
  const userMatrix = await buildUserItemMatrix();
  const itemMatrix = await buildItemUserMatrix();
  const targetRatings = userMatrix.get(userId);
  if (!targetRatings) return [];

  const eligibleProducts = await getEligibleProducts();
  const eligibleSet = new Set(eligibleProducts);

  const ratedProducts = new Set(targetRatings.keys());
  const candidates = new Set<number>();

  const similarities: Map<number, Map<number, number>> = new Map();

  for (const ratedPid of ratedProducts) {
    const ratedVector = itemMatrix.get(ratedPid);
    if (!ratedVector) continue;

    for (const [otherPid, otherVector] of itemMatrix) {
      if (otherPid === ratedPid) continue;
      if (!eligibleSet.has(otherPid)) continue;
      if (ratedProducts.has(otherPid)) continue;

      const sim = cosineSimilarity(ratedVector, otherVector);
      if (sim > 0) {
        candidates.add(otherPid);
        if (!similarities.has(ratedPid)) {
          similarities.set(ratedPid, new Map());
        }
        similarities.get(ratedPid)?.set(otherPid, sim);
      }
    }
  }

  const scores: Map<number, number> = new Map();
  const weightSum: Map<number, number> = new Map();

  for (const [ratedPid, userScore] of targetRatings) {
    const itemSims = similarities.get(ratedPid);
    if (!itemSims) continue;
    for (const [candidatePid, sim] of itemSims) {
      scores.set(candidatePid, (scores.get(candidatePid) || 0) + sim * userScore);
      weightSum.set(candidatePid, (weightSum.get(candidatePid) || 0) + sim);
    }
  }

  const result: { productId: number; predicted: number }[] = [];
  for (const [pid, total] of scores) {
    const sum = weightSum.get(pid) || 1;
    result.push({ productId: pid, predicted: total / sum });
  }

  result.sort((a, b) => b.predicted - a.predicted);
  return result.slice(0, limit).map(r => r.productId);
}

async function combinedRecommend(userId: number, limit: number = 20): Promise<number[]> {
  const [userRecs, itemRecs] = await Promise.all([
    userBasedRecommend(userId, limit),
    itemBasedRecommend(userId, limit)
  ]);

  const combined = new Set<number>();
  for (const pid of userRecs) combined.add(pid);
  for (const pid of itemRecs) combined.add(pid);

  const final: number[] = [];
  const seen = new Set<number>();
  for (const pid of userRecs) {
    if (!seen.has(pid)) {
      final.push(pid);
      seen.add(pid);
    }
  }
  for (const pid of itemRecs) {
    if (!seen.has(pid)) {
      final.push(pid);
      seen.add(pid);
    }
  }

  return final.slice(0, limit);
}

export {
  cosineSimilarity,
  buildUserItemMatrix,
  buildItemUserMatrix,
  userBasedRecommend,
  itemBasedRecommend,
  combinedRecommend
};
