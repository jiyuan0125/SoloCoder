package com.orgchart.server.service;

import com.orgchart.common.dto.OperationLogDTO;
import com.orgchart.server.entity.OperationLog;
import com.orgchart.server.repository.OperationLogRepository;
import com.orgchart.server.util.IdGenerator;
import org.springframework.beans.BeanUtils;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.List;
import java.util.stream.Collectors;

@Service
public class OperationLogService {

    private final OperationLogRepository operationLogRepository;

    public OperationLogService(OperationLogRepository operationLogRepository) {
        this.operationLogRepository = operationLogRepository;
    }

    public void logCreate(String targetType, String targetId, String targetName, String newValue) {
        createLog("CREATE", targetType, targetId, targetName, null, newValue);
    }

    public void logUpdate(String targetType, String targetId, String targetName, String oldValue, String newValue) {
        createLog("UPDATE", targetType, targetId, targetName, oldValue, newValue);
    }

    public void logDelete(String targetType, String targetId, String targetName, String oldValue) {
        createLog("DELETE", targetType, targetId, targetName, oldValue, null);
    }

    public void logTransfer(String targetType, String targetId, String targetName, String oldValue, String newValue) {
        createLog("TRANSFER", targetType, targetId, targetName, oldValue, newValue);
    }

    public void logMerge(String targetType, String targetId, String targetName, String oldValue, String newValue) {
        createLog("MERGE", targetType, targetId, targetName, oldValue, newValue);
    }

    private void createLog(String operationType, String targetType, String targetId, 
                           String targetName, String oldValue, String newValue) {
        OperationLog log = new OperationLog();
        log.setId(IdGenerator.generateOperationLogId());
        log.setOperationType(operationType);
        log.setTargetType(targetType);
        log.setTargetId(targetId);
        log.setTargetName(targetName);
        log.setOldValue(oldValue);
        log.setNewValue(newValue);
        log.setOperator("SYSTEM");
        log.setOperationTime(LocalDateTime.now());
        operationLogRepository.save(log);
    }

    public List<OperationLogDTO> getAllLogs() {
        return operationLogRepository.findAll().stream()
                .map(this::toDTO)
                .collect(Collectors.toList());
    }

    public List<OperationLogDTO> getLogsByTarget(String targetType, String targetId) {
        return operationLogRepository.findByTargetTypeAndTargetId(targetType, targetId).stream()
                .map(this::toDTO)
                .collect(Collectors.toList());
    }

    private OperationLogDTO toDTO(OperationLog entity) {
        OperationLogDTO dto = new OperationLogDTO();
        BeanUtils.copyProperties(entity, dto);
        return dto;
    }
}
