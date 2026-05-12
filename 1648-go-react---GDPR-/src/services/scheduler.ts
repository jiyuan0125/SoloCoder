import cron from 'node-cron';
import { store } from '../store';
import { daysBetween, isOverdue } from '../utils';

export function startScheduler(): void {
  cron.schedule('0 * * * *', () => {
    console.log(`[${new Date().toISOString()}] Running retention period scan...`);
    scanRetentionPeriod();
  });

  cron.schedule('0 * * * *', () => {
    console.log(`[${new Date().toISOString()}] Running overdue request check...`);
    checkOverdueRequests();
  });

  console.log('Scheduled tasks started: retention scan and overdue check (every hour)');
}

function scanRetentionPeriod(): void {
  const categories = store.getAllDataCategories();
  const categoryRetentionMap = new Map<string, number>();
  for (const cat of categories) {
    categoryRetentionMap.set(cat.id, cat.retentionDays);
  }

  const allData = store.getAllUserData();
  const now = new Date();
  let marked = 0;

  for (const record of allData) {
    if (record.status !== 'active') continue;

    const retentionDays = categoryRetentionMap.get(record.categoryId);
    if (!retentionDays) continue;

    const age = daysBetween(record.createdAt, now);
    if (age >= retentionDays) {
      store.updateUserData(record.id, { status: 'pending_cleanup' });
      marked++;
    }
  }

  if (marked > 0) {
    console.log(`Marked ${marked} records as pending_cleanup due to retention period expiration`);
  }
}

function checkOverdueRequests(): void {
  const requests = store.getAllRequests();
  let marked = 0;

  for (const req of requests) {
    if (req.status === 'completed' || req.status === 'overdue') continue;

    if (isOverdue(req.deadline)) {
      store.updateRequest(req.id, { status: 'overdue' });
      marked++;
    }
  }

  if (marked > 0) {
    console.log(`Marked ${marked} requests as overdue`);
  }
}
