import { Channel, RoutingStrategy } from './types';
import { calculateFee, listNonDegradedChannels } from './channelService';
import { getChannelById } from './db';

export interface RoutingResult {
  channel: Channel;
}

export class ChannelNotFoundError extends Error {}
export class AllChannelsDegradedError extends Error {
  constructor() {
    super('所有渠道不可用');
  }
}

function getSuccessRateEstimate(channel: Channel): number {
  if (channel.isDegraded) return 0;
  if (channel.degradedAt) return 0.9;
  return 0.95 + Math.random() * 0.05;
}

export function routePayment(
  amount: number,
  strategy: RoutingStrategy,
  specifiedChannelId?: string
): RoutingResult {
  if (strategy === 'SPECIFIED') {
    if (!specifiedChannelId) {
      throw new ChannelNotFoundError('未指定渠道');
    }
    const ch = getChannelById(specifiedChannelId);
    if (!ch) {
      throw new ChannelNotFoundError('指定渠道不存在');
    }
    if (ch.isDegraded) {
      const alternatives = listNonDegradedChannels();
      if (alternatives.length === 0) {
        throw new AllChannelsDegradedError();
      }
      return { channel: alternatives[0] };
    }
    return { channel: ch };
  }

  const channels = listNonDegradedChannels();
  if (channels.length === 0) {
    throw new AllChannelsDegradedError();
  }

  if (strategy === 'RATE') {
    channels.sort((a, b) => calculateFee(a, amount) - calculateFee(b, amount));
    return { channel: channels[0] };
  }

  if (strategy === 'SUCCESS_RATE') {
    channels.sort((a, b) => getSuccessRateEstimate(b) - getSuccessRateEstimate(a));
    return { channel: channels[0] };
  }

  throw new Error('未知路由策略');
}
