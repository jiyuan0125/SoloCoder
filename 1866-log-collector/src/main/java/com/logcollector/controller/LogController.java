package com.logcollector.controller;

import com.logcollector.model.LogEntry;
import com.logcollector.model.LogQueryParam;
import com.logcollector.model.RawLogEntry;
import com.logcollector.service.LogStorageService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
@RequestMapping("/logs")
public class LogController {

    private final LogStorageService logStorageService;

    public LogController(LogStorageService logStorageService) {
        this.logStorageService = logStorageService;
    }

    @PostMapping
    public ResponseEntity<LogEntry> ingest(@RequestBody RawLogEntry raw) {
        LogEntry saved = logStorageService.save(raw);
        return ResponseEntity.ok(saved);
    }

    @PostMapping("/batch")
    public ResponseEntity<List<LogEntry>> ingestBatch(@RequestBody List<RawLogEntry> raws) {
        List<LogEntry> saved = raws.stream()
                .map(logStorageService::save)
                .toList();
        return ResponseEntity.ok(saved);
    }

    @GetMapping
    public ResponseEntity<List<LogEntry>> query(
            @RequestParam(required = false) String service,
            @RequestParam(required = false) List<String> level,
            @RequestParam(required = false) Long start_time,
            @RequestParam(required = false) Long end_time,
            @RequestParam(defaultValue = "0") int page,
            @RequestParam(defaultValue = "20") int page_size) {

        LogQueryParam param = new LogQueryParam();
        param.setService(service);
        param.setLevels(level);
        param.setStartTime(start_time);
        param.setEndTime(end_time);
        param.setPage(page);
        param.setPageSize(page_size);

        List<LogEntry> result = logStorageService.query(param);
        return ResponseEntity.ok(result);
    }

    @GetMapping("/trace/{requestId}")
    public ResponseEntity<List<LogEntry>> getTrace(@PathVariable String requestId) {
        List<LogEntry> trace = logStorageService.getTrace(requestId);
        return ResponseEntity.ok(trace);
    }
}
