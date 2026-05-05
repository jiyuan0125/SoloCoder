package com.employee.server.service;

import com.employee.common.dto.OperationLogDTO;
import com.employee.server.context.UserContext;
import com.employee.server.entity.OperationLog;
import com.employee.server.repository.OperationLogRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.stream.Collectors;

@Service
public class OperationLogService {
    private final OperationLogRepository operationLogRepository;

    @Autowired
    public OperationLogService(OperationLogRepository operationLogRepository) {
        this.operationLogRepository = operationLogRepository;
    }

    public void logChange(String targetEmployeeId, String targetEmployeeName,
                           String fieldName, String oldValue, String newValue) {
        if (oldValue == null && newValue == null) {
            return;
        }
        if (oldValue != null && oldValue.equals(newValue)) {
            return;
        }

        OperationLog log = new OperationLog();
        log.setOperatorId(UserContext.getCurrentUserId());
        log.setOperatorName(UserContext.getCurrentUserName());
        log.setTargetEmployeeId(targetEmployeeId);
        log.setTargetEmployeeName(targetEmployeeName);
        log.setFieldName(fieldName);
        log.setOldValue(oldValue);
        log.setNewValue(newValue);
        operationLogRepository.save(log);
    }

    public List<OperationLogDTO> getAllLogs() {
        return operationLogRepository.findAll().stream()
                .map(this::toDTO)
                .collect(Collectors.toList());
    }

    public List<OperationLogDTO> getLogsByEmployee(String employeeId) {
        return operationLogRepository.findByTargetEmployeeId(employeeId).stream()
                .map(this::toDTO)
                .collect(Collectors.toList());
    }

    private OperationLogDTO toDTO(OperationLog log) {
        OperationLogDTO dto = new OperationLogDTO();
        dto.setLogId(log.getLogId());
        dto.setOperatorId(log.getOperatorId());
        dto.setOperatorName(log.getOperatorName());
        dto.setTargetEmployeeId(log.getTargetEmployeeId());
        dto.setTargetEmployeeName(log.getTargetEmployeeName());
        dto.setFieldName(log.getFieldName());
        dto.setOldValue(log.getOldValue());
        dto.setNewValue(log.getNewValue());
        dto.setOperationTime(log.getOperationTime());
        return dto;
    }
}
