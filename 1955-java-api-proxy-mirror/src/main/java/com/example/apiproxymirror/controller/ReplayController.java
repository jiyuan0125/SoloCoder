package com.example.apiproxymirror.controller;

import com.example.apiproxymirror.model.RecordedExchange;
import com.example.apiproxymirror.service.RecordingService;
import jakarta.servlet.http.HttpServletRequest;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpHeaders;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.Map;
import java.util.Optional;

@RestController
@RequestMapping("/replay")
@RequiredArgsConstructor
public class ReplayController {

    private final RecordingService recordingService;

    @GetMapping("/{recordId}")
    public ResponseEntity<?> replayById(@PathVariable String recordId) {
        Optional<RecordedExchange> recordOpt = recordingService.getById(recordId);
        if (recordOpt.isEmpty()) {
            Map<String, String> error = new HashMap<>();
            error.put("error", "Record not found");
            error.put("id", recordId);
            return ResponseEntity.status(404).body(error);
        }

        return buildReplayResponse(recordOpt.get());
    }

    @GetMapping("/latest")
    public ResponseEntity<?> replayLatest(
            @RequestParam(required = false) String method,
            @RequestParam(required = false) String path) {
        Optional<RecordedExchange> recordOpt = recordingService.getLatest(method, path);
        if (recordOpt.isEmpty()) {
            Map<String, String> error = new HashMap<>();
            error.put("error", "No matching record found");
            if (method != null) {
                error.put("method", method);
            }
            if (path != null) {
                error.put("path", path);
            }
            return ResponseEntity.status(404).body(error);
        }

        return buildReplayResponse(recordOpt.get());
    }

    private ResponseEntity<String> buildReplayResponse(RecordedExchange record) {
        HttpHeaders headers = new HttpHeaders();
        if (record.getResponseHeaders() != null) {
            record.getResponseHeaders().forEach(headers::set);
        }

        String contentType = headers.getFirst(HttpHeaders.CONTENT_TYPE);
        if (contentType == null) {
            headers.setContentType(MediaType.APPLICATION_JSON);
        }

        headers.set("X-Replayed", "true");
        headers.set("X-Record-Id", record.getId());

        String body = record.getResponseBody() != null ? record.getResponseBody() : "";

        return ResponseEntity
                .status(record.getResponseStatus())
                .headers(headers)
                .body(body);
    }
}
