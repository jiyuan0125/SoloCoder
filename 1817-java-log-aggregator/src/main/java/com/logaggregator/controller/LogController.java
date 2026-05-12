package com.logaggregator.controller;

import com.logaggregator.dto.*;
import com.logaggregator.model.LogLevel;
import com.logaggregator.service.LogService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.format.annotation.DateTimeFormat;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.time.LocalDateTime;
import java.util.List;

@RestController
@RequestMapping("/api/logs")
@RequiredArgsConstructor
public class LogController {

    private final LogService logService;

    @PostMapping("/batch")
    public ResponseEntity<BatchLogResponse> batchUpload(@Valid @RequestBody BatchLogRequest request) {
        BatchLogResponse response = logService.batchSave(request);
        return ResponseEntity.ok(response);
    }

    @GetMapping
    public ResponseEntity<PageResponse<LogResponse>> queryLogs(
            @RequestParam(required = false) String service,
            @RequestParam(required = false) String keyword,
            @RequestParam(required = false) List<LogLevel> levels,
            @RequestParam(required = false) @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) LocalDateTime startTime,
            @RequestParam(required = false) @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) LocalDateTime endTime,
            @RequestParam(defaultValue = "0") Integer page,
            @RequestParam(required = false) Integer size) {

        LogQueryRequest request = LogQueryRequest.builder()
                .service(service)
                .keyword(keyword)
                .levels(levels)
                .startTime(startTime)
                .endTime(endTime)
                .page(page)
                .size(size)
                .build();

        PageResponse<LogResponse> response = logService.queryLogs(request);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/trace/{traceId}")
    public ResponseEntity<List<LogResponse>> getTraceChain(@PathVariable String traceId) {
        List<LogResponse> response = logService.getTraceChain(traceId);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/stats")
    public ResponseEntity<AggregationStats> getAggregationStats() {
        AggregationStats response = logService.getAggregationStats();
        return ResponseEntity.ok(response);
    }
}
