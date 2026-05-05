package com.workshift.server.controller;

import com.workshift.common.dto.ShiftDTO;
import com.workshift.common.enums.ErrorCode;
import com.workshift.common.request.CreateShiftRequest;
import com.workshift.common.response.ApiResponse;
import com.workshift.server.service.ShiftService;
import java.time.LocalDate;
import java.util.List;
import java.util.Optional;
import org.springframework.format.annotation.DateTimeFormat;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api/shifts")
public class ShiftController {

    private final ShiftService shiftService;

    public ShiftController(ShiftService shiftService) {
        this.shiftService = shiftService;
    }

    @PostMapping
    public ApiResponse<ShiftDTO> createShift(@RequestBody CreateShiftRequest request) {
        if (request.getEmployeeId() == null || request.getDate() == null || request.getShiftType() == null) {
            return ApiResponse.error(ErrorCode.INVALID_PARAMETER.getCode(), "参数不完整");
        }

        Optional<ErrorCode> result = shiftService.createShift(
                request.getEmployeeId(), request.getDate(), request.getShiftType());
        
        if (result.isPresent()) {
            return ApiResponse.error(result.get().getCode(), result.get().getMessage());
        }

        ShiftDTO shift = shiftService.getShiftByEmployeeAndDate(
                request.getEmployeeId(), request.getDate());
        return ApiResponse.success(shift);
    }

    @PostMapping("/{id}/publish")
    public ApiResponse<Void> publishShift(@PathVariable String id) {
        Optional<ErrorCode> result = shiftService.publishShift(id);
        if (result.isPresent()) {
            return ApiResponse.error(result.get().getCode(), result.get().getMessage());
        }
        return ApiResponse.success();
    }

    @PostMapping("/publish-batch")
    public ApiResponse<Void> publishShifts(@RequestBody List<String> shiftIds) {
        Optional<ErrorCode> result = shiftService.publishAllShifts(shiftIds);
        if (result.isPresent()) {
            return ApiResponse.error(result.get().getCode(), result.get().getMessage());
        }
        return ApiResponse.success();
    }

    @GetMapping("/{id}")
    public ApiResponse<ShiftDTO> getShift(@PathVariable String id) {
        ShiftDTO shift = shiftService.getShiftById(id);
        if (shift == null) {
            return ApiResponse.error(ErrorCode.SHIFT_NOT_FOUND.getCode(), ErrorCode.SHIFT_NOT_FOUND.getMessage());
        }
        return ApiResponse.success(shift);
    }

    @GetMapping("/employee/{employeeId}")
    public ApiResponse<List<ShiftDTO>> getShiftsByEmployee(@PathVariable String employeeId) {
        List<ShiftDTO> shifts = shiftService.getShiftsByEmployee(employeeId);
        return ApiResponse.success(shifts);
    }

    @GetMapping("/employee/{employeeId}/range")
    public ApiResponse<List<ShiftDTO>> getShiftsByEmployeeAndDateRange(
            @PathVariable String employeeId,
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate startDate,
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate endDate) {
        
        if (startDate.isAfter(endDate)) {
            return ApiResponse.error(ErrorCode.INVALID_DATE_RANGE.getCode(), ErrorCode.INVALID_DATE_RANGE.getMessage());
        }
        List<ShiftDTO> shifts = shiftService.getShiftsByEmployeeAndDateRange(employeeId, startDate, endDate);
        return ApiResponse.success(shifts);
    }

    @GetMapping("/range")
    public ApiResponse<List<ShiftDTO>> getShiftsByDateRange(
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate startDate,
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate endDate) {
        
        if (startDate.isAfter(endDate)) {
            return ApiResponse.error(ErrorCode.INVALID_DATE_RANGE.getCode(), ErrorCode.INVALID_DATE_RANGE.getMessage());
        }
        List<ShiftDTO> shifts = shiftService.getShiftsByDateRange(startDate, endDate);
        return ApiResponse.success(shifts);
    }

    @DeleteMapping("/{id}")
    public ApiResponse<Void> deleteShift(@PathVariable String id) {
        Optional<ErrorCode> result = shiftService.deleteShift(id);
        if (result.isPresent()) {
            return ApiResponse.error(result.get().getCode(), result.get().getMessage());
        }
        return ApiResponse.success();
    }
}