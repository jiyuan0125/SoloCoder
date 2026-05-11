package com.company.payroll.service;

import com.company.payroll.config.AllowanceConfig;
import com.company.payroll.model.*;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.time.LocalDate;
import java.util.*;
import java.util.stream.Collectors;

public class PayrollDataStore {
    private final Map<String, Department> departments = new HashMap<>();
    private final Map<String, Employee> employees = new HashMap<>();
    private final List<PaySlip> paySlips = new ArrayList<>();

    public void addDepartment(Department department) {
        departments.put(department.getId(), department);
    }

    public Department getDepartment(String id) {
        return departments.get(id);
    }

    public Collection<Department> getAllDepartments() {
        return departments.values();
    }

    public void addEmployee(Employee employee) {
        employees.put(employee.getId(), employee);
    }

    public Employee getEmployee(String id) {
        return employees.get(id);
    }

    public Collection<Employee> getAllEmployees() {
        return employees.values();
    }

    public List<Employee> getEmployeesByDepartment(String departmentId) {
        return employees.values().stream()
                .filter(e -> departmentId.equals(e.getDepartmentId()))
                .collect(Collectors.toList());
    }

    public void addPaySlip(PaySlip paySlip) {
        paySlips.add(paySlip);
    }

    public List<PaySlip> getPaySlipsByEmployee(String employeeId) {
        return paySlips.stream()
                .filter(p -> employeeId.equals(p.getEmployeeId()))
                .sorted(Comparator.comparing(PaySlip::getYear).thenComparing(PaySlip::getMonth))
                .collect(Collectors.toList());
    }

    public List<PaySlip> getPaySlipsByMonth(int year, int month) {
        return paySlips.stream()
                .filter(p -> p.getYear() == year && p.getMonth() == month)
                .collect(Collectors.toList());
    }

    public List<PaySlip> getPreviousPaySlips(String employeeId, int year, int month) {
        return paySlips.stream()
                .filter(p -> employeeId.equals(p.getEmployeeId()))
                .filter(p -> {
                    if (p.getYear() < year) return true;
                    return p.getYear() == year && p.getMonth() < month;
                })
                .collect(Collectors.toList());
    }

    public List<PaySlip> getPaySlipsByDepartmentAndMonth(String departmentId, int year, int month) {
        return paySlips.stream()
                .filter(p -> p.getYear() == year && p.getMonth() == month)
                .filter(p -> {
                    Employee emp = employees.get(p.getEmployeeId());
                    return emp != null && departmentId.equals(emp.getDepartmentId());
                })
                .collect(Collectors.toList());
    }
}
