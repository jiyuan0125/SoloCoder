package com.gateway.auth.controller;

import com.gateway.auth.model.AuditLog;
import com.gateway.auth.service.AuditLogService;
import com.gateway.auth.service.JwtService;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.time.Instant;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/admin")
public class AdminController {

    private final AuditLogService auditLogService;
    private final JwtService jwtService;

    public AdminController(AuditLogService auditLogService, JwtService jwtService) {
        this.auditLogService = auditLogService;
        this.jwtService = jwtService;
    }

    @GetMapping("/logs")
    public List<AuditLog> getLogs(
            @RequestParam(required = false) String clientId,
            @RequestParam(required = false) String startTime,
            @RequestParam(required = false) String endTime) {

        Instant start = parseInstant(startTime);
        Instant end = parseInstant(endTime);

        return auditLogService.getLogs(clientId, start, end);
    }

    @GetMapping("/key-status")
    public Map<String, Object> getKeyStatus() {
        Map<String, Object> status = new HashMap<>();

        status.put("primaryKeyPresent", jwtService.getPrimaryKey() != null);
        status.put("oldKeyPresent", jwtService.getOldKey() != null);
        status.put("oldKeyValid", jwtService.isOldKeyValid());
        status.put("oldKeyExpireAt", jwtService.getOldKeyExpireTime());

        return status;
    }

    private Instant parseInstant(String value) {
        if (value == null || value.trim().isEmpty()) {
            return null;
        }
        try {
            return Instant.parse(value);
        } catch (Exception e) {
            return null;
        }
    }
}
