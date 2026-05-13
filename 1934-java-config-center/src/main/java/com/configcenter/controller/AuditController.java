package com.configcenter.controller;

import com.configcenter.exception.ForbiddenException;
import com.configcenter.model.AuditLog;
import com.configcenter.service.AuditService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.time.Instant;
import java.util.List;

@RestController
@RequestMapping("/audit/logs")
@RequiredArgsConstructor
public class AuditController {

    private final AuditService auditService;

    private String getUserId(String userId) {
        if (userId == null || userId.isEmpty()) {
            throw new IllegalArgumentException("X-User-Id header is required");
        }
        return userId;
    }

    @GetMapping
    public ResponseEntity<List<AuditLog>> queryLogs(
            @RequestParam(required = false) String userId,
            @RequestParam(required = false) String appName,
            @RequestParam(required = false) String startTime,
            @RequestParam(required = false) String endTime,
            @RequestHeader(value = "X-User-Id", required = false) String callerUserId) {
        
        String actualCallerUserId = getUserId(callerUserId);
        
        Instant startInstant = null;
        Instant endInstant = null;
        
        if (startTime != null && !startTime.isEmpty()) {
            startInstant = Instant.parse(startTime);
        }
        if (endTime != null && !endTime.isEmpty()) {
            endInstant = Instant.parse(endTime);
        }
        
        List<AuditLog> logs = auditService.queryLogs(userId, appName, startInstant, endInstant);
        
        return ResponseEntity.ok(logs);
    }
}
