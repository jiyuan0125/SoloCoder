package com.payroll.server.service;

import com.payroll.common.dto.*;
import com.payroll.server.entity.ArchiveStatus;
import com.payroll.server.entity.DeductionDetail;
import com.payroll.server.entity.PayrollArchive;
import com.payroll.server.repository.PayrollArchiveRepository;
import com.payroll.server.util.ChecksumUtil;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.time.LocalDateTime;
import java.time.YearMonth;
import java.util.*;
import java.util.stream.Collectors;

@Service
public class PayrollArchiveService {

    private final PayrollArchiveRepository repository;

    @Autowired
    public PayrollArchiveService(PayrollArchiveRepository repository) {
        this.repository = repository;
    }

    public PayrollArchiveDTO createArchive(ArchiveRequestDTO request) {
        if (repository.existsByEmployeeIdAndYearMonth(
                request.getEmployeeId(), request.getYear(), request.getMonth())) {
            throw new IllegalArgumentException("该员工当月归档记录已存在");
        }

        PayrollArchive archive = convertToEntity(request);
        
        BigDecimal totalDeduction = calculateTotalDeduction(archive.getDeductionDetails());
        archive.setTotalDeduction(totalDeduction);
        archive.setNetPay(archive.getBaseSalary().subtract(totalDeduction));
        
        archive.setArchiveId(generateArchiveId(request.getEmployeeId(), request.getYear(), request.getMonth()));
        archive.setChecksum(ChecksumUtil.calculateChecksum(archive));
        
        repository.save(archive);
        
        return convertToDTO(archive, true);
    }

    public PayrollArchiveDTO confirmArchive(String archiveId) {
        PayrollArchive archive = repository.findByArchiveId(archiveId)
                .orElseThrow(() -> new IllegalArgumentException("归档记录不存在"));
        
        if (archive.getStatus() == ArchiveStatus.CONFIRMED) {
            throw new IllegalArgumentException("该归档已确认，不可重复确认");
        }
        
        archive.setStatus(ArchiveStatus.CONFIRMED);
        archive.setConfirmedAt(LocalDateTime.now());
        
        repository.save(archive);
        
        return convertToDTO(archive, true);
    }

    public PayrollArchiveDTO addRemark(String archiveId, String remark) {
        PayrollArchive archive = repository.findByArchiveId(archiveId)
                .orElseThrow(() -> new IllegalArgumentException("归档记录不存在"));
        
        if (remark == null || remark.trim().isEmpty()) {
            throw new IllegalArgumentException("备注内容不能为空");
        }
        
        archive.getRemarks().add(remark);
        repository.save(archive);
        
        return convertToDTO(archive, true);
    }

    public PayrollArchiveDTO getByArchiveId(String archiveId) {
        PayrollArchive archive = repository.findByArchiveId(archiveId)
                .orElseThrow(() -> new IllegalArgumentException("归档记录不存在"));
        
        boolean isValid = ChecksumUtil.verifyChecksum(archive, archive.getChecksum());
        
        return convertToDTO(archive, isValid);
    }

    public List<PayrollArchiveDTO> queryByEmployee(String employeeId) {
        List<PayrollArchive> archives = repository.findByEmployeeId(employeeId);
        return archives.stream()
                .map(a -> convertToDTO(a, ChecksumUtil.verifyChecksum(a, a.getChecksum())))
                .collect(Collectors.toList());
    }

    public List<PayrollArchiveDTO> queryByYear(int year) {
        List<PayrollArchive> archives = repository.findByYear(year);
        return archives.stream()
                .map(a -> convertToDTO(a, ChecksumUtil.verifyChecksum(a, a.getChecksum())))
                .collect(Collectors.toList());
    }

    public List<PayrollArchiveDTO> queryByYearMonth(int year, int month) {
        List<PayrollArchive> archives = repository.findByYearAndMonth(year, month);
        return archives.stream()
                .map(a -> convertToDTO(a, ChecksumUtil.verifyChecksum(a, a.getChecksum())))
                .collect(Collectors.toList());
    }

    public List<PayrollArchiveDTO> queryByEmployeeAndYear(String employeeId, int year) {
        List<PayrollArchive> archives = repository.findByEmployeeIdAndYear(employeeId, year);
        return archives.stream()
                .map(a -> convertToDTO(a, ChecksumUtil.verifyChecksum(a, a.getChecksum())))
                .collect(Collectors.toList());
    }

    public List<PayrollArchiveDTO> queryByYears(List<Integer> years) {
        List<PayrollArchive> archives = repository.findByYears(years);
        return archives.stream()
                .map(a -> convertToDTO(a, ChecksumUtil.verifyChecksum(a, a.getChecksum())))
                .collect(Collectors.toList());
    }

