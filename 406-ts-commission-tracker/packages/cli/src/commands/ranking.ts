import { apiClient } from '../client';
import { formatRankingList } from '../utils/formatter';

export async function handleGetRankings(args: string[]): Promise<void> {
  const dimension = args[0] || 'monthly';
  
  if (!['monthly', 'quarterly', 'yearly'].includes(dimension)) {
    console.log('用法: commission ranking <维度> [月份/季度] [年份]');
    console.log('维度: monthly | quarterly | yearly');
    console.log('示例:');
    console.log('  commission ranking monthly 5 2026');
    console.log('  commission ranking quarterly 2 2026');
    console.log('  commission ranking yearly 2026');
    return;
  }

  let path = `/api/rankings?dimension=${dimension}`;
  
  if (dimension === 'monthly' && args.length >= 2) {
    path += `&month=${args[1]}`;
  }
  if (dimension === 'quarterly' && args.length >= 2) {
    path += `&quarter=${args[1]}`;
  }
  if (args.length >= 3) {
    path += `&year=${args[2]}`;
  }

  const response = await apiClient.get(path);

  if (!response.success) {
    console.error('获取排名失败:', response.error?.message || '未知错误');
    return;
  }

  console.log(formatRankingList(response.data as any[]));
}
