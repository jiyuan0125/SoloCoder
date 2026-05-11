package com.company.payroll.service;

import com.company.payroll.model.Department;
import com.company.payroll.model.DepartmentSalarySummary;
import com.company.payroll.model.Employee;
import com.company.payroll.model.PaySlip;

import java.util.ArrayList;
import java.util.List;

public class DepartmentSummaryService {
    private final PayrollDataStore dataStore;

    public DepartmentSummaryService(PayrollDataStore dataStore) {
        this.dataStore = dataStore;
    }

    public DepartmentSalarySummary calculateDepartmentSummary(String departmentId, int year, int month) {
        Department department = dataStore.getDepartment(departmentId);
        if (department == null) {
            throw new IllegalArgumentException("部门不存在: " + departmentId);
        }

        DepartmentSalarySummary summary = new DepartmentSalarySummary();
        summary.setDepartmentId(departmentId);
        summary.setDepartmentName(department.getName());
        summary.setYear(year);
        summary.setMonth(month);

        List<PaySlip> paySlips = dataStore.getPaySlipsByDepartmentAndMonth(departmentId, year, month);

        for (PaySlip paySlip : paySlips) {
            summary.addPaySlip(paySlip);
        }

        return summary;
    }

    public List<DepartmentSalarySummary> calculateAllDepartmentsSummary(int year, int month) {
        List<DepartmentSalarySummary> summaries = new ArrayList<>();

        for (Department department : dataStore.getAllDepartments()) {
            summaries.add(calculateDepartmentSummary(department.getId(), year, month));
        }

        return summaries;
    }

    public String formatSummary(DepartmentSalarySummary summary) {
        StringBuilder sb = new StringBuilder();
        sb.append("==================== 部门薪资汇总 ====================\n");
        sb.append("部门: ").append(summary.getDepartmentName()).append("\n");
        sb.append("月份: ").append(summary.getYear()).append("年").append(summary.getMonth()).append("月\n");
        sb.append("员工数: ").append(summary.getEmployeeCount()).append("\n");
        sb.append("------------------------------------------------\n");
        sb.append("薪资总额: ¥").append(summary.getTotalGrossSalary()).append("\n");
        sb.append("平均薪资: ¥").append(summary.getAverageGrossSalary()).append("\n");
        sb.append("个税总额: ¥").append(summary.getTotalTax()).append("\n");
        sb.append("五险一金总额: ¥").append(summary.getTotalSocialSecurity()).append("\n");
        sb.append("实发总额: ¥").append(summary.getTotalNetSalary()).append("\n");
        sb.append("================================================\n");
        return sb.toString();
    }
}
