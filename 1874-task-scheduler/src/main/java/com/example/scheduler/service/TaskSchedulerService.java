package com.example.scheduler.service;

import com.example.scheduler.dto.TaskRequest;
import com.example.scheduler.dto.TaskUpdateRequest;
import com.example.scheduler.model.CallbackLog;
import com.example.scheduler.model.ConcurrencyPolicy;
import com.example.scheduler.model.ExecutionHistory;
import com.example.scheduler.model.ExecutionStatus;
import com.example.scheduler.model.Task;
import com.example.scheduler.model.TaskType;
import com.example.scheduler.repository.JsonFileRepository;
import com.fasterxml.jackson.databind.ObjectMapper;
import jakarta.annotation.PostConstruct;
import jakarta.annotation.PreDestroy;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.time.Instant;
import java.time.LocalDateTime;
import java.time.ZoneId;
import java.util.List;
import java.util.Map;
import java.util.Queue;
import java.util.Set;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentLinkedQueue;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.ScheduledFuture;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicBoolean;

@Service
public class TaskSchedulerService {

    private final JsonFileRepository repository;
    private final ObjectMapper objectMapper;
    private final HttpClient httpClient;

    @Value("${scheduler.max-retries:3}")
    private int maxRetries;

    @Value("${scheduler.max-queue-size:5}")
    private int maxQueueSize;

    private final ScheduledExecutorService scheduler;
    private final Map<String, ScheduledFuture<?>> scheduledTasks = new ConcurrentHashMap<>();
    private final Set<String> runningTasks = ConcurrentHashMap.newKeySet();
    private final Map<String, Queue<Runnable>> taskQueues = new ConcurrentHashMap<>();
    private final Map<String, Integer> retryCounts = new ConcurrentHashMap<>();

    public TaskSchedulerService(JsonFileRepository repository) {
        this.repository = repository;
        this.objectMapper = new ObjectMapper();
        this.objectMapper.findAndRegisterModules();
        this.httpClient = HttpClient.newBuilder()
            .connectTimeout(Duration.ofSeconds(30))
            .build();
        this.scheduler = Executors.newScheduledThreadPool(Runtime.getRuntime().availableProcessors());
    }

    @PostConstruct
    public void init() {
        List<Task> tasks = repository.findAllTasks();
        for (Task task : tasks) {
            rescheduleTask(task);
        }
    }

    @PreDestroy
    public void shutdown() {
        scheduler.shutdown();
        try {
            if (!scheduler.awaitTermination(60, TimeUnit.SECONDS)) {
                scheduler.shutdownNow();
            }
        } catch (InterruptedException e) {
            scheduler.shutdownNow();
            Thread.currentThread().interrupt();
        }
    }

    public Task createTask(TaskRequest request) throws IOException {
        validateTaskRequest(request);
        
        Task task = new Task();
        task.setName(request.getName());
        task.setType(request.getType());
        task.setExpression(request.getExpression());
        task.setDelaySeconds(request.getDelaySeconds());
        task.setCallbackUrl(request.getCallbackUrl());
        task.setParameters(request.getParameters());
        task.setAllowRetry(request.isAllowRetry());
        task.setConcurrencyPolicy(request.getConcurrencyPolicy());
        
        task = repository.saveTask(task);
        scheduleTask(task);
        
        return task;
    }

    public Task updateTask(String taskId, TaskUpdateRequest request) throws IOException {
        Task existingTask = repository.findTaskById(taskId);
        if (existingTask == null) {
            return null;
        }

        if (request.getName() != null) {
            existingTask.setName(request.getName());
        }
        if (request.getType() != null) {
            existingTask.setType(request.getType());
        }
        if (request.getExpression() != null) {
            existingTask.setExpression(request.getExpression());
        }
        if (request.getDelaySeconds() != null) {
            existingTask.setDelaySeconds(request.getDelaySeconds());
        }
        if (request.getCallbackUrl() != null) {
            existingTask.setCallbackUrl(request.getCallbackUrl());
        }
        if (request.getParameters() != null) {
            existingTask.setParameters(request.getParameters());
        }
        if (request.getAllowRetry() != null) {
            existingTask.setAllowRetry(request.getAllowRetry());
        }
        if (request.getConcurrencyPolicy() != null) {
            existingTask.setConcurrencyPolicy(request.getConcurrencyPolicy());
        }
        
        existingTask.setUpdatedAt(Instant.now());
        
        cancelScheduledTask(taskId);
        existingTask = repository.saveTask(existingTask);
        scheduleTask(existingTask);
        
        return existingTask;
    }

