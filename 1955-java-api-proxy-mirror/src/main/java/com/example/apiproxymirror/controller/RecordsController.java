package com.example.apiproxymirror.controller;

import com.example.apiproxymirror.model.RecordedExchange;
import com.example.apiproxymirror.service.RecordingService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpHeaders;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;

@RestController
@RequestMapping("/records")
@RequiredArgsConstructor
public class RecordsController {

    private final RecordingService recordingService;

    @GetMapping
    public ResponseEntity<List<RecordedExchange>> listRecords(
            @RequestParam(required = false) String path,
            @RequestParam(required = false) Integer status) {
        List<RecordedExchange> records = recordingService.list(path, status);
        return ResponseEntity.ok(records);
    }

    @GetMapping("/{id}")
    public ResponseEntity<?> getRecordById(@PathVariable String id) {
        Optional<RecordedExchange> recordOpt = recordingService.getById(id);
        if (recordOpt.isPresent()) {
            return ResponseEntity.ok(recordOpt.get());
        }

        Map<String, String> error = new HashMap<>();
        error.put("error", "Record not found");
        error.put("id", id);
        return ResponseEntity.status(404).body(error);
    }

    @PostMapping("/clear")
    public ResponseEntity<Map<String, Object>> clearAllRecords() {
        recordingService.clearAllRecords();
        Map<String, Object> response = new HashMap<>();
        response.put("status", "cleared");
        response.put("recordCount", recordingService.getRecordCount());
        return ResponseEntity.ok(response);
    }
}
