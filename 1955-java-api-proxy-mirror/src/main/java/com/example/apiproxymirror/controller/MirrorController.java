package com.example.apiproxymirror.controller;

import com.example.apiproxymirror.model.RecordingSettings;
import com.example.apiproxymirror.service.RecordingService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.Map;

@RestController
@RequestMapping("/mirror")
@RequiredArgsConstructor
public class MirrorController {

    private final RecordingService recordingService;

    @PostMapping("/start")
    public ResponseEntity<Map<String, Object>> startRecording(@RequestBody(required = false) RecordingSettings settings) {
        if (settings == null) {
            settings = new RecordingSettings();
        }
        recordingService.startRecording(settings);

        Map<String, Object> response = new HashMap<>();
        response.put("status", "started");
        response.put("settings", recordingService.getSettings());
        response.put("recordCount", recordingService.getRecordCount());
        return ResponseEntity.ok(response);
    }

    @PostMapping("/stop")
    public ResponseEntity<Map<String, Object>> stopRecording() {
        recordingService.stopRecording();

        Map<String, Object> response = new HashMap<>();
        response.put("status", "stopped");
        response.put("settings", recordingService.getSettings());
        response.put("recordCount", recordingService.getRecordCount());
        return ResponseEntity.ok(response);
    }

    @GetMapping("/status")
    public ResponseEntity<Map<String, Object>> getStatus() {
        Map<String, Object> response = new HashMap<>();
        response.put("recording", recordingService.isRecordingEnabled());
        response.put("settings", recordingService.getSettings());
        response.put("recordCount", recordingService.getRecordCount());
        return ResponseEntity.ok(response);
    }
}
