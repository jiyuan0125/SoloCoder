package com.logaggregate.controller;

import com.logaggregate.model.LogEntry;
import com.logaggregate.service.ingest.IngestService;
import jakarta.validation.Valid;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/logs")
public class LogIngestController {

    private final IngestService ingestService;

    public LogIngestController(IngestService ingestService) {
        this.ingestService = ingestService;
    }

    @PostMapping
    public ResponseEntity<Map<String, Object>> ingestSingle(@RequestBody @Valid LogEntry entry) {
        IngestService.IngestResult result = ingestService.ingestSingle(entry);
        return buildResponse(result);
    }

    @PostMapping("/batch")
    public ResponseEntity<Map<String, Object>> ingestBatch(@RequestBody List<@Valid LogEntry> entries) {
        if (entries == null) {
            entries = List.of();
        }
        IngestService.IngestResult result = ingestService.ingestBatch(entries);
        return buildResponse(result);
    }

    private ResponseEntity<Map<String, Object>> buildResponse(IngestService.IngestResult result) {
        if (result.getWarning() != null) {
            return ResponseEntity.status(HttpStatus.ACCEPTED).body(Map.of(
                    "status", "accepted",
                    "warning", result.getWarning(),
                    "received", result.getReceived(),
                    "ingested", result.getIngested(),
                    "truncated", result.isTruncated()
            ));
        }

        return ResponseEntity.status(HttpStatus.ACCEPTED).body(Map.of(
                "status", "accepted",
                "received", result.getReceived(),
                "ingested", result.getIngested()
        ));
    }
}
