package com.workshift.server.service;

import com.workshift.common.dto.ShiftDTO;
import com.workshift.common.enums.ErrorCode;
import com.workshift.common.enums.ShiftType;
import com.workshift.server.entity.ShiftEntity;
import com.workshift.server.repository.DataStore;
import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;
import org.springframework.stereotype.Service;

@Service
public class ShiftService {

    private final DataStore dataStore;
    private final ShiftRuleValidator ruleValidator;

    public ShiftService(DataStore dataStore, ShiftRuleValidator ruleValidator) {
        this.dataStore = dataStore;
        this.ruleValidator = ruleValidator;
    }

    public Optional<ErrorCode> createShift(String employeeId, LocalDate date, ShiftType shiftType) {
        if (dataStore.getEmployee(employeeId) == null) {
            return Optional.of(ErrorCode.EMPLOYEE_NOT_FOUND);
        }

        List<ShiftEntity> allShifts = dataStore.getAllShifts();
        Optional<ErrorCode> validationResult = ruleValidator.validateNewShift(
                employeeId, date, shiftType, allShifts);
        
        if (validationResult.isPresent()) {
            return validationResult;
        }

        boolean isHoliday = dataStore.isHoliday(date);
        double workHours = shiftType.getEndHour() - shiftType.getStartHour();
        if (shiftType.getEndHour() == 24 && shiftType.getStartHour() == 0) {
            workHours = 8;
        }

        ShiftEntity shift = new ShiftEntity();
        shift.setId(dataStore.generateId());
        shift.setEmployeeId(employeeId);
        shift.setDate(date);
        shift.setShiftType(shiftType);
        shift.setPublished(false);
        shift.setHoliday(isHoliday);
        shift.setWorkHours(workHours);

        dataStore.saveShift(shift);
        return Optional.empty();
    }

    public Optional<ErrorCode> publishShift(String shiftId) {
        ShiftEntity shift = dataStore.getShift(shiftId);
        if (shift == null) {
            return Optional.of(ErrorCode.SHIFT_NOT_FOUND);
        }

        shift.setPublished(true);
        shift.setPublishTime(LocalDateTime.now());
        shift.setUpdatedAt(LocalDateTime.now());
        dataStore.saveShift(shift);

        return Optional.empty();
    }

    public Optional<ErrorCode> publishAllShifts(List<String> shiftIds) {
        for (String shiftId : shiftIds) {
            Optional<ErrorCode> result = publishShift(shiftId);
            if (result.isPresent()) {
                return result;
            }
        }
        return Optional.empty();
    }

    public List<ShiftDTO> getShiftsByEmployee(String employeeId) {
        List<ShiftEntity> entities = dataStore.getShiftsByEmployee(employeeId);
        return entities.stream().map(this::toDTO).collect(Collectors.toList());
    }

    public List<ShiftDTO> getShiftsByEmployeeAndDateRange(String employeeId, LocalDate startDate, LocalDate endDate) {
        List<ShiftEntity> entities = dataStore.getShiftsByEmployeeAndDateRange(employeeId, startDate, endDate);
        return entities.stream().map(this::toDTO).collect(Collectors.toList());
    }

    public List<ShiftDTO> getShiftsByDateRange(LocalDate startDate, LocalDate endDate) {
        List<ShiftEntity> entities = dataStore.getShiftsByDateRange(startDate, endDate);
        return entities.stream().map(this::toDTO).collect(Collectors.toList());
    }

    public ShiftDTO getShiftById(String id) {
        ShiftEntity entity = dataStore.getShift(id);
        return entity != null ? toDTO(entity) : null;
    }

    public Optional<ErrorCode> deleteShift(String shiftId) {
        ShiftEntity shift = dataStore.getShift(shiftId);
        if (shift == null) {
            return Optional.of(ErrorCode.SHIFT_NOT_FOUND);
        }

        if (shift.isPublished()) {
            return Optional.of(ErrorCode.SHIFT_NOT_FOUND);
        }

        dataStore.removeShift(shiftId);
        return Optional.empty();
    }

    public void updateShiftEntity(ShiftEntity shift) {
        shift.setUpdatedAt(LocalDateTime.now());
        dataStore.saveShift(shift);
    }

    public ShiftEntity getShiftEntityByEmployeeAndDate(String employeeId, LocalDate date) {
        return dataStore.getShiftByEmployeeAndDate(employeeId, date);
    }

    public ShiftDTO getShiftByEmployeeAndDate(String employeeId, LocalDate date) {
        ShiftEntity entity = dataStore.getShiftByEmployeeAndDate(employeeId, date);
        return entity != null ? toDTO(entity) : null;
    }

    public List<ShiftEntity> getAllShiftEntities() {
        return dataStore.getAllShifts();
    }

    private ShiftDTO toDTO(ShiftEntity entity) {
        ShiftDTO dto = new ShiftDTO();
        dto.setId(entity.getId());
        dto.setEmployeeId(entity.getEmployeeId());
        dto.setDate(entity.getDate());
        dto.setShiftType(entity.getShiftType());
        dto.setPublished(entity.isPublished());
        dto.setPublishTime(entity.getPublishTime());
        dto.setHoliday(entity.isHoliday());
        dto.setWorkHours(entity.getWorkHours());
        return dto;
    }
}