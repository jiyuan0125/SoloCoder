package com.company.leave.repository;

import com.company.leave.entity.Employee;
import org.springframework.stereotype.Repository;

import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.stream.Collectors;

@Repository
public class EmployeeRepository {
    private final Map<Long, Employee> employees = new HashMap<>();
    private Long nextId = 1L;

    public Employee save(Employee employee) {
        if (employee.getId() == null) {
            employee.setId(nextId++);
        }
        employees.put(employee.getId(), employee);
        return employee;
    }

    public Optional<Employee> findById(Long id) {
        return Optional.ofNullable(employees.get(id));
    }

    public List<Employee> findAll() {
        return List.copyOf(employees.values());
    }

    public List<Employee> findByManagerId(Long managerId) {
        return employees.values().stream()
                .filter(e -> managerId.equals(e.getManagerId()))
                .collect(Collectors.toList());
    }

    public boolean existsById(Long id) {
        return employees.containsKey(id);
    }
}
