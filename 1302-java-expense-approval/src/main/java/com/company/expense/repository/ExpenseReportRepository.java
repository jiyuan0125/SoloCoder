package com.company.expense.repository;

import com.company.expense.entity.Employee;
import com.company.expense.entity.ExpenseReport;
import com.company.expense.enums.ExpenseStatus;
import com.company.expense.enums.ExpenseType;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.LocalDateTime;
import java.util.Optional;

@Repository
public interface ExpenseReportRepository extends JpaRepository<ExpenseReport, Long> {

    Optional<ExpenseReport> findByReportNo(String reportNo);

    Page<ExpenseReport> findByApplicantOrderBySubmittedAtDesc(Employee applicant, Pageable pageable);

    Page<ExpenseReport> findByApplicantAndStatusOrderBySubmittedAtDesc(Employee applicant, ExpenseStatus status, Pageable pageable);

    Page<ExpenseReport> findByApplicantAndExpenseTypeOrderBySubmittedAtDesc(Employee applicant, ExpenseType expenseType, Pageable pageable);

    Page<ExpenseReport> findByApplicantAndSubmittedAtBetweenOrderBySubmittedAtDesc(Employee applicant, LocalDateTime startDate, LocalDateTime endDate, Pageable pageable);

    @Query("SELECT e FROM ExpenseReport e WHERE e.applicant = :applicant " +
           "AND (:status IS NULL OR e.status = :status) " +
           "AND (:expenseType IS NULL OR e.expenseType = :expenseType) " +
           "AND (:startDate IS NULL OR e.submittedAt >= :startDate) " +
           "AND (:endDate IS NULL OR e.submittedAt <= :endDate) " +
           "ORDER BY e.submittedAt DESC")
    Page<ExpenseReport> findByApplicantWithFilters(
            @Param("applicant") Employee applicant,
            @Param("status") ExpenseStatus status,
            @Param("expenseType") ExpenseType expenseType,
            @Param("startDate") LocalDateTime startDate,
            @Param("endDate") LocalDateTime endDate,
            Pageable pageable);

    @Query("SELECT e FROM ExpenseReport e WHERE " +
           "(:status IS NULL OR e.status = :status) " +
           "AND (:expenseType IS NULL OR e.expenseType = :expenseType) " +
           "AND (:startDate IS NULL OR e.submittedAt >= :startDate) " +
           "AND (:endDate IS NULL OR e.submittedAt <= :endDate) " +
           "ORDER BY e.submittedAt DESC")
    Page<ExpenseReport> findAllWithFilters(
            @Param("status") ExpenseStatus status,
            @Param("expenseType") ExpenseType expenseType,
            @Param("startDate") LocalDateTime startDate,
            @Param("endDate") LocalDateTime endDate,
            Pageable pageable);

    @Query("SELECT e FROM ExpenseReport e WHERE e.currentApprover = :approver " +
           "AND (e.status = 'PENDING' OR e.status = 'APPROVING') " +
           "ORDER BY e.submittedAt DESC")
    Page<ExpenseReport> findPendingByApprover(@Param("approver") Employee approver, Pageable pageable);
}
