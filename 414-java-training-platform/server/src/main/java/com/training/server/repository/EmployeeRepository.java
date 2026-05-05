package com.training.server.repository;

import com.training.server.entity.Employee;
import org.springframework.stereotype.Repository;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;

@Repository
public class EmployeeRepository {

    private final Map<String, Employee> employees = new HashMap<>();

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
        List<Employee> result = new ArrayList<>();
        for (Employee employee : employees.values()) {
            if (departmentId.equals(employee.getDepartmentId())) {
                result.add(employee);
            }
        }
        return result;
    }

    public int countByDepartmentId(String departmentId) {
        int count = 0;
        for (Employee employee : employees.values()) {
            if (departmentId.equals(employee.getDepartmentId())) {
                count++;
            }
        }
        return count;
    }
}
