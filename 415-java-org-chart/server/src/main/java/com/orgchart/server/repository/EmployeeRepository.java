package com.orgchart.server.repository;

import com.orgchart.server.entity.Employee;
import org.springframework.stereotype.Repository;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

@Repository
public class EmployeeRepository {

    private final Map<String, Employee> employees = new ConcurrentHashMap<>();

    public Employee save(Employee employee) {
        employees.put(employee.getId(), employee);
        return employee;
    }

    public Optional<Employee> findById(String id) {
        return Optional.ofNullable(employees.get(id));
    }

    public List<Employee> findAll() {
        return new ArrayList<>(employees.values());
    }

    public void deleteById(String id) {
        employees.remove(id);
    }

    public boolean existsById(String id) {
        return employees.containsKey(id);
    }

    public List<Employee> findByDepartmentId(String departmentId) {
        return employees.values().stream()
                .filter(emp -> departmentId.equals(emp.getDepartmentId()))
                .collect(Collectors.toList());
    }

    public List<Employee> findByManagerId(String managerId) {
        return employees.values().stream()
                .filter(emp -> managerId.equals(emp.getManagerId()))
                .collect(Collectors.toList());
    }

    public boolean existsByEmail(String email) {
        if (email == null) {
            return false;
        }
        return employees.values().stream()
                .anyMatch(emp -> email.equals(emp.getEmail()));
    }

    public boolean existsByEmailExcludingId(String email, String excludeId) {
        if (email == null) {
            return false;
        }
        return employees.values().stream()
                .anyMatch(emp -> email.equals(emp.getEmail()) && !excludeId.equals(emp.getId()));
    }
}