    public AnnualReportDTO generateAnnualReport(int year) {
        List<PayrollArchive> archives = repository.findByYear(year);
        List<PayrollArchive> confirmedArchives = archives.stream()
                .filter(a -> a.getStatus() == ArchiveStatus.CONFIRMED)
                .collect(Collectors.toList());

        if (confirmedArchives.isEmpty()) {
            AnnualReportDTO report = new AnnualReportDTO();
            report.setYear(year);
            report.setTotalExpense(BigDecimal.ZERO);
            report.setAverageSalary(BigDecimal.ZERO);
            report.setMaxSalary(BigDecimal.ZERO);
            report.setMinSalary(BigDecimal.ZERO);
            report.setTotalRecords(0);
            report.setTotalEmployees(0);
            report.setYearOverYearChange(BigDecimal.ZERO);
            report.setMonthlySummaries(new ArrayList<>());
            return report;
        }

        BigDecimal totalExpense = confirmedArchives.stream()
                .map(PayrollArchive::getNetPay)
                .reduce(BigDecimal.ZERO, BigDecimal::add);

        Set<String> uniqueEmployees = confirmedArchives.stream()
                .map(PayrollArchive::getEmployeeId)
                .collect(Collectors.toSet());

        BigDecimal averageSalary = totalExpense.divide(
                BigDecimal.valueOf(uniqueEmployees.size()), 2, RoundingMode.HALF_UP);

        List<BigDecimal> netPays = confirmedArchives.stream()
                .map(PayrollArchive::getNetPay)
                .sorted()
                .collect(Collectors.toList());

        BigDecimal maxSalary = netPays.get(netPays.size() - 1);
        BigDecimal minSalary = netPays.get(0);

        List<MonthlySummaryDTO> monthlySummaries = new ArrayList<>();
        for (int month = 1; month <= 12; month++) {
            int finalMonth = month;
            List<PayrollArchive> monthArchives = confirmedArchives.stream()
                    .filter(a -> a.getMonth() == finalMonth)
                    .collect(Collectors.toList());
            
            if (!monthArchives.isEmpty()) {
                BigDecimal monthTotal = monthArchives.stream()
                        .map(PayrollArchive::getNetPay)
                        .reduce(BigDecimal.ZERO, BigDecimal::add);
                monthlySummaries.add(new MonthlySummaryDTO(month, monthTotal, monthArchives.size()));
            }
        }

        BigDecimal yearOverYearChange = calculateYearOverYearChange(year);

        AnnualReportDTO report = new AnnualReportDTO();
        report.setYear(year);
        report.setTotalExpense(totalExpense);
        report.setAverageSalary(averageSalary);
        report.setMaxSalary(maxSalary);
        report.setMinSalary(minSalary);
        report.setTotalRecords(confirmedArchives.size());
        report.setTotalEmployees(uniqueEmployees.size());
        report.setYearOverYearChange(yearOverYearChange);
        report.setMonthlySummaries(monthlySummaries);

        return report;
    }

    public YearComparisonDTO compareYears(int baseYear, int compareYear) {
        AnnualReportDTO baseReport = generateAnnualReport(baseYear);
        AnnualReportDTO compareReport = generateAnnualReport(compareYear);

        BigDecimal totalChange = compareReport.getTotalExpense().subtract(baseReport.getTotalExpense());
        BigDecimal totalChangeRate = BigDecimal.ZERO;
        if (baseReport.getTotalExpense().compareTo(BigDecimal.ZERO) > 0) {
            totalChangeRate = totalChange.divide(baseReport.getTotalExpense(), 4, RoundingMode.HALF_UP)
                    .multiply(BigDecimal.valueOf(100));
        }

        List<MonthlyComparisonDTO> monthlyComparisons = new ArrayList<>();
        for (int month = 1; month <= 12; month++) {
            MonthlySummaryDTO baseMonth = findMonthlySummary(baseReport.getMonthlySummaries(), month);
            MonthlySummaryDTO compareMonth = findMonthlySummary(compareReport.getMonthlySummaries(), month);

            BigDecimal baseAmount = baseMonth != null ? baseMonth.getTotalExpense() : BigDecimal.ZERO;
            BigDecimal compareAmount = compareMonth != null ? compareMonth.getTotalExpense() : BigDecimal.ZERO;

            if (baseAmount.compareTo(BigDecimal.ZERO) > 0 || compareAmount.compareTo(BigDecimal.ZERO) > 0) {
                BigDecimal change = compareAmount.subtract(baseAmount);
                BigDecimal changeRate = BigDecimal.ZERO;
                if (baseAmount.compareTo(BigDecimal.ZERO) > 0) {
                    changeRate = change.divide(baseAmount, 4, RoundingMode.HALF_UP)
                            .multiply(BigDecimal.valueOf(100));
                }

                MonthlyComparisonDTO monthlyComparison = new MonthlyComparisonDTO();
                monthlyComparison.setMonth(month);
                monthlyComparison.setBaseYearAmount(baseAmount);
                monthlyComparison.setCompareYearAmount(compareAmount);
                monthlyComparison.setChange(change);
                monthlyComparison.setChangeRate(changeRate);
                monthlyComparisons.add(monthlyComparison);
            }
        }

        YearComparisonDTO comparison = new YearComparisonDTO();
        comparison.setBaseYear(baseYear);
        comparison.setCompareYear(compareYear);
        comparison.setBaseYearTotal(baseReport.getTotalExpense());
        comparison.setCompareYearTotal(compareReport.getTotalExpense());
        comparison.setTotalChange(totalChange);
        comparison.setTotalChangeRate(totalChangeRate);
        comparison.setMonthlyComparisons(monthlyComparisons);

        return comparison;
    }

