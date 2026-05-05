package com.reimbursement.server.service;

import com.reimbursement.common.dto.MonthlyReportDTO;
import com.reimbursement.server.entity.CostCenter;
import com.reimbursement.server.entity.Reimbursement;
import com.reimbursement.server.entity.ReimbursementAllocation;
import com.reimbursement.server.repository.InMemoryDataStore;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.time.YearMonth;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

@Service
public class MonthlyReportService {
    private final InMemoryDataStore dataStore;
    private final Map<String, List<MonthlyReportEntry>> reportData = new ConcurrentHashMap<>();

    public MonthlyReportService(InMemoryDataStore dataStore) {
        this.dataStore = dataStore;
    }

    public void addAllocationToReport(Reimbursement reimbursement) {
        LocalDateTime now = LocalDateTime.now();
        int year = now.getYear();
        int month = now.getMonthValue();

        for (ReimbursementAllocation allocation : reimbursement.getAllocations()) {
            String key = buildKey(year, month, allocation.getCostCenterId());
            List<MonthlyReportEntry> entries = reportData.computeIfAbsent(key, k -> new ArrayList<>());
            
            MonthlyReportEntry entry = new MonthlyReportEntry();
            entry.setReimbursementId(reimbursement.getId());
            entry.setEmployeeName(reimbursement.getEmployeeName());
            entry.setReimbursementType(reimbursement.getReimbursementType());
            entry.setAmount(allocation.getAllocatedAmount());
            entry.setTime(now);
            entries.add(entry);
        }
    }

    public MonthlyReportDTO getMonthlyReport(int year, int month, String costCenterId) {
        CostCenter costCenter = dataStore.getCostCenterById(costCenterId);
        if (costCenter == null) {
            return null;
        }

        String key = buildKey(year, month, costCenterId);
        List<MonthlyReportEntry> entries = reportData.getOrDefault(key, new ArrayList<>());

        BigDecimal totalAmount = entries.stream()
                .map(MonthlyReportEntry::getAmount)
                .reduce(BigDecimal.ZERO, BigDecimal::add);

        BigDecimal budgetAmount = costCenter.getMonthlyBudget();
        BigDecimal remainingBudget = budgetAmount.subtract(totalAmount);
        boolean isOverBudget = remainingBudget.compareTo(BigDecimal.ZERO) < 0;

        MonthlyReportDTO dto = new MonthlyReportDTO();
        dto.setYear(year);
        dto.setMonth(month);
        dto.setCostCenterId(costCenterId);
        dto.setCostCenterName(costCenter.getName());
        dto.setTotalAmount(totalAmount);
        dto.setBudgetAmount(budgetAmount);
        dto.setRemainingBudget(remainingBudget);
        dto.setOverBudget(isOverBudget);

        List<MonthlyReportDTO.ReimbursementSummaryDTO> summaries = new ArrayList<>();
        for (MonthlyReportEntry entry : entries) {
            MonthlyReportDTO.ReimbursementSummaryDTO summary = new MonthlyReportDTO.ReimbursementSummaryDTO();
            summary.setReimbursementId(entry.getReimbursementId());
            summary.setEmployeeName(entry.getEmployeeName());
            summary.setReimbursementType(entry.getReimbursementType());
            summary.setAmount(entry.getAmount());
            summaries.add(summary);
        }
        dto.setReimbursements(summaries);

        return dto;
    }

    public List<MonthlyReportDTO> getAllMonthlyReports(int year, int month) {
        List<MonthlyReportDTO> reports = new ArrayList<>();
        for (CostCenter costCenter : dataStore.getCostCenterMap().values()) {
            MonthlyReportDTO report = getMonthlyReport(year, month, costCenter.getId());
            if (report != null) {
                reports.add(report);
            }
        }
        return reports;
    }

    public boolean checkBudgetWarning(String costCenterId) {
        YearMonth currentMonth = YearMonth.now();
        MonthlyReportDTO report = getMonthlyReport(
            currentMonth.getYear(), 
            currentMonth.getMonthValue(), 
            costCenterId
        );
        
        if (report == null) {
            return false;
        }
        
        BigDecimal threshold = report.getBudgetAmount().multiply(new BigDecimal("0.9"));
        return report.getTotalAmount().compareTo(threshold) >= 0;
    }

    private String buildKey(int year, int month, String costCenterId) {
        return year + "-" + String.format("%02d", month) + "-" + costCenterId;
    }

    private static class MonthlyReportEntry {
        private String reimbursementId;
        private String employeeName;
        private String reimbursementType;
        private BigDecimal amount;
        private LocalDateTime time;

        public String getReimbursementId() {
            return reimbursementId;
        }

        public void setReimbursementId(String reimbursementId) {
            this.reimbursementId = reimbursementId;
        }

        public String getEmployeeName() {
            return employeeName;
        }

        public void setEmployeeName(String employeeName) {
            this.employeeName = employeeName;
        }

        public String getReimbursementType() {
            return reimbursementType;
        }

        public void setReimbursementType(String reimbursementType) {
            this.reimbursementType = reimbursementType;
        }

        public BigDecimal getAmount() {
            return amount;
        }

        public void setAmount(BigDecimal amount) {
            this.amount = amount;
        }

        public LocalDateTime getTime() {
            return time;
        }

        public void setTime(LocalDateTime time) {
            this.time = time;
        }
    }
}
