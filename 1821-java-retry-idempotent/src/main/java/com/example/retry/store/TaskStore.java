package com.example.retry.store;

import com.example.retry.model.Task;
import com.example.retry.model.TaskStatus;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

@Component
public class TaskStore {

    private final Map<String, Task> taskById = new ConcurrentHashMap<>();
    private final Map<String, Task> taskByOperationId = new ConcurrentHashMap<>();

    public Task save(Task task) {
        taskById.put(task.getId(), task);
        taskByOperationId.put(task.getOperationId(), task);
        return task;
    }

    public Optional<Task> findById(String id) {
        return Optional.ofNullable(taskById.get(id));
    }

    public Optional<Task> findByOperationId(String operationId) {
        return Optional.ofNullable(taskByOperationId.get(operationId));
    }

    public List<Task> findByStatus(TaskStatus status) {
        return taskById.values().stream()
                .filter(task -> task.getStatus() == status)
                .collect(Collectors.toList());
    }

    public boolean existsByOperationId(String operationId) {
        return taskByOperationId.containsKey(operationId);
    }
}
