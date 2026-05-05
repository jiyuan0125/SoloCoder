package com.orgchart.server.controller;

import com.orgchart.common.dto.ApiResponse;
import com.orgchart.common.dto.OperationLogDTO;
import com.orgchart.server.service.OperationLogService;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/operation-logs")
public class OperationLogController {

    private final OperationLogService operationLogService;

    public OperationLogController(OperationLogService operationLogService) {
        this.operationLogService = operationLogService;
    }

    @GetMapping
    public ApiResponse<List<OperationLogDTO>> getAllLogs() {
        List<OperationLogDTO> logs = operationLogService.getAllLogs();
        return ApiResponse.success(logs);
    }

    @GetMapping("/{targetType}/{targetId}")
    public ApiResponse<List<OperationLogDTO>> getLogsByTarget(
            @PathVariable String targetType,
            @PathVariable String targetId) {
        List<OperationLogDTO> logs = operationLogService.getLogsByTarget(targetType, targetId);
        return ApiResponse.success(logs);
    }
}
