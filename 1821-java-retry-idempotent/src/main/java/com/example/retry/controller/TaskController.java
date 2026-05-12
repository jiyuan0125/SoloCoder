package com.example.retry.controller;

import com.example.retry.dto.CreateTaskRequest;
import com.example.retry.dto.TaskResponse;
import com.example.retry.model.TaskStatus;
import com.example.retry.service.TaskService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/tasks")
@RequiredArgsConstructor
public class TaskController {

    private final TaskService taskService;

    @PostMapping
    public ResponseEntity<TaskResponse> createTask(@RequestBody CreateTaskRequest request) {
        String operationId = request.getOperationId();
        if (operationId == null || operationId.trim().isEmpty()) {
            return ResponseEntity.badRequest().build();
        }
        if (operationId.length() > 64) {
            return ResponseEntity.badRequest().build();
        }

        TaskResponse response = taskService.createTask(request);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/{id}")
    public ResponseEntity<TaskResponse> getTaskById(@PathVariable String id) {
        return taskService.getTaskById(id)
                .map(ResponseEntity::ok)
                .orElse(ResponseEntity.notFound().build());
    }

    @GetMapping
    public ResponseEntity<List<TaskResponse>> getTasksByStatus(@RequestParam(required = false) TaskStatus status) {
        if (status == null) {
            return ResponseEntity.ok().build();
        }
        List<TaskResponse> tasks = taskService.getTasksByStatus(status);
        return ResponseEntity.ok(tasks);
    }

    @PostMapping("/{id}/retry")
    public ResponseEntity<TaskResponse> manualRetry(@PathVariable String id) {
        return taskService.manualRetry(id)
                .map(ResponseEntity::ok)
                .orElse(ResponseEntity.notFound().build());
    }
}
