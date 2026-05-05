package com.employee.server.repository;

import com.employee.server.entity.Employee;
import org.springframework.stereotype.Repository;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

@Repository
public class EmployeeRepository {
    private final Map<String, Employee> employeeMap = new ConcurrentHashMap<>();

    public Employee save(Employee employee) {
        employeeMap.put(employee.getEmployeeId(), employee);
        return employee;
    }

    public Optional<Employee> findById(String employeeId) {
        return Optional.ofNullable(employeeMap.get(employeeId));
    }

    public List<Employee> findAll() {
        return new ArrayList<>(employeeMap.values());
    }

    public List<Employee> findByResigned(boolean resigned) {
        return employeeMap.values().stream()
                .filter(e -> e.isResigned() == resigned)
                .collect(Collectors.toList());
    }

    public boolean existsById(String employeeId) {
        return employeeMap.containsKey(employeeId);
    }

    public void deleteById(String employeeId) {
        employeeMap.remove(employeeId);
    }

    public boolean isPhoneUsedByOtherActiveEmployee(String phone, String excludeEmployeeId) {
        return employeeMap.values().stream()
                .anyMatch(e -> !e.isResigned()
                        && phone.equals(e.getPhone())
                        && !e.getEmployeeId().equals(excludeEmployeeId));
    }

    public List<Employee> searchByKeyword(String keyword, boolean includeResigned) {
        String lowerKeyword = keyword.toLowerCase();
        return employeeMap.values().stream()
                .filter(e -> includeResigned || !e.isResigned())
                .filter(e -> matchesKeyword(e, lowerKeyword))
                .sorted((e1, e2) -> compareByRelevance(e1, e2, lowerKeyword))
                .collect(Collectors.toList());
    }

    private boolean matchesKeyword(Employee employee, String lowerKeyword) {
        return employee.getName().toLowerCase().contains(lowerKeyword)
                || employee.getDepartment().toLowerCase().contains(lowerKeyword)
                || employee.getPosition().toLowerCase().contains(lowerKeyword);
    }

    private int compareByRelevance(Employee e1, Employee e2, String lowerKeyword) {
        int score1 = calculateRelevanceScore(e1, lowerKeyword);
        int score2 = calculateRelevanceScore(e2, lowerKeyword);
        return Integer.compare(score2, score1);
    }

    private int calculateRelevanceScore(Employee employee, String lowerKeyword) {
        int score = 0;
        
        if (lowerKeyword.equals(employee.getName().toLowerCase())) {
            score += 100;
        } else if (employee.getName().toLowerCase().contains(lowerKeyword)) {
            score += 50;
        }
        
        if (lowerKeyword.equals(employee.getDepartment().toLowerCase())) {
            score += 100;
        } else if (employee.getDepartment().toLowerCase().contains(lowerKeyword)) {
            score += 50;
        }
        
        if (lowerKeyword.equals(employee.getPosition().toLowerCase())) {
            score += 100;
        } else if (employee.getPosition().toLowerCase().contains(lowerKeyword)) {
            score += 50;
        }
        
        return score;
    }
}
