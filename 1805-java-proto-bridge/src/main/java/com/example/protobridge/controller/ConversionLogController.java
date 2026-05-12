package com.example.protobridge.controller;

import com.example.protobridge.log.ConversionLog;
import com.example.protobridge.log.ConversionLogService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/admin/logs")
@RequiredArgsConstructor
public class ConversionLogController {

    private final ConversionLogService logService;

    @GetMapping
    public ResponseEntity<List<ConversionLog>> getLogs(
            @RequestParam(required = false) String direction,
            @RequestParam(required = false) Integer limit) {
        if (direction != null && !direction.isEmpty()) {
            return ResponseEntity.ok(logService.getLogsByDirection(direction, limit));
        }
        return ResponseEntity.ok(logService.getLogs(limit));
    }

    @GetMapping("/count")
    public ResponseEntity<Map<String, Object>> getLogCount() {
        Map<String, Object> result = new HashMap<>();
        result.put("count", logService.getLogCount());
        result.put("max", 1000);
        return ResponseEntity.ok(result);
    }

    @DeleteMapping
    public ResponseEntity<Map<String, Object>> clearLogs() {
        Map<String, Object> result = new HashMap<>();
        result.put("success", true);
        result.put("message", "日志已清空");
        logService.clearLogs();
        return ResponseEntity.ok(result);
    }
}
