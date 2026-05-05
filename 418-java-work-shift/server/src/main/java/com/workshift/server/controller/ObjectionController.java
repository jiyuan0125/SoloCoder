package com.workshift.server.controller;

import com.workshift.common.dto.ObjectionDTO;
import com.workshift.common.enums.ErrorCode;
import com.workshift.common.enums.ObjectionStatus;
import com.workshift.common.request.CreateObjectionRequest;
import com.workshift.common.response.ApiResponse;
import com.workshift.server.service.ObjectionService;
import java.util.List;
import java.util.Optional;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api/objections")
public class ObjectionController {

    private final ObjectionService objectionService;

    public ObjectionController(ObjectionService objectionService) {
        this.objectionService = objectionService;
    }

    @PostMapping
    public ApiResponse<ObjectionDTO> createObjection(@RequestBody CreateObjectionRequest request) {
        if (request.getShiftId() == null || request.getEmployeeId() == null || request.getReason() == null) {
            return ApiResponse.error(ErrorCode.INVALID_PARAMETER.getCode(), "参数不完整");
        }

        Optional<ErrorCode> result = objectionService.createObjection(
                request.getShiftId(), request.getEmployeeId(), request.getReason());

        if (result.isPresent()) {
            return ApiResponse.error(result.get().getCode(), result.get().getMessage());
        }

        List<ObjectionDTO> objections = objectionService.getObjectionsByEmployee(request.getEmployeeId());
        ObjectionDTO latest = objections.get(objections.size() - 1);
        return ApiResponse.success(latest);
    }

    @PutMapping("/{id}/resolve")
    public ApiResponse<Void> resolveObjection(@PathVariable String id, @RequestParam String handlerNote) {
        Optional<ErrorCode> result = objectionService.handleObjection(
                id, ObjectionStatus.RESOLVED, handlerNote);
        if (result.isPresent()) {
            return ApiResponse.error(result.get().getCode(), result.get().getMessage());
        }
        return ApiResponse.success();
    }

    @PutMapping("/{id}/dismiss")
    public ApiResponse<Void> dismissObjection(@PathVariable String id, @RequestParam String handlerNote) {
        Optional<ErrorCode> result = objectionService.handleObjection(
                id, ObjectionStatus.DISMISSED, handlerNote);
        if (result.isPresent()) {
            return ApiResponse.error(result.get().getCode(), result.get().getMessage());
        }
        return ApiResponse.success();
    }

    @GetMapping("/{id}")
    public ApiResponse<ObjectionDTO> getObjection(@PathVariable String id) {
        ObjectionDTO objection = objectionService.getObjectionById(id);
        if (objection == null) {
            return ApiResponse.error(ErrorCode.OBJECTION_NOT_FOUND.getCode(), ErrorCode.OBJECTION_NOT_FOUND.getMessage());
        }
        return ApiResponse.success(objection);
    }

    @GetMapping("/employee/{employeeId}")
    public ApiResponse<List<ObjectionDTO>> getObjectionsByEmployee(@PathVariable String employeeId) {
        List<ObjectionDTO> objections = objectionService.getObjectionsByEmployee(employeeId);
        return ApiResponse.success(objections);
    }

    @GetMapping
    public ApiResponse<List<ObjectionDTO>> getAllObjections() {
        List<ObjectionDTO> objections = objectionService.getAllObjections();
        return ApiResponse.success(objections);
    }
}