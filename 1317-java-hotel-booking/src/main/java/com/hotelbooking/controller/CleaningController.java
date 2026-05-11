package com.hotelbooking.controller;

import com.hotelbooking.exception.ResourceNotFoundException;
import com.hotelbooking.model.entity.CleaningTask;
import com.hotelbooking.model.enums.CleaningPriority;
import com.hotelbooking.service.CleaningService;
import lombok.RequiredArgsConstructor;
import org.springframework.format.annotation.DateTimeFormat;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.time.LocalDateTime;
import java.util.List;

@RestController
@RequestMapping("/api/cleaning")
@RequiredArgsConstructor
public class CleaningController {

    private final CleaningService cleaningService;

    @PostMapping("/tasks")
    public ResponseEntity<CleaningTask> createTask(
            @RequestParam Long roomId,
            @RequestParam(defaultValue = "NORMAL") CleaningPriority priority,
            @RequestParam(required = false) Long assignedTo) {
        CleaningTask task = cleaningService.createCleaningTask(roomId, priority, assignedTo);
        return ResponseEntity.ok(task);
    }

    @GetMapping("/tasks/pending")
    public ResponseEntity<List<CleaningTask>> getPendingTasks() {
        List<CleaningTask> tasks = cleaningService.getPendingTasks();
        return ResponseEntity.ok(tasks);
    }

    @GetMapping("/tasks/assignee/{assigneeId}")
    public ResponseEntity<List<CleaningTask>> getTasksByAssignee(@PathVariable Long assigneeId) {
        List<CleaningTask> tasks = cleaningService.getTasksByAssignee(assigneeId);
        return ResponseEntity.ok(tasks);
    }

    @GetMapping("/tasks/{taskId}")
    public ResponseEntity<CleaningTask> getTask(@PathVariable Long taskId) {
        CleaningTask task = cleaningService.findById(taskId)
                .orElseThrow(() -> new ResourceNotFoundException("清洁任务不存在"));
        return ResponseEntity.ok(task);
    }

    @PutMapping("/tasks/{taskId}/complete")
    public ResponseEntity<CleaningTask> completeTask(@PathVariable Long taskId) {
        CleaningTask task = cleaningService.completeCleaningTask(taskId);
        return ResponseEntity.ok(task);
    }

    @PutMapping("/tasks/{taskId}/estimate")
    public ResponseEntity<CleaningTask> updateEstimate(
            @PathVariable Long taskId,
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) LocalDateTime newEstimate) {
        CleaningTask task = cleaningService.updateCleaningEstimate(taskId, newEstimate);
        return ResponseEntity.ok(task);
    }

    @PutMapping("/tasks/{taskId}/rush")
    public ResponseEntity<CleaningTask> rushTask(@PathVariable Long taskId) {
        CleaningTask task = cleaningService.rushCleaning(taskId);
        return ResponseEntity.ok(task);
    }
}
