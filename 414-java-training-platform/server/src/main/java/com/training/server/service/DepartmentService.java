package com.training.server.service;

import com.training.common.dto.request.CreateDepartmentRequest;
import com.training.common.dto.response.DepartmentDTO;
import com.training.common.enums.ErrorCode;
import com.training.server.entity.Department;
import com.training.server.repository.DepartmentRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Service
public class DepartmentService {

    @Autowired
    private DepartmentRepository departmentRepository;

    public DepartmentDTO createDepartment(CreateDepartmentRequest request) {
        Department department = new Department();
        department.setId(UUID.randomUUID().toString());
        department.setName(request.getName());
        department.setManagerId(request.getManagerId());
        
        Department saved = departmentRepository.save(department);
        return toDTO(saved);
    }

    public DepartmentDTO getDepartmentById(String id) {
        Optional<Department> departmentOpt = departmentRepository.findById(id);
        if (departmentOpt.isEmpty()) {
            throw new RuntimeException(ErrorCode.DEPARTMENT_NOT_FOUND.getMessage());
        }
        return toDTO(departmentOpt.get());
    }

    public List<DepartmentDTO> getAllDepartments() {
        List<Department> departments = departmentRepository.findAll();
        List<DepartmentDTO> dtoList = new ArrayList<>();
        for (Department department : departments) {
            dtoList.add(toDTO(department));
        }
        return dtoList;
    }

    public DepartmentDTO updateDepartment(String id, CreateDepartmentRequest request) {
        Optional<Department> departmentOpt = departmentRepository.findById(id);
        if (departmentOpt.isEmpty()) {
            throw new RuntimeException(ErrorCode.DEPARTMENT_NOT_FOUND.getMessage());
        }
        
        Department department = departmentOpt.get();
        if (request.getName() != null) {
            department.setName(request.getName());
        }
        if (request.getManagerId() != null) {
            department.setManagerId(request.getManagerId());
        }
        
        Department saved = departmentRepository.save(department);
        return toDTO(saved);
    }

    public void deleteDepartment(String id) {
        if (!departmentRepository.existsById(id)) {
            throw new RuntimeException(ErrorCode.DEPARTMENT_NOT_FOUND.getMessage());
        }
        departmentRepository.deleteById(id);
    }

    private DepartmentDTO toDTO(Department department) {
        DepartmentDTO dto = new DepartmentDTO();
        dto.setId(department.getId());
        dto.setName(department.getName());
        dto.setManagerId(department.getManagerId());
        dto.setCreatedAt(department.getCreatedAt());
        return dto;
    }
}
