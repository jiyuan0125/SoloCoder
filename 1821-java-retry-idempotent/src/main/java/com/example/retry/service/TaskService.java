package com.example.retry.service;

import com.example.retry.dto.CreateTaskRequest;
import com.example.retry.dto.TaskResponse;
import com.example.retry.model.Task;
import com.example.retry.model.TaskStatus;
import com.example.retry.store.TaskStore;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.util.List;
import java.util.Optional;
import java.util.UUID;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.stream.Collectors;

@Slf4j
@Service
@RequiredArgsConstructor
public class TaskService {

    public static final int DEFAULT_MAX_RETRIES = 5;
    public static final long DEFAULT_INITIAL_INTERVAL_MS = 1000L;
    public static final long DEFAULT_MAX_INTERVAL_MS = 60000L;

    private final TaskStore taskStore;
    private final ExternalPaymentService externalPaymentService;
    private final ScheduledExecutorService scheduledExecutorService;

    public TaskResponse createTask(CreateTaskRequest request) {
        String operationId = request.getOperationId();

        Optional<Task> existingTask = taskStore.findByOperationId(operationId);
        if (existingTask.isPresent()) {
            return toResponse(existingTask.get());
        }

        int maxRetries = request.getMaxRetries() != null ? request.getMaxRetries() : DEFAULT_MAX_RETRIES;
        long initialIntervalMs = request.getInitialIntervalMs() != null ? request.getInitialIntervalMs() : DEFAULT_INITIAL_INTERVAL_MS;
        long maxIntervalMs = request.getMaxIntervalMs() != null ? request.getMaxIntervalMs() : DEFAULT_MAX_INTERVAL_MS;

        Instant now = Instant.now();
        Task task = Task.builder()
                .id(UUID.randomUUID().toString())
                .operationId(operationId)
                .status(TaskStatus.QUEUED)
                .retryCount(0)
                .maxRetries(maxRetries)
                .initialIntervalMs(initialIntervalMs)
                .maxIntervalMs(maxIntervalMs)
                .createdAt(now)
                .updatedAt(now)
                .build();

        taskStore.save(task);

        executeTaskAsync(task.getId(), request.getPayload());

        return toResponse(task);
    }

    @Async
    public void executeTaskAsync(String taskId, String payload) {
        Task task = taskStore.findById(taskId).orElse(null);
        if (task == null) {
            return;
        }

        task.setStatus(TaskStatus.RUNNING);
        task.setUpdatedAt(Instant.now());
        taskStore.save(task);

        try {
            String result = externalPaymentService.executePayment(task.getOperationId(), payload);
            task.setStatus(TaskStatus.SUCCESS);
            task.setResult(result);
            task.setUpdatedAt(Instant.now());
            taskStore.save(task);
            log.info("Task {} completed successfully", taskId);
        } catch (Exception e) {
            handleFailure(task, e.getMessage());
        }
    }

    private void handleFailure(Task task, String errorMessage) {
        int retryCount = task.getRetryCount();
        int maxRetries = task.getMaxRetries();

        if (retryCount >= maxRetries) {
            task.setStatus(TaskStatus.DEAD);
            task.setErrorMessage(errorMessage);
            task.setUpdatedAt(Instant.now());
            taskStore.save(task);
            log.warn("Task {} reached max retries, marked as DEAD", task.getId());
            return;
        }

        task.setRetryCount(retryCount + 1);
        task.setStatus(TaskStatus.RETRYING);
        task.setErrorMessage(errorMessage);

        long nextDelayMs = calculateNextDelay(task);
        Instant nextRetryTime = Instant.now().plusMillis(nextDelayMs);
        task.setNextRetryTime(nextRetryTime);
        task.setUpdatedAt(Instant.now());
        taskStore.save(task);

        log.info("Task {} scheduled for retry in {}ms (attempt {}/{})",
                task.getId(), nextDelayMs, task.getRetryCount(), maxRetries);

        scheduledExecutorService.schedule(() -> retryTask(task.getId()), nextDelayMs, TimeUnit.MILLISECONDS);
    }

    private long calculateNextDelay(Task task) {
        long baseDelay = task.getInitialIntervalMs();
        int retryCount = task.getRetryCount();
        long delay = baseDelay * (long) Math.pow(2, retryCount - 1);
        return Math.min(delay, task.getMaxIntervalMs());
    }

    private void retryTask(String taskId) {
        Task task = taskStore.findById(taskId).orElse(null);
        if (task == null || task.getStatus() != TaskStatus.RETRYING) {
            return;
        }

        task.setStatus(TaskStatus.RUNNING);
        task.setNextRetryTime(null);
        task.setUpdatedAt(Instant.now());
        taskStore.save(task);

        try {
            String result = externalPaymentService.executePayment(task.getOperationId(), null);
            task.setStatus(TaskStatus.SUCCESS);
            task.setResult(result);
            task.setErrorMessage(null);
            task.setUpdatedAt(Instant.now());
            taskStore.save(task);
            log.info("Task {} retry attempt {} succeeded", taskId, task.getRetryCount());
        } catch (Exception e) {
            handleFailure(task, e.getMessage());
        }
    }

    public Optional<TaskResponse> getTaskById(String id) {
        return taskStore.findById(id).map(this::toResponse);
    }

    public List<TaskResponse> getTasksByStatus(TaskStatus status) {
        return taskStore.findByStatus(status).stream()
                .map(this::toResponse)
                .collect(Collectors.toList());
    }

    public Optional<TaskResponse> manualRetry(String taskId) {
        Task task = taskStore.findById(taskId).orElse(null);
        if (task == null) {
            return Optional.empty();
        }

        TaskStatus currentStatus = task.getStatus();
        if (currentStatus != TaskStatus.FAILED && currentStatus != TaskStatus.DEAD) {
            return Optional.of(toResponse(task));
        }

        task.setRetryCount(0);
        task.setStatus(TaskStatus.QUEUED);
        task.setErrorMessage(null);
        task.setNextRetryTime(null);
        task.setUpdatedAt(Instant.now());
        taskStore.save(task);

        executeTaskAsync(task.getId(), null);

        return Optional.of(toResponse(task));
    }

    private TaskResponse toResponse(Task task) {
        return TaskResponse.builder()
                .id(task.getId())
                .operationId(task.getOperationId())
                .status(task.getStatus())
                .result(task.getResult())
                .errorMessage(task.getErrorMessage())
                .retryCount(task.getRetryCount())
                .nextRetryTime(task.getNextRetryTime())
                .createdAt(task.getCreatedAt())
                .updatedAt(task.getUpdatedAt())
                .build();
    }
}
