package com.training.server.repository;

import com.training.server.entity.Department;
import org.springframework.stereotype.Repository;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;

@Repository
public class DepartmentRepository {

    private final Map<String, Department> departments = new HashMap<>();

    public Department save(Department department) {
        departments.put(department.getId(), department);
        return department;
    }

    public Optional<Department> findById(String id) {
        return Optional.ofNullable(departments.get(id));
    }

    public List<Department> findAll() {
        return new ArrayList<>(departments.values());
    }

    public void deleteById(String id) {
        departments.remove(id);
    }

    public boolean existsById(String id) {
        return departments.containsKey(id);
    }

    public Optional<Department> findByName(String name) {
        for (Department department : departments.values()) {
            if (name.equals(department.getName())) {
                return Optional.of(department);
            }
        }
        return Optional.empty();
    }
}