    public boolean deleteTask(String taskId) throws IOException {
        Task task = repository.findTaskById(taskId);
        if (task == null) {
            return false;
        }
        
        cancelScheduledTask(taskId);
        runningTasks.remove(taskId);
        taskQueues.remove(taskId);
        retryCounts.remove(taskId);
        repository.deleteTask(taskId);
        
        return true;
    }

    public List<Task> getAllTasks() {
        return repository.findAllTasks();
    }

    public Task getTaskById(String taskId) {
        return repository.findTaskById(taskId);
    }

    public List<ExecutionHistory> getTaskHistory(String taskId) {
        return repository.findExecutionHistoryByTaskId(taskId);
    }

    public List<CallbackLog> getTaskCallbacks(String taskId) {
        return repository.findCallbackLogsByTaskId(taskId);
    }

    private void validateTaskRequest(TaskRequest request) {
        if (request.getType() == TaskType.CRON && request.getExpression() == null) {
            throw new IllegalArgumentException("expression is required for CRON tasks");
        }
        if (request.getType() == TaskType.ONCE && request.getDelaySeconds() == null) {
            throw new IllegalArgumentException("delaySeconds is required for ONCE tasks");
        }
    }

    private void scheduleTask(Task task) {
        if (task.getType() == TaskType.CRON) {
            scheduleCronTask(task);
        } else {
            scheduleOnceTask(task);
        }
    }

    private void scheduleOnceTask(Task task) {
        Instant nextRun = task.getNextExecutionTime();
        if (nextRun == null) {
            nextRun = Instant.now().plusSeconds(task.getDelaySeconds());
            task.setNextExecutionTime(nextRun);
            try {
                repository.saveTask(task);
            } catch (IOException e) {
                throw new RuntimeException("Failed to update task", e);
            }
        }

        long delay = Duration.between(Instant.now(), nextRun).toMillis();
        if (delay < 0) {
            delay = 0;
        }

        ScheduledFuture<?> future = scheduler.schedule(
            () -> executeTask(task),
            delay,
            TimeUnit.MILLISECONDS
        );
        scheduledTasks.put(task.getId(), future);
    }

    private void scheduleCronTask(Task task) {
        Instant nextRun = task.getNextExecutionTime();
        if (nextRun == null) {
            nextRun = calculateNextCronTime(task.getExpression(), Instant.now());
            task.setNextExecutionTime(nextRun);
            try {
                repository.saveTask(task);
            } catch (IOException e) {
                throw new RuntimeException("Failed to update task", e);
            }
        }

        long delay = Duration.between(Instant.now(), nextRun).toMillis();
        if (delay < 0) {
            delay = 0;
        }

        ScheduledFuture<?> future = scheduler.schedule(
            () -> executeCronTask(task),
            delay,
            TimeUnit.MILLISECONDS
        );
        scheduledTasks.put(task.getId(), future);
    }

    private void rescheduleTask(Task task) {
        if (task.getType() == TaskType.CRON) {
            Instant nextRun = task.getNextExecutionTime();
            if (nextRun == null || nextRun.isBefore(Instant.now())) {
                nextRun = calculateNextCronTime(task.getExpression(), Instant.now());
                task.setNextExecutionTime(nextRun);
                try {
                    repository.saveTask(task);
                } catch (IOException e) {
                    throw new RuntimeException("Failed to update task", e);
                }
            }
            scheduleCronTask(task);
        } else {
            Instant nextRun = task.getNextExecutionTime();
            if (nextRun != null && nextRun.isAfter(Instant.now())) {
                scheduleOnceTask(task);
            }
        }
    }

    private void cancelScheduledTask(String taskId) {
        ScheduledFuture<?> future = scheduledTasks.remove(taskId);
        if (future != null) {
            future.cancel(false);
        }
    }

    private void executeTask(Task task) {
        handleExecution(task, false);
    }