    private BigDecimal calculateYearOverYearChange(int year) {
        int previousYear = year - 1;
        List<PayrollArchive> previousYearArchives = repository.findByYear(previousYear).stream()
                .filter(a -> a.getStatus() == ArchiveStatus.CONFIRMED)
                .collect(Collectors.toList());

        if (previousYearArchives.isEmpty()) {
            return BigDecimal.ZERO;
        }

        BigDecimal previousTotal = previousYearArchives.stream()
                .map(PayrollArchive::getNetPay)
                .reduce(BigDecimal.ZERO, BigDecimal::add);

        List<PayrollArchive> currentYearArchives = repository.findByYear(year).stream()
                .filter(a -> a.getStatus() == ArchiveStatus.CONFIRMED)
                .collect(Collectors.toList());

        BigDecimal currentTotal = currentYearArchives.stream()
                .map(PayrollArchive::getNetPay)
                .reduce(BigDecimal.ZERO, BigDecimal::add);

        if (previousTotal.compareTo(BigDecimal.ZERO) == 0) {
            return BigDecimal.ZERO;
        }

        return currentTotal.subtract(previousTotal)
                .divide(previousTotal, 4, RoundingMode.HALF_UP)
                .multiply(BigDecimal.valueOf(100));
    }

    private MonthlySummaryDTO findMonthlySummary(List<MonthlySummaryDTO> summaries, int month) {
        return summaries.stream()
                .filter(s -> s.getMonth() == month)
                .findFirst()
                .orElse(null);
    }

    private BigDecimal calculateTotalDeduction(List<DeductionDetail> details) {
        if (details == null || details.isEmpty()) {
            return BigDecimal.ZERO;
        }
        return details.stream()
                .map(DeductionDetail::getAmount)
                .reduce(BigDecimal.ZERO, BigDecimal::add);
    }

    private String generateArchiveId(String employeeId, int year, int month) {
        return String.format("ARC-%s-%04d%02d-%s",
                employeeId, year, month, UUID.randomUUID().toString().substring(0, 8).toUpperCase());
    }

    private PayrollArchive convertToEntity(ArchiveRequestDTO request) {
        PayrollArchive archive = new PayrollArchive();
        archive.setEmployeeId(request.getEmployeeId());
        archive.setEmployeeName(request.getEmployeeName());
        archive.setIdCard(request.getIdCard());
        archive.setYear(request.getYear());
        archive.setMonth(request.getMonth());
        archive.setBaseSalary(request.getBaseSalary());
        
        if (request.getDeductionDetails() != null) {
            List<DeductionDetail> details = request.getDeductionDetails().stream()
                    .map(dto -> new DeductionDetail(dto.getDeductionType(), dto.getAmount(), dto.getDescription()))
                    .collect(Collectors.toList());
            archive.setDeductionDetails(details);
        }
        
        return archive;
    }

    private PayrollArchiveDTO convertToDTO(PayrollArchive archive, boolean isValid) {
        PayrollArchiveDTO dto = new PayrollArchiveDTO();
        dto.setArchiveId(archive.getArchiveId());
        dto.setEmployeeId(archive.getEmployeeId());
        dto.setEmployeeName(archive.getEmployeeName());
        dto.setIdCard(archive.getIdCard());
        dto.setYear(archive.getYear());
        dto.setMonth(archive.getMonth());
        dto.setBaseSalary(archive.getBaseSalary());
        
        if (archive.getDeductionDetails() != null) {
            List<DeductionDetailDTO> details = archive.getDeductionDetails().stream()
                    .map(e -> new DeductionDetailDTO(e.getDeductionType(), e.getAmount(), e.getDescription()))
                    .collect(Collectors.toList());
            dto.setDeductionDetails(details);
        }
        
        dto.setTotalDeduction(archive.getTotalDeduction());
        dto.setNetPay(archive.getNetPay());
        dto.setChecksum(archive.getChecksum());
        dto.setDataValid(isValid);
        dto.setStatus(archive.getStatus().name());
        dto.setRemarks(new ArrayList<>(archive.getRemarks()));
        dto.setCreatedAt(archive.getCreatedAt());
        dto.setConfirmedAt(archive.getConfirmedAt());
        
        return dto;
    }
}
