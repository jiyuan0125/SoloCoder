from typing import Dict, List, Optional
from datetime import datetime
import asyncio

from models import Transaction, TransactionStatus, Step, StepStatus
from operations import get_executor, get_compensator


class TransactionEngine:
    def __init__(self):
        self.transactions: Dict[str, Transaction] = {}
        self._lock = asyncio.Lock()

    async def create_transaction(self, steps_data: List[Dict]) -> Transaction:
        transaction = Transaction.create(steps_data)
        async with self._lock:
            self.transactions[transaction.id] = transaction
        return transaction

    def get_transaction(self, transaction_id: str) -> Optional[Transaction]:
        return self.transactions.get(transaction_id)

    def list_transactions(self, status_filter: Optional[str] = None) -> List[Transaction]:
        transactions = list(self.transactions.values())
        if status_filter:
            transactions = [
                t for t in transactions
                if t.status.value == status_filter
            ]
        return sorted(transactions, key=lambda t: t.created_at, reverse=True)

    async def execute_transaction(self, transaction_id: str) -> Transaction:
        transaction = self.get_transaction(transaction_id)
        if not transaction:
            raise ValueError(f"事务不存在: {transaction_id}")

        async with self._lock:
            if transaction.status != TransactionStatus.PENDING:
                raise ValueError(f"事务状态不是 pending，无法执行: {transaction.status.value}")
            transaction.status = TransactionStatus.EXECUTING
            transaction.updated_at = datetime.utcnow()

        failed_step_index = None

        for i, step in enumerate(transaction.steps):
            async with self._lock:
                step.status = StepStatus.PENDING
            try:
                executor = get_executor(step.operation_type)
                result = await executor(step.params)
                async with self._lock:
                    step.status = StepStatus.SUCCESS
                    step.result = result
                    step.executed_at = datetime.utcnow()
            except Exception as e:
                async with self._lock:
                    step.status = StepStatus.FAILED
                    step.error_message = str(e)
                    step.executed_at = datetime.utcnow()
                failed_step_index = i
                async with self._lock:
                    transaction.error_message = f"步骤 {i} ({step.operation_type}) 执行失败: {str(e)}"
                break

        if failed_step_index is None:
            async with self._lock:
                transaction.status = TransactionStatus.COMPLETED
                transaction.updated_at = datetime.utcnow()
        else:
            await self._compensate_transaction(transaction, failed_step_index)

        return transaction

    async def _compensate_transaction(
        self,
        transaction: Transaction,
        failed_step_index: int
    ) -> None:
        async with self._lock:
            transaction.status = TransactionStatus.COMPENSATING
            transaction.updated_at = datetime.utcnow()

        compensation_failed = False

        for i in range(failed_step_index - 1, -1, -1):
            step = transaction.steps[i]
            
            if step.status != StepStatus.SUCCESS:
                continue

            try:
                compensator = get_compensator(step.operation_type)
                result = await compensator(step.params)
                async with self._lock:
                    step.compensation_status = "success"
                    step.result = result
                    step.compensated_at = datetime.utcnow()
            except Exception as e:
                async with self._lock:
                    step.compensation_status = "failed"
                    step.compensation_error = str(e)
                compensation_failed = True
                break

        async with self._lock:
            if compensation_failed:
                transaction.status = TransactionStatus.COMPENSATION_FAILED
            else:
                transaction.status = TransactionStatus.COMPENSATED
            transaction.updated_at = datetime.utcnow()

    async def retry_compensation(self, transaction_id: str) -> Transaction:
        transaction = self.get_transaction(transaction_id)
        if not transaction:
            raise ValueError(f"事务不存在: {transaction_id}")

        async with self._lock:
            if transaction.status != TransactionStatus.COMPENSATION_FAILED:
                raise ValueError(f"事务状态不是 compensation_failed，无法重试补偿: {transaction.status.value}")
            transaction.status = TransactionStatus.COMPENSATING
            transaction.updated_at = datetime.utcnow()

        all_compensated = True

        for step in reversed(transaction.steps):
            if step.status != StepStatus.SUCCESS:
                continue
            
            if step.compensation_status == "success":
                continue

            if step.compensation_status is None and step.compensation_error is None:
                continue

            try:
                compensator = get_compensator(step.operation_type)
                result = await compensator(step.params)
                async with self._lock:
                    step.compensation_status = "success"
                    step.compensation_error = None
                    step.result = result
                    step.compensated_at = datetime.utcnow()
            except Exception as e:
                async with self._lock:
                    step.compensation_status = "failed"
                    step.compensation_error = str(e)
                all_compensated = False
                break

        async with self._lock:
            if all_compensated:
                transaction.status = TransactionStatus.COMPENSATED
            else:
                transaction.status = TransactionStatus.COMPENSATION_FAILED
            transaction.updated_at = datetime.utcnow()

        return transaction
