package com.example.scheduler.controller;

import com.example.scheduler.dto.TaskRequest;
import com.example.scheduler.dto.TaskUpdateRequest;
import com.example.scheduler.model.CallbackLog;
import com.example.scheduler.model.ExecutionHistory;
import com.example.scheduler.model.Task;
import com.example.scheduler.service.TaskSchedulerService;
import jakarta.validation.Valid;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.io.IOException;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/tasks")
public class TaskController {

    private final TaskSchedulerService schedulerService;

    public TaskController(TaskSchedulerService schedulerService) {
        this.schedulerService = schedulerService;
    }

    @PostMapping
    public ResponseEntity<?> createTask(@Valid @RequestBody TaskRequest request) {
        try {
            Task task = schedulerService.createTask(request);
            return ResponseEntity.status(HttpStatus.CREATED).body(task);
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
        } catch (IOException e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to save task: " + e.getMessage()));
        }
    }

    @GetMapping
    public ResponseEntity<List<Task>> getAllTasks() {
        return ResponseEntity.ok(schedulerService.getAllTasks());
    }

    @GetMapping("/{id}")
    public ResponseEntity<?> getTaskById(@PathVariable String id) {
        Task task = schedulerService.getTaskById(id);
        if (task == null) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(task);
    }

    @GetMapping("/{id}/history")
    public ResponseEntity<?> getTaskHistory(@PathVariable String id) {
        Task task = schedulerService.getTaskById(id);
        if (task == null) {
            return ResponseEntity.notFound().build();
        }
        List<ExecutionHistory> history = schedulerService.getTaskHistory(id);
        return ResponseEntity.ok(history);
    }

    @GetMapping("/{id}/callbacks")
    public ResponseEntity<?> getTaskCallbacks(@PathVariable String id) {
        Task task = schedulerService.getTaskById(id);
        if (task == null) {
            return ResponseEntity.notFound().build();
        }
        List<CallbackLog> logs = schedulerService.getTaskCallbacks(id);
        return ResponseEntity.ok(logs);
    }

    @PutMapping("/{id}")
    public ResponseEntity<?> updateTask(@PathVariable String id, @RequestBody TaskUpdateRequest request) {
        try {
            Task task = schedulerService.updateTask(id, request);
            if (task == null) {
                return ResponseEntity.notFound().build();
            }
            return ResponseEntity.ok(task);
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
        } catch (IOException e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to update task: " + e.getMessage()));
        }
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<?> deleteTask(@PathVariable String id) {
        try {
            boolean deleted = schedulerService.deleteTask(id);
            if (!deleted) {
                return ResponseEntity.notFound().build();
            }
            return ResponseEntity.noContent().build();
        } catch (IOException e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                .body(Map.of("error", "Failed to delete task: " + e.getMessage()));
        }
    }
}
