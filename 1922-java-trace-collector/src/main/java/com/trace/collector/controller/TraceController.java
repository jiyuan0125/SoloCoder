package com.trace.collector.controller;

import com.trace.collector.model.Span;
import com.trace.collector.model.TraceTree;
import com.trace.collector.service.TraceService;
import jakarta.validation.Valid;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.format.annotation.DateTimeFormat;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.time.Instant;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping
public class TraceController {

    @Autowired
    private TraceService traceService;

    @PostMapping("/spans")
    public ResponseEntity<Map<String, Object>> createSpans(@RequestBody @Valid List<Span> spans) {
        if (spans.size() > 100) {
            Map<String, Object> error = new HashMap<>();
            error.put("error", "Maximum 100 spans allowed per batch");
            return ResponseEntity.badRequest().body(error);
        }

        traceService.saveSpans(spans);

        Map<String, Object> result = new HashMap<>();
        result.put("success", true);
        result.put("count", spans.size());

        return ResponseEntity.ok(result);
    }

    @GetMapping("/traces/{traceId}")
    public ResponseEntity<TraceTree> getTraceById(@PathVariable String traceId) {
        TraceTree traceTree = traceService.getTraceTree(traceId);
        if (traceTree == null) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(traceTree);
    }

    @GetMapping("/traces")
    public ResponseEntity<List<TraceTree>> searchTraces(
            @RequestParam(required = false) String service,
            @RequestParam(required = false) @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) Instant from,
            @RequestParam(required = false) @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) Instant to) {

        List<TraceTree> traces = traceService.searchTraces(service, from, to);
        return ResponseEntity.ok(traces);
    }

    @GetMapping("/services")
    public ResponseEntity<List<String>> getAllServices() {
        return ResponseEntity.ok(traceService.getAllServices());
    }

    @GetMapping("/services/{name}/operations")
    public ResponseEntity<List<String>> getOperationsByService(@PathVariable String name) {
        List<String> operations = traceService.getOperationsByService(name);
        return ResponseEntity.ok(operations);
    }
}
