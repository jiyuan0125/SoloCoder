package com.crondispatch.controller;

import com.crondispatch.model.CallbackLog;
import com.crondispatch.model.ExecutionHistory;
import com.crondispatch.model.Task;
import com.crondispatch.service.HistoryService;
import com.crondispatch.service.TaskSchedulerService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/tasks")
public class TaskController {

    private final TaskSchedulerService taskSchedulerService;
    private final HistoryService historyService;

    public TaskController(TaskSchedulerService taskSchedulerService, HistoryService historyService) {
        this.taskSchedulerService = taskSchedulerService;
        this.historyService = historyService;
    }

    @PostMapping
    public ResponseEntity<?> registerTask(@RequestBody Task task) {
        try {
            Task registered = taskSchedulerService.registerTask(task);
            return ResponseEntity.ok(registered);
        } catch (IllegalArgumentException e) {
            Map<String, String> error = new HashMap<>();
            error.put("error", e.getMessage());
            return ResponseEntity.badRequest().body(error);
        }
    }

    @GetMapping
    public ResponseEntity<List<Task>> getTasks() {
        List<Task> tasks = taskSchedulerService.getAllTasks();
        return ResponseEntity.ok(tasks);
    }

    @GetMapping("/{id}")
    public ResponseEntity<?> getTask(@PathVariable String id) {
        Task task = taskSchedulerService.getTask(id);
        if (task == null) {
            Map<String, String> error = new HashMap<>();
            error.put("error", "任务不存在");
            return ResponseEntity.status(404).body(error);
        }
        return ResponseEntity.ok(task);
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<?> deleteTask(@PathVariable String id) {
        boolean deleted = taskSchedulerService.deleteTask(id);
        if (!deleted) {
            Map<String, String> error = new HashMap<>();
            error.put("error", "任务不存在");
            return ResponseEntity.status(404).body(error);
        }
        Map<String, String> result = new HashMap<>();
        result.put("message", "任务已删除");
        return ResponseEntity.ok(result);
    }

    @GetMapping("/{id}/history")
    public ResponseEntity<?> getTaskHistory(@PathVariable String id) {
        if (!taskSchedulerService.getAllTasks().stream().anyMatch(t -> t.getId().equals(id))) {
            Map<String, String> error = new HashMap<>();
            error.put("error", "任务不存在");
            return ResponseEntity.status(404).body(error);
        }
        List<ExecutionHistory> history = historyService.getTaskHistory(id);
        return ResponseEntity.ok(history);
    }

    @GetMapping("/{id}/logs")
    public ResponseEntity<?> getTaskLogs(@PathVariable String id) {
        if (!taskSchedulerService.getAllTasks().stream().anyMatch(t -> t.getId().equals(id))) {
            Map<String, String> error = new HashMap<>();
            error.put("error", "任务不存在");
            return ResponseEntity.status(404).body(error);
        }
        List<CallbackLog> logs = historyService.getTaskLogs(id);
        return ResponseEntity.ok(logs);
    }
}
