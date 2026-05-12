import { registerRateUpdateCallback } from './exchangeRate';
import { TransactionStatus } from './types';
import { 
  getPendingTransactionsAwaitingRate, 
  getTransaction, 
  processTransaction, 
  updateTransactionStatus
} from './transaction';

let isProcessing = false;

const processPendingTransactions = async (): Promise<void> => {
  if (isProcessing) {
    return;
  }
  
  isProcessing = true;
  
  try {
    const pendingTransactions = await getPendingTransactionsAwaitingRate();
    
    for (const transaction of pendingTransactions) {
      const currentTransaction = await getTransaction(transaction.id);
      
      if (!currentTransaction) {
        continue;
      }
      
      if (currentTransaction.status === TransactionStatus.CANCELLED) {
        continue;
      }
      
      const result = await processTransaction(currentTransaction);
      
      await updateTransactionStatus(
        currentTransaction.id,
        result.status,
        result.targetAmount,
        result.rateInfo
      );
    }
  } catch (error) {
    console.error('处理待处理交易时出错:', error);
  } finally {
    isProcessing = false;
  }
};

export const initWakeUpMechanism = (): void => {
  registerRateUpdateCallback(() => {
    processPendingTransactions();
  });
};
