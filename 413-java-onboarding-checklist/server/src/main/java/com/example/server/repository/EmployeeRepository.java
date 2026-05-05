package com.example.server.repository;

import com.example.common.dto.EmployeeDTO;
import com.example.common.enums.OnboardingStatus;
import org.springframework.stereotype.Repository;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

@Repository
public class EmployeeRepository {
    private final Map<String, EmployeeDTO> employees = new ConcurrentHashMap<>();

    public EmployeeDTO save(EmployeeDTO employee) {
        employees.put(employee.getId(), employee);
        return employee;
    }

    public Optional<EmployeeDTO> findById(String id) {
        return Optional.ofNullable(employees.get(id));
    }

    public List<EmployeeDTO> findAll() {
        return new ArrayList<>(employees.values());
    }

    public void deleteById(String id) {
        employees.remove(id);
    }

    public boolean existsById(String id) {
        return employees.containsKey(id);
    }

    public List<EmployeeDTO> findByStatus(OnboardingStatus status) {
        return employees.values().stream()
                .filter(e -> e.getStatus() == status)
                .collect(Collectors.toList());
    }

    public long countByStatus(OnboardingStatus status) {
        return employees.values().stream()
                .filter(e -> e.getStatus() == status)
                .count();
    }
}
