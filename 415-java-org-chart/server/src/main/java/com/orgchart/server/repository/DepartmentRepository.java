package com.orgchart.server.repository;

import com.orgchart.server.entity.Department;
import org.springframework.stereotype.Repository;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

@Repository
public class DepartmentRepository {

    private final Map<String, Department> departments = new ConcurrentHashMap<>();

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

    public List<Department> findByParentId(String parentId) {
        return departments.values().stream()
                .filter(dep -> {
                    if (parentId == null) {
                        return dep.getParentId() == null;
                    }
                    return parentId.equals(dep.getParentId());
                })
                .collect(Collectors.toList());
    }

    public boolean existsByNameAndParentId(String name, String parentId) {
        return departments.values().stream()
                .anyMatch(dep -> {
                    boolean nameMatch = name.equals(dep.getName());
                    boolean parentMatch;
                    if (parentId == null) {
                        parentMatch = dep.getParentId() == null;
                    } else {
                        parentMatch = parentId.equals(dep.getParentId());
                    }
                    return nameMatch && parentMatch;
                });
    }

    public List<Department> findAllDescendants(String departmentId) {
        List<Department> descendants = new ArrayList<>();
        collectDescendants(departmentId, descendants);
        return descendants;
    }

    private void collectDescendants(String parentId, List<Department> result) {
        List<Department> children = findByParentId(parentId);
        for (Department child : children) {
            result.add(child);
            collectDescendants(child.getId(), result);
        }
    }
}
