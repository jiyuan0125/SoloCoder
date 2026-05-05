package com.performance.server.service;

import com.performance.common.dto.DepartmentDTO;
import com.performance.server.entity.Department;
import com.performance.server.repository.DepartmentRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

@Service
public class DepartmentService {

    @Autowired
    private DepartmentRepository departmentRepository;

    public DepartmentDTO createDepartment(DepartmentDTO dto) {
        Department department = new Department();
        department.setName(dto.getName());
        department.setCode(dto.getCode());
        department.setManagerId(dto.getManagerId());
        
        Department saved = departmentRepository.save(department);
        return toDTO(saved);
    }

    public Optional<DepartmentDTO> getDepartmentById(Long id) {
        Optional<Department> deptOpt = departmentRepository.findById(id);
        if (deptOpt.isPresent()) {
            return Optional.of(toDTO(deptOpt.get()));
        }
        return Optional.empty();
    }

    public List<DepartmentDTO> getAllDepartments() {
        List<Department> departments = departmentRepository.findAll();
        List<DepartmentDTO> dtos = new ArrayList<>();
        for (Department department : departments) {
            dtos.add(toDTO(department));
        }
        return dtos;
    }

    public Optional<DepartmentDTO> updateDepartmentManager(Long departmentId, Long managerId) {
        Optional<Department> deptOpt = departmentRepository.findById(departmentId);
        if (deptOpt.isPresent()) {
            Department department = deptOpt.get();
            department.setManagerId(managerId);
            Department saved = departmentRepository.save(department);
            return Optional.of(toDTO(saved));
        }
        return Optional.empty();
    }

    private DepartmentDTO toDTO(Department department) {
        DepartmentDTO dto = new DepartmentDTO();
        dto.setId(department.getId());
        dto.setName(department.getName());
        dto.setCode(department.getCode());
        dto.setManagerId(department.getManagerId());
        return dto;
    }
}