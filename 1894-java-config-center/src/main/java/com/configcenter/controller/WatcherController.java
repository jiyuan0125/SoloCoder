package com.configcenter.controller;

import com.configcenter.dto.WatcherRequest;
import com.configcenter.entity.Watcher;
import com.configcenter.service.WatcherService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/applications/{applicationId}/watchers")
@RequiredArgsConstructor
public class WatcherController {

    private final WatcherService watcherService;

    @GetMapping
    public ResponseEntity<List<Watcher>> getWatchers(@PathVariable Long applicationId) {
        return ResponseEntity.ok(watcherService.getWatchersByApplication(applicationId));
    }

    @PostMapping("/register")
    public ResponseEntity<Watcher> registerWatcher(
            @PathVariable Long applicationId,
            @Valid @RequestBody WatcherRequest request) {
        Watcher registered = watcherService.registerWatcher(applicationId, request);
        return new ResponseEntity<>(registered, HttpStatus.OK);
    }

    @PostMapping("/unregister")
    public ResponseEntity<Void> unregisterWatcher(
            @PathVariable Long applicationId,
            @Valid @RequestBody WatcherRequest request) {
        watcherService.unregisterWatcher(applicationId, request);
        return ResponseEntity.noContent().build();
    }
}
