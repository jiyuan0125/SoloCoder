package com.company.expense.repository;

import com.company.expense.entity.ApprovalRecord;
import com.company.expense.entity.ExpenseReport;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface ApprovalRecordRepository extends JpaRepository<ApprovalRecord, Long> {

    List<ApprovalRecord> findByExpenseReportOrderByCreatedAt(ExpenseReport expenseReport);
}