    private void executeCronTask(Task task) {
        handleExecution(task, true);
        
        Task currentTask = repository.findTaskById(task.getId());
        if (currentTask != null) {
            Instant nextRun = calculateNextCronTime(currentTask.getExpression(), Instant.now());
            currentTask.setNextExecutionTime(nextRun);
            try {
                repository.saveTask(currentTask);
            } catch (IOException e) {
                throw new RuntimeException("Failed to update task", e);
            }
            scheduleCronTask(currentTask);
        }
    }

    private void handleExecution(Task task, boolean isCron) {
        if (runningTasks.contains(task.getId())) {
            handleConcurrency(task);
            return;
        }

        AtomicBoolean shouldQueue = new AtomicBoolean(false);
        if (task.getConcurrencyPolicy() == ConcurrencyPolicy.QUEUE) {
            Queue<Runnable> queue = taskQueues.computeIfAbsent(task.getId(), k -> new ConcurrentLinkedQueue<>());
            synchronized (queue) {
                if (runningTasks.contains(task.getId())) {
                    if (queue.size() < maxQueueSize) {
                        queue.add(() -> doExecute(task, isCron));
                        return;
                    } else {
                        recordSkipped(task);
                        return;
                    }
                }
            }
        }

        if (!runningTasks.add(task.getId())) {
            handleConcurrency(task);
            return;
        }

        try {
            doExecute(task, isCron);
        } finally {
            runningTasks.remove(task.getId());
            processQueue(task, isCron);
        }
    }

    private void handleConcurrency(Task task) {
        if (task.getConcurrencyPolicy() == ConcurrencyPolicy.SKIP) {
            recordSkipped(task);
        } else {
            Queue<Runnable> queue = taskQueues.computeIfAbsent(task.getId(), k -> new ConcurrentLinkedQueue<>());
            synchronized (queue) {
                if (queue.size() < maxQueueSize) {
                    queue.add(() -> doExecute(task, false));
                } else {
                    recordSkipped(task);
                }
            }
        }
    }

    private void processQueue(Task task, boolean isCron) {
        Queue<Runnable> queue = taskQueues.get(task.getId());
        if (queue != null) {
            Runnable nextTask = queue.poll();
            if (nextTask != null && runningTasks.add(task.getId())) {
                try {
                    nextTask.run();
                } finally {
                    runningTasks.remove(task.getId());
                    processQueue(task, isCron);
                }
            }
        }
    }

    private void doExecute(Task task, boolean isCron) {
        int retryCount = retryCounts.getOrDefault(task.getId(), 0);
        
        ExecutionHistory history = new ExecutionHistory();
        history.setTaskId(task.getId());
        history.setStatus(ExecutionStatus.RUNNING);
        history.setRetryCount(retryCount);
        
        try {
            repository.addExecutionHistory(history);
        } catch (IOException e) {
            e.printStackTrace();
        }

        CallbackLog callbackLog = new CallbackLog();
        callbackLog.setTaskId(task.getId());
        callbackLog.setExecutionHistoryId(history.getId());
        callbackLog.setRequestUrl(task.getCallbackUrl());
        callbackLog.setRequestTime(Instant.now());

        try {
            String requestBody = objectMapper.writeValueAsString(Map.of(
                "taskId", task.getId(),
                "taskName", task.getName(),
                "parameters", task.getParameters() != null ? task.getParameters() : Map.of()
            ));
            callbackLog.setRequestBody(requestBody);

            HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(task.getCallbackUrl()))
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(requestBody))
                .timeout(Duration.ofSeconds(60))
                .build();

            HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());
            callbackLog.setResponseTime(Instant.now());
            callbackLog.setResponseStatus(response.statusCode());
            callbackLog.setResponseBody(response.body());
            callbackLog.setResponseHeaders(response.headers().toString());

            try {
                repository.addCallbackLog(callbackLog);
            } catch (IOException e) {
                e.printStackTrace();
            }

            if (response.statusCode() >= 200 && response.statusCode() < 300) {
                history.setStatus(ExecutionStatus.SUCCESS);
                retryCounts.remove(task.getId());
                if (!isCron) {
                    try {
                        repository.deleteTask(task.getId());
                    } catch (IOException e) {
                        e.printStackTrace();
                    }
                }
            } else {
                handleFailure(task, history, callbackLog, "HTTP " + response.statusCode(), retryCount, isCron);
            }

        } catch (java.net.http.HttpConnectTimeoutException e) {
            callbackLog.setResponseTime(Instant.now());
            callbackLog.setResponseStatus(-1);
            callbackLog.setErrorReason("Connection timeout");
            try {
                repository.addCallbackLog(callbackLog);
            } catch (IOException ex) {
                ex.printStackTrace();
            }
            handleFailure(task, history, callbackLog, "Connection timeout", retryCount, isCron);
        } catch (java.net.UnknownHostException e) {
            callbackLog.setResponseTime(Instant.now());
            callbackLog.setResponseStatus(-1);
            callbackLog.setErrorReason("DNS resolution failed");
            try {
                repository.addCallbackLog(callbackLog);
            } catch (IOException ex) {
                ex.printStackTrace();
            }
            handleFailure(task, history, callbackLog, "DNS resolution failed: " + e.getMessage(), retryCount, isCron);
        } catch (Exception e) {
            callbackLog.setResponseTime(Instant.now());
            callbackLog.setResponseStatus(-1);
            callbackLog.setErrorReason(e.getMessage());
            try {
                repository.addCallbackLog(callbackLog);
            } catch (IOException ex) {
                ex.printStackTrace();
            }
            handleFailure(task, history, callbackLog, e.getClass().getSimpleName() + ": " + e.getMessage(), retryCount, isCron);
        }

        try {
            repository.addExecutionHistory(history);
        } catch (IOException e) {
            e.printStackTrace();
        }
    }

    private void handleFailure(Task task, ExecutionHistory history, CallbackLog callbackLog, 
                              String errorReason, int retryCount, boolean isCron) {
        history.setErrorReason(errorReason);
        history.setStatus(ExecutionStatus.FAILED);

        if (task.isAllowRetry() && retryCount < maxRetries) {
            int newRetryCount = retryCount + 1;
            retryCounts.put(task.getId(), newRetryCount);
            long delay = (long) Math.pow(2, newRetryCount) * 1000;
            
            scheduler.schedule(() -> {
                runningTasks.remove(task.getId());
                handleExecution(task, isCron);
            }, delay, TimeUnit.MILLISECONDS);
        } else {
            retryCounts.remove(task.getId());
            if (!isCron) {
                try {
                    repository.deleteTask(task.getId());
                } catch (IOException e) {
                    e.printStackTrace();
                }
            }
        }
    }

    private void recordSkipped(Task task) {
        ExecutionHistory history = new ExecutionHistory();
        history.setTaskId(task.getId());
        history.setStatus(ExecutionStatus.SKIPPED);
        history.setRetryCount(0);

        try {
            repository.addExecutionHistory(history);
        } catch (IOException e) {
            e.printStackTrace();
        }
    }

    private Instant calculateNextCronTime(String expression, Instant from) {
        LocalDateTime now = LocalDateTime.ofInstant(from, ZoneId.systemDefault());
        
        String[] parts = expression.trim().split("\\s+");
        if (parts.length != 5) {
            throw new IllegalArgumentException("Invalid cron expression: " + expression);
        }

        int[] cronFields = new int[5];
        for (int i = 0; i < 5; i++) {
            cronFields[i] = parseCronField(parts[i], i);
        }

        LocalDateTime next = now.plusMinutes(1);
        for (int i = 0; i < 1000000; i++) {
            if (matchesCron(next, cronFields)) {
                return next.atZone(ZoneId.systemDefault()).toInstant();
            }
            next = next.plusMinutes(1);
        }

        throw new IllegalArgumentException("Cannot calculate next execution time for: " + expression);
    }

    private int parseCronField(String field, int index) {
        if ("*".equals(field)) {
            return -1;
        }
        try {
            return Integer.parseInt(field);
        } catch (NumberFormatException e) {
            throw new IllegalArgumentException("Invalid cron field at index " + index + ": " + field);
        }
    }

    private boolean matchesCron(LocalDateTime dt, int[] fields) {
        if (fields[0] != -1 && dt.getMinute() != fields[0]) return false;
        if (fields[1] != -1 && dt.getHour() != fields[1]) return false;
        if (fields[2] != -1 && dt.getDayOfMonth() != fields[2]) return false;
        if (fields[3] != -1 && dt.getMonthValue() != fields[3]) return false;
        if (fields[4] != -1 && dt.getDayOfWeek().getValue() % 7 != fields[4]) return false;
        return true;
    }
}
