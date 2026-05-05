package com.performance.server.repository;

import com.performance.server.entity.Department;
import org.springframework.stereotype.Repository;

import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicLong;

@Repository
public class DepartmentRepository {
    private final Map<Long, Department> departments = new ConcurrentHashMap<>();
    private final AtomicLong idGenerator = new AtomicLong(1);

    public Department save(Department department) {
        if (department.getId() == null) {
            department.setId(idGenerator.getAndIncrement());
            department.setCreatedAt(LocalDateTime.now());
        }
        department.setUpdatedAt(LocalDateTime.now());
        departments.put(department.getId(), department);
        return department;
    }

    public Optional<Department> findById(Long id) {
        return Optional.ofNullable(departments.get(id));
    }

    public List<Department> findAll() {
        return new ArrayList<>(departments.values());
    }

    public Optional<Department> findByCode(String code) {
        for (Department department : departments.values()) {
            if (code.equals(department.getCode())) {
                return Optional.of(department);
            }
        }
        return Optional.empty();
    }

    public boolean existsById(Long id) {
        return departments.containsKey(id);
    }

    public void deleteById(Long id) {
        departments.remove(id);
    }
}