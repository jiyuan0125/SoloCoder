package com.workshift.server.service;

import com.workshift.common.dto.EmployeeDTO;
import com.workshift.common.enums.ErrorCode;
import com.workshift.server.entity.EmployeeEntity;
import com.workshift.server.repository.DataStore;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;
import org.springframework.stereotype.Service;

@Service
public class EmployeeService {

    private final DataStore dataStore;

    public EmployeeService(DataStore dataStore) {
        this.dataStore = dataStore;
    }

    public EmployeeDTO createEmployee(EmployeeDTO dto) {
        EmployeeEntity entity = new EmployeeEntity();
        entity.setId(dataStore.generateId());
        entity.setName(dto.getName());
        entity.setDepartment(dto.getDepartment());
        entity.setHourlyWage(dto.getHourlyWage());

        dataStore.saveEmployee(entity);
        return toDTO(entity);
    }

    public Optional<ErrorCode> updateEmployee(String id, EmployeeDTO dto) {
        EmployeeEntity entity = dataStore.getEmployee(id);
        if (entity == null) {
            return Optional.of(ErrorCode.EMPLOYEE_NOT_FOUND);
        }

        if (dto.getName() != null) {
            entity.setName(dto.getName());
        }
        if (dto.getDepartment() != null) {
            entity.setDepartment(dto.getDepartment());
        }
        if (dto.getHourlyWage() != null) {
            entity.setHourlyWage(dto.getHourlyWage());
        }

        dataStore.saveEmployee(entity);
        return Optional.empty();
    }

    public Optional<ErrorCode> deleteEmployee(String id) {
        if (dataStore.getEmployee(id) == null) {
            return Optional.of(ErrorCode.EMPLOYEE_NOT_FOUND);
        }
        dataStore.removeEmployee(id);
        return Optional.empty();
    }

    public EmployeeDTO getEmployee(String id) {
        EmployeeEntity entity = dataStore.getEmployee(id);
        return entity != null ? toDTO(entity) : null;
    }

    public List<EmployeeDTO> getAllEmployees() {
        List<EmployeeEntity> entities = dataStore.getAllEmployees();
        return entities.stream().map(this::toDTO).collect(Collectors.toList());
    }

    private EmployeeDTO toDTO(EmployeeEntity entity) {
        EmployeeDTO dto = new EmployeeDTO();
        dto.setId(entity.getId());
        dto.setName(entity.getName());
        dto.setDepartment(entity.getDepartment());
        dto.setHourlyWage(entity.getHourlyWage());
        return dto;
    }
}