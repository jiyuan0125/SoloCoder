package com.performance.server.repository;

import com.performance.server.entity.Employee;
import org.springframework.stereotype.Repository;

import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicLong;

@Repository
public class EmployeeRepository {
    private final Map<Long, Employee> employees = new ConcurrentHashMap<>();
    private final AtomicLong idGenerator = new AtomicLong(1);

    public Employee save(Employee employee) {
        if (employee.getId() == null) {
            employee.setId(idGenerator.getAndIncrement());
            employee.setCreatedAt(LocalDateTime.now());
        }
        employee.setUpdatedAt(LocalDateTime.now());
        employees.put(employee.getId(), employee);
        return employee;
    }

    public Optional<Employee> findById(Long id) {
        return Optional.ofNullable(employees.get(id));
    }

    public List<Employee> findAll() {
        return new ArrayList<>(employees.values());
    }

    public List<Employee> findByDepartmentId(Long departmentId) {
        List<Employee> result = new ArrayList<>();
        for (Employee employee : employees.values()) {
            if (departmentId.equals(employee.getDepartmentId())) {
                result.add(employee);
            }
        }
        return result;
    }

    public List<Employee> findByManagerId(Long managerId) {
        List<Employee> result = new ArrayList<>();
        for (Employee employee : employees.values()) {
            if (managerId.equals(employee.getManagerId())) {
                result.add(employee);
            }
        }
        return result;
    }

    public boolean existsById(Long id) {
        return employees.containsKey(id);
    }

    public void deleteById(Long id) {
        employees.remove(id);
    }
}