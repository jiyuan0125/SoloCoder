package com.example.server.controller;

import com.example.common.dto.AlertDTO;
import com.example.common.enums.ErrorCode;
import com.example.common.response.ApiResponse;
import com.example.server.service.AlertService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import java.util.List;
import java.util.Optional;

@RestController
@RequestMapping("/api/alerts")
public class AlertController {

    @Autowired
    private AlertService alertService;

    @GetMapping
    public ApiResponse<List<AlertDTO>> getAllAlerts() {
        List<AlertDTO> alerts = alertService.getAllAlerts();
        return ApiResponse.success(alerts);
    }

    @GetMapping("/unread")
    public ApiResponse<List<AlertDTO>> getUnreadAlerts() {
        List<AlertDTO> alerts = alertService.getUnreadAlerts();
        return ApiResponse.success(alerts);
    }

    @GetMapping("/unread/count")
    public ApiResponse<Long> getUnreadAlertCount() {
        long count = alertService.getUnreadAlertCount();
        return ApiResponse.success(count);
    }

    @GetMapping("/{id}")
    public ApiResponse<AlertDTO> getAlertById(@PathVariable String id) {
        Optional<AlertDTO> alert = alertService.getAlertById(id);
        if (alert.isPresent()) {
            return ApiResponse.success(alert.get());
        }
        return ApiResponse.error(ErrorCode.ALERT_NOT_FOUND);
    }

    @PutMapping("/{id}/read")
    public ApiResponse<Void> markAsRead(@PathVariable String id) {
        boolean marked = alertService.markAsRead(id);
        if (marked) {
            return ApiResponse.success();
        }
        return ApiResponse.error(ErrorCode.ALERT_NOT_FOUND);
    }

    @DeleteMapping("/{id}")
    public ApiResponse<Void> deleteAlert(@PathVariable String id) {
        boolean deleted = alertService.deleteAlert(id);
        if (deleted) {
            return ApiResponse.success();
        }
        return ApiResponse.error(ErrorCode.ALERT_NOT_FOUND);
    }
}
