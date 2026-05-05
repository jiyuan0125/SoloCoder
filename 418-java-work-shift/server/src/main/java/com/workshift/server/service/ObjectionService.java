package com.workshift.server.service;

import com.workshift.common.dto.ObjectionDTO;
import com.workshift.common.enums.ErrorCode;
import com.workshift.common.enums.ObjectionStatus;
import com.workshift.server.entity.ObjectionEntity;
import com.workshift.server.entity.ShiftEntity;
import com.workshift.server.repository.DataStore;
import java.time.Duration;
import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;
import org.springframework.stereotype.Service;

@Service
public class ObjectionService {

    private static final long OBJECTION_WINDOW_HOURS = 24;

    private final DataStore dataStore;

    public ObjectionService(DataStore dataStore) {
        this.dataStore = dataStore;
    }

    public Optional<ErrorCode> createObjection(String shiftId, String employeeId, String reason) {
        ShiftEntity shift = dataStore.getShift(shiftId);
        if (shift == null) {
            return Optional.of(ErrorCode.SHIFT_NOT_FOUND);
        }

        if (!shift.isPublished()) {
            return Optional.of(ErrorCode.SHIFT_NOT_PUBLISHED);
        }

        if (!shift.getEmployeeId().equals(employeeId)) {
            return Optional.of(ErrorCode.EMPLOYEE_NOT_FOUND);
        }

        if (shift.getPublishTime() != null) {
            Duration elapsed = Duration.between(shift.getPublishTime(), LocalDateTime.now());
            if (elapsed.toHours() > OBJECTION_WINDOW_HOURS) {
                return Optional.of(ErrorCode.OBJECTION_TIME_EXPIRED);
            }
        }

        ObjectionEntity existing = dataStore.getObjectionByShiftAndEmployee(shiftId, employeeId);
        if (existing != null) {
            return Optional.of(ErrorCode.OBJECTION_ALREADY_SUBMITTED);
        }

        ObjectionEntity objection = new ObjectionEntity();
        objection.setId(dataStore.generateId());
        objection.setShiftId(shiftId);
        objection.setEmployeeId(employeeId);
        objection.setReason(reason);
        objection.setStatus(ObjectionStatus.PENDING);

        dataStore.saveObjection(objection);
        return Optional.empty();
    }

    public Optional<ErrorCode> handleObjection(String objectionId, ObjectionStatus newStatus, String handlerNote) {
        ObjectionEntity objection = dataStore.getObjection(objectionId);
        if (objection == null) {
            return Optional.of(ErrorCode.OBJECTION_NOT_FOUND);
        }

        if (objection.getStatus() != ObjectionStatus.PENDING) {
            return Optional.of(ErrorCode.OBJECTION_NOT_FOUND);
        }

        objection.setStatus(newStatus);
        objection.setHandlerNote(handlerNote);
        objection.setHandleTime(LocalDateTime.now());

        dataStore.saveObjection(objection);
        return Optional.empty();
    }

    public List<ObjectionDTO> getAllObjections() {
        List<ObjectionEntity> entities = dataStore.getAllObjections();
        return entities.stream().map(this::toDTO).collect(Collectors.toList());
    }

    public List<ObjectionDTO> getObjectionsByEmployee(String employeeId) {
        List<ObjectionEntity> entities = dataStore.getAllObjections().stream()
                .filter(o -> employeeId.equals(o.getEmployeeId()))
                .collect(Collectors.toList());
        return entities.stream().map(this::toDTO).collect(Collectors.toList());
    }

    public ObjectionDTO getObjectionById(String id) {
        ObjectionEntity entity = dataStore.getObjection(id);
        return entity != null ? toDTO(entity) : null;
    }

    public int getObjectionCountByEmployeeAndMonth(String employeeId, int year, int month) {
        return (int) dataStore.getAllObjections().stream()
                .filter(o -> employeeId.equals(o.getEmployeeId()))
                .filter(o -> o.getCreateTime().getYear() == year)
                .filter(o -> o.getCreateTime().getMonthValue() == month)
                .count();
    }

    private ObjectionDTO toDTO(ObjectionEntity entity) {
        ObjectionDTO dto = new ObjectionDTO();
        dto.setId(entity.getId());
        dto.setShiftId(entity.getShiftId());
        dto.setEmployeeId(entity.getEmployeeId());
        dto.setReason(entity.getReason());
        dto.setStatus(entity.getStatus());
        dto.setHandlerNote(entity.getHandlerNote());
        dto.setCreateTime(entity.getCreateTime());
        dto.setHandleTime(entity.getHandleTime());
        return dto;
    }
}