package com.company.vehicledispatch.service;

import com.company.vehicledispatch.entity.Department;
import com.company.vehicledispatch.exception.BusinessException;
import com.company.vehicledispatch.repository.DepartmentRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Optional;

@Service
public class DepartmentService {

    @Autowired
    private DepartmentRepository departmentRepository;

    public List<Department> getAllDepartments() {
        return departmentRepository.findAll();
    }

    public Optional<Department> getDepartmentById(Long id) {
        return departmentRepository.findById(id);
    }

    public Optional<Department> getDepartmentByName(String name) {
        return departmentRepository.findByName(name);
    }

    @Transactional
    public Department createDepartment(Department department) {
        if (departmentRepository.findByName(department.getName()).isPresent()) {
            throw new BusinessException("部门名称已存在: " + department.getName());
        }
        return departmentRepository.save(department);
    }

    @Transactional
    public Department updateDepartment(Long id, Department departmentDetails) {
        Department department = departmentRepository.findById(id)
                .orElseThrow(() -> new BusinessException("部门不存在: " + id));

        if (!department.getName().equals(departmentDetails.getName()) &&
                departmentRepository.findByName(departmentDetails.getName()).isPresent()) {
            throw new BusinessException("部门名称已存在: " + departmentDetails.getName());
        }

        department.setName(departmentDetails.getName());
        department.setDescription(departmentDetails.getDescription());

        return departmentRepository.save(department);
    }

    @Transactional
    public void deleteDepartment(Long id) {
        if (!departmentRepository.existsById(id)) {
            throw new BusinessException("部门不存在: " + id);
        }
        departmentRepository.deleteById(id);
    }
}
