import { Channel } from './types';
import { getAllChannels, getActiveChannels } from './db';

export function calculateFee(channel: Channel, amount: number): number {
  const tiers = [...channel.feeTiers].sort((a, b) => a.maxAmount - b.maxAmount);
  for (const tier of tiers) {
    if (amount <= tier.maxAmount) {
      return Math.floor(amount * tier.rate);
    }
  }
  const lastTier = tiers[tiers.length - 1];
  return Math.floor(amount * lastTier.rate);
}

export function listAllChannels(): Channel[] {
  return getAllChannels();
}

export function listNonDegradedChannels(): Channel[] {
  return getActiveChannels();
}
