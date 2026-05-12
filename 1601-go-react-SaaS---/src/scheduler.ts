import cron from 'node-cron';
import { billingService } from './services/billingService';

export function startScheduler(): void {
  cron.schedule('0 0 1 * *', () => {
    console.log('[Scheduler] 开始生成月度账单...');
    const bills = billingService.generateMonthlyBills();
    console.log(`[Scheduler] 生成了 ${bills.length} 条账单`);
  });

  cron.schedule('0 0 * * *', () => {
    console.log('[Scheduler] 检查逾期账单...');
    billingService.processOverdueBills();
  });

  console.log('[Scheduler] 定时任务已启动');
}
