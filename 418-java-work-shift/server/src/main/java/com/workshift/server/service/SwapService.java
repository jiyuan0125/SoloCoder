package com.workshift.server.service;

import com.workshift.common.dto.ShiftSwapDTO;
import com.workshift.common.enums.ErrorCode;
import com.workshift.common.enums.ShiftType;
import com.workshift.common.enums.SwapStatus;
import com.workshift.server.entity.ShiftEntity;
import com.workshift.server.entity.ShiftSwapEntity;
import com.workshift.server.repository.DataStore;
import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;
import org.springframework.stereotype.Service;

@Service
public class SwapService {

    private final DataStore dataStore;
    private final ShiftRuleValidator ruleValidator;
    private final ShiftService shiftService;

    public SwapService(DataStore dataStore, ShiftRuleValidator ruleValidator, ShiftService shiftService) {
        this.dataStore = dataStore;
        this.ruleValidator = ruleValidator;
        this.shiftService = shiftService;
    }

    public Optional<ErrorCode> createSwap(String requesterEmployeeId, String targetEmployeeId,
            LocalDate requesterDate, LocalDate targetDate, String requesterSignature) {
        
        if (requesterEmployeeId.equals(targetEmployeeId)) {
            return Optional.of(ErrorCode.CANNOT_SWAP_SELF);
        }

        if (dataStore.getEmployee(requesterEmployeeId) == null || 
            dataStore.getEmployee(targetEmployeeId) == null) {
            return Optional.of(ErrorCode.EMPLOYEE_NOT_FOUND);
        }

        ShiftEntity requesterShift = dataStore.getShiftByEmployeeAndDate(requesterEmployeeId, requesterDate);
        ShiftEntity targetShift = dataStore.getShiftByEmployeeAndDate(targetEmployeeId, targetDate);

        if (requesterShift == null || targetShift == null) {
            return Optional.of(ErrorCode.TARGET_SHIFT_NOT_FOUND);
        }

        if (!requesterShift.isPublished() || !targetShift.isPublished()) {
            return Optional.of(ErrorCode.SHIFT_NOT_PUBLISHED);
        }

        Optional<ErrorCode> validationResult = ruleValidator.validateSwapResult(
                requesterEmployeeId, requesterShift,
                targetEmployeeId, targetShift,
                dataStore.getAllShifts());

        if (validationResult.isPresent()) {
            return Optional.of(ErrorCode.SHIFT_SWAP_RULE_VIOLATION);
        }

        ShiftSwapEntity swap = new ShiftSwapEntity();
        swap.setId(dataStore.generateId());
        swap.setRequesterEmployeeId(requesterEmployeeId);
        swap.setTargetEmployeeId(targetEmployeeId);
        swap.setRequesterDate(requesterDate);
        swap.setRequesterShiftType(requesterShift.getShiftType());
        swap.setTargetDate(targetDate);
        swap.setTargetShiftType(targetShift.getShiftType());
        swap.setStatus(SwapStatus.PENDING);
        swap.setRequesterSignature(requesterSignature);
        swap.setRequesterSignatureTime(LocalDateTime.now());

        dataStore.saveSwap(swap);
        return Optional.empty();
    }

    public Optional<ErrorCode> confirmSwap(String swapId, String targetEmployeeId, 
            boolean confirmed, String targetSignature, String rejectReason) {
        
        ShiftSwapEntity swap = dataStore.getSwap(swapId);
        if (swap == null) {
            return Optional.of(ErrorCode.SWAP_NOT_FOUND);
        }

        if (swap.getStatus() != SwapStatus.PENDING) {
            return Optional.of(ErrorCode.SWAP_ALREADY_PROCESSED);
        }

        if (!swap.getTargetEmployeeId().equals(targetEmployeeId)) {
            return Optional.of(ErrorCode.EMPLOYEE_NOT_FOUND);
        }

        if (confirmed) {
            ShiftEntity requesterShift = dataStore.getShiftByEmployeeAndDate(
                    swap.getRequesterEmployeeId(), swap.getRequesterDate());
            ShiftEntity targetShift = dataStore.getShiftByEmployeeAndDate(
                    swap.getTargetEmployeeId(), swap.getTargetDate());

            if (requesterShift == null || targetShift == null) {
                return Optional.of(ErrorCode.TARGET_SHIFT_NOT_FOUND);
            }

            ShiftType requesterOriginalType = requesterShift.getShiftType();
            ShiftType targetOriginalType = targetShift.getShiftType();

            requesterShift.setShiftType(targetOriginalType);
            targetShift.setShiftType(requesterOriginalType);

            shiftService.updateShiftEntity(requesterShift);
            shiftService.updateShiftEntity(targetShift);

            swap.setStatus(SwapStatus.CONFIRMED);
            swap.setTargetSignature(targetSignature);
            swap.setTargetSignatureTime(LocalDateTime.now());
        } else {
            swap.setStatus(SwapStatus.REJECTED);
            swap.setRejectReason(rejectReason);
        }

        dataStore.saveSwap(swap);
        return Optional.empty();
    }

    public List<ShiftSwapDTO> getSwapsByEmployee(String employeeId) {
        List<ShiftSwapEntity> entities = dataStore.getSwapsByEmployee(employeeId);
        return entities.stream().map(this::toDTO).collect(Collectors.toList());
    }

    public List<ShiftSwapDTO> getAllSwaps() {
        List<ShiftSwapEntity> entities = dataStore.getAllSwaps();
        return entities.stream().map(this::toDTO).collect(Collectors.toList());
    }

    public ShiftSwapDTO getSwapById(String id) {
        ShiftSwapEntity entity = dataStore.getSwap(id);
        return entity != null ? toDTO(entity) : null;
    }

    public int getSwapCountByEmployeeAndMonth(String employeeId, int year, int month) {
        LocalDate startDate = LocalDate.of(year, month, 1);
        LocalDate endDate = startDate.plusMonths(1).minusDays(1);

        return (int) dataStore.getAllSwaps().stream()
                .filter(s -> s.getStatus() == SwapStatus.CONFIRMED)
                .filter(s -> s.getRequesterEmployeeId().equals(employeeId) || 
                              s.getTargetEmployeeId().equals(employeeId))
                .filter(s -> !s.getRequesterDate().isBefore(startDate) && 
                              !s.getRequesterDate().isAfter(endDate))
                .count();
    }

    private ShiftSwapDTO toDTO(ShiftSwapEntity entity) {
        ShiftSwapDTO dto = new ShiftSwapDTO();
        dto.setId(entity.getId());
        dto.setRequesterEmployeeId(entity.getRequesterEmployeeId());
        dto.setTargetEmployeeId(entity.getTargetEmployeeId());
        dto.setRequesterDate(entity.getRequesterDate());
        dto.setRequesterShiftType(entity.getRequesterShiftType());
        dto.setTargetDate(entity.getTargetDate());
        dto.setTargetShiftType(entity.getTargetShiftType());
        dto.setStatus(entity.getStatus());
        dto.setRequesterSignature(entity.getRequesterSignature());
        dto.setRequesterSignatureTime(entity.getRequesterSignatureTime());
        dto.setTargetSignature(entity.getTargetSignature());
        dto.setTargetSignatureTime(entity.getTargetSignatureTime());
        dto.setCreateTime(entity.getCreateTime());
        dto.setRejectReason(entity.getRejectReason());
        return dto;
    }
}