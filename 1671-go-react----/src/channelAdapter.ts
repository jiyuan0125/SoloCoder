import { v4 as uuidv4 } from 'uuid';

export interface ChannelResult {
  success: boolean;
  transactionNo: string;
  responseTimeMs: number;
}

export interface ChannelRefundResult {
  success: boolean;
  responseTimeMs: number;
}

export async function chargeViaChannel(channelId: string, amount: number): Promise<ChannelResult> {
  const start = Date.now();
  await new Promise(resolve => setTimeout(resolve, 50 + Math.random() * 100));
  const success = Math.random() > 0.02;
  return {
    success,
    transactionNo: uuidv4(),
    responseTimeMs: Date.now() - start,
  };
}

export async function refundViaChannel(
  channelId: string,
  channelTransactionNo: string,
  amount: number
): Promise<ChannelRefundResult> {
  const start = Date.now();
  await new Promise(resolve => setTimeout(resolve, 30 + Math.random() * 80));
  return {
    success: true,
    responseTimeMs: Date.now() - start,
  };
}
